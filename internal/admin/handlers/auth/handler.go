// Package auth provides HTTP handlers for authenticating administrator
// sessions.
package auth

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/coderianx/NextGate/internal/admin/password"
	"github.com/coderianx/NextGate/internal/config"
	"github.com/coderianx/NextGate/internal/models"
	"github.com/coderianx/NextGate/internal/store"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// AdminLoginHandler authenticates an administrator against the bcrypt hash
// stored in the system keyring and returns a signed JWT. It responds with 400
// for malformed requests, 409 for invalid credentials, and 500 when the
// database, keyring, or token signing is unavailable.
func AdminLoginHandler(
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

	var req models.LoginAdminReq

	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "invalid_request",
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

	adminPasswordKeyring, err := password.GetAdminPassword()

	if err != nil {
		log.Printf("[ADMIN] Keyring Admin Password: %w\n", err)

		w.WriteHeader(500)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "keyring_admin_error",
		})

		return
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(adminPasswordKeyring),
		[]byte(req.Password),
	)

	if err != nil {
		w.WriteHeader(409)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "invalid_credentials",
		})

		return
	}

	claims := models.Claims{
		UserID:   0,
		Username: "admin",
		Role:     "admin",

		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt: jwt.NewNumericDate(
				time.Now(),
			),

			ExpiresAt: jwt.NewNumericDate(
				time.Now().Add(60 * time.Minute),
			),
		},
	}

	secretJwt := config.GetEnv("JWT_SECRET")

	access := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	accessToken, err := access.SignedString(
		[]byte(secretJwt),
	)

	if err != nil {

		log.Printf("[ADMIN] JWT error: %w\n", err)

		w.WriteHeader(500)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "jwt_error",
		})
		return
	}

	w.WriteHeader(200)
	json.NewEncoder(w).Encode(map[string]any{
		"message":     "admin_logined",
		"accessToken": accessToken,
	})
}
