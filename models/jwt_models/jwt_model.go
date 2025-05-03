package jwt_models

import "github.com/golang-jwt/jwt/v5"

type JWTClaims struct {
	UserID int    `json:"user_id"`
	Email  string `json:"email"`
	Roles  string `json:"role"`
	jwt.RegisteredClaims
}
