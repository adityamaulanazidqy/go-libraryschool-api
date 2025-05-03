package profile_routes

import (
	"database/sql"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"go-libraryschool/controllers/profile_controller"
	"go-libraryschool/helpers"
	"go-libraryschool/middlewares"
	"net/http"
)

func ProfileRoute(mux *http.ServeMux, db *sql.DB, logLogrus *logrus.Logger, rdb *redis.Client) {
	controller := profile_controller.NewProfileController(db, logLogrus, rdb)

	registerRoute := func(path string, method string, roles []string, handlerFunc func(http.ResponseWriter, *http.Request)) {
		mux.Handle(path, middlewares.JWTMiddleware(roles...)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != method {
				helpers.SendJson(w, http.StatusMethodNotAllowed, helpers.ApiResponse{
					Message: "Method Not Allowed",
				})
				return
			}
			handlerFunc(w, r)
		})))
	}

	registerRoute("/profile", http.MethodGet, []string{"Manager", "Librarian", "Student"}, controller.GetProfile)
	registerRoute("/profile/update-profile", http.MethodPut, []string{"Manager", "Librarian", "Student"}, controller.UpdateProfile)
}
