package main

import (
	"bufio"
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

type RunContainerRequest struct {
	Image        string            `json:"image"`
	Name         string            `json:"name,omitempty"`
	Ports        map[string]string `json:"ports,omitempty"`
	Environment  []string          `json:"environment,omitempty"`
	Volumes      []string          `json:"volumes,omitempty"`
	Command      []string          `json:"command,omitempty"`
	WorkingDir   string            `json:"working_dir,omitempty"`
	RestartPolicy string           `json:"restart_policy,omitempty"`
}

type PullImageRequest struct {
	Image string `json:"image"`
}

type SearchResponse struct {
	Local     []LocalImageResult `json:"local"`
	DockerHub []HubImageResult   `json:"dockerhub"`
}

type LocalImageResult struct {
	ID       string   `json:"id"`
	RepoTags []string `json:"repo_tags"`
	Size     int64    `json:"size"`
	Created  int64    `json:"created"`
}

type HubImageResult struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Stars       int    `json:"star_count"`
	Official    bool   `json:"is_official"`
	Automated   bool   `json:"is_automated"`
}

func main() {
	// Load environment variables from .env file
	loadEnvFile()

	// Setup logging
	logLevel := getEnvOrDefault("LOG_LEVEL", "info")
	if level, err := logrus.ParseLevel(logLevel); err == nil {
		logrus.SetLevel(level)
	}
	logrus.SetFormatter(&logrus.JSONFormatter{})

	// Initialize database
	if err := initDatabase(); err != nil {
		logrus.WithError(err).Warn("Failed to initialize database, continuing without persistence")
	}

	// Initialize authentication
	initAuth()

	logrus.Info("Docker service starting...")

	// Setup router
	router := mux.NewRouter()
	setupRoutes(router)

	// Get port from environment variable or use default
	port := getEnvOrDefault("PORT", "8081")
	
	// Get frontend URL from environment
	frontendURL := getEnvOrDefault("FRONTEND_URL", "http://localhost:3000")
	
	// Setup CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{frontendURL, "http://127.0.0.1:3000", "http://localhost:" + port, "http://127.0.0.1:" + port},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	// Create server
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      c.Handler(router),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logrus.WithField("port", port).Info("Starting Docker service")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.WithError(err).Fatal("Server failed to start")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logrus.Info("Shutting down server...")

	// Close database connection
	closeDatabase()

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logrus.WithError(err).Fatal("Server forced to shutdown")
	}

	logrus.Info("Server exited")
}

// loadEnvFile loads environment variables from .env file
func loadEnvFile() {
	file, err := os.Open(".env")
	if err != nil {
		logrus.Debug("No .env file found, using system environment variables")
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			
			// Only set if not already set in environment
			if os.Getenv(key) == "" {
				os.Setenv(key, value)
			}
		}
	}
	
	if err := scanner.Err(); err != nil {
		logrus.WithError(err).Warn("Error reading .env file")
	}
}

func setupRoutes(router *mux.Router) {
	// Public routes (no auth required)
	router.HandleFunc("/health", healthCheck).Methods("GET")
	router.HandleFunc("/auth/login", loginHandler).Methods("POST")

	// Protected routes (auth required)
	router.HandleFunc("/auth/logout", authMiddleware(logoutHandler)).Methods("POST")
	router.HandleFunc("/auth/me", authMiddleware(meHandler)).Methods("GET")
	router.HandleFunc("/auth/change-password", authMiddleware(changePasswordHandler)).Methods("POST")

	// Container routes
	router.HandleFunc("/containers", authMiddleware(listContainers)).Methods("GET")
	router.HandleFunc("/containers/run", authMiddleware(runContainer)).Methods("POST")
	router.HandleFunc("/containers/{id}/start", authMiddleware(startContainer)).Methods("POST")
	router.HandleFunc("/containers/{id}/stop", authMiddleware(stopContainer)).Methods("POST")
	router.HandleFunc("/containers/{id}/restart", authMiddleware(restartContainer)).Methods("POST")
	router.HandleFunc("/containers/{id}", authMiddleware(deleteContainer)).Methods("DELETE")
	router.HandleFunc("/containers/{id}/stats", authMiddleware(getContainerStats)).Methods("GET")
	router.HandleFunc("/containers/{id}/logs", authMiddleware(getContainerLogs)).Methods("GET")

	// Image routes
	router.HandleFunc("/images", authMiddleware(listImages)).Methods("GET")
	router.HandleFunc("/images/search", authMiddleware(searchImages)).Methods("GET")
	router.HandleFunc("/images/pull", authMiddleware(pullImage)).Methods("POST")
	router.HandleFunc("/images/{id}", authMiddleware(deleteImage)).Methods("DELETE")
	router.HandleFunc("/images/{id}/inspect", authMiddleware(inspectImage)).Methods("GET")

	// Volume routes
	router.HandleFunc("/volumes", authMiddleware(listVolumes)).Methods("GET")
	router.HandleFunc("/volumes/{name}", authMiddleware(deleteVolume)).Methods("DELETE")

	// Network routes
	router.HandleFunc("/networks", authMiddleware(listNetworks)).Methods("GET")
	router.HandleFunc("/networks/{id}", authMiddleware(deleteNetwork)).Methods("DELETE")

	// System info and metrics
	router.HandleFunc("/system/info", authMiddleware(getSystemInfo)).Methods("GET")
	router.HandleFunc("/system/metrics", authMiddleware(getSystemMetrics)).Methods("GET")
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

func runContainer(w http.ResponseWriter, r *http.Request) {
	var req RunContainerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	containerID, err := dockerRun(req)
	if err != nil {
		logrus.WithError(err).WithField("image", req.Image).Error("Failed to run container")
		http.Error(w, "Failed to run container: "+err.Error(), http.StatusInternalServerError)
		return
	}

	logrus.WithFields(logrus.Fields{
		"container": containerID,
		"image":     req.Image,
	}).Info("Container created and started")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message":     "Container created and started successfully",
		"containerID": containerID,
	})
}

func searchImages(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Query parameter 'q' is required", http.StatusBadRequest)
		return
	}

	// First search local images
	localImages, err := searchLocalImages(query)
	if err != nil {
		logrus.WithError(err).Error("Failed to search local images")
	}

	// Then search Docker Hub
	hubImages, err := searchDockerHub(query)
	if err != nil {
		logrus.WithError(err).Error("Failed to search Docker Hub")
	}

	response := SearchResponse{
		Local:     localImages,
		DockerHub: hubImages,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func pullImage(w http.ResponseWriter, r *http.Request) {
	var req PullImageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := dockerPull(req.Image); err != nil {
		logrus.WithError(err).WithField("image", req.Image).Error("Failed to pull image")
		http.Error(w, "Failed to pull image: "+err.Error(), http.StatusInternalServerError)
		return
	}

	logrus.WithField("image", req.Image).Info("Image pulled successfully")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Image pulled successfully"})
}

func inspectImage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	inspection, err := dockerInspectImage(id)
	if err != nil {
		logrus.WithError(err).WithField("image", id).Error("Failed to inspect image")
		http.Error(w, "Failed to inspect image: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(inspection)
}

func getSystemMetrics(w http.ResponseWriter, r *http.Request) {
	metrics, err := getRealSystemMetrics()
	if err != nil {
		logrus.WithError(err).Error("Failed to get system metrics")
		http.Error(w, "Failed to get system metrics: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}
