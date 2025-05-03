package auth_routes

import (
	"go-libraryschool/controllers/auth_controller"
	"go-libraryschool/helpers"
	"net/http"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

func LogoutRoute(mux *http.ServeMux, rdb *redis.Client, logLogrus *logrus.Logger) {
	mux.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			controller := auth_controller.NewLogoutController(rdb, logLogrus)
			controller.Logout(w, r)
		} else {
			logrus.WithFields(logrus.Fields{
				"message": "Method not allowed",
			}).Error("Method not allowed")

			helpers.SendJson(w, http.StatusMethodNotAllowed, helpers.ApiResponse{
				Message: "Method Not Allowed",
				Data:    nil,
			})
			return
		}
	})
}
