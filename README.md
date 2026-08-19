# Password Manager

Password Manager is a small command-line application written in Go.
The project was created as a learning project for:

- encrypting and decrypting passwords
- storing passwords in an SQLite database
- working with command-line flags and terminal input

## Features

- First-run setup with a master password (at least 8 characters)
- Passwords stored in an SQLite database
- Database encryption at rest using Argon2id and AES-256-GCM
- Add, edit, delete and search password entries
- Copy a selected password to the system clipboard when multiple matches exist

## Requirements

- Go 1.25 or newer
- A terminal that supports hidden password input
- A working system clipboard (required when using the clipboard functionality)

## Running the application

Run the application from the project directory:

```bash
go run .
```

On the first run, the application creates `database.db` and asks for a master
password. On subsequent runs, enter the same password to decrypt the database.
The database is encrypted again when the application exits.

Available options:

```text
-n              Save a new password
-e <website>    Edit a password for a website
-d <website>    Delete a password for a website
-w <website>    Search for a password by website
```

Examples:

```bash
go run . -n
go run . -w example.com
go run . -e example.com
go run . -d example.com
```

## Building a binary

Create an executable with:

```bash
go build -o passwordmanager .
```

Then run it with:

```bash
./passwordmanager -w example.com
```

## How it works

The application stores the entries in a SQLite table named `passwords` with
the columns `id`, `website`, `name` and `password`.

Before the database is opened, the master password is used to derive a
32-byte key with Argon2id. The database file is then encrypted with AES-GCM.
Each encrypted file contains the 16-byte salt, the GCM nonce and the encrypted
database contents. After an operation completes, the plaintext SQLite file is
encrypted again.

Password input is hidden while entering the master password and new passwords.
When a website has multiple entries, the user can select one by its displayed
index; the selected password is copied to the clipboard.

## Example output

```text
Enter password for your password database
>
Searching the website:  example.com
0: website: example.com, name: alice
Copied to clipboard
```

## Project structure

```text
.
├── main.go       # 
├── go.mod        # Go module definition
├── go.sum        # Dependency checksums
└── README.md     # Project documentation
```

## Current limitations

- This is a learning project and has not been audited for security.
- The database is decrypted to `database.db` while the application is running.
- There is no password strength validation beyond the eight-character minimum
  for the master password.
- The command-line input and database error handling still need improvement.
- No automated tests or backup mechanism are included.
