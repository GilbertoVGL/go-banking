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
	GetAccountBySecretAndCPF(context.Context, LoginRequest) (Account, error)
}

type Service interface {
	LoginUser(context.Context, LoginRequest) (LoginReponse, error)
}

type service struct {
	r Repository
}

func New(r Repository) *service {
	return &service{r}
}

func (s *service) LoginUser(ctx context.Context, loginReq LoginRequest) (LoginReponse, error) {
	var login LoginReponse
	accountCh := make(chan Account)
	errCh := make(chan error)

	go func() {
		if err := validateValues(loginReq); err != nil {
			errCh <- err
			return
		}

		loginReq.Secret = fmt.Sprintf("%x", sha256.Sum256([]byte(loginReq.Secret+os.Getenv("SALT"))))

		account, err := s.r.GetAccountBySecretAndCPF(ctx, loginReq)
		if err != nil {
			errCh <- err
			return
		}

		accountCh <- account
	}()

	select {
	case account := <-accountCh:
		if !account.Active {
			return login, errors.New("this account is inactive")
		}
		var err error
		login.Token, err = generateToken(account)

		if err != nil {
			return login, errors.New("failed to create user token")
		}

		return login, nil
	case err := <-errCh:
		return login, err
	case <-ctx.Done():
		return login, ctx.Err()
	}
}

func generateToken(account Account) (string, error) {
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"authorized": true,
		"userId":     account.Id,
		"exp":        time.Now().Add(time.Minute * 15).Unix(),
	})
	return at.SignedString([]byte(os.Getenv("JWT_SECRET")))
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
