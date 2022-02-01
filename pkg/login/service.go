package login

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/GilbertoVGL/go-banking/pkg/apperrors"
	"github.com/GilbertoVGL/go-banking/pkg/validators"
	"github.com/golang-jwt/jwt"
)

type Repository interface {
	GetAccountBySecretAndCPF(context.Context, LoginRequest, chan Account, chan error)
}

type Service interface {
	LoginUser(context.Context, LoginRequest, chan LoginReponse, chan error)
}

type service struct {
	r Repository
}

func New(r Repository) *service {
	return &service{r}
}

func (s *service) LoginUser(ctx context.Context, loginReq LoginRequest, loginCh chan LoginReponse, errorCh chan error) {
	if err := validateValues(loginReq); err != nil {
		errorCh <- err
		return
	}

	accountCh := make(chan Account)
	errCh := make(chan error)
	loginReq.Secret = fmt.Sprintf("%x", sha256.Sum256([]byte(loginReq.Secret+os.Getenv("SALT"))))

	go s.r.GetAccountBySecretAndCPF(ctx, loginReq, accountCh, errCh)

	select {
	case account := <-accountCh:
		if !account.Active {
			errorCh <- errors.New("this account is inactive")
			return
		}

		at := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"authorized": true,
			"userId":     account.Id,
			"exp":        time.Now().Add(time.Minute * 15).Unix(),
		})
		token, err := at.SignedString([]byte(os.Getenv("JWT_SECRET")))

		if err != nil {
			errorCh <- errors.New("failed to create user token")
			return
		}

		loginCh <- LoginReponse{Token: token}

		return
	case err := <-errCh:
		errorCh <- err
	case <-ctx.Done():
		errorCh <- ctx.Err()
	}
}

func validateValues(l LoginRequest) error {
	var invalid []string

	if l.Cpf == "" {
		invalid = append(invalid, "CPF")
	}
	if l.Secret == "" {
		invalid = append(invalid, "Secret")
	}
	if len(invalid) > 0 {
		return apperrors.NewArgumentError("missing values", strings.Join(invalid, ", "))
	}

	if err := validators.ValidateCPF(l.Cpf); err != nil {
		return apperrors.NewArgumentError("invalid CPF", err.Error())
	}

	return nil
}
