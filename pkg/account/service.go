package account

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"os"

	"github.com/GilbertoVGL/go-banking/pkg/apperrors"
	"github.com/GilbertoVGL/go-banking/pkg/validators"
)

type Repository interface {
	ListAccount(ListAccountQuery) (ListAccountsReponse, error)
	AddAccount(NewAccountRequest) error
	GetAccountBalance(uint64) (int64, error)
}

type Service interface {
	List(ListAccountQuery) (ListAccountsReponse, error)
	NewAccount(NewAccountRequest) error
	GetBalance(uint64) (BalanceResponse, error)
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

	if err := validateAccountValues(newAccount); err != nil {
		return err
	}

	newAccount.Secret = fmt.Sprintf("%x", sha256.Sum256([]byte(newAccount.Secret+os.Getenv("SALT"))))

	if err := s.r.AddAccount(newAccount); err != nil {
		return err
	}

	return nil
}

func (s *service) GetBalance(userId uint64) (BalanceResponse, error) {
	var balanceResponse BalanceResponse
	balance, err := s.r.GetAccountBalance(userId)
	balanceResponse.Balance = balance

	if err != nil {
		return balanceResponse, err
	}

	return balanceResponse, nil
}

func validateAccountValues(a NewAccountRequest) error {
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
	if a.Balance < 0 {
		invalid = append(invalid, "balance")
	}

	if len(invalid) > 0 {
		return &apperrors.ArgumentError{Context: invalid, Err: errors.New("invalid values")}
	}

	if err := validators.ValidateCPF(a.Cpf); err != nil {
		invalid = append(invalid, err.Error())
	}

	if len(a.Secret) < 8 || len(a.Secret) > 16 {
		invalid = append(invalid, "secret must be between 8 and 16 characters")
	}

	if a.Balance < 0 {
		invalid = append(invalid, "invalid balance")
	}

	if len(invalid) > 0 {
		return &apperrors.ArgumentError{Context: invalid, Err: errors.New("invalid values")}
	}

	return nil
}
