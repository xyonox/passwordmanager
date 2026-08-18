package main

import (
	"bufio"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"

	_ "modernc.org/sqlite"
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

	_, err := os.Stat(dbFile)
	if err != nil {
		return ErrFirstRun
	}

	db, err := sql.Open("sqlite", dbFile)
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

func firstRun() error {

	fmt.Println("Enter password for your password database")
	fmt.Println("Password must be at least 8 characters")
	fmt.Print("> ")

	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return errors.New("inable to read firstRun input")
	}

	input := strings.TrimSpace(scanner.Text())

	if len(input) < 8 {
		return errors.New("password must be at least 8 characters")
	}

	//os.Create(dbFile)

	return nil
}

func run() error {

	// Password for database scanner input
	// if frist time run, set password

	err := loadSQL()

	if err == ErrFirstRun {
		err = firstRun()
		if err != nil {
			return err
		}
	}

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
