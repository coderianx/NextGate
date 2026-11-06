// Package auth provides registration, login, and refresh-token handlers for
// the public API.
package auth

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/coderianx/NextGate/internal/config"
	"github.com/coderianx/NextGate/internal/models"
	"github.com/coderianx/NextGate/internal/store"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// RegisterHandler creates a user from the JSON request body and stores the
// password as a bcrypt hash. It responds with 201 on success and 409 when the
// username is already registered.
func RegisterHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set("Content-Type", "application/json")

	if store.DB == nil {
		w.WriteHeader(500)

		json.NewEncoder(w).Encode(map[any]string{
			"error": "database_not_connected",
		})

		return
	}

	var req models.RegisterReq

	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "invalid_request",
		})

		return
	}

	if req.Username == "" {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "username_is_required",
		})

		return
	}

	if len(req.Username) > 20 {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "username_is_too_long",
		})
	}

	if req.Password == "" {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "password_is_required",
		})

		return
	}

	if len(req.Password) > 72 {

		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "password_must_not_exceed_72_bytes",
		})

		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(req.Password),
		12,
	)

	if err != nil {
		w.WriteHeader(500)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "bcrypt_error",
		})

		return
	}

	_, err = store.DB.Exec(
		`
		INSERT INTO users (username, password)
		VALUES ($1,$2)
		`,
		req.Username,
		string(hashedPassword),
	)

	if err != nil {
		if store.IsUniqueError(err) {
			w.WriteHeader(409)

			json.NewEncoder(w).Encode(map[string]any{
				"error": "username_already_exists",
			})

			return
		}

		w.WriteHeader(500)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "database_error",
		})

		return
	}

	w.WriteHeader(201)

	json.NewEncoder(w).Encode(map[string]any{
		"message": "user_registered",
	})
}

// LoginHandler verifies a user's credentials and returns a short-lived access
// JWT together with a refresh token. Only a SHA-256 hash of the refresh token
// is stored in the database.
func LoginHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set("Content-Type", "application/json")

	if store.DB == nil {
		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "database_not_connected",
		})

		return
	}

	var req models.LoginReq

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "invalid_request",
		})

		return
	}

	if req.Username == "" {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "username_is_required",
		})

		return
	}

	if req.Password == "" {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "password_is_required",
		})

		return
	}

	var (
		userID         int64
		username       string
		hashedPassword string
	)

	err := store.DB.QueryRow(
		`
		SELECT id, username, password
		FROM users
		WHERE username = $1
		`,
		req.Username,
	).Scan(&userID, &username, &hashedPassword)

	if errors.Is(err, sql.ErrNoRows) {
		w.WriteHeader(http.StatusUnauthorized)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "invalid_credentials",
		})

		return
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "database_error",
		})

		return
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(hashedPassword),
		[]byte(req.Password),
	); err != nil {
		w.WriteHeader(http.StatusUnauthorized)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "invalid_credentials",
		})

		return
	}

	jwtSecret := config.GetEnv("JWT_SECRET")

	if jwtSecret == "" {
		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "jwt_secret_not_configured",
		})

		return
	}

	now := time.Now()
	claims := models.Claims{
		UserID:   userID,
		Username: username,
		Role:     "user",
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(60 * time.Minute)),
		},
	}

	accessToken, err := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	).SignedString([]byte(jwtSecret))

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "jwt_error",
		})

		return
	}

	randomToken, err := generateRefreshToken()

	if err != nil {
		w.WriteHeader(500)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "generate_refresh_token",
		})

		return
	}

	refreshTokenHash := sha256.Sum256([]byte(randomToken))
	expiresAt := time.Now().Add(refreshTokenLifetime)

	_, err = store.DB.Exec(
		`
		INSERT INTO refresh_tokens (user_id, token, expires_at)
		VALUES ($1,$2,$3)
		`,
		userID,
		hex.EncodeToString(refreshTokenHash[:]),
		expiresAt,
	)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "refresh_token_save_error",
		})

		return
	}

	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(map[string]any{
		"message":      "user_logged_in",
		"accessToken":  accessToken,
		"refreshToken": randomToken,
	})
}

// RefreshHandler validates and rotates a refresh token, returning a new access
// JWT and refresh token. The old token is deleted and the replacement is stored
// in the same database transaction.
func RefreshHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set("Content-Type", "application/json")

	if store.DB == nil {
		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "database_not_connected",
		})

		return
	}

	var req models.RefreshReq

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "invalid_request",
		})

		return
	}

	if req.RefreshToken == "" {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "refresh_token_is_required",
		})

		return
	}

	jwtSecret := config.GetEnv("JWT_SECRET")
	if jwtSecret == "" {
		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "jwt_secret_not_configured",
		})

		return
	}

	oldTokenHash := sha256.Sum256([]byte(req.RefreshToken))

	tx, err := store.DB.BeginTx(r.Context(), nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "database_error",
		})

		return
	}
	defer tx.Rollback()

	var (
		userID    int64
		username  string
		expiresAt time.Time
	)

	err = tx.QueryRow(
		`
		SELECT users.id, users.username, refresh_tokens.expires_at
		FROM refresh_tokens
		JOIN users ON users.id = refresh_tokens.user_id
		WHERE refresh_tokens.token = $1
		`,
		hex.EncodeToString(oldTokenHash[:]),
	).Scan(&userID, &username, &expiresAt)

	if errors.Is(err, sql.ErrNoRows) {
		w.WriteHeader(http.StatusUnauthorized)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "invalid_refresh_token",
		})

		return
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "database_error",
		})

		return
	}

	now := time.Now()
	if !expiresAt.After(now) {
		w.WriteHeader(http.StatusUnauthorized)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "refresh_token_expired",
		})

		return
	}

	newRefreshToken, err := generateRefreshToken()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "generate_refresh_token",
		})

		return
	}

	newRefreshTokenHash := sha256.Sum256([]byte(newRefreshToken))
	newExpiresAt := now.Add(refreshTokenLifetime)

	claims := models.Claims{
		UserID:   userID,
		Username: username,
		Role:     "user",
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(60 * time.Minute)),
		},
	}

	accessToken, err := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	).SignedString([]byte(jwtSecret))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "jwt_error",
		})

		return
	}

	if _, err := tx.Exec(
		`
		DELETE FROM refresh_tokens
		WHERE token = $1
		`,
		hex.EncodeToString(oldTokenHash[:]),
	); err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "database_error",
		})

		return
	}

	if _, err := tx.Exec(
		`
		INSERT INTO refresh_tokens (user_id, token, expires_at)
		VALUES ($1,$2,$3)
		`,
		userID,
		hex.EncodeToString(newRefreshTokenHash[:]),
		newExpiresAt,
	); err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "refresh_token_save_error",
		})

		return
	}

	if err := tx.Commit(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "database_error",
		})

		return
	}

	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(map[string]any{
		"message":      "token_refreshed",
		"accessToken":  accessToken,
		"refreshToken": newRefreshToken,
	})
}
