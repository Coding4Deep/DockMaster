package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// DockerContainer represents a Docker container
type DockerContainer struct {
	ID      string `json:"ID"`
	Names   string `json:"Names"`
	Image   string `json:"Image"`
	Command string `json:"Command"`
	Created string `json:"CreatedAt"`
	Ports   string `json:"Ports"`
	Labels  string `json:"Labels"`
	State   string `json:"State"`
	Status  string `json:"Status"`
	Mounts  string `json:"Mounts"`
	Size    string `json:"Size"`
}

type CreateContainerRequest struct {
	Name         string            `json:"name"`
	Image        string            `json:"image"`
	Ports        []PortMapping     `json:"ports"`
	Environment  map[string]string `json:"environment"`
	Volumes      []VolumeMapping   `json:"volumes"`
	Command      []string          `json:"command"`
	WorkingDir   string            `json:"working_dir"`
	RestartPolicy string           `json:"restart_policy"`
}

type PortMapping struct {
	HostPort      string `json:"host_port"`
	ContainerPort string `json:"container_port"`
	Protocol      string `json:"protocol"`
}

type VolumeMapping struct {
	HostPath      string `json:"host_path"`
	ContainerPath string `json:"container_path"`
	ReadOnly      bool   `json:"read_only"`
}

// convertToFrontendFormat converts raw Docker container data to frontend format
func convertToFrontendFormat(raw DockerContainer) map[string]interface{} {
	// Parse names
	names := []string{}
	if raw.Names != "" {
		names = []string{raw.Names}
	}

	// Parse ports
	ports := []map[string]interface{}{}
	if raw.Ports != "" {
		portParts := strings.Split(raw.Ports, ",")
		for _, portStr := range portParts {
			portStr = strings.TrimSpace(portStr)
			if strings.Contains(portStr, "->") {
				parts := strings.Split(portStr, "->")
				if len(parts) == 2 {
					publicPart := strings.TrimSpace(parts[0])
					privatePart := strings.TrimSpace(parts[1])

					port := map[string]interface{}{
						"PrivatePort": privatePart,
						"PublicPort":  publicPart,
						"Type":        "tcp",
					}

					if strings.Contains(publicPart, ":") {
						ipPort := strings.Split(publicPart, ":")
						if len(ipPort) == 2 {
							port["IP"] = ipPort[0]
							if publicPortNum, err := strconv.Atoi(ipPort[1]); err == nil {
								port["PublicPort"] = publicPortNum
							}
						}
					}

					ports = append(ports, port)
				}
			}
		}
	}

	// Parse created time
	created, _ := time.Parse("2006-01-02 15:04:05 -0700 MST", raw.Created)

	return map[string]interface{}{
		"Id":      raw.ID,
		"Names":   names,
		"Image":   raw.Image,
		"Command": raw.Command,
		"Created": created.Unix(),
		"Ports":   ports,
		"Labels":  map[string]string{},
		"State":   raw.State,
		"Status":  raw.Status,
		"Mounts":  []interface{}{},
		"SizeRw":  0,
		"SizeRootFs": 0,
	}
}

func getRealContainers(all bool) ([]map[string]interface{}, error) {
	args := []string{"ps", "--format", "json"}
	if all {
		args = append(args, "-a")
	}

	cmd := exec.Command("docker", args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute docker ps: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var containers []map[string]interface{}
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var containerJSON map[string]interface{}
		if err := json.Unmarshal([]byte(line), &containerJSON); err != nil {
			logrus.WithError(err).WithField("line", line).Warn("Failed to parse container JSON")
			continue
		}

		// Convert to our expected format
		container := map[string]interface{}{
			"id":      containerJSON["ID"],
			"name":    strings.TrimPrefix(fmt.Sprintf("%v", containerJSON["Names"]), "/"),
			"image":   containerJSON["Image"],
			"command": containerJSON["Command"],
			"created": containerJSON["CreatedAt"],
			"status":  containerJSON["Status"],
			"state":   containerJSON["State"],
			"ports":   containerJSON["Ports"],
		}

		containers = append(containers, container)
	}

	return containers, nil
}

