package transfer

import (
	"context"
	"strings"

	"github.com/GilbertoVGL/go-banking/pkg/account"
	"github.com/GilbertoVGL/go-banking/pkg/apperrors"
)

type Service interface {
	GetTransfers(context.Context, uint64, ListTransferQuery) (ListTransferResponse, error)
	DoTransfer(context.Context, TransferRequest) error
}

type Repository interface {
	AddTransfer(context.Context, TransferRequest) error
	GetTransfers(context.Context, uint64, ListTransferQuery) (ListTransferResponse, error)
	GetAccountBalance(context.Context, uint64) (int64, error)
	GetAccountById(context.Context, uint64) (account.Account, error)
}

type service struct {
	r Repository
}

func New(r Repository) *service {
	return &service{r}
}

func (s *service) GetTransfers(ctx context.Context, id uint64, l ListTransferQuery) (ListTransferResponse, error) {
	transferListCh := make(chan ListTransferResponse)
	errCh := make(chan error)

	go func() {
		transferList, err := s.r.GetTransfers(ctx, id, l)
		if err != nil {
			errCh <- err
			return
		}
		transferListCh <- transferList
	}()

	select {
	case transferList := <-transferListCh:
		return transferList, nil
	case err := <-errCh:
		return ListTransferResponse{}, err
	case <-ctx.Done():
		return ListTransferResponse{}, ctx.Err()
	}
}

func (s *service) DoTransfer(ctx context.Context, t TransferRequest) error {
	transferCh := make(chan bool)
	errCh := make(chan error)

	go func() {
		var invalid []string

		if t.Amount == nil {
			invalid = append(invalid, "amount")
		}

		if t.Destination == nil {
			invalid = append(invalid, "destination")
		}

		if len(invalid) > 0 {
			errCh <- apperrors.NewArgumentError(strings.Join(invalid, ", "))
		}

		if *t.Amount < 1 {
			invalid = append(invalid, "amount")
			errCh <- apperrors.NewArgumentError(strings.Join(invalid, ", "))
			return
		}

		originBalance, err := s.r.GetAccountBalance(ctx, t.Origin)

		if err != nil {
			errCh <- err
			return
		}

		if originBalance < *t.Amount {
			errCh <- apperrors.NewTransferRequestError("not enough funds")
			return
		}

		if _, err := s.r.GetAccountById(ctx, *t.Destination); err != nil {
			if _, ok := err.(*apperrors.AccountNotFoundError); ok {
				errCh <- apperrors.NewTransferRequestError("destination account not found", err.Error())
			}

			errCh <- err
		}

		if err := s.r.AddTransfer(ctx, t); err != nil {
			errCh <- err
		}

		transferCh <- true
	}()

	select {
	case <-transferCh:
		return nil
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}
