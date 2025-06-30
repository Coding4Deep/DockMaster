package main

import (
	"net/http"

	"github.com/sirupsen/logrus"
)

// Simple middleware that checks for user headers set by API gateway
func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check if user info is provided by API gateway
		username := r.Header.Get("X-User")
		if username == "" {
			logrus.Warn("No user information provided by API gateway")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		logrus.WithField("username", username).Debug("Request authenticated by API gateway")
		next(w, r)
	}
}
