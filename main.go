package main

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
	"golang.org/x/term"

	_ "modernc.org/sqlite"
)

// For env variables
// go get github.com/joho/godotenv

// For database
// go get modernc.org/sqlite

// For clipboard
// go get golang.design/x/clipboard@latest

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
		return fmt.Errorf("file data too short")
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

	websiteFlag := flag.String("w", "", "Website to search your password (last prio)")

	saveNewPasswordFlag := flag.Bool("n", false, "Save a password (first prio)")

	editPasswordFlag := flag.String("e", "", "Edit a password, value has to be the website of the password (second prio)")

	deleteFlag := flag.String("d", "", "Delete a password, value has to be the website of the password (third prio)")

	flag.Parse()

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

	/*_, err = db.Exec("INSERT INTO passwords (website, name ,password) VALUES (?, ?,?)", "test.com", "testa testo", "pw")
	if err != nil {
		return err
	}

	*/

	if *saveNewPasswordFlag {
		fmt.Println("Enter password the new password ")
		fmt.Print("> ")

		pw, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			return err
		}
		fmt.Println("")

		fmt.Println("Enter website name")
		fmt.Print("> ")

		scanner := bufio.NewScanner(os.Stdin)
		if !scanner.Scan() {
			return errors.New("no input")
		}
		website := strings.TrimSpace(scanner.Text())

		fmt.Println("")

		fmt.Println("Enter a username for the password")
		fmt.Print("> ")

		if !scanner.Scan() {
			return errors.New("no input")
		}
		name := strings.TrimSpace(scanner.Text())
		fmt.Println("")

		_, err = db.Exec("INSERT INTO passwords (website, name ,password) VALUES (?, ?,?)", website, name, pw)
		if err != nil {
			return err
		}

	}

	if *editPasswordFlag != "" {
		fmt.Println("Searching the website: ", *editPasswordFlag)
		rows, err := db.Query(
			"SELECT id, name FROM passwords WHERE website = ?",
			*editPasswordFlag)
		if err != nil {
			return err
		}

		var id int
		var name string

		length := 0
		ids := []int{}

		for rows.Next() {
			err := rows.Scan(&id, &name)
			if err != nil {
				return err
			}

			fmt.Printf("%v: website: %v, name: %v\n", length, *editPasswordFlag, name)
			ids = append(ids, id)
			length++
		}

		finalId := -1
		scanner := bufio.NewScanner(os.Stdin)

		if length == 0 {
			fmt.Println("Website not found")
		} else if length == 1 {
			finalId = ids[0]
		} else if length > 1 {
			fmt.Println("multiple websites found, please select one:")
			fmt.Print("> ")
			if !scanner.Scan() {
				return errors.New("no input")
			}
			input := strings.TrimSpace(scanner.Text())
			inputInt, err := strconv.Atoi(input)
			if err != nil {
				return err
			}
			finalId = ids[inputInt]
		}

		fmt.Println(finalId)
		fmt.Println("Enter new password")
		fmt.Print("> ")
		pw, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			return err
		}
		fmt.Println("")
		_, err = db.Exec("UPDATE passwords SET password = ? WHERE id = ?", pw, finalId)
		if err != nil {
			return err
		}

	}

	if *deleteFlag != "" {
		fmt.Println("Searching the website: ", *deleteFlag)
		rows, err := db.Query(
			"SELECT id, name FROM passwords WHERE website = ?",
			*deleteFlag)
		if err != nil {
			return err
		}

		var id int
		var name string

		length := 0
		ids := []int{}

		for rows.Next() {
			err := rows.Scan(&id, &name)
			if err != nil {
				return err
			}

			fmt.Printf("%v: website: %v, name: %v\n", length, *deleteFlag, name)
			ids = append(ids, id)
			length++
		}

		finalId := -1
		scanner := bufio.NewScanner(os.Stdin)

		if length == 0 {
			fmt.Println("Website not found")
		} else if length == 1 {
			finalId = ids[0]
		} else if length > 1 {
			fmt.Println("multiple websites found, please select one:")
			fmt.Print("> ")
			if !scanner.Scan() {
				return errors.New("no input")
			}
			input := strings.TrimSpace(scanner.Text())
			inputInt, err := strconv.Atoi(input)
			if err != nil {
				return err
			}
			finalId = ids[inputInt]
		}

		_, err = db.Exec("DELETE FROM passwords WHERE id = ?", finalId)
		if err != nil {
			return err
		}
		fmt.Println("Website deleted")
	}

	if *websiteFlag != "" {
		var password string
		var name string

		rows, err := db.Query(
			"SELECT password, name FROM passwords WHERE website = ?",
			*websiteFlag)
		if err != nil {
			return err
		}

		length := 0
		passwords := []string{}

		for rows.Next() {
			err := rows.Scan(&password, &name)
			if err != nil {
				return err
			}

			fmt.Printf("%v: website: %v, name: %v\n", length, *websiteFlag, name)
			passwords = append(passwords, password)
			length++
		}

		if length == 0 {
			fmt.Println("No passwords found")
		} else if length == 1 {
			fmt.Println("Password: ", passwords[0])
		} else if length > 1 {
			fmt.Println("multiple passwords found, please select one:")
			scanner := bufio.NewScanner(os.Stdin)
			fmt.Print("> ")
			if !scanner.Scan() {
				return errors.New("no input")
			}
			input := strings.TrimSpace(scanner.Text())
			inputInt, err := strconv.Atoi(input)
			if err != nil {
				return err
			}
			fmt.Println("Password: ", passwords[inputInt])
		}
	}

	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			fmt.Println(err)
			return
		}
	}(db)

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
