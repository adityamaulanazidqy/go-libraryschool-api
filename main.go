package main

import (
	"go-libraryschool/config"
	"go-libraryschool/docs"
	"go-libraryschool/middlewares"
	"go-libraryschool/routes"
	"log"
	"net/http"
)

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

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal(err)
	}
}
