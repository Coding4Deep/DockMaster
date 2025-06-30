package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/signal"
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

	logrus.Info("Container service starting...")

	// Setup router
	router := mux.NewRouter()
	setupRoutes(router)

	// Setup CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://127.0.0.1:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	// Create server
	srv := &http.Server{
		Addr:         ":8082",
		Handler:      c.Handler(router),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logrus.Info("Starting Container service on port 8082")
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

	// Container routes
	router.HandleFunc("/containers", authMiddleware(listContainers)).Methods("GET")
	router.HandleFunc("/containers", authMiddleware(createContainer)).Methods("POST")
	router.HandleFunc("/containers/{id}/start", authMiddleware(startContainer)).Methods("POST")
	router.HandleFunc("/containers/{id}/stop", authMiddleware(stopContainer)).Methods("POST")
	router.HandleFunc("/containers/{id}/restart", authMiddleware(restartContainer)).Methods("POST")
	router.HandleFunc("/containers/{id}", authMiddleware(deleteContainer)).Methods("DELETE")
	router.HandleFunc("/containers/{id}/stats", authMiddleware(getContainerStats)).Methods("GET")
	router.HandleFunc("/containers/{id}/logs", authMiddleware(getContainerLogs)).Methods("GET")
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "healthy",
		"service": "container-service",
	})
}
