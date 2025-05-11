package auth_controller

import (
	context2 "context"
	"database/sql"
	"encoding/json"
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
// @Accept json
// @Produce json
// @Param request body auth_models.RegisterRequest true "Register Request"
// @Success 201 {object} helpers.ApiResponse
// @Failure 400 {object} helpers.ApiResponse
// @Failure 401 {object} helpers.ApiResponse
// @Failure 500 {object} helpers.ApiResponse
// @Router /register [post]
func (RC *RegisterController) Register(w http.ResponseWriter, r *http.Request) {
	var req auth_models.RegisterRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		helpers.SendJson(w, http.StatusBadRequest, helpers.ApiResponse{
			Message: "Invalid request",
		})
		return
	}

	if req.Username == "" {
		helpers.SendJson(w, http.StatusBadRequest, helpers.ApiResponse{
			Message: "Username is required",
		})
	}

	if !strings.HasSuffix(req.Email, "@gmail.com") && !strings.Contains(req.Email, "@") {
		http.Error(w, "Please input your email with google account.", http.StatusBadRequest)
		return
	}

	ctx, cancel := context2.WithTimeout(context2.Background(), 4*time.Second)
	defer cancel()

	query := "SELECT email FROM users WHERE email = ?"

	var tmpEmail string
	err = RC.Db.QueryRowContext(ctx, query, req.Email).Scan(&tmpEmail)

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

	if req.Password == "" {
		helpers.SendJson(w, http.StatusBadRequest, helpers.ApiResponse{
			Message: "Please input your password!",
		})
		return
	}

	if len(req.Password) < 6 {
		helpers.SendJson(w, http.StatusBadRequest, helpers.ApiResponse{
			Message: "Password must be at least 6 characters!",
		})
		return
	}

	hashPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		helpers.SendJson(w, http.StatusInternalServerError, helpers.ApiResponse{
			Message: "Something went wrong",
			Data:    nil,
		})
		return
	}

	var roleID = 3

	queryInsert := "INSERT INTO users (username, email, password, roleID, created_at) VALUES (?, ?, ?, ?, ?)"
	stmt, err := RC.Db.PrepareContext(ctx, queryInsert)
	if err != nil {
		helpers.SendJson(w, http.StatusInternalServerError, helpers.ApiResponse{
			Message: "Something went wrong",
		})
		return
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, req.Username, req.Email, hashPassword, roleID, time.Now())
	if err != nil {
		helpers.SendJson(w, http.StatusInternalServerError, helpers.ApiResponse{
			Message: "Something went wrong",
		})
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
		Username: req.Username,
		Email:    req.Email,
		Role:     role,
	}

	helpers.SendJson(w, http.StatusCreated, helpers.ApiResponse{
		Message: "User successfully registered!",
		Data:    user,
	})
}
