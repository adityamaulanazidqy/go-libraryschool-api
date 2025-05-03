package auth_routes

import (
	"database/sql"
	"github.com/sirupsen/logrus"
	"go-libraryschool/controllers/auth_controller"
	"go-libraryschool/helpers"
	"net/http"
)

func UpdatePasswordRoute(mux *http.ServeMux, db *sql.DB, logLogrus *logrus.Logger) {
	mux.HandleFunc("/update-password", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			controller := auth_controller.NewUpdatePasswordController(db, logLogrus)
			controller.UpdatePassword(w, r)
		} else {
			logLogrus.WithFields(logrus.Fields{
				"message": "Method Not Allowed",
			}).Error("Method Not Allowed")

			helpers.SendJson(w, http.StatusMethodNotAllowed, helpers.ApiResponse{
				Message: "Method Not Allowed",
				Data:    nil,
			})
			return
		}
	})
}
