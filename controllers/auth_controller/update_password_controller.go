package auth_controller

import (
	context2 "context"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/sirupsen/logrus"
	"go-libraryschool/helpers"
	"go-libraryschool/middlewares"
	"go-libraryschool/models/jwt_models"
	"go-libraryschool/models/request_models"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"time"
)

type UpdatePasswordController struct {
	Db        *sql.DB
	logLogrus *logrus.Logger
}

func NewUpdatePasswordController(db *sql.DB, logLogrus *logrus.Logger) *UpdatePasswordController {
	return &UpdatePasswordController{
		Db:        db,
		logLogrus: logLogrus,
	}
}

// UpdatePassword godoc
// @Summary UpdatePassword user
// @Description Used for user which want for update password
// @Tags Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body request_models.UpdatePasswordRequest false "Update Password Request"
// @Success 200 {object} helpers.ApiResponse
// @Failure 400 {object} helpers.ApiResponse
// @Failure 500 {object} helpers.ApiResponse
// @Router /update-password [put]
func (controller *UpdatePasswordController) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	var updatePassword request_models.UpdatePasswordRequest

	err := json.NewDecoder(r.Body).Decode(&updatePassword)
	if err != nil {
		controller.logLogrus.WithFields(logrus.Fields{
			"error":   err,
			"message": "Invalid request body",
		}).Error("Error in Update Password")

		helpers.SendJson(w, http.StatusBadRequest, helpers.ApiResponse{
			Message: "Request body not compatible with request body is want.",
		})
		return
	}

	claims := r.Context().Value(middlewares.UserContextKey).(*jwt_models.JWTClaims)

	var userID = claims.UserID

	if userID <= 0 {
		err := errors.New("invalid user ID")
		controller.logLogrus.WithFields(logrus.Fields{
			"error":   err,
			"message": "Invalid user id",
		}).Error("Error in Update Password")

		helpers.SendJson(w, http.StatusBadRequest, helpers.ApiResponse{
			Message: "Invalid user id",
		})
		return
	}

	if updatePassword.Password == "" {
		err := errors.New("invalid user password")
		controller.logLogrus.WithFields(logrus.Fields{
			"error":   err,
			"message": "Invalid user password",
		}).Error("Error in Update Password")

		helpers.SendJson(w, http.StatusBadRequest, helpers.ApiResponse{
			Message: "Invalid user password",
		})
		return
	}

	ctx, cancel := context2.WithTimeout(context2.Background(), 4*time.Second)
	defer cancel()

	query := "UPDATE users SET password = ? WHERE id = ?"
	stmt, err := controller.Db.PrepareContext(ctx, query)
	if err != nil {
		controller.logLogrus.WithFields(logrus.Fields{
			"error":   err,
			"message": "Failed to prepare statement",
		}).Error("Failed to prepare statement")

		helpers.SendJson(w, http.StatusInternalServerError, helpers.ApiResponse{
			Message: "Failed to prepare statement",
		})
		return
	}
	defer stmt.Close()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(updatePassword.Password), bcrypt.DefaultCost)
	if err != nil {
		controller.logLogrus.WithFields(logrus.Fields{
			"error":   err,
			"message": "Failed to hash password",
		}).Error("Failed to hash password")

		helpers.SendJson(w, http.StatusInternalServerError, helpers.ApiResponse{
			Message: "Failed to hash password",
		})
		return
	}

	result, err := stmt.ExecContext(ctx, string(hashedPassword), userID)
	if err != nil {
		controller.logLogrus.WithFields(logrus.Fields{
			"error":   err,
			"message": "Failed to update password",
		}).Error("Failed to update password")

		helpers.SendJson(w, http.StatusInternalServerError, helpers.ApiResponse{
			Message: "Failed to update password",
		})
		return
	}

	rowsEffect, err := result.RowsAffected()
	if err != nil {
		controller.logLogrus.WithFields(logrus.Fields{
			"error":   err,
			"message": "Failed to update password",
		}).Error("Failed to update password")

		helpers.SendJson(w, http.StatusInternalServerError, helpers.ApiResponse{
			Message: "Failed to update password",
		})
		return
	}

	controller.logLogrus.Info("Successfully updated password and e=rows effect:", rowsEffect)

	helpers.SendJson(w, http.StatusOK, helpers.ApiResponse{
		Message: "Successfully updated password.",
		Data:    nil,
	})
	return
}
