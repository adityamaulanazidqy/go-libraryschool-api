package config

import (
	"database/sql"
	"log"
	"os"

	"github.com/joho/godotenv"

	_ "github.com/go-sql-driver/mysql"
)

var logger *log.Logger

func ConnDB() (*sql.DB, error) {
	err := godotenv.Load(".env")
	if err != nil {
		logger.Fatalf("Error loading .env file: %v", err)
		return nil, err
	}

	dsn := os.Getenv("MYSQL_USERNAME") + ":" + os.Getenv("MYSQL_PASSWORD") + "@tcp(" + os.Getenv("MYSQL_HOST") + ":" + os.Getenv("MYSQL_PORT") + ")/" + os.Getenv("MYSQL_DATABASE")
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		logger.Fatalf("Error connecting to database: %v", err)
		return nil, err
	}

	return db, nil
}
