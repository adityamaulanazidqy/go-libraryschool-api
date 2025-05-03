package auth_controller

import (
	context2 "context"
	"database/sql"
	"encoding/json"
	"errors"
	"go-libraryschool/helpers"
	"go-libraryschool/models/auth_models"
	"go-libraryschool/models/identity"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"strings"
	"time"
)

type LoginController struct {
	Db *sql.DB
}

func NewLoginController(db *sql.DB) *LoginController {
	return &LoginController{db}
}

// Login godoc
// @Summary Login user
// @Description Login with email and password (only email gmail make used)
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body auth_models.LoginRequest true "Login Request"
// @Success 200 {object} helpers.ApiResponseAuthorization
// @Failure 400 {object} helpers.ApiResponse
// @Failure 401 {object} helpers.ApiResponse
// @Failure 404 {object} helpers.ApiResponse
// @Router /login [post]
func (controller LoginController) Login(w http.ResponseWriter, r *http.Request) {
	var user auth_models.LoginRequest

	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil || user.Password == "" || user.Email == "" {
		http.Error(w, "Please fill all field!", http.StatusBadRequest)
		return
	}

	if !strings.HasSuffix(user.Email, "@gmail.com") && !strings.Contains(user.Email, "@") {
		http.Error(w, "Please input your email with google account.", http.StatusBadRequest)
		return
	}

	ctx, cancel := context2.WithTimeout(context2.Background(), 4*time.Second)
	defer cancel()

	query := "SELECT id, username, email, password, roleID, created_at FROM users WHERE email = ?"

	var u identity.User
	err = controller.Db.QueryRowContext(ctx, query, user.Email).Scan(&u.Id, &u.Username, &u.Email, &u.Password, &u.RoleID, &u.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "User not found!", http.StatusNotFound)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(user.Password))
	if err != nil {
		helpers.SendJson(w, http.StatusUnauthorized, helpers.ApiResponse{
			Message: "Wrong Password!",
		})
		return
	}

	queryRole := "SELECT role FROM roles WHERE id = ?"
	err = controller.Db.QueryRowContext(ctx, queryRole, u.RoleID).Scan(&u.Role)
	if err != nil {
		helpers.SendJson(w, http.StatusBadRequest, helpers.ApiResponse{
			Message: "Role Not Found!",
		})
	}

	token, err := helpers.GenerateToken(u.Id, u.Email, u.Role)
	if err != nil {
		helpers.SendJson(w, http.StatusInternalServerError, helpers.ApiResponse{
			Message: "Failed to generate token!",
		})
	}

	var responseLogin = auth_models.LoginResponse{
		Username:  u.Username,
		Email:     u.Email,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
	}

	helpers.SendJsonAuthorization(w, http.StatusOK, helpers.ApiResponseAuthorization{
		Message: "Success!",
		Data:    responseLogin,
		Token:   token,
	})
}
