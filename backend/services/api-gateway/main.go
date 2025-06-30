package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
)

func main() {
	// Setup logging
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&logrus.JSONFormatter{})

	logrus.Info("API Gateway starting...")

	// Setup router
	router := mux.NewRouter()
	setupRoutes(router)

	// Setup CORS - Allow any origin on port 3000 for development
	c := cors.New(cors.Options{
		AllowOriginFunc: func(origin string) bool {
			// Allow localhost, 127.0.0.1, and any IP on port 3000
			return strings.HasSuffix(origin, ":3000") || 
				   origin == "http://localhost:3000" || 
				   origin == "http://127.0.0.1:3000"
		},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	// Create server
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      c.Handler(router),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logrus.Info("Starting API Gateway on port 8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.WithError(err).Fatal("Server failed to start")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logrus.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logrus.WithError(err).Fatal("Server forced to shutdown")
	}

	logrus.Info("Server exited")
}

func setupRoutes(router *mux.Router) {
	// Health check
	router.HandleFunc("/health", healthCheck).Methods("GET")

	// Auth routes (proxy to auth service) - no auth required for login
	router.HandleFunc("/auth/login", proxyToService("auth-service:8081")).Methods("POST")
	router.HandleFunc("/auth/logout", authMiddleware(proxyToService("auth-service:8081"))).Methods("POST")
	router.HandleFunc("/auth/change-password", authMiddleware(proxyToService("auth-service:8081"))).Methods("POST")
	router.HandleFunc("/auth/me", authMiddleware(proxyToService("auth-service:8081"))).Methods("GET")

	// Container routes (proxy to container service) - auth required
	router.PathPrefix("/containers").HandlerFunc(authMiddleware(proxyToService("container-service:8082")))

	// Image routes (proxy to image service) - auth required
	router.PathPrefix("/images").HandlerFunc(authMiddleware(proxyToService("image-service:8083")))

	// Volume routes (proxy to volume service) - auth required
	router.PathPrefix("/volumes").HandlerFunc(authMiddleware(proxyToService("volume-service:8084")))

	// Network routes (proxy to network service) - auth required
	router.PathPrefix("/networks").HandlerFunc(authMiddleware(proxyToService("network-service:8085")))

	// System info route - auth required
	router.HandleFunc("/system/info", authMiddleware(getSystemInfo)).Methods("GET")

	// Logs route (proxy to auth service) - auth required
	router.HandleFunc("/logs", authMiddleware(proxyToService("auth-service:8081"))).Methods("GET")
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "healthy",
		"service": "api-gateway",
	})
}
