package db

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

// Open establishes a database connection.
func Open(databaseURL string) (*sql.DB, error) {
	return sql.Open("mysql", databaseURL)
}