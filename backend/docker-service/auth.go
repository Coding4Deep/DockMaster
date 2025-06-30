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

var (
	jwtSecret []byte
	users     = make(map[string]User)
)

type User struct {
	Username     string `json:"username"`
	PasswordHash string `json:"password_hash"`
	Role         string `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
	User      UserInfo `json:"user"`
}

type UserInfo struct {
	Username string `json:"username"`
	Role     string `json:"role"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
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
	
	if len(users) == 0 {
		createUser(adminUsername, adminPassword, "admin")
		logrus.WithFields(logrus.Fields{
			"username": adminUsername,
			"password": adminPassword,
		}).Warn("Created default admin user - CHANGE PASSWORD IMMEDIATELY!")
	}
}

func createUser(username, password, role string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := User{
		Username:     username,
		PasswordHash: string(hashedPassword),
		Role:         role,
		CreatedAt:    time.Now(),
	}

	users[username] = user

	// Save to database if available
	if db != nil {
		if err := saveUserToDB(user); err != nil {
			logrus.WithError(err).Warn("Failed to save user to database")
		}
	}

	return nil
}

func authenticateUser(username, password string) (*User, error) {
	user, exists := users[username]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("invalid password")
	}

	return &user, nil
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
		// Skip auth for login and health endpoints
		if r.URL.Path == "/auth/login" || r.URL.Path == "/health" {
			next(w, r)
			return
		}

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
		logrus.WithFields(logrus.Fields{
			"username": req.Username,
			"error":    err.Error(),
		}).Warn("Failed login attempt")
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token, expiresAt, err := generateToken(user)
	if err != nil {
		logrus.WithError(err).Error("Failed to generate token")
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	logrus.WithField("username", user.Username).Info("User logged in successfully")

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

// Logout handler (client-side token removal)
func logoutHandler(w http.ResponseWriter, r *http.Request) {
	username := r.Header.Get("X-User")
	logrus.WithField("username", username).Info("User logged out")
	
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

	// Validate current password
	user, exists := users[username]
	if !exists {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword))
	if err != nil {
		logrus.WithField("username", username).Warn("Invalid current password for password change")
		http.Error(w, "Invalid current password", http.StatusUnauthorized)
		return
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		logrus.WithError(err).Error("Failed to hash new password")
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	// Update password
	user.PasswordHash = string(hashedPassword)
	users[username] = user

	// Update in database if available
	if db != nil {
		if err := updateUserPassword(username, string(hashedPassword)); err != nil {
			logrus.WithError(err).Error("Failed to update password in database")
			http.Error(w, "Failed to update password in database", http.StatusInternalServerError)
			return
		}
	}

	logrus.WithField("username", username).Info("Password changed successfully")
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Password changed successfully"})
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
