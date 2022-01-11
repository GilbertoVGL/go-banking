package apperrors

import (
	"fmt"
	"strings"
)

type ArgumentError struct {
	Context string
	Err     string
}

func (e *ArgumentError) Error() string {
	return fmt.Sprintf("%s: %s", e.Err, e.Context)
}

func NewArgumentError(context ...string) error {
	return &ArgumentError{Context: strings.Join(context, ": "), Err: "invalid argument"}
}

type TransferRequestError struct {
	Context string
	Err     string
}

func (e *TransferRequestError) Error() string {
	return fmt.Sprintf("%s: %s", e.Err, e.Context)
}

func NewTransferRequestError(context ...string) error {
	return &TransferRequestError{Context: strings.Join(context, ": "), Err: "transfer error"}
}

type AccountNotFoundError struct {
	Context string
	Err     string
}

func (e *AccountNotFoundError) Error() string {
	return fmt.Sprintf("%s: %s", e.Err, e.Context)
}

func NewAccountNotFoundError(context ...string) error {
	return &AccountNotFoundError{Context: strings.Join(context, ": "), Err: "database error"}
}

type DatabaseError struct {
	Context string
	Err     string
}

func (e *DatabaseError) Error() string {
	return fmt.Sprintf("%s: %s", e.Err, e.Context)
}

func NewDatabaseError(context ...string) error {
	return &DatabaseError{Context: strings.Join(context, ": "), Err: "database error"}
}
