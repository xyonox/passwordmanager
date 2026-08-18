package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"

	"golang.org/x/crypto/argon2"
	"golang.org/x/term"

	_ "modernc.org/sqlite"
)

// For env variables
// go get github.com/joho/godotenv

// For database
// go get modernc.org/sqlite

var ErrFirstRun = errors.New("first run")

const (
	dbFile = "test.txt"
)

func DervireKey(password []byte, salt []byte) []byte {
	var time uint32 = 1
	var memory uint32 = 64 * 1024
	var threads uint8 = 4
	var keyLength uint32 = 32

	return argon2.IDKey(password, salt, time, memory, threads, keyLength)
}

func encryptFile(password string) error {

	salt := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return err
	}

	key := DervireKey([]byte(password), salt)

	plain, err := os.ReadFile(dbFile)
	if err != nil {
		return err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	nonce := make([]byte, gcm.NonceSize())

	ciphertext := gcm.Seal(nonce, nonce, plain, nil)

	var fileData []byte
	fileData = append(fileData, salt...)
	fileData = append(fileData, nonce...)
	fileData = append(fileData, ciphertext...)

	return os.WriteFile(dbFile+".enc", fileData, 0644)
}

func decryptFile(password []byte) error {

	return nil
}

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

	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return errors.New("not a terminal")
	}

	fmt.Println("Enter password for your password database")
	fmt.Println("Password must be at least 8 characters")
	fmt.Print("> ")

	pw, err := term.ReadPassword(int(os.Stdin.Fd()))

	if err != nil {
		return err
	}

	// Print password TODO REMOVE
	for _, c := range pw {
		fmt.Printf("%c", c)
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

	fmt.Print(encryptFile("HeyHuHa"))

	/*if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	*/
}

func executeQuery(db *sql.DB, query string) {

}
