package profile_repository

import (
	context2 "context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"go-libraryschool/helpers"
	request "go-libraryschool/models/request_models"
	response "go-libraryschool/models/response_models"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type ProfileRepository struct {
	Db        *sql.DB
	logLogrus *logrus.Logger
	rdb       *redis.Client
}

func NewProfileRepository(db *sql.DB, logLogrus *logrus.Logger, rdb *redis.Client) *ProfileRepository {
	return &ProfileRepository{
		Db:        db,
		logLogrus: logLogrus,
		rdb:       rdb,
	}
}

func (repository *ProfileRepository) GetProfileRepository(userID int) (helpers.ApiResponse, int, error) {
	var profile response.ProfileResponse

	ctx, cancel := context2.WithTimeout(context2.Background(), 4*time.Second)
	defer cancel()

	profileRedis, err := repository.rdb.Get(ctx, fmt.Sprintf("profile: %d", userID)).Result()
	if err != nil {
		responseRepo, code, err := repository.getProfileDBMysql(userID)
		return responseRepo, code, err
	}

	err = json.Unmarshal([]byte(profileRedis), &profile)
	if err != nil {
		repository.logLogrus.WithFields(logrus.Fields{
			"error":   err,
			"message": "Failed to Json Unmarshal",
		}).Error("Failed to Json Unmarshal")

		return helpers.ApiResponse{Message: "Failed to Json Unmarshal"}, http.StatusInternalServerError, err
	}

	return helpers.ApiResponse{Message: "Success get data profile retrieved from redis.", Data: profile}, http.StatusOK, nil
}

func (repository *ProfileRepository) getProfileDBMysql(userID int) (helpers.ApiResponse, int, error) {
	var profile response.ProfileResponse

	ctx, cancel := context2.WithTimeout(context2.Background(), 4*time.Second)
	defer cancel()

	query := "SELECT username, email, roleID FROM users WHERE id = ?"
	stmt, err := repository.Db.PrepareContext(ctx, query)
	if err != nil {
		repository.logLogrus.WithFields(logrus.Fields{
			"error":   err,
			"message": "failed to prepare statement",
		}).Error("failed to prepare statement")

		return helpers.ApiResponse{Message: "Failed to prepare statement"}, http.StatusInternalServerError, err
	}
	defer stmt.Close()

	err = stmt.QueryRowContext(ctx, userID).Scan(&profile.Username, &profile.Email, &profile.RoleID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			repository.logLogrus.WithFields(logrus.Fields{
				"error":   err,
				"message": "User not found",
			}).Error("User not found")

			return helpers.ApiResponse{Message: "User not found"}, http.StatusNotFound, err
		}

		repository.logLogrus.WithFields(logrus.Fields{
			"error":   err,
			"message": "failed to query row",
		}).Error("failed to query row")

		return helpers.ApiResponse{Message: "failed to query row"}, http.StatusInternalServerError, err
	}

	queryGenre := "SELECT role FROM roles WHERE id = ?"
	stmt, err = repository.Db.PrepareContext(ctx, queryGenre)
	if err != nil {
		repository.logLogrus.WithFields(logrus.Fields{
			"error":   err,
			"message": "failed to prepare statement",
		}).Error("failed to prepare statement")

		return helpers.ApiResponse{Message: "Failed to prepare statement"}, http.StatusInternalServerError, err
	}
	defer stmt.Close()

	err = stmt.QueryRowContext(ctx, profile.RoleID).Scan(&profile.Role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			repository.logLogrus.WithFields(logrus.Fields{
				"error":   err,
				"message": "Role not found",
			}).Error("Role not found")

			return helpers.ApiResponse{Message: "Role not found"}, http.StatusNotFound, err
		}

		repository.logLogrus.WithFields(logrus.Fields{
			"error":   err,
			"message": "failed to query row",
		}).Error("failed to query row")

		return helpers.ApiResponse{Message: "failed to query row"}, http.StatusInternalServerError, err
	}

	profileJSON, err := json.Marshal(profile)
	if err != nil {
		repository.logLogrus.WithFields(logrus.Fields{
			"error":   err,
			"message": "failed to marshal profile",
		}).Error("failed to marshal profile")

		return helpers.ApiResponse{Message: "failed to marshal profile"}, http.StatusInternalServerError, err
	}

	err = repository.rdb.Set(ctx, fmt.Sprintf("profile: %d", userID), profileJSON, 10*time.Second).Err()
	if err != nil {
		repository.logLogrus.WithFields(logrus.Fields{
			"error":   err,
			"message": "failed to set profile in redis",
		}).Error("failed to set profile in redis")

		return helpers.ApiResponse{Message: "failed to set profile in redis"}, http.StatusInternalServerError, err
	}

	return helpers.ApiResponse{Message: "Success get data profile retrieved from database.", Data: profile}, http.StatusOK, nil
}

