package routes

import (
	"database/sql"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	httpSwagger "github.com/swaggo/http-swagger"
	"go-libraryschool/controllers"
	"go-libraryschool/middlewares"
	AuthRoutes "go-libraryschool/routes/auth_routes"
	ManagementBookRoutes "go-libraryschool/routes/management_book_routes"
	OtpRoutes "go-libraryschool/routes/otp_email_routes"
	ProfileRoutes "go-libraryschool/routes/profile_routes"
	"net/http"
)

func Router(mux *http.ServeMux, db *sql.DB, logLogrus *logrus.Logger, rdb *redis.Client) *http.ServeMux {
	AuthRoutes.LoginRoute(mux, db)
	AuthRoutes.RegisterRoute(mux, db, logLogrus)
	AuthRoutes.LogoutRoute(mux, rdb, logLogrus)
	AuthRoutes.UpdatePasswordRoute(mux, db, logLogrus)

	OtpRoutes.OtpEmailRoute(mux, db, logLogrus, rdb)

	mux.Handle("/user/profile", middlewares.JWTMiddleware("Student", "Manager", "Librarian")(http.HandlerFunc(controllers.GetProfile)))

	mux.HandleFunc("/swagger/", httpSwagger.WrapHandler)

	ManagementBookRoutes.ManagementBookRoute(mux, db, logLogrus, rdb)

	ProfileRoutes.ProfileRoute(mux, db, logLogrus, rdb)

	return mux
}
