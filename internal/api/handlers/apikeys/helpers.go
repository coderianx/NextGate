// Package apikeys provides handlers and helpers for managing API keys issued to
// authenticated users.
package apikeys

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

const (
	apiKeySize   = 32
	apiKeyPrefix = "sk-ng-v1-"
)

// generateAPIKey returns a cryptographically random, URL-safe API key.
func generateAPIKey() (string, error) {
	key := make([]byte, apiKeySize)

	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return "", fmt.Errorf("apikeys: generate API key: %w", err)
	}

	return apiKeyPrefix + base64.RawURLEncoding.EncodeToString(key), nil
}
