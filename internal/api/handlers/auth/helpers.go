package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"time"
)

const refreshTokenSize = 32

const refreshTokenLifetime = 30 * 24 * time.Hour

// generateRefreshToken returns a cryptographically random, URL-safe refresh
// token suitable for one-time rotation.
func generateRefreshToken() (string, error) {
	token := make([]byte, refreshTokenSize)

	if _, err := io.ReadFull(rand.Reader, token); err != nil {
		return "", fmt.Errorf("api: generate refresh token: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(token), nil
}
