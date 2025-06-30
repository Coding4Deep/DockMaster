package main

import (
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
)

var db *sql.DB

// initDatabase initializes the SQLite database
func initDatabase() error {
	// Create data directory if it doesn't exist
	dataDir := "./data"
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return err
	}

	// Open database
	dbPath := filepath.Join(dataDir, "dockmaster.db")
	var err error
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}

	// Test connection
	if err = db.Ping(); err != nil {
		return err
	}

	// Create tables
	if err = createTables(); err != nil {
		return err
	}

	// Load users from database
	if err = loadUsersFromDB(); err != nil {
		logrus.WithError(err).Warn("Failed to load users from database")
	}

	logrus.Info("Database initialized successfully")
	return nil
}

// createTables creates the necessary database tables
func createTables() error {
	// Users table
	usersTable := `
	CREATE TABLE IF NOT EXISTS users (
		username TEXT PRIMARY KEY,
		password_hash TEXT NOT NULL,
		role TEXT NOT NULL DEFAULT 'user',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	// Settings table for application settings
	settingsTable := `
	CREATE TABLE IF NOT EXISTS settings (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	// Execute table creation
	if _, err := db.Exec(usersTable); err != nil {
		return err
	}

	if _, err := db.Exec(settingsTable); err != nil {
		return err
	}

	return nil
}

// saveUserToDB saves a user to the database
func saveUserToDB(user User) error {
	query := `
	INSERT OR REPLACE INTO users (username, password_hash, role, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?)`

	_, err := db.Exec(query, user.Username, user.PasswordHash, user.Role, user.CreatedAt, time.Now())
	return err
}

// loadUsersFromDB loads all users from the database
func loadUsersFromDB() error {
	query := `SELECT username, password_hash, role, created_at FROM users`
	rows, err := db.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var user User
		err := rows.Scan(&user.Username, &user.PasswordHash, &user.Role, &user.CreatedAt)
		if err != nil {
			logrus.WithError(err).Error("Failed to scan user row")
			continue
		}
		users[user.Username] = user
	}

	return rows.Err()
}

// updateUserPassword updates a user's password in the database
func updateUserPassword(username, newPasswordHash string) error {
	query := `UPDATE users SET password_hash = ?, updated_at = ? WHERE username = ?`
	_, err := db.Exec(query, newPasswordHash, time.Now(), username)
	return err
}

// saveSetting saves a setting to the database
func saveSetting(key, value string) error {
	query := `INSERT OR REPLACE INTO settings (key, value, updated_at) VALUES (?, ?, ?)`
	_, err := db.Exec(query, key, value, time.Now())
	return err
}

// getSetting gets a setting from the database
func getSetting(key string) (string, error) {
	query := `SELECT value FROM settings WHERE key = ?`
	var value string
	err := db.QueryRow(query, key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return value, err
}

// saveSettingJSON saves a JSON setting to the database
func saveSettingJSON(key string, value interface{}) error {
	jsonData, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return saveSetting(key, string(jsonData))
}

// getSettingJSON gets a JSON setting from the database
func getSettingJSON(key string, dest interface{}) error {
	value, err := getSetting(key)
	if err != nil || value == "" {
		return err
	}
	return json.Unmarshal([]byte(value), dest)
}

// closeDatabase closes the database connection
func closeDatabase() {
	if db != nil {
		db.Close()
	}
}
