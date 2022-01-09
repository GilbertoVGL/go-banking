package apperrors

import (
	"fmt"
	"strings"
)

type ArgumentError struct {
	Context []string
	Err     string
}

func (e *ArgumentError) Error() string {
	return fmt.Sprintf("%s: %s", e.Err, strings.Join(e.Context, ", "))
}

type TransferRequestError struct {
	Context string
	Err     string
}

func (e *TransferRequestError) Error() string {
	return fmt.Sprintf("%s: %s", e.Err, e.Context)
}

type AccountNotFoundError struct {
	Context string
	Err     string
}

func (e *AccountNotFoundError) Error() string {
	return fmt.Sprintf("%s: %s", e.Err, e.Context)
}

type DatabaseError struct {
	Context string
	Err     string
}

func (e *DatabaseError) Error() string {
	return fmt.Sprintf("%s: %s", e.Err, e.Context)
}
