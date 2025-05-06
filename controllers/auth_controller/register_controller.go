package auth_controller

import (
	context2 "context"
	"database/sql"
	"encoding/json"
	"errors"
	"go-libraryschool/helpers"
	"go-libraryschool/models/auth_models"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"strings"
	"time"
)

type RegisterController struct {
	Db *sql.DB
}

func NewRegisterController(db *sql.DB) *RegisterController {
	return &RegisterController{db}
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
	var user auth_models.RegisterRequest

	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil || user.Username == "" || user.Email == "" || user.Password == "" {
		http.Error(w, "Please fill all field!", http.StatusBadRequest)
		return
	}

	if !strings.HasSuffix(user.Email, "@gmail.com") && !strings.Contains(user.Email, "@") {
		http.Error(w, "Please input your email with google account.", http.StatusBadRequest)
		return
	}

	ctx, cancel := context2.WithTimeout(context2.Background(), 4*time.Second)
	defer cancel()

	query := "SELECT email FROM users WHERE email = ?"

	var tmpEmail string
	err = RC.Db.QueryRowContext(ctx, query, user.Email).Scan(&tmpEmail)

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

	hashPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		helpers.SendJson(w, http.StatusInternalServerError, helpers.ApiResponse{
			Message: "Something went wrong",
			Data:    nil,
		})
		return
	}

	var roleID = 3

	queryInsert := "INSERT INTO users (username, email, password, roleID, created_at) VALUES (?, ?, ?, ?, ?)"
	_, err = RC.Db.ExecContext(ctx, queryInsert, user.Username, user.Email, hashPassword, roleID, time.Now())
	if err != nil {
		http.Error(w, "Failed to insert data users!", http.StatusInternalServerError)
		return
	}

	helpers.SendJson(w, http.StatusCreated, helpers.ApiResponse{
		Message: "User successfully registered!",
		Data:    user,
	})
}
