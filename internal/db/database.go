package db

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

var DB *sql.DB

func ensureAdmin() {
	var count int

	err := DB.QueryRow("SELECT COUNT(*) FROM users WHERE role = 'admin'").Scan(&count)
	if err != nil {
		log.Fatal(err)
	}

	if count > 0 {
		return // admin already exists
	}

	password := os.Getenv("ADMIN_PASSWORD")
	if password == "" {
		log.Fatal("ADMIN_PASSWORD environment variable not set")
	}

	// hash the password
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}

	_, err = DB.Exec(`
		INSERT INTO users (email, username, password, role)
		VALUES ($1, $2, $3, $4)
	`,
		"ziad.skafi12@gmail.com",
		"ziad",
		hashed,
		"admin",
	)

	if err != nil {
		log.Fatal(err)
	}

	log.Println("Default admin account created")
}

func InitDB() {
	var err error

	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		log.Fatal("DATABASE_URL environment variable not set")
	}

	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	err = DB.Ping()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	createTables()

	ensureAdmin()

	log.Println("Database successfully initialized")
}

func createTables() {
	createTables := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		email TEXT NOT NULL UNIQUE,
		username TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL,
		role TEXT NOT NULL DEFAULT 'user'
	);

	CREATE TABLE IF NOT EXISTS password_resets (
		email TEXT NOT NULL,
		token_hash TEXT NOT NULL,
		expires_at BIGINT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS notes (
		id SERIAL PRIMARY KEY,
		user_id INT NOT NULL REFERENCES users(id),
		username TEXT NOT NULL,
		title TEXT,
		content TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS reminders (
		id SERIAL PRIMARY KEY,
		user_id INT NOT NULL REFERENCES users(id),
		username TEXT NOT NULL,
		title TEXT,
		content TEXT,
		due TIMESTAMP NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS expenses (
		id SERIAL PRIMARY KEY,
		user_id INT NOT NULL REFERENCES users(id),
		username TEXT NOT NULL,
		amount DOUBLE PRECISION NOT NULL,
		category TEXT NOT NULL,
		date TIMESTAMP NOT NULL,
		note TEXT
	);

	CREATE TABLE IF NOT EXISTS leaderboard (
		id SERIAL PRIMARY KEY,
		user_id INT NOT NULL REFERENCES users(id),
		username TEXT NOT NULL,
		section TEXT NOT NULL,
		name TEXT NOT NULL,
		points INT NOT NULL
	);
	`

	_, err := DB.Exec(createTables)
	if err != nil {
		log.Fatal("Failed to create tables:", err)
	}
}
