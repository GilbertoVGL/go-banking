package login

import (
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
	GetAccountBySecretAndCPF(LoginRequest) (Account, error)
}

type Service interface {
	LoginUser(LoginRequest) (LoginReponse, error)
}

type service struct {
	r Repository
}

func New(r Repository) *service {
	return &service{r}
}

func (s *service) LoginUser(login LoginRequest) (LoginReponse, error) {
	var loginResponse LoginReponse

	if err := validateValues(login); err != nil {
		return loginResponse, err
	}

	login.Secret = fmt.Sprintf("%x", sha256.Sum256([]byte(login.Secret+os.Getenv("SALT"))))
	account, err := s.r.GetAccountBySecretAndCPF(login)

	if err != nil {
		return loginResponse, err
	}

	if !account.Active {
		return loginResponse, errors.New("this account is inactive")
	}

	at := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"authorized": true,
		"userId":     account.Id,
		"exp":        time.Now().Add(time.Minute * 15).Unix(),
	})
	token, err := at.SignedString([]byte(os.Getenv("JWT_SECRET")))

	if err != nil {
		return loginResponse, errors.New("failed to create user token")
	}

	return LoginReponse{Token: token}, nil
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
