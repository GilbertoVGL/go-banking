package apperrors

import (
	"fmt"
	"strings"
)

const DB_ERROR_PREFIX string = "database error"
const ARGUMENT_ERROR_PREFIX string = "invalid argument"
const TRANSFER_ERROR_PREFIX string = "transfer error"
const CONFIG_ERROR_PREFIX string = "configuration error"
const AUTH_ERROR_PREFIX string = "authentication error"
const VALIDATOR_ERROR_PREFIX string = "validator error"
const INTERNAL_ERROR_PREFIX string = "server error"

type ArgumentError struct {
	Context string
	Err     string
}

type TransferRequestError struct {
	Context string
	Err     string
}

type AccountNotFoundError struct {
	Context string
	Err     string
}

type DatabaseError struct {
	Context string
	Err     string
}

type EnvVarError struct {
	Context string
	Err     string
}

type AuthError struct {
	Context string
	Err     string
}

type ValidatorError struct {
	Context string
	Err     string
}

type InternalServerError struct {
	Context string
	Err     string
}

type RestError struct {
	Err string `json:"error"`
}

func (e *ArgumentError) Error() string {
	return fmt.Sprintf("%s: %s", e.Err, e.Context)
}

func (e *TransferRequestError) Error() string {
	return fmt.Sprintf("%s: %s", e.Err, e.Context)
}

func (e *AccountNotFoundError) Error() string {
	return fmt.Sprintf("%s: %s", e.Err, e.Context)
}

func (e *DatabaseError) Error() string {
	return fmt.Sprintf("%s: %s", e.Err, e.Context)
}

func (e *EnvVarError) Error() string {
	return fmt.Sprintf("%s: %s", e.Err, e.Context)
}

func (e *AuthError) Error() string {
	return fmt.Sprintf("%s: %s", e.Err, e.Context)
}

func (e *ValidatorError) Error() string {
	return fmt.Sprintf("%s: %s", e.Err, e.Context)
}

func (e *InternalServerError) Error() string {
	return fmt.Sprintf("%s: %s", e.Err, e.Context)
}

func NewArgumentError(context ...string) error {
	return &ArgumentError{Context: strings.Join(context, ": "), Err: ARGUMENT_ERROR_PREFIX}
}

func NewTransferRequestError(context ...string) error {
	return &TransferRequestError{Context: strings.Join(context, ": "), Err: TRANSFER_ERROR_PREFIX}
}

func NewAccountNotFoundError(context ...string) error {
	return &AccountNotFoundError{Context: strings.Join(context, ": "), Err: DB_ERROR_PREFIX}
}

func NewDatabaseError(context ...string) error {
	return &DatabaseError{Context: strings.Join(context, ": "), Err: DB_ERROR_PREFIX}
}

func NewEnvVarError(context ...string) error {
	return &EnvVarError{Context: strings.Join(context, ": "), Err: CONFIG_ERROR_PREFIX}
}

func NewAuthError(context ...string) error {
	return &AuthError{Context: strings.Join(context, ": "), Err: AUTH_ERROR_PREFIX}
}

func NewValidatorError(context ...string) error {
	return &ValidatorError{Context: strings.Join(context, ": "), Err: VALIDATOR_ERROR_PREFIX}
}

func NewInternalServerError(context ...string) error {
	return &InternalServerError{Context: strings.Join(context, ": "), Err: INTERNAL_ERROR_PREFIX}
}
