package middlewares

import (
	"context"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"go-libraryschool/helpers"
	"go-libraryschool/models/jwt_models"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

var (
	jwtKey = []byte(os.Getenv("JWT_KEY"))
	rdb    *redis.Client
)

type contextKey string

const UserContextKey = contextKey("user")

func SetRedisClientMiddleware(redisClient *redis.Client) {
	rdb = redisClient
}

func JWTMiddleware(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Missing token", http.StatusUnauthorized)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "Invalid token format", http.StatusUnauthorized)
				return
			}

			tokenStr := parts[1]
			claims := &jwt_models.JWTClaims{}

			token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
				return jwtKey, nil
			})
			if err != nil || !token.Valid {
				http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
				return
			}

			ctxRedis := context.Background()
			if rdb != nil {
				blacklisted, err := rdb.Get(ctxRedis, "blacklist:"+tokenStr).Result()
				if err == nil && blacklisted == "true" {
					helpers.SendJson(w, http.StatusUnauthorized, helpers.ApiResponse{
						Message: "Token has been logged out",
					})
					return
				}
			}

			if len(allowedRoles) > 0 {
				roleMatch := false
				for _, role := range allowedRoles {
					if claims.Roles == role {
						roleMatch = true
						break
					}
				}
				if !roleMatch {
					http.Error(w, "Forbidden", http.StatusForbidden)
					return
				}
			}

			ctx := context.WithValue(r.Context(), UserContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func ExtractTokenFromHeader(r *http.Request) (string, error) {
	bearerToken := r.Header.Get("Authorization")
	parts := strings.Split(bearerToken, " ")
	if len(parts) == 2 && parts[0] == "Bearer" {
		return parts[1], nil
	}

	return "", nil
}

func VerifyToken(tokenStr string) (*jwt_models.JWTClaims, error) {
	claims := &jwt_models.JWTClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		fmt.Println("invalid token or expired token")
		return nil, errors.New("invalid token or expired token")
	}

	return claims, nil
}
