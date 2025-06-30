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
	logrus.SetLevel(logrus.InfoLevel)
	logrus.SetFormatter(&logrus.JSONFormatter{})

	// Initialize authentication
	initAuth()

	logrus.Info("Docker service starting...")

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
		Addr:         ":8080",
		Handler:      c.Handler(router),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logrus.Info("Starting Docker service on port 8080")
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
	// Public routes (no auth required)
	router.HandleFunc("/health", healthCheck).Methods("GET")
	router.HandleFunc("/auth/login", loginHandler).Methods("POST")

	// Protected routes (auth required)
	router.HandleFunc("/auth/logout", authMiddleware(logoutHandler)).Methods("POST")
	router.HandleFunc("/auth/me", authMiddleware(meHandler)).Methods("GET")

	// Container routes
	router.HandleFunc("/containers", authMiddleware(listContainers)).Methods("GET")
	router.HandleFunc("/containers/{id}/start", authMiddleware(startContainer)).Methods("POST")
	router.HandleFunc("/containers/{id}/stop", authMiddleware(stopContainer)).Methods("POST")
	router.HandleFunc("/containers/{id}/restart", authMiddleware(restartContainer)).Methods("POST")
	router.HandleFunc("/containers/{id}", authMiddleware(deleteContainer)).Methods("DELETE")
	router.HandleFunc("/containers/{id}/stats", authMiddleware(getContainerStats)).Methods("GET")
	router.HandleFunc("/containers/{id}/logs", authMiddleware(getContainerLogs)).Methods("GET")

	// Image routes
	router.HandleFunc("/images", authMiddleware(listImages)).Methods("GET")
	router.HandleFunc("/images/{id}", authMiddleware(deleteImage)).Methods("DELETE")

	// Volume routes
	router.HandleFunc("/volumes", authMiddleware(listVolumes)).Methods("GET")
	router.HandleFunc("/volumes/{name}", authMiddleware(deleteVolume)).Methods("DELETE")

	// Network routes
	router.HandleFunc("/networks", authMiddleware(listNetworks)).Methods("GET")
	router.HandleFunc("/networks/{id}", authMiddleware(deleteNetwork)).Methods("DELETE")

	// System info
	router.HandleFunc("/system/info", authMiddleware(getSystemInfo)).Methods("GET")
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "healthy",
		"service": "docker-service",
	})
}

// Real Docker API implementations
func listContainers(w http.ResponseWriter, r *http.Request) {
	all := r.URL.Query().Get("all") == "true"
	
	containers, err := getRealContainers(all)
	if err != nil {
		logrus.WithError(err).Error("Failed to get containers")
		http.Error(w, "Failed to get containers: "+err.Error(), http.StatusInternalServerError)
		return
	}

	logrus.WithField("count", len(containers)).Info("Listed containers")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(containers)
}

func startContainer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := dockerStart(id); err != nil {
		logrus.WithError(err).WithField("container", id).Error("Failed to start container")
		http.Error(w, "Failed to start container: "+err.Error(), http.StatusInternalServerError)
		return
	}

	logrus.WithField("container", id).Info("Container started")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Container started successfully"})
}

func stopContainer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := dockerStop(id); err != nil {
		logrus.WithError(err).WithField("container", id).Error("Failed to stop container")
		http.Error(w, "Failed to stop container: "+err.Error(), http.StatusInternalServerError)
		return
	}

	logrus.WithField("container", id).Info("Container stopped")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Container stopped successfully"})
}

func restartContainer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := dockerRestart(id); err != nil {
		logrus.WithError(err).WithField("container", id).Error("Failed to restart container")
		http.Error(w, "Failed to restart container: "+err.Error(), http.StatusInternalServerError)
		return
	}

	logrus.WithField("container", id).Info("Container restarted")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Container restarted successfully"})
}

func deleteContainer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	force := r.URL.Query().Get("force") == "true"

	if err := dockerRemove(id, force); err != nil {
		logrus.WithError(err).WithField("container", id).Error("Failed to delete container")
		http.Error(w, "Failed to delete container: "+err.Error(), http.StatusInternalServerError)
		return
	}

	logrus.WithField("container", id).Info("Container deleted")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Container deleted successfully"})
}

func getContainerStats(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	stats, err := getRealContainerStats(id)
	if err != nil {
		logrus.WithError(err).WithField("container", id).Error("Failed to get container stats")
		http.Error(w, "Failed to get container stats: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func getContainerLogs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	tail := r.URL.Query().Get("tail")
	if tail == "" {
		tail = "100"
	}

	logs, err := dockerLogs(id, tail)
	if err != nil {
		logrus.WithError(err).WithField("container", id).Error("Failed to get container logs")
		http.Error(w, "Failed to get container logs: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}

func listImages(w http.ResponseWriter, r *http.Request) {
	images, err := getRealImages()
	if err != nil {
		logrus.WithError(err).Error("Failed to get images")
		http.Error(w, "Failed to get images: "+err.Error(), http.StatusInternalServerError)
		return
	}

	logrus.WithField("count", len(images)).Info("Listed images")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(images)
}

func deleteImage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	force := r.URL.Query().Get("force") == "true"

	if err := dockerRemoveImage(id, force); err != nil {
		logrus.WithError(err).WithField("image", id).Error("Failed to delete image")
		http.Error(w, "Failed to delete image: "+err.Error(), http.StatusInternalServerError)
		return
	}

	logrus.WithField("image", id).Info("Image deleted")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Image deleted successfully"})
}

func listVolumes(w http.ResponseWriter, r *http.Request) {
	volumes, err := getRealVolumes()
	if err != nil {
		logrus.WithError(err).Error("Failed to get volumes")
		http.Error(w, "Failed to get volumes: "+err.Error(), http.StatusInternalServerError)
		return
	}

	logrus.WithField("count", len(volumes)).Info("Listed volumes")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(volumes)
}

func deleteVolume(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	force := r.URL.Query().Get("force") == "true"

	if err := dockerRemoveVolume(name, force); err != nil {
		logrus.WithError(err).WithField("volume", name).Error("Failed to delete volume")
		http.Error(w, "Failed to delete volume: "+err.Error(), http.StatusInternalServerError)
		return
	}

	logrus.WithField("volume", name).Info("Volume deleted")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Volume deleted successfully"})
}

func listNetworks(w http.ResponseWriter, r *http.Request) {
	networks, err := getRealNetworks()
	if err != nil {
		logrus.WithError(err).Error("Failed to get networks")
		http.Error(w, "Failed to get networks: "+err.Error(), http.StatusInternalServerError)
		return
	}

	logrus.WithField("count", len(networks)).Info("Listed networks")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(networks)
}

func deleteNetwork(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := dockerRemoveNetwork(id); err != nil {
		logrus.WithError(err).WithField("network", id).Error("Failed to delete network")
		http.Error(w, "Failed to delete network: "+err.Error(), http.StatusInternalServerError)
		return
	}

	logrus.WithField("network", id).Info("Network deleted")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Network deleted successfully"})
}

func getSystemInfo(w http.ResponseWriter, r *http.Request) {
	info, err := getRealSystemInfo()
	if err != nil {
		logrus.WithError(err).Error("Failed to get system info")
		http.Error(w, "Failed to get system info: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}
