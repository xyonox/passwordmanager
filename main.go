package main

import (
	"errors"
	"fmt"
	"os"
)

// For env variables
// go get github.com/joho/godotenv

var ErrFirstRun = errors.New("first run")

func loadSQL() error {

	return ErrFirstRun

	return nil
}

func run() error {

	// Password for database scanner input
	// if frist time run, set password

	if errors.Is(loadSQL(), ErrFirstRun) {
		fmt.Println("First run erkannt")
		return nil
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
