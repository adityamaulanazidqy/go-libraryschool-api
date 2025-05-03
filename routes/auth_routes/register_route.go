package auth_routes

import (
	"database/sql"
	"go-libraryschool/controllers/auth_controller"
	"go-libraryschool/helpers"
	"net/http"
)

func RegisterRoute(mux *http.ServeMux, db *sql.DB) {
	mux.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			controller := auth_controller.NewRegisterController(db)
			controller.Register(w, r)
		} else {
			helpers.SendJson(w, http.StatusMethodNotAllowed, helpers.ApiResponse{
				Message: "Invalid Method",
			})
		}
	})
}
