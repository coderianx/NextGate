package config

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"github.com/zalando/go-keyring"
)

var nextgatePath string

var requiredEnvNames = []string{"JWT_SECRET"}

const (
	keyringService = "nextgate"
	aesKeyName     = "aes-256-key"
	aes256KeySize  = 32
)

// ErrAESKeyNotFound is returned when no AES-256 key is stored in the system
// keyring.
var ErrAESKeyNotFound = errors.New("config: AES-256 key not found")

// GetAESKey returns the AES-256 key stored in the system keyring.
func GetAESKey() ([]byte, error) {
	storedKey, err := keyring.Get(keyringService, aesKeyName)
	if errors.Is(err, keyring.ErrNotFound) {
		return nil, ErrAESKeyNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("config: keyring get error: %w", err)
	}

	key, err := base64.StdEncoding.DecodeString(storedKey)
	if err != nil {
		return nil, fmt.Errorf("config: decode AES key: %w", err)
	}

	if len(key) != aes256KeySize {
		return nil, fmt.Errorf("config: invalid AES-256 key length: %d", len(key))
	}

	return key, nil
}

// CreateAESKey generates and stores a new random AES-256 key. If a key already
// exists in the system keyring, it leaves it unchanged and returns nil.
func CreateAESKey() error {
	_, err := keyring.Get(keyringService, aesKeyName)
	if err == nil {
		return nil
	}

	if !errors.Is(err, keyring.ErrNotFound) {
		return fmt.Errorf("config: keyring get error: %w", err)
	}

	key := make([]byte, aes256KeySize)
	if _, err := rand.Read(key); err != nil {
		return fmt.Errorf("config: generate AES-256 key: %w", err)
	}

	if err := keyring.Set(keyringService, aesKeyName, base64.StdEncoding.EncodeToString(key)); err != nil {
		return fmt.Errorf("config: keyring set error: %w", err)
	}

	return nil
}

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
