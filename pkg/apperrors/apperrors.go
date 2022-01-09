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
