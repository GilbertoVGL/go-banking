package apperrors

import (
	"fmt"
	"strings"
)

type ArgumentError struct {
	Context []string
	Err     error
}

func (e *ArgumentError) Error() string {
	return fmt.Sprintf("%s: %s", e.Err.Error(), strings.Join(e.Context, ", "))
}

type TransferRequestError struct {
	Context string
	Err     error
}

func (e *TransferRequestError) Error() string {
	return fmt.Sprintf("%s: %s", e.Err.Error(), e.Context)
}

type AccountNotFoundError struct {
	Context string
	Err     error
}

func (e *AccountNotFoundError) Error() string {
	return fmt.Sprintf("%s: %s", e.Err.Error(), e.Context)
}

type DatabaseError struct {
	Context string
	Err     error
}

func (e *DatabaseError) Error() string {
	return fmt.Sprintf("%s: %s", e.Err.Error(), e.Context)
}
