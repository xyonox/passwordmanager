package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"database/sql"
	"errors"
	"flag"
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
	dbFile = "database.db"
)

func DervireKey(password []byte, salt []byte) []byte {
	var time uint32 = 1
	var memory uint32 = 64 * 1024
	var threads uint8 = 4
	var keyLength uint32 = 32

	return argon2.IDKey(password, salt, time, memory, threads, keyLength)
}

func encryptFile(password []byte) error {

	salt := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return err
	}

	key := DervireKey(password, salt)

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

	ciphertext := gcm.Seal(nil, nonce, plain, nil)

	var fileData []byte
	fileData = append(fileData, salt...)
	fileData = append(fileData, nonce...)
	fileData = append(fileData, ciphertext...)

	return os.WriteFile(dbFile, fileData, 0644)
}

func decryptFile(password []byte) error {
	fileData, err := os.ReadFile(dbFile)
	if err != nil {
		return err
	}

	if len(fileData) < 16 {
		return fmt.Errorf("filedata too short")
	}
	salt := fileData[:16]

	key := DervireKey(password, salt)

	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	nonceStart := 16
	nonceEnd := nonceStart + gcm.NonceSize()
	if len(fileData) < nonceEnd {
		return fmt.Errorf("nonce not found in file data")
	}
	nonce := fileData[nonceStart:nonceEnd]
	ciphertext := fileData[nonceEnd:]

	// 4. Entschlüsseln
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return fmt.Errorf("FAILED: Wrong password? or corrupted file? " + err.Error())
	}

	return os.WriteFile(dbFile, plaintext, 0644)
}

func loadSQL() (*sql.DB, error) {
	_, err := os.Stat(dbFile)
	if errors.Is(err, os.ErrNotExist) {
		return nil, ErrFirstRun
	} else if err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		return nil, err
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			fmt.Println(err)
			return
		}
	}(db)

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	fmt.Println("PONG")

	return db, nil
}

func firstRun() error {
	fmt.Println("Enter password for your password database")
	fmt.Println("Password must be at least 8 characters")
	fmt.Print("> ")

	pw, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return err
	}

	if len(pw) < 8 {
		return errors.New("password must be at least 8 characters")
	}

	_, err = os.Create(dbFile)
	if err != nil {
		return err
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

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS passwords (id INTEGER PRIMARY KEY AUTOINCREMENT, website TEXT ,name TEXT, password TEXT)")
	if err != nil {
		return err
	}

	err = db.Close()
	if err != nil {
		return err
	}

	err = encryptFile(pw)
	if err != nil {
		return err
	}

	return nil
}

func run() error {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return errors.New("not a terminal")
	}

	websiteFlag := flag.String("w", "", "Website to search your password (REQUIRED)")

	flag.Parse()

	fmt.Println("Website: ", *websiteFlag)

	if *websiteFlag == "" {
		return errors.New("no website specified (Flag -w, for help use -h)")
	}

	db, err := loadSQL()

	if errors.Is(err, ErrFirstRun) {
		err = firstRun()
		if err != nil {
			return err
		}
	}

	if err != nil {
		secondErr := db.Close()
		if secondErr != nil {
			fmt.Println(err)
			return secondErr
		}
		return err
	}

	row := db.QueryRow("SELECT password FROM passwords WHERE website = ?", *websiteFlag)

	// Should not print out. instead, in copy
	fmt.Println("Password: ", row)

	err = db.Close()
	if err != nil {
		return err
	}

	return nil
}

func main() {
	_, err := os.Stat(dbFile)

	var pw []byte

	if err == nil {
		fmt.Println("Enter password for your password database")
		fmt.Print("> ")

		pw, err = term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		err = decryptFile(pw)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Println("")
	}

	if err := run(); err != nil {
		fmt.Println(err)
	}

	if pw == nil {
		return
	}

	err = encryptFile(pw)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
