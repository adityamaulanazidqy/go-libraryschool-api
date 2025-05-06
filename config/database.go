package config

import (
	"database/sql"
	"github.com/joho/godotenv"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

var logger *log.Logger

func ConnDB() *sql.DB {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	dsn := os.Getenv("MYSQL_USERNAME") + ":" + os.Getenv("MYSQL_PASSWORD") + "@tcp(" + os.Getenv("MYSQL_HOST") + ":" + os.Getenv("MYSQL_PORT") + ")/" + os.Getenv("MYSQL_DATABASE")
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		logger.Fatalf("Error connecting to database: %v", err)
	}

	return db
}