func listContainers(w http.ResponseWriter, r *http.Request) {
	// Show all containers by default (both running and stopped)
	// Only show running containers if explicitly requested with all=false
	all := r.URL.Query().Get("all") != "false"

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

func createContainer(w http.ResponseWriter, r *http.Request) {
	var req CreateContainerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Build docker run command
	args := []string{"run", "-d"}

	// Add name if provided
	if req.Name != "" {
		args = append(args, "--name", req.Name)
	}

	// Add port mappings
	for _, port := range req.Ports {
		portMapping := fmt.Sprintf("%s:%s", port.HostPort, port.ContainerPort)
		if port.Protocol != "" && port.Protocol != "tcp" {
			portMapping += "/" + port.Protocol
		}
		args = append(args, "-p", portMapping)
	}

	// Add environment variables
	for key, value := range req.Environment {
		args = append(args, "-e", fmt.Sprintf("%s=%s", key, value))
	}

	// Add volume mappings
	for _, volume := range req.Volumes {
		volumeMapping := fmt.Sprintf("%s:%s", volume.HostPath, volume.ContainerPath)
		if volume.ReadOnly {
			volumeMapping += ":ro"
		}
		args = append(args, "-v", volumeMapping)
	}

	// Add working directory
	if req.WorkingDir != "" {
		args = append(args, "-w", req.WorkingDir)
	}

	// Add restart policy
	if req.RestartPolicy != "" {
		args = append(args, "--restart", req.RestartPolicy)
	}

	// Add image
	args = append(args, req.Image)

	// Add command if provided
	if len(req.Command) > 0 {
		args = append(args, req.Command...)
	}

	cmd := exec.Command("docker", args...)
	output, err := cmd.Output()
	if err != nil {
		logrus.WithError(err).WithField("args", args).Error("Failed to create container")
		http.Error(w, "Failed to create container: "+err.Error(), http.StatusInternalServerError)
		return
	}

	containerID := strings.TrimSpace(string(output))
	logrus.WithFields(logrus.Fields{
		"container_id": containerID,
		"image":        req.Image,
		"name":         req.Name,
	}).Info("Container created successfully")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message":      "Container created successfully",
		"container_id": containerID,
	})
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

// Docker command helpers
func dockerStart(containerID string) error {
	cmd := exec.Command("docker", "start", containerID)
	return cmd.Run()
}

func dockerStop(containerID string) error {
	cmd := exec.Command("docker", "stop", containerID)
	return cmd.Run()
}

func dockerRestart(containerID string) error {
	cmd := exec.Command("docker", "restart", containerID)
	return cmd.Run()
}

func dockerRemove(containerID string, force bool) error {
	args := []string{"rm"}
	if force {
		args = append(args, "-f")
	}
	args = append(args, containerID)

	cmd := exec.Command("docker", args...)
	return cmd.Run()
}

func getRealContainerStats(containerID string) (map[string]interface{}, error) {
	cmd := exec.Command("docker", "stats", "--no-stream", "--format", "table {{.Container}}\\t{{.CPUPerc}}\\t{{.MemUsage}}\\t{{.NetIO}}\\t{{.BlockIO}}", containerID)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) < 2 {
		return map[string]interface{}{}, nil
	}

	line := strings.TrimSpace(lines[1])
	parts := strings.Split(line, "\t")
	if len(parts) >= 5 {
		return map[string]interface{}{
			"container": parts[0],
			"cpu":       parts[1],
			"memory":    parts[2],
			"network":   parts[3],
			"block_io":  parts[4],
		}, nil
	}

	return map[string]interface{}{}, nil
}

func dockerLogs(containerID, tail string) (map[string]interface{}, error) {
	cmd := exec.Command("docker", "logs", "--tail", tail, containerID)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	logs := strings.Split(string(output), "\n")
	return map[string]interface{}{
		"logs": logs,
	}, nil
}
