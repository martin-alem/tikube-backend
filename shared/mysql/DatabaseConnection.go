package mysql

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"os"
	"time"
)

func ConnectToDatabase() *sql.DB {

	dns := os.Getenv("MSQL_DB")

	// Open a new database connection.
	db, err := sql.Open("mysql", dns)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}

	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	// Check if the database is accessible by pinging it
	if pingErr := db.Ping(); pingErr != nil {
		log.Fatalf("Error pinging database: %v", pingErr)
	}

	return db

}
