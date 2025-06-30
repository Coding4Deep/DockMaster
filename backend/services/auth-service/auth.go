package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret []byte

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

type LoginResponse struct {
	Token     string   `json:"token"`
	ExpiresAt int64    `json:"expires_at"`
	User      UserInfo `json:"user"`
}

type UserInfo struct {
	Username string `json:"username"`
	Role     string `json:"role"`
}

type Claims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func initAuth() {
	// Generate JWT secret if not provided
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		jwtSecret = []byte(secret)
	} else {
		jwtSecret = make([]byte, 32)
		rand.Read(jwtSecret)
		logrus.Warn("JWT_SECRET not provided, using random secret (tokens will be invalid after restart)")
	}

	// Create default admin user if no users exist
	adminUsername := getEnvOrDefault("ADMIN_USERNAME", "admin")
	adminPassword := getEnvOrDefault("ADMIN_PASSWORD", "admin123")

	// Check if admin user exists
	if _, err := Storage.GetUser(adminUsername); err != nil {
		// User doesn't exist, create it
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
		if err != nil {
			logrus.WithError(err).Fatal("Failed to hash admin password")
		}

		if err := Storage.CreateUser(adminUsername, string(hashedPassword), "admin"); err != nil {
			logrus.WithError(err).Fatal("Failed to create admin user")
		}

		Storage.LogEntry("info", "Default admin user created", "auth-service", map[string]string{
			"username": adminUsername,
			"password": adminPassword,
		})

		logrus.WithFields(logrus.Fields{
			"username": adminUsername,
			"password": adminPassword,
		}).Warn("Created default admin user - CHANGE PASSWORD IMMEDIATELY!")
	}
}

func authenticateUser(username, password string) (*User, error) {
	user, err := Storage.GetUser(username)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("invalid password")
	}

	return user, nil
}

func generateToken(user *User) (string, int64, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "dockmaster",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", 0, err
	}

	return tokenString, expirationTime.Unix(), nil
}

func validateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

// Middleware for authentication
func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			http.Error(w, "Bearer token required", http.StatusUnauthorized)
			return
		}

		claims, err := validateToken(tokenString)
		if err != nil {
			logrus.WithError(err).Warn("Invalid token")
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Add user info to request context
		r.Header.Set("X-User", claims.Username)
		r.Header.Set("X-Role", claims.Role)

		next(w, r)
	}
}

// Login handler
func loginHandler(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := authenticateUser(req.Username, req.Password)
	if err != nil {
		Storage.LogEntry("warn", "Failed login attempt", "auth-service", map[string]string{
			"username": req.Username,
			"error":    err.Error(),
		})
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token, expiresAt, err := generateToken(user)
	if err != nil {
		logrus.WithError(err).Error("Failed to generate token")
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	Storage.LogEntry("info", "User logged in successfully", "auth-service", map[string]string{
		"username": user.Username,
	})

	response := LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User: UserInfo{
			Username: user.Username,
			Role:     user.Role,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Logout handler
func logoutHandler(w http.ResponseWriter, r *http.Request) {
	username := r.Header.Get("X-User")
	Storage.LogEntry("info", "User logged out", "auth-service", map[string]string{
		"username": username,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Logged out successfully"})
}

// Get current user info
func meHandler(w http.ResponseWriter, r *http.Request) {
	username := r.Header.Get("X-User")
	role := r.Header.Get("X-Role")

	response := UserInfo{
		Username: username,
		Role:     role,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Change password handler
func changePasswordHandler(w http.ResponseWriter, r *http.Request) {
	username := r.Header.Get("X-User")

	var req ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Verify current password
	user, err := Storage.GetUser(username)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword)); err != nil {
		Storage.LogEntry("warn", "Failed password change attempt", "auth-service", map[string]string{
			"username": username,
			"reason":   "invalid current password",
		})
		http.Error(w, "Current password is incorrect", http.StatusBadRequest)
		return
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	// Update password in database
	if err := Storage.UpdateUserPassword(username, string(hashedPassword)); err != nil {
		logrus.WithError(err).Error("Failed to update password")
		http.Error(w, "Failed to update password", http.StatusInternalServerError)
		return
	}

	Storage.LogEntry("info", "Password changed successfully", "auth-service", map[string]string{
		"username": username,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Password changed successfully"})
}

// Get logs handler
func getLogsHandler(w http.ResponseWriter, r *http.Request) {
	// Only admin can view logs
	role := r.Header.Get("X-Role")
	if role != "admin" {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	service := r.URL.Query().Get("service")
	limit := 100 // Default limit

	logs, err := Storage.GetLogs(limit, service)
	if err != nil {
		logrus.WithError(err).Error("Failed to get logs")
		http.Error(w, "Failed to get logs", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
