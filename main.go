package main

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
)

// For env variables
// go get github.com/joho/godotenv

// For database
// go get modernc.org/sqlite

var ErrFirstRun = errors.New("first run")

const (
	dbFile = "sql/database.sql"
)

func loadSQL() error {

	//return ErrFirstRun

	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		return err
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			fmt.Println(err)
			return
		}
	}(db)

	fmt.Println(db.Ping())

	return nil
}

func run() error {

	// Password for database scanner input
	// if frist time run, set password

	err := loadSQL()
	if err != nil {
		return err
	}

	return nil
}

func main() {
	// First steps
	// Load sqlite database
	// Create tables
	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func executeQuery(db *sql.DB, query string) {

}
