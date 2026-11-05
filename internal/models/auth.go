package models

import "github.com/golang-jwt/jwt/v5"

type LoginAdminReq struct {
	Password string `json:"password"`
}

type Claims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`

	jwt.RegisteredClaims
}
