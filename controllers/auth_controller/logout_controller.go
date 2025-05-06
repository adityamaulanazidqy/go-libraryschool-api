package auth_controller

import (
	"context"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"go-libraryschool/helpers"
	"go-libraryschool/middlewares"
	"net/http"
	"time"
)

type LogoutController struct {
	rdb       *redis.Client
	logLogrus *logrus.Logger
}

func NewLogoutController(rdb *redis.Client, logLogrus *logrus.Logger) *LogoutController {
	return &LogoutController{
		rdb:       rdb,
		logLogrus: logLogrus,
	}
}

// Logout godoc
// @Summary Logout user
// @Description User Must first log in for used feature Logout
// @Tags Auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} helpers.ApiResponse
// @Failure 401 {object} helpers.ApiResponse
// @Failure 500 {object} helpers.ApiResponse
// @Router /logout [post]
func (controller *LogoutController) Logout(w http.ResponseWriter, r *http.Request) {
	token, err := middlewares.ExtractTokenFromHeader(r)
	if err != nil {
		controller.logLogrus.WithFields(logrus.Fields{
			"error":   err,
			"token":   token,
			"message": "Error extracting token",
		})

		helpers.SendJson(w, http.StatusInternalServerError, helpers.ApiResponse{
			Message: "You are not authorized to access this resource.",
		})
		return
	}

	if token == "" {
		controller.logLogrus.WithFields(logrus.Fields{
			"error":   err,
			"token":   token,
			"message": "Token is empty.",
		})
		helpers.SendJson(w, http.StatusUnauthorized, helpers.ApiResponse{
			Message: "Missing or invalid token.",
		})
		return
	}

	claims, err := middlewares.VerifyToken(token)
	if err != nil {
		controller.logLogrus.WithFields(logrus.Fields{
			"error":   err,
			"token":   token,
			"message": "Error verifying token",
		})

		helpers.SendJson(w, http.StatusUnauthorized, helpers.ApiResponse{
			Message: "Unauthorized.",
		})
		return
	}

	expDuration := time.Until(claims.ExpiresAt.Time)
	if expDuration <= 0 {
		expDuration = time.Minute * 1
	}

	ctx := context.Background()
	err = controller.rdb.Set(ctx, "blacklist:"+token, "true", expDuration).Err()
	if err != nil {
		controller.logLogrus.WithFields(logrus.Fields{
			"error":   err,
			"token":   token,
			"message": "Error setting blacklist",
		})

		helpers.SendJson(w, http.StatusInternalServerError, helpers.ApiResponse{
			Message: "Failed to logout.",
		})
		return
	}

	controller.logLogrus.Info("Successfully logged out.")

	helpers.SendJson(w, http.StatusOK, helpers.ApiResponse{
		Message: "You have successfully logged out.",
	})
	return
}
