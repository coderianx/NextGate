// Package middlewares provides HTTP middleware for authentication and
// authorization.
package middlewares

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/coderianx/NextGate/internal/config"
	"github.com/coderianx/NextGate/internal/models"
	"github.com/golang-jwt/jwt/v5"
)

const (
	roleAll   = "all"
	roleAdmin = "admin"
	roleUser  = "user"
)

type claimsContextKey struct{}

// AuthMiddleware validates an HS256 JWT supplied as a Bearer token and
// authorizes its role. The allowed value must be "all", "admin", or "user";
// "all" permits both admin and user tokens. Invalid tokens receive 401
// Unauthorized, insufficient roles receive 403 Forbidden, and an unsupported
// allowed value receives 500 Internal Server Error without calling next.
func AuthMiddleware(allowed string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !validAllowedRole(allowed) {
				writeAuthError(w, http.StatusInternalServerError, "invalid_role_configuration")
				return
			}

			authHeader := r.Header.Get("Authorization")
			parts := strings.Fields(authHeader)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
				writeAuthError(w, http.StatusUnauthorized, "invalid_authorization_header")
				return
			}

			secret := config.GetEnv("JWT_SECRET")
			if secret == "" {
				writeAuthError(w, http.StatusInternalServerError, "jwt_secret_not_configured")
				return
			}

			claims := &models.Claims{}
			token, err := jwt.ParseWithClaims(
				parts[1],
				claims,
				func(token *jwt.Token) (any, error) {
					return []byte(secret), nil
				},
				jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
				jwt.WithExpirationRequired(),
			)
			if err != nil || !token.Valid {
				writeAuthError(w, http.StatusUnauthorized, "invalid_token")
				return
			}

			if !roleAllowed(allowed, claims.Role) {
				writeAuthError(w, http.StatusForbidden, "insufficient_role")
				return
			}

			ctx := context.WithValue(r.Context(), claimsContextKey{}, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetClaims returns the verified JWT claims stored in the request context by
// [AuthMiddleware]. The boolean is false when the middleware has not populated
// the context.
func GetClaims(r *http.Request) (*models.Claims, bool) {
	claims, ok := r.Context().Value(claimsContextKey{}).(*models.Claims)
	return claims, ok
}

// validAllowedRole reports whether role is a supported authorization policy.
func validAllowedRole(role string) bool {
	switch role {
	case roleAll, roleAdmin, roleUser:
		return true
	default:
		return false
	}
}

// roleAllowed reports whether an authenticated role satisfies the requested
// policy. The "all" policy permits both administrator and user roles.
func roleAllowed(allowed, role string) bool {
	if role != roleAdmin && role != roleUser {
		return false
	}

	return allowed == roleAll || allowed == role
}

// writeAuthError writes a JSON error response with the supplied status code.
func writeAuthError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
