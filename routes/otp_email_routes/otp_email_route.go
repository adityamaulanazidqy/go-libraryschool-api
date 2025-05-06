package otp_email_routes

import (
	"database/sql"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"go-libraryschool/controllers/otp_email_controller"
	"go-libraryschool/helpers"
	"net/http"
)

func OtpEmailRoute(mux *http.ServeMux, db *sql.DB, logLogrus *logrus.Logger, rdb *redis.Client) {
	controller := otp_email_controller.NewOtpEmailController(db, logLogrus, rdb)

	mux.HandleFunc("/otp/send-otp", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			controller.OtpEmail(w, r)
		} else {
			helpers.SendJson(w, http.StatusMethodNotAllowed, helpers.ApiResponse{
				Message: "Invalid Method",
			})
		}
	})

	mux.HandleFunc("/otp/verify-otp", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			controller.VerifyOtp(w, r)
		} else {
			helpers.SendJson(w, http.StatusMethodNotAllowed, helpers.ApiResponse{
				Message: "Invalid Method",
			})
		}
	})
}
