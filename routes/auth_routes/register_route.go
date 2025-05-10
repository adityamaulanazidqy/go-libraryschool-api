package auth_routes

import (
	"database/sql"
	"github.com/sirupsen/logrus"
	"go-libraryschool/controllers/auth_controller"
	"go-libraryschool/helpers"
	"net/http"
)

func RegisterRoute(mux *http.ServeMux, db *sql.DB, logLogrus *logrus.Logger) {
	mux.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			controller := auth_controller.NewRegisterController(db, logLogrus)
			controller.Register(w, r)
		} else {
			helpers.SendJson(w, http.StatusMethodNotAllowed, helpers.ApiResponse{
				Message: "Invalid Method",
			})
		}
	})
}
