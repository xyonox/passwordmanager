package main

import (
	"fmt"
	"os"
)

// For env variables
// go get github.com/joho/godotenv

func loadSQL() {

}

func run() error {

	// Password for database scanner input
	// if frist time run, set password

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
