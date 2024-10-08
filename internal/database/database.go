package database

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

type Database struct {
	db *sql.DB
}

func NewDatabase() (*Database, error) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)

	db, err := sql.Open(os.Getenv("DB_NAME"), connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &Database{db: db}, nil
}

func (d *Database) AllowUser(userID int64) error {
	_, err := d.db.Exec("INSERT INTO allowed_users (user_id) VALUES ($1) ON CONFLICT (user_id) DO NOTHING", userID)
	return err
}

func (d *Database) IsUserAllowed(userID int64) bool {
	var exists bool
	err := d.db.QueryRow("SELECT EXISTS(SELECT 1 FROM allowed_users WHERE user_id = $1)", userID).Scan(&exists)
	if err != nil {
		return false
	}
	return exists
}

func (d *Database) LogUserActivity(userID int64, command string) error {
	_, err := d.db.Exec("INSERT INTO user_activity (user_id, command) VALUES ($1, $2)", userID, command)
	return err
}