func (repository *ProfileRepository) UpdateProfileRepository(profile *request.ProfileUpdate, profileID *int) (helpers.ApiResponse, int, error) {
	var existingProfile request.ProfileUpdate

	ctx, cancel := context2.WithTimeout(context2.Background(), 4*time.Second)
	defer cancel()

	query := "SELECT username, email FROM users WHERE id = ?"
	stmt, err := repository.Db.PrepareContext(ctx, query)
	if err != nil {
		repository.logLogrus.WithFields(logrus.Fields{
			"error":   err,
			"message": "failed to prepare statement",
		}).Error("failed to prepare statement")

		return helpers.ApiResponse{Message: "Failed to prepare statement"}, http.StatusInternalServerError, err
	}

	err = stmt.QueryRowContext(ctx, profileID).Scan(&existingProfile.Username, &existingProfile.Email)
	stmt.Close()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			repository.logLogrus.WithFields(logrus.Fields{
				"error":   err,
				"message": "User not found",
			}).Error("User not found")

			return helpers.ApiResponse{Message: "User not found"}, http.StatusNotFound, err
		}

		repository.logLogrus.WithFields(logrus.Fields{
			"error":   err,
			"message": "failed to query row",
		}).Error("failed to query row")
		return helpers.ApiResponse{Message: "Failed to query row"}, http.StatusInternalServerError, err
	}

	if strings.TrimSpace(profile.Username) == "" {
		profile.Username = existingProfile.Username
	}
	if strings.TrimSpace(profile.Email) == "" {
		profile.Email = existingProfile.Email
	}

	if existingProfile.Email != profile.Email {
		queryCheckEmail := "SELECT id FROM users WHERE email = ? AND id != ?"
		stmt, err = repository.Db.PrepareContext(ctx, queryCheckEmail)
		if err != nil {
			repository.logLogrus.WithFields(logrus.Fields{
				"error":   err,
				"message": "failed to prepare statement",
			}).Error("failed to prepare statement")

			return helpers.ApiResponse{Message: "Failed to prepare statement"}, http.StatusInternalServerError, err
		}
		var tempId int
		err = stmt.QueryRowContext(ctx, profile.Email, profileID).Scan(&tempId)
		stmt.Close()
		if err == nil {
			repository.logLogrus.WithFields(logrus.Fields{
				"error":   err,
				"message": "Email already exists",
			}).Error("Email already exists")

			return helpers.ApiResponse{Message: "Email already exists"}, http.StatusBadRequest, errors.New("email already exists")
		} else if err != nil && !errors.Is(err, sql.ErrNoRows) {
			repository.logLogrus.WithFields(logrus.Fields{
				"error":   err,
				"message": "Failed to check if email already exists",
			})

			return helpers.ApiResponse{Message: "Failed to check email"}, http.StatusInternalServerError, err
		}
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@gmail\.com$`)
	if !emailRegex.MatchString(profile.Email) {
		err = errors.New("invalid email format")
		repository.logLogrus.WithFields(logrus.Fields{
			"error":   err,
			"message": "invalid email format",
		}).Error("invalid email format")

		return helpers.ApiResponse{Message: "Invalid email format"}, http.StatusBadRequest, err
	}

	updateQuery := "UPDATE users SET username = ?, email = ? WHERE id = ?"
	stmt, err = repository.Db.PrepareContext(ctx, updateQuery)
	if err != nil {
		repository.logLogrus.WithFields(logrus.Fields{
			"error":   err,
			"message": "failed to prepare statement",
		}).Error("failed to prepare statement")

		return helpers.ApiResponse{Message: "Failed to prepare statement"}, http.StatusInternalServerError, err
	}
	result, err := stmt.ExecContext(ctx, profile.Username, profile.Email, profileID)
	stmt.Close()
	if err != nil {
		repository.logLogrus.WithFields(logrus.Fields{
			"error":   err,
			"message": "failed to update profile",
		}).Error("failed to update profile")

		return helpers.ApiResponse{Message: "Failed to update profile"}, http.StatusInternalServerError, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		repository.logLogrus.WithFields(logrus.Fields{
			"error":   err,
			"message": "Failed to get rows affected",
		}).Error("Failed to get rows affected")

		return helpers.ApiResponse{Message: "Failed to get rows affected"}, http.StatusInternalServerError, err
	}

	if rowsAffected == 0 {
		err := errors.New("no rows updated")
		repository.logLogrus.WithFields(logrus.Fields{
			"error":   err,
			"message": "No rows affected when updating profile",
		}).Error("No rows affected when updating profile")

		return helpers.ApiResponse{Message: "No rows affected"}, http.StatusInternalServerError, err
	}

	repository.logLogrus.Info("Successfully updated profile userID: ", profileID)

	return helpers.ApiResponse{
		Message: "Success update profile",
		Data:    profile,
	}, http.StatusOK, nil
}
