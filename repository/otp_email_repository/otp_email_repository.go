package otp_email_repository

import (
	context2 "context"
	"database/sql"
	"errors"
	"github.com/sirupsen/logrus"
	"go-libraryschool/helpers"
	"net/http"
	"time"
)

type OtpEmailRepository struct {
	Db        *sql.DB
	logLogrus *logrus.Logger
}

func NewOtpEmailRepository(db *sql.DB, logLogrus *logrus.Logger) *OtpEmailRepository {
	return &OtpEmailRepository{
		Db:        db,
		logLogrus: logLogrus,
	}
}

func (repository *OtpEmailRepository) VerificationEmail(email string) (helpers.ApiResponse, int, error) {
	var response helpers.ApiResponse

	ctx, cancel := context2.WithTimeout(context2.Background(), 4*time.Second)
	defer cancel()

	query := "SELECT email FROM users WHERE email = ?"

	stmt, err := repository.Db.PrepareContext(ctx, query)
	if err != nil {
		response = helpers.ApiResponse{
			Message: "Something went wrong in prepare verification email!",
			Data:    nil,
		}

		return response, http.StatusInternalServerError, err
	}
	defer stmt.Close()

	err = stmt.QueryRowContext(ctx, email).Scan(&email)
	if !errors.Is(err, sql.ErrNoRows) {
		response = helpers.ApiResponse{
			Message: "Email already exists!",
			Data:    nil,
		}
		return response, http.StatusConflict, err
	}

	return response, http.StatusOK, nil
}
