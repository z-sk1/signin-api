package db

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func InitDB() {
	var err error
	DB, err = sql.Open("sqlite3", "./data.db")
	if err != nil {
		log.Fatal(err)
	}

	createTables := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL UNIQUE,
	);

	CREATE TABLE IF NOT EXISTS notes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		title TEXT,
		content TEXT,
		FOREIGN KEY(user_id) REFERENCES users(id),	
	);

	CREATE TABLE IF NOT EXISTS reminders (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		title TEXT,
		content TEXT,
		due_date DATETIME,
		FOREIGN KEY(user_id) REFERENCES users(id),
	);
	`

	_, err = DB.Exec(createTables)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("database succesfully initalised")
}
