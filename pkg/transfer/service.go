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

func (s *service) DoTransfer(transfer TransferRequest) error {
	if transfer.Amount <= 0 {
		return &apperrors.ArgumentError{Context: []string{"transfer value should be greater than zero"}, Err: errors.New("invalid values")}
	}

	originBalance, err := s.r.GetAccountBalance(transfer.Origin)

	if err != nil {
		return err
	}

	if originBalance < transfer.Amount {
		return errors.New("insuficient funds")
	}

	if _, err := s.r.GetAccountById(transfer.Destination); err != nil {
		return errors.New("destination " + err.Error())
	}

	if err := s.r.AddTransfer(transfer); err != nil {
		return err
	}

	return nil
}
