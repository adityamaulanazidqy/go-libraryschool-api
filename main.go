package main

import (
	"go-libraryschool/config"
	"go-libraryschool/controllers/otp_email_controller"
	"go-libraryschool/docs"
	"go-libraryschool/middlewares"
	"go-libraryschool/routes"
	"log"
	"net/http"
)

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and your JWT token.
func main() {
	docs.SwaggerInfo.Title = "Go Library School API"
	docs.SwaggerInfo.Description = "This is a school API"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = "localhost:8080"
	docs.SwaggerInfo.Schemes = []string{"http"}

	db := config.ConnDB()

	rdb := config.ConnRedis()

	logLogrus := config.LogrusLogger()

	mux := http.NewServeMux()
	routes.Router(mux, db, logLogrus, rdb)

	middlewares.SetRedisClientMiddleware(rdb)

	otp_email_controller.SetOtpEmail()

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal(err)
	}
}
