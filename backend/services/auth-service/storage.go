package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
)

type FileStorage struct {
	dataDir string
}

type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"password_hash"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type LogEntry struct {
	ID        int         `json:"id"`
	Level     string      `json:"level"`
	Message   string      `json:"message"`
	Service   string      `json:"service"`
	Data      interface{} `json:"data"`
	CreatedAt time.Time   `json:"created_at"`
}

var Storage *FileStorage

func InitStorage() error {
	// Create data directory if it doesn't exist
	dataDir := "./data"
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %v", err)
	}

	Storage = &FileStorage{dataDir: dataDir}
	logrus.Info("File storage initialized successfully")
	return nil
}

func (fs *FileStorage) CreateUser(username, passwordHash, role string) error {
	users, err := fs.loadUsers()
	if err != nil {
		return err
	}

	// Check if user already exists
	for _, user := range users {
		if user.Username == username {
			return fmt.Errorf("user already exists")
		}
	}

	newUser := User{
		ID:           len(users) + 1,
		Username:     username,
		PasswordHash: passwordHash,
		Role:         role,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	users = append(users, newUser)
	return fs.saveUsers(users)
}

func (fs *FileStorage) GetUser(username string) (*User, error) {
	users, err := fs.loadUsers()
	if err != nil {
		return nil, err
	}

	for _, user := range users {
		if user.Username == username {
			return &user, nil
		}
	}

	return nil, fmt.Errorf("user not found")
}

func (fs *FileStorage) UpdateUserPassword(username, newPasswordHash string) error {
	users, err := fs.loadUsers()
	if err != nil {
		return err
	}

	for i, user := range users {
		if user.Username == username {
			users[i].PasswordHash = newPasswordHash
			users[i].UpdatedAt = time.Now()
			return fs.saveUsers(users)
		}
	}

	return fmt.Errorf("user not found")
}

func (fs *FileStorage) LogEntry(level, message, service string, data interface{}) error {
	logs, err := fs.loadLogs()
	if err != nil {
		// If logs file doesn't exist, start with empty slice
		logs = []LogEntry{}
	}

	newLog := LogEntry{
		ID:        len(logs) + 1,
		Level:     level,
		Message:   message,
		Service:   service,
		Data:      data,
		CreatedAt: time.Now(),
	}

	logs = append(logs, newLog)
	
	// Keep only last 1000 logs to prevent file from growing too large
	if len(logs) > 1000 {
		logs = logs[len(logs)-1000:]
	}

	return fs.saveLogs(logs)
}

func (fs *FileStorage) GetLogs(limit int, service string) ([]LogEntry, error) {
	logs, err := fs.loadLogs()
	if err != nil {
		return []LogEntry{}, nil
	}

	// Filter by service if specified
	var filteredLogs []LogEntry
	if service != "" {
		for _, log := range logs {
			if log.Service == service {
				filteredLogs = append(filteredLogs, log)
			}
		}
	} else {
		filteredLogs = logs
	}

	// Return last 'limit' entries
	if len(filteredLogs) > limit {
		return filteredLogs[len(filteredLogs)-limit:], nil
	}

	return filteredLogs, nil
}

func (fs *FileStorage) loadUsers() ([]User, error) {
	usersFile := filepath.Join(fs.dataDir, "users.json")
	
	if _, err := os.Stat(usersFile); os.IsNotExist(err) {
		return []User{}, nil
	}

	data, err := os.ReadFile(usersFile)
	if err != nil {
		return nil, err
	}

	var users []User
	if err := json.Unmarshal(data, &users); err != nil {
		return nil, err
	}

	return users, nil
}

func (fs *FileStorage) saveUsers(users []User) error {
	usersFile := filepath.Join(fs.dataDir, "users.json")
	
	data, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(usersFile, data, 0644)
}

func (fs *FileStorage) loadLogs() ([]LogEntry, error) {
	logsFile := filepath.Join(fs.dataDir, "logs.json")
	
	if _, err := os.Stat(logsFile); os.IsNotExist(err) {
		return []LogEntry{}, nil
	}

	data, err := os.ReadFile(logsFile)
	if err != nil {
		return nil, err
	}

	var logs []LogEntry
	if err := json.Unmarshal(data, &logs); err != nil {
		return nil, err
	}

	return logs, nil
}

func (fs *FileStorage) saveLogs(logs []LogEntry) error {
	logsFile := filepath.Join(fs.dataDir, "logs.json")
	
	data, err := json.MarshalIndent(logs, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(logsFile, data, 0644)
}

func (fs *FileStorage) Close() error {
	// Nothing to close for file storage
	return nil
}
