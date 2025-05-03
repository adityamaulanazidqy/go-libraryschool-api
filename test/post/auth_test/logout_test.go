package auth_test

import (
	"context"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"go-libraryschool/controllers/auth_controller"
	"go-libraryschool/models/jwt_models"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestLogoutHandler(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6380",
		DB:   10,
	})
	defer rdb.FlushDB(context.Background())

	log := logrus.New()
	controller := auth_controller.NewLogoutController(rdb, log)

	claims := &jwt_models.JWTClaims{
		UserID: 4,
		Email:  "handayani@gmail.com",
		Roles:  "Manager",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	jwtKey := []byte(os.Getenv("JWT_KEY"))
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(jwtKey)
	if err != nil {
		t.Fatal("Failed to sign token:", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)

	rr := httptest.NewRecorder()

	controller.Logout(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200 OK, got %d", rr.Code)
	}

	ctx := context.Background()
	val, err := rdb.Get(ctx, "blacklist:"+tokenStr).Result()
	if err != nil {
		t.Errorf("Expected token to be blacklisted in Redis, got error: %v", err)
	}
	if val != "true" {
		t.Errorf("Expected blacklist value to be 'true', got %s", val)
	}
}
