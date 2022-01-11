package config

import (
	"os"
	"path/filepath"
	"strings"

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
}

func Load(path string) error {
	realPath := filepath.FromSlash(path)
	godotenv.Load(realPath)

	if err := verifyEnv(); err != nil {
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
