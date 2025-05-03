package auth_routes

import (
	"database/sql"
	"go-libraryschool/controllers/auth_controller"
	"go-libraryschool/helpers"
	"net/http"
)

func LoginRoute(mux *http.ServeMux, db *sql.DB) {
	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			controller := auth_controller.NewLoginController(db)
			controller.Login(w, r)
		} else {
			helpers.SendJson(w, http.StatusMethodNotAllowed, helpers.ApiResponse{
				Message: "Method Not Allowed",
				Data:    nil,
			})
			return
		}
	})
}
