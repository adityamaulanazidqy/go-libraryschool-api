package auth_controller

import (
	context2 "context"
	"database/sql"
	"errors"
	"github.com/sirupsen/logrus"
	"go-libraryschool/helpers"
	"go-libraryschool/models/auth_models"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"strings"
	"time"
)

type RegisterController struct {
	Db        *sql.DB
	logLogrus *logrus.Logger
}

func NewRegisterController(db *sql.DB, logLogrus *logrus.Logger) *RegisterController {
	return &RegisterController{db, logLogrus}
}

// Register godoc
// @Summary Register user
// @Description Register with username, email, password (only email gmail make used)
// @Tags Auth
// @Accept multipart/form-data
// @Produce json
// @Param request body auth_models.RegisterRequest true "Register Request"
// @Success 201 {object} helpers.ApiResponse
// @Failure 400 {object} helpers.ApiResponse
// @Failure 401 {object} helpers.ApiResponse
// @Failure 500 {object} helpers.ApiResponse
// @Router /register [post]
func (RC *RegisterController) Register(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		RC.logLogrus.WithFields(logrus.Fields{
			"error":   err,
			"message": "Failed to parse multipart form",
		}).Error("Failed to parse multipart form")

		helpers.SendJson(w, http.StatusBadRequest, helpers.ApiResponse{
			Message: "Failed to parse multipart form",
		})
		return
	}

	file, handler, err := r.FormFile("profile")
	if err != nil {
		RC.logLogrus.WithFields(logrus.Fields{
			"error":   err,
			"message": "Failed to get profile image",
		}).Error("Failed to get profile image")

		helpers.SendJson(w, http.StatusBadRequest, helpers.ApiResponse{
			Message: "Failed to get profile image",
		})
		return
	}
	defer file.Close()

	filename, err := helpers.SaveImages().Profile(file, handler, "_")
	if err != nil {
		RC.logLogrus.WithFields(logrus.Fields{
			"error":   err,
			"message": "Failed to save profile image",
		}).Error("Failed to save profile image")

		helpers.SendJson(w, http.StatusBadRequest, helpers.ApiResponse{
			Message: "Failed to save profile image",
		})
		return
	}

	username := r.FormValue("username")
	if username == "" {
		helpers.SendJson(w, http.StatusBadRequest, helpers.ApiResponse{
			Message: "Username is required",
		})
	}

	email := r.FormValue("email")
	if !strings.HasSuffix(email, "@gmail.com") && !strings.Contains(email, "@") {
		http.Error(w, "Please input your email with google account.", http.StatusBadRequest)
		return
	}

	ctx, cancel := context2.WithTimeout(context2.Background(), 4*time.Second)
	defer cancel()

	query := "SELECT email FROM users WHERE email = ?"

	var tmpEmail string
	err = RC.Db.QueryRowContext(ctx, query, email).Scan(&tmpEmail)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		helpers.SendJson(w, http.StatusInternalServerError, helpers.ApiResponse{
			Message: "Database error",
		})
		return
	}

	if err == nil {
		helpers.SendJson(w, http.StatusUnauthorized, helpers.ApiResponse{
			Message: "Email already exists!",
		})
		return
	}

	password := r.FormValue("password")
	if password == "" {
		helpers.SendJson(w, http.StatusBadRequest, helpers.ApiResponse{
			Message: "Please input your password!",
		})
		return
	}

	if len(password) < 6 {
		helpers.SendJson(w, http.StatusBadRequest, helpers.ApiResponse{
			Message: "Password must be at least 6 characters!",
		})
		return
	}

	hashPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		helpers.SendJson(w, http.StatusInternalServerError, helpers.ApiResponse{
			Message: "Something went wrong",
			Data:    nil,
		})
		return
	}

	var roleID = 3

	queryInsert := "INSERT INTO users (username, email, password, profile, roleID, created_at) VALUES (?, ?, ?, ?, ?, ?)"
	_, err = RC.Db.ExecContext(ctx, queryInsert, username, email, hashPassword, filename, roleID, time.Now())
	if err != nil {
		http.Error(w, "Failed to insert data users!", http.StatusInternalServerError)
		return
	}

	var role string
	queryRole := "SELECT role FROM roles WHERE id = ?"
	err = RC.Db.QueryRowContext(ctx, queryRole, roleID).Scan(&role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			helpers.SendJson(w, http.StatusNotFound, helpers.ApiResponse{
				Message: "Role not found!",
			})
			return
		}

		helpers.SendJson(w, http.StatusInternalServerError, helpers.ApiResponse{
			Message: "Something went wrong",
		})
		return
	}

	user := auth_models.RegisterResponse{
		Username: username,
		Email:    email,
		Role:     role,
	}

	helpers.SendJson(w, http.StatusCreated, helpers.ApiResponse{
		Message: "User successfully registered!",
		Data:    user,
	})
}
