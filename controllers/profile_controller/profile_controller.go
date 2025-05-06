package profile_controller

import (
	"database/sql"
	"encoding/json"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"go-libraryschool/helpers"
	"go-libraryschool/middlewares"
	"go-libraryschool/models/jwt_models"
	profileModel "go-libraryschool/models/request_models"
	repository "go-libraryschool/repository/profile_repository"
	"net/http"
)

type ProfileController struct {
	Db          *sql.DB
	logLogrus   *logrus.Logger
	profileRepo *repository.ProfileRepository
	rdb         *redis.Client
}

func NewProfileController(db *sql.DB, logLogrus *logrus.Logger, rdb *redis.Client) *ProfileController {
	return &ProfileController{
		Db:          db,
		logLogrus:   logLogrus,
		rdb:         rdb,
		profileRepo: repository.NewProfileRepository(db, logLogrus, rdb),
	}
}

// GetProfile godoc
// @Summary GetProfile user
// @Description view data user example username, email and role.
// @Tags Profile
// @Produce json
// @Security BearerAuth
// @Success 200 {object} helpers.ApiResponse
// @Failure 404 {object} helpers.ApiResponse
// @Failure 500 {object} helpers.ApiResponse
// @Router /profile [get]
func (controller *ProfileController) GetProfile(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(middlewares.UserContextKey).(*jwt_models.JWTClaims)

	responseRepo, code, err := controller.profileRepo.GetProfileRepository(claims.UserID)
	if err != nil {
		helpers.SendJson(w, code, responseRepo)
		return
	}

	helpers.SendJson(w, code, responseRepo)
}

// UpdateProfile godoc
// @Summary Update Profile user
// @Description Used for update profile example username or email.
// @Tags Profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body request_models.ProfileUpdate false "Update Profile"
// @Success 200 {object} helpers.ApiResponse
// @Failure 400 {object} helpers.ApiResponse
// @Failure 404 {object} helpers.ApiResponse
// @Failure 500 {object} helpers.ApiResponse
// @Router /profile/update-profile [put]
func (controller *ProfileController) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(middlewares.UserContextKey).(*jwt_models.JWTClaims)

	var profile profileModel.ProfileUpdate

	err := json.NewDecoder(r.Body).Decode(&profile)
	if err != nil {
		controller.logLogrus.WithFields(logrus.Fields{
			"error":   err,
			"message": "failed to decode body",
		}).Error("failed to decode body")

		helpers.SendJson(w, http.StatusBadRequest, helpers.ApiResponse{
			Message: "invalid request body",
		})
		return
	}

	var profileID = claims.UserID

	responseRepo, code, err := controller.profileRepo.UpdateProfileRepository(&profile, &profileID)
	if err != nil {
		helpers.SendJson(w, code, responseRepo)
		return
	}

	helpers.SendJson(w, code, responseRepo)
}
