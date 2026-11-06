// Package secret provides authenticated encryption and decryption for secrets
// using AES-256-GCM.
package secret

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
)

const aes256KeySize = 32

// ErrInvalidKey is returned when the provided key is not 32 bytes long.
var ErrInvalidKey = errors.New("secret: AES-256 key must be 32 bytes")

// Encrypt encrypts plaintext with key using AES-256-GCM. The returned string
// contains the nonce followed by the ciphertext, encoded with base64.
func Encrypt(key []byte, plaintext string) (string, error) {
	gcm, err := newGCM(key)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("secret: generate nonce: %w", err)
	}

	encrypted := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// Decrypt decrypts a base64-encoded ciphertext produced by [Encrypt] using
// key. It returns an error when the key is invalid or the ciphertext has been
// modified.
func Decrypt(key []byte, ciphertext string) (string, error) {
	gcm, err := newGCM(key)
	if err != nil {
		return "", err
	}

	encrypted, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("secret: decode ciphertext: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(encrypted) < nonceSize {
		return "", errors.New("secret: ciphertext is too short")
	}

	nonce, ciphertextBytes := encrypted[:nonceSize], encrypted[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", fmt.Errorf("secret: decrypt data: %w", err)
	}

	return string(plaintext), nil
}

// newGCM returns an AES-GCM AEAD for a 32-byte AES-256 key.
func newGCM(key []byte) (cipher.AEAD, error) {
	if len(key) != aes256KeySize {
		return nil, ErrInvalidKey
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("secret: create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("secret: create GCM: %w", err)
	}

	return gcm, nil
}
