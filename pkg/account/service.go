package account

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"strings"

	"github.com/GilbertoVGL/go-banking/pkg/apperrors"
	"github.com/GilbertoVGL/go-banking/pkg/validators"
)

type Repository interface {
	ListAccount(context.Context, ListAccountQuery) (ListAccountsReponse, error)
	AddAccount(context.Context, NewAccountRequest) error
	GetAccountBalance(context.Context, uint64) (int64, error)
}

type Service interface {
	List(context.Context, ListAccountQuery) (ListAccountsReponse, error)
	NewAccount(context.Context, NewAccountRequest) error
	GetBalance(context.Context, uint64) (BalanceResponse, error)
}

type service struct {
	r Repository
}

func New(r Repository) *service {
	return &service{r}
}

func (s *service) List(ctx context.Context, q ListAccountQuery) (ListAccountsReponse, error) {
	accountsCh := make(chan ListAccountsReponse)
	errCh := make(chan error)

	go func() {
		accounts, err := s.r.ListAccount(ctx, q)
		if err != nil {
			errCh <- err
			return
		}
		accountsCh <- accounts
	}()

	select {
	case accounts := <-accountsCh:
		return accounts, nil
	case err := <-errCh:
		return ListAccountsReponse{}, err
	case <-ctx.Done():
		return ListAccountsReponse{}, ctx.Err()
	}
}

func (s *service) NewAccount(ctx context.Context, newAccount NewAccountRequest) error {
	accountCh := make(chan bool)
	errCh := make(chan error)

	go func() {
		if err := validateAccountValues(newAccount); err != nil {
			errCh <- err
			return
		}

		newAccount.Secret = fmt.Sprintf("%x", sha256.Sum256([]byte(newAccount.Secret+os.Getenv("SALT"))))

		if err := s.r.AddAccount(ctx, newAccount); err != nil {
			errCh <- err
			return
		}

		accountCh <- true
	}()

	select {
	case <-accountCh:
		return nil
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *service) GetBalance(ctx context.Context, userId uint64) (BalanceResponse, error) {
	var balanceResponse BalanceResponse
	balanceCh := make(chan int64)
	errCh := make(chan error)

	go func() {
		balance, err := s.r.GetAccountBalance(ctx, userId)
		if err != nil {
			errCh <- err
			return
		}
		balanceCh <- balance
	}()

	select {
	case balance := <-balanceCh:
		balanceResponse.Balance = balance
		return balanceResponse, nil
	case err := <-errCh:
		return balanceResponse, err
	case <-ctx.Done():
		return balanceResponse, ctx.Err()
	}
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
		return apperrors.NewArgumentError(strings.Join(invalid, ", "))
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
		return apperrors.NewArgumentError(strings.Join(invalid, ", "))
	}

	return nil
}
