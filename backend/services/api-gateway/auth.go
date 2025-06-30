package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
)

type Claims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func getJWTSecret() string {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "dockmaster-super-secret-key-2024" // Default secret - should match auth service
	}
	return secret
}

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Skip auth for login and health endpoints
		if strings.HasPrefix(r.URL.Path, "/auth/login") || 
		   strings.HasPrefix(r.URL.Path, "/health") ||
		   strings.HasPrefix(r.URL.Path, "/system/info") {
			next(w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			logrus.Warn("No authorization header provided")
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			logrus.Warn("Bearer token not found in authorization header")
			http.Error(w, "Bearer token required", http.StatusUnauthorized)
			return
		}

		logrus.WithField("token_length", len(tokenString)).Debug("Parsing token")
		
		token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			secret := getJWTSecret()
			logrus.WithField("secret_length", len(secret)).Debug("Using JWT secret")
			return []byte(secret), nil
		})

		if err != nil {
			logrus.WithError(err).Warn("Invalid token")
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		if claims, ok := token.Claims.(*Claims); ok && token.Valid {
			logrus.WithField("username", claims.Username).Debug("Token validated successfully")
			// Add user info to request headers for downstream services
			r.Header.Set("X-User", claims.Username)
			r.Header.Set("X-Role", claims.Role)
			next(w, r)
		} else {
			logrus.Warn("Token claims invalid")
			http.Error(w, "Invalid token", http.StatusUnauthorized)
		}
	}
}
