package auth_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"go-libraryschool/controllers/auth_controller"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"golang.org/x/crypto/bcrypt"
)

func TestLoginHandler(t *testing.T) {
	log.Println("Starting TestLoginHandler")

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	email := "testuser@gmail.com"
	password := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	log.Println("Setting up user and role mock rows")
	rows := sqlmock.NewRows([]string{"id", "username", "email", "password", "roleID", "created_at"}).
		AddRow(1, "TestUser", email, string(hashedPassword), 2, time.Now())
	mock.ExpectQuery("SELECT id, username, email, password, roleID, created_at FROM users WHERE email = ?").
		WithArgs(email).WillReturnRows(rows)

	roleRows := sqlmock.NewRows([]string{"role"}).AddRow("Admin")
	mock.ExpectQuery("SELECT role FROM roles WHERE id = ?").
		WithArgs(2).WillReturnRows(roleRows)

	controller := auth_controller.NewLoginController(db)

	log.Println("Creating login request")
	loginBody := map[string]string{
		"email":    email,
		"password": password,
	}
	jsonBody, _ := json.Marshal(loginBody)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	controller.Login(rr, req)

	log.Printf("Response status: %d\n", rr.Code)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("expected status 200, got %d", status)
	}

	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("response is not valid JSON: %v", err)
	}

	log.Printf("Response body: %v\n", response)

	if response["message"] != "Success!" {
		t.Errorf("expected message Success!, got %v", response["message"])
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	log.Println("Starting TestLogin_UserNotFound")

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	email := "notfound@gmail.com"
	password := "somepassword"

	log.Println("Expecting no user row returned")
	mock.ExpectQuery("SELECT id, username, email, password, roleID, created_at FROM users WHERE email = ?").
		WithArgs(email).
		WillReturnError(sql.ErrNoRows)

	controller := auth_controller.NewLoginController(db)

	loginBody := map[string]string{
		"email":    email,
		"password": password,
	}
	jsonBody, _ := json.Marshal(loginBody)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	controller.Login(rr, req)

	log.Printf("Response status: %d\n", rr.Code)
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 for user not found, got %d", rr.Code)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	log.Println("Starting TestLogin_WrongPassword")

	db, mock, _ := sqlmock.New()
	defer db.Close()

	email := "test@gmail.com"
	correctPassword := "correct123"
	wrongPassword := "wrong123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(correctPassword), bcrypt.DefaultCost)

	log.Println("Mocking user row with correct password")
	mock.ExpectQuery("SELECT id, username, email, password, roleID, created_at FROM users WHERE email = ?").
		WithArgs(email).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "username", "email", "password", "roleID", "created_at",
		}).AddRow(1, "User", email, hashedPassword, 1, time.Now()))

	controller := auth_controller.NewLoginController(db)

	loginBody := map[string]string{
		"email":    email,
		"password": wrongPassword,
	}
	jsonBody, _ := json.Marshal(loginBody)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	controller.Login(rr, req)

	log.Printf("Response status: %d\n", rr.Code)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for wrong password, got %d", rr.Code)
	}
}

func TestLogin_InvalidEmailFormat(t *testing.T) {
	log.Println("Starting TestLogin_InvalidEmailFormat")

	db, _, _ := sqlmock.New()
	defer db.Close()

	controller := auth_controller.NewLoginController(db)

	loginBody := map[string]string{
		"email":    "invalidemail",
		"password": "anypassword",
	}
	jsonBody, _ := json.Marshal(loginBody)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	controller.Login(rr, req)

	log.Printf("Response status: %d\n", rr.Code)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid email format, got %d", rr.Code)
	}
}
