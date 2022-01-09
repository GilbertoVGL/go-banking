package transfer

import (
	"errors"

	"github.com/GilbertoVGL/go-banking/pkg/account"
	"github.com/GilbertoVGL/go-banking/pkg/apperrors"
)

type Service interface {
	GetTransfers(uint64, ListTransferQuery) (ListTransferReponse, error)
	DoTransfer(TransferRequest) error
}

type Repository interface {
	AddTransfer(TransferRequest) error
	GetTransfers(uint64, ListTransferQuery) (ListTransferReponse, error)
	GetAccountBalance(uint64) (int64, error)
	GetAccountById(id uint64) (account.Account, error)
}

type service struct {
	r Repository
}

func New(r Repository) *service {
	return &service{r}
}

func (s *service) GetTransfers(id uint64, l ListTransferQuery) (ListTransferReponse, error) {
	return s.r.GetTransfers(id, l)
}

func (s *service) DoTransfer(t TransferRequest) error {
	var invalid []string

	if t.Amount == nil {
		invalid = append(invalid, "amount")
	}

	if t.Destination == nil {
		invalid = append(invalid, "destination")
	}

	if len(invalid) > 0 {
		return &apperrors.ArgumentError{Context: invalid, Err: errors.New("missing values")}
	}

	if *t.Amount < 0 {
		invalid = append(invalid, "amount")
		return &apperrors.ArgumentError{Context: invalid, Err: errors.New("invalid values")}
	}

	originBalance, err := s.r.GetAccountBalance(t.Origin)

	if err != nil {
		return err
	}

	if originBalance < *t.Amount {
		return &apperrors.TransferRequestError{Context: "not enough funds", Err: errors.New("transfer error")}
	}

	if _, err := s.r.GetAccountById(*t.Destination); err != nil {
		if _, ok := err.(*apperrors.AccountNotFoundError); ok {
			return &apperrors.TransferRequestError{Context: "destination account not found", Err: errors.New(err.Error() + ": transfer error")}
		}

		return err
	}

	if err := s.r.AddTransfer(t); err != nil {
		return err
	}

	return nil
}
