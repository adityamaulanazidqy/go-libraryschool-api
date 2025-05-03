package helpers

import (
	"github.com/golang-jwt/jwt/v5"
	"go-libraryschool/models/jwt_models"
	"os"
	"time"
)

var jwtKey = []byte(os.Getenv("JWT_KEY"))

func GenerateToken(userID int, email, role string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)

	claims := &jwt_models.JWTClaims{
		UserID: userID,
		Email:  email,
		Roles:  role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}
