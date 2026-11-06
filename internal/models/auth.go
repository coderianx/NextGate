// Package models defines the JSON request types and JWT claims shared by the
// NextGate HTTP APIs.
package models

import "github.com/golang-jwt/jwt/v5"

// LoginAdminReq is the JSON body accepted by the administrator login endpoint.
type LoginAdminReq struct {
	// Password is the administrator password to verify.
	Password string `json:"password"`
}

// Claims is the JWT payload used for authenticated administrators and API
// users.
type Claims struct {
	// UserID is the authenticated user's database identifier. Administrator
	// tokens use zero.
	UserID int64 `json:"user_id"`

	// Username is the authenticated account name.
	Username string `json:"username"`

	// Role is the authorization role, either "admin" or "user".
	Role string `json:"role"`

	// RegisteredClaims carries the standard JWT timestamps and validation
	// fields.
	jwt.RegisteredClaims
}

// RegisterReq is the JSON body accepted by the public registration endpoint.
type RegisterReq struct {
	// Username is the unique account name to create.
	Username string `json:"username"`

	// Password is the plaintext password to hash and store.
	Password string `json:"password"`
}

// LoginReq is the JSON body accepted by the public login endpoint.
type LoginReq struct {
	// Username is the account name to authenticate.
	Username string `json:"username"`

	// Password is the plaintext password to verify.
	Password string `json:"password"`
}

// RefreshReq is the JSON body accepted by the token refresh endpoint.
type RefreshReq struct {
	// RefreshToken is the opaque refresh token issued by a previous login or
	// refresh response.
	RefreshToken string `json:"refresh_token"`
}
