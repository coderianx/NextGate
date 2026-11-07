package middlewares

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"

	"github.com/coderianx/NextGate/internal/models"
	"github.com/coderianx/NextGate/internal/store"
)

const apiKeyHeader = "X-API-Key"

// APIKey describes an API key authenticated by [APIKeyMiddleware]. It never
// contains the raw key material; the middleware only stores the validated
// database record.
type APIKey struct {
	// ID is the API key's database identifier.
	ID int64

	// UserID is the identifier of the user that owns the key.
	UserID int64

	// Username is the owner's account name.
	Username string

	// Title is the user-supplied label for the key.
	Title string
}

type apiKeyContextKey struct{}

// APIKeyMiddleware authenticates a request with an API key issued by the public
// API. The key may be sent as a Bearer token in the Authorization header or in
// the X-API-Key header; the Authorization header takes precedence.
//
// The middleware hashes the supplied key with SHA-256, looks the digest up in
// the apikeys table, and stores the owning user's claims plus the validated
// [APIKey] in the request context. Missing or unknown keys receive 401
// Unauthorized, and a missing database receives 500 Internal Server Error
// without calling next.
func APIKeyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if store.DB == nil {
			writeAuthError(w, http.StatusInternalServerError, "database_not_connected")
			return
		}

		apiKey, ok := extractAPIKey(r)
		if !ok {
			writeAuthError(w, http.StatusUnauthorized, "missing_api_key")
			return
		}

		digest := sha256.Sum256([]byte(apiKey))

		var key APIKey

		err := store.DB.QueryRowContext(
			r.Context(),
			`
			SELECT apikeys.id, apikeys.user_id, apikeys.title, users.username
			FROM apikeys
			JOIN users ON users.id = apikeys.user_id
			WHERE apikeys.api_key = $1
			`,
			hex.EncodeToString(digest[:]),
		).Scan(&key.ID, &key.UserID, &key.Title, &key.Username)

		if errors.Is(err, sql.ErrNoRows) {
			writeAuthError(w, http.StatusUnauthorized, "invalid_api_key")
			return
		}

		if err != nil {
			writeAuthError(w, http.StatusInternalServerError, "database_error")
			return
		}

		claims := &models.Claims{
			UserID:   key.UserID,
			Username: key.Username,
			Role:     roleUser,
		}

		ctx := context.WithValue(r.Context(), apiKeyContextKey{}, &key)
		ctx = context.WithValue(ctx, claimsContextKey{}, claims)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetAPIKey returns the API key stored in the request context by
// [APIKeyMiddleware]. The boolean is false when the middleware has not
// populated the context.
func GetAPIKey(r *http.Request) (*APIKey, bool) {
	apiKey, ok := r.Context().Value(apiKeyContextKey{}).(*APIKey)
	return apiKey, ok
}

// extractAPIKey returns the API key supplied in the Authorization or X-API-Key
// header. The Authorization header takes precedence when both are present.
func extractAPIKey(r *http.Request) (string, bool) {
	if authHeader := r.Header.Get("Authorization"); authHeader != "" {
		parts := strings.Fields(authHeader)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
			return "", false
		}

		return parts[1], true
	}

	key := strings.TrimSpace(r.Header.Get(apiKeyHeader))
	if key == "" {
		return "", false
	}

	return key, true
}
