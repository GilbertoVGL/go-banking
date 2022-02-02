package config

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/GilbertoVGL/go-banking/pkg/apperrors"
	"github.com/joho/godotenv"
)

var requiredEnvs []string = []string{
	"JWT_SECRET",
	"APP_PORT",
	"DB_PW",
	"DB_PORT",
	"DB_USER",
	"DB_HOST",
	"DB_NAME",
	"DB_MAX_IDLE_CONN",
	"DB_MAX_POOL",
	"RESQUEST_TIMEOUT_S",
	"SERVER_WRITE_TIMEOUT_S",
	"SERVER_READ_TIMEOUT_S",
	"SERVER_IDLE_TIMEOUT_S",
	"SERVER_READ_HEADER_TIMEOUT_S",
	"ORIGIN_ALLOWED",
}

var ServerWriteTimeout time.Duration
var ServerReadTimeout time.Duration
var ServerIdleTimeout time.Duration
var ServerReadHeaderTimeout time.Duration
var RequestTimeout time.Duration

func Load(path string) error {
	realPath := filepath.FromSlash(path)
	godotenv.Load(realPath)

	if err := verifyEnv(); err != nil {
		return err
	}

	if err := fillTimeOutValues(); err != nil {
		return err
	}

	return nil
}

func verifyEnv() error {
	var missing []string

	for _, env := range requiredEnvs {
		if v := os.Getenv(env); v == "" {
			missing = append(missing, env)
		}
	}

	if len(missing) > 0 {
		return apperrors.NewEnvVarError("Missing required variables", strings.Join(missing, ", "))
	}

	return nil
}

func fillTimeOutValues() error {
	var invalid []string

	rt, err := strconv.Atoi(os.Getenv("RESQUEST_TIMEOUT_S"))

	if err != nil {
		invalid = append(invalid, "RESQUEST_TIMEOUT_S")
	}

	swt, err := strconv.Atoi(os.Getenv("SERVER_WRITE_TIMEOUT_S"))

	if err != nil {
		invalid = append(invalid, "SERVER_WRITE_TIMEOUT_S")
	}

	srt, err := strconv.Atoi(os.Getenv("SERVER_READ_TIMEOUT_S"))

	if err != nil {
		invalid = append(invalid, "SERVER_READ_TIMEOUT_S")
	}

	sit, err := strconv.Atoi(os.Getenv("SERVER_IDLE_TIMEOUT_S"))

	if err != nil {
		invalid = append(invalid, "SERVER_IDLE_TIMEOUT_S")
	}

	srht, err := strconv.Atoi(os.Getenv("SERVER_READ_HEADER_TIMEOUT_S"))

	if err != nil {
		invalid = append(invalid, "SERVER_READ_HEADER_TIMEOUT_S")
	}

	if len(invalid) > 0 {
		return apperrors.NewEnvVarError("invalid env value", strings.Join(invalid, ", "))
	}

	RequestTimeout = (time.Duration(rt) * time.Second)
	ServerWriteTimeout = (time.Duration(swt) * time.Second)
	ServerReadTimeout = (time.Duration(srt) * time.Second)
	ServerIdleTimeout = (time.Duration(sit) * time.Second)
	ServerReadHeaderTimeout = (time.Duration(srht) * time.Second)

	return nil
}
