package account

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/GilbertoVGL/go-banking/pkg/validators"
)

type Repository interface {
	ListAccount(ListAccountQuery) (ListAccountsReponse, error)
	AddAccount(NewAccountRequest) error
	GetAccountBalance(UserId) (BalanceResponse, error)
}

type Service interface {
	List(ListAccountQuery) (ListAccountsReponse, error)
	NewAccount(NewAccountRequest) error
	GetBalance(UserId) (BalanceResponse, error)
}

type service struct {
	r Repository
}

func New(r Repository) *service {
	return &service{r}
}

func (s *service) List(q ListAccountQuery) (ListAccountsReponse, error) {
	accounts, err := s.r.ListAccount(q)

	if err != nil {
		return accounts, err
	}

	return accounts, err
}

func (s *service) NewAccount(newAccount NewAccountRequest) error {

	if err := validateValues(newAccount); err != nil {
		return err
	}

	newAccount.Secret = fmt.Sprintf("%x", sha256.Sum256([]byte(newAccount.Secret+os.Getenv("SALT"))))

	if err := s.r.AddAccount(newAccount); err != nil {
		return err
	}

	return nil
}

func (s *service) GetBalance(a UserId) (BalanceResponse, error) {
	balance, err := s.r.GetAccountBalance(a)

	if err != nil {
		return balance, err
	}

	return balance, nil
}

func validateValues(a NewAccountRequest) error {
	var invalid []string

	if a.Secret == "" {
		invalid = append(invalid, "secret")
	}
	if a.Cpf == "" {
		invalid = append(invalid, "cpf")
	}
	if a.Name == "" {
		invalid = append(invalid, "name")
	}

	if len(invalid) > 0 {
		return errors.New(fmt.Sprintf("missing values: %s", strings.Join(invalid, ", ")))
	}

	if err := validators.ValidateCPF(a.Cpf); err != nil {
		invalid = append(invalid, err.Error())
	}

	if len(a.Secret) < 8 || len(a.Secret) > 16 {
		invalid = append(invalid, "secret must be between 8 and 16 characters")
	}

	if a.Balance < 0 {
		invalid = append(invalid, "invalid balance value")
	}

	if len(invalid) > 0 {
		return errors.New(fmt.Sprintf("invalid input(s): %s", strings.Join(invalid, " and ")))
	}

	return nil
}
