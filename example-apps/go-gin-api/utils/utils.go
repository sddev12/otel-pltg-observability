package utils

import (
	"errors"
	"io/fs"
	"os"

	"github.com/joho/godotenv"
)

func LoadEnvVars() error {
	if err := godotenv.Load(); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}

	requiredEnvVars := []string{"LOG_LEVEL"}
	for _, envVar := range requiredEnvVars {
		if _, exists := os.LookupEnv(envVar); !exists {
			return errors.New("missing required environment variable: " + envVar)
		}
	}

	return nil
}
