package config

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

var logger *log.Logger

func ConnDB() *sql.DB {
	dsn := "root:@tcp(127.0.0.1:3306)/go_libraryschool"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		logger.Fatalf("Error connecting to database: %v", err)
	}

	return db
}
