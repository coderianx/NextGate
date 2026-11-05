package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)

var nextgatePath string

var requiredEnvNames = []string{"JWT_SECRET"}

func Init() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("config: get user home dir: %w", err)
	}

	nextgatePath = filepath.Join(home, ".nextgate")

	if err := os.MkdirAll(nextgatePath, 0755); err != nil {
		return fmt.Errorf("config: create nextgate folder: %w", err)
	}

	envPath := filepath.Join(nextgatePath, ".env")

	file, err := os.OpenFile(
		envPath,
		os.O_CREATE|os.O_EXCL|os.O_WRONLY,
		0600,
	)

	if err != nil && !errors.Is(err, os.ErrExist) {
		return fmt.Errorf("config: create env file: %w", err)
	}

	if file != nil {
		if _, err := file.WriteString("JWT_SECRET="); err != nil {
			file.Close()
			return fmt.Errorf("config: write env file: %w", err)
		}

		if err := file.Close(); err != nil {
			return fmt.Errorf("config: close env file: %w", err)
		}
	}

	if err := godotenv.Load(envPath); err != nil {
		return fmt.Errorf("config: load env file: %w", err)
	}

	return nil
}

func GetNextGatePath() string {
	return nextgatePath
}

// GetEnv returns the value of an environment variable. Init loads the NextGate
// env file without overriding variables already present in the process
// environment.
func GetEnv(envName string) string {
	return os.Getenv(envName)
}

// CheckEnvs verifies that every required environment variable has a value.
func CheckEnvs() error {
	missing := make([]string, 0, len(requiredEnvNames))

	for _, envName := range requiredEnvNames {
		if GetEnv(envName) == "" {
			missing = append(missing, envName)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf(
			"config: missing required environment variables: %s",
			strings.Join(missing, ", "),
		)
	}

	return nil
}
