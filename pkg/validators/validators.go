package validators

import (
	"errors"
	"regexp"
	"strconv"
	"strings"

	"github.com/GilbertoVGL/go-banking/pkg/apperrors"
)

var CPFRegex = regexp.MustCompile(`^(\d{3}.\d{3}.\d{3}-\d{2})$`)

func getVerifyingDigit(starts int, uniqueDigits string) int {
	sum := 0
	for i := 0; i < len(uniqueDigits); i++ {
		v, _ := strconv.Atoi(string(uniqueDigits[i]))
		sum += starts * v
		starts++
	}

	if r := sum % 11; r == 10 {
		return 0
	} else {
		return r
	}
}

func ValidateCPF(cpf string) error {
	if !CPFRegex.MatchString(cpf) {
		return &apperrors.ArgumentError{Context: []string{"invalid CPF format or value"}, Err: "invalid argument"}
	}

	cpf = strings.ReplaceAll(cpf, ".", "")
	cpf = strings.ReplaceAll(cpf, "-", "")

	if _, err := strconv.Atoi(cpf); err != nil {
		return err
	}

	firstDigit, _ := strconv.Atoi(string(cpf[9]))

	if firstVerifier := getVerifyingDigit(1, cpf[0:9]); firstVerifier != firstDigit {
		return errors.New("invalid CPF")
	}

	secondDigit, _ := strconv.Atoi(string(cpf[10]))

	if secondVerifier := getVerifyingDigit(0, cpf[0:10]); secondVerifier != secondDigit {
		return errors.New("invalid CPF")
	}

	return nil
}
