package apperrors

import (
	"fmt"
	"strings"
)

type ArgumentError struct {
	Arguments []string
	Err       error
}

func (e *ArgumentError) Error() string {
	return fmt.Sprintf("%s: %s", e.Err.Error(), strings.Join(e.Arguments, ", "))
}
