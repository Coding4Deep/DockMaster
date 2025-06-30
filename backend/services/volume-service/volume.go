package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// DockerVolume represents a Docker volume
type DockerVolume struct {
	Driver     string `json:"Driver"`
	Labels     string `json:"Labels"`
	Mountpoint string `json:"Mountpoint"`
	Name       string `json:"Name"`
	Options    string `json:"Options"`
	Scope      string `json:"Scope"`
	Size       string `json:"Size"`
}

type CreateVolumeRequest struct {
	Name       string            `json:"name"`
	Driver     string            `json:"driver"`
	DriverOpts map[string]string `json:"driver_opts"`
	Labels     map[string]string `json:"labels"`
}

type VolumeInspectResult struct {
	Name       string            `json:"Name"`
	Driver     string            `json:"Driver"`
	Mountpoint string            `json:"Mountpoint"`
	Labels     map[string]string `json:"Labels"`
	Options    map[string]string `json:"Options"`
	Scope      string            `json:"Scope"`
	CreatedAt  string            `json:"CreatedAt"`
}

// convertToFrontendFormat converts raw Docker volume data to frontend format
func convertToFrontendFormat(raw DockerVolume) map[string]interface{} {
	// Parse labels
	labels := make(map[string]string)
	if raw.Labels != "" && raw.Labels != "{}" {
		// Simple label parsing - in production you'd want more robust parsing
		labelParts := strings.Split(raw.Labels, ",")
		for _, label := range labelParts {
			if strings.Contains(label, "=") {
				parts := strings.SplitN(label, "=", 2)
				if len(parts) == 2 {
					labels[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
				}
			}
		}
	}

	return map[string]interface{}{
		"Name":       raw.Name,
		"Driver":     raw.Driver,
		"Mountpoint": raw.Mountpoint,
		"Labels":     labels,
		"Options":    map[string]string{},
		"Scope":      raw.Scope,
		"CreatedAt":  time.Now().Format(time.RFC3339),
	}
}

func getRealVolumes() ([]map[string]interface{}, error) {
	cmd := exec.Command("docker", "volume", "ls", "--format", "json")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute docker volume ls: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var volumes []map[string]interface{}
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var volumeJSON map[string]interface{}
		if err := json.Unmarshal([]byte(line), &volumeJSON); err != nil {
			logrus.WithError(err).WithField("line", line).Warn("Failed to parse volume JSON")
			continue
		}

		// Convert to our expected format
		volume := map[string]interface{}{
			"name":       volumeJSON["Name"],
			"driver":     volumeJSON["Driver"],
			"mountpoint": volumeJSON["Mountpoint"],
			"scope":      volumeJSON["Scope"],
			"labels":     volumeJSON["Labels"],
		}

		volumes = append(volumes, volume)
	}

	return volumes, nil
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
	json.NewEncoder(w).Encode(map[string]interface{}{
		"Volumes": volumes,
	})
}

func createVolume(w http.ResponseWriter, r *http.Request) {
	var req CreateVolumeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Build docker volume create command
	args := []string{"volume", "create"}

	// Add driver if specified
	if req.Driver != "" {
		args = append(args, "--driver", req.Driver)
	}

	// Add driver options
	for key, value := range req.DriverOpts {
		args = append(args, "--opt", fmt.Sprintf("%s=%s", key, value))
	}

	// Add labels
	for key, value := range req.Labels {
		args = append(args, "--label", fmt.Sprintf("%s=%s", key, value))
	}

	// Add volume name
	if req.Name != "" {
		args = append(args, req.Name)
	}

	cmd := exec.Command("docker", args...)
	output, err := cmd.Output()
	if err != nil {
		logrus.WithError(err).WithField("args", args).Error("Failed to create volume")
		http.Error(w, "Failed to create volume: "+err.Error(), http.StatusInternalServerError)
		return
	}

	volumeName := strings.TrimSpace(string(output))
	logrus.WithFields(logrus.Fields{
		"volume_name": volumeName,
		"driver":      req.Driver,
	}).Info("Volume created successfully")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Volume created successfully",
		"name":    volumeName,
	})
}

func deleteVolume(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	force := r.URL.Query().Get("force") == "true"

	// Check if volume is in use by any container
	if !force {
		inUse, err := isVolumeInUse(name)
		if err != nil {
			logrus.WithError(err).WithField("volume", name).Error("Failed to check if volume is in use")
			http.Error(w, "Failed to check volume usage: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if inUse {
			http.Error(w, "Volume is currently in use by one or more containers. Use force=true to delete anyway.", http.StatusConflict)
			return
		}
	}

	if err := dockerRemoveVolume(name, force); err != nil {
		logrus.WithError(err).WithField("volume", name).Error("Failed to delete volume")
		http.Error(w, "Failed to delete volume: "+err.Error(), http.StatusInternalServerError)
		return
	}

	logrus.WithField("volume", name).Info("Volume deleted")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Volume deleted successfully"})
}

func inspectVolume(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	cmd := exec.Command("docker", "volume", "inspect", name)
	output, err := cmd.Output()
	if err != nil {
		logrus.WithError(err).WithField("volume", name).Error("Failed to inspect volume")
		http.Error(w, "Failed to inspect volume: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var inspectResult []VolumeInspectResult
	if err := json.Unmarshal(output, &inspectResult); err != nil {
		logrus.WithError(err).Error("Failed to parse volume inspect output")
		http.Error(w, "Failed to parse volume inspect output", http.StatusInternalServerError)
		return
	}

	if len(inspectResult) == 0 {
		http.Error(w, "Volume not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(inspectResult[0])
}

func isVolumeInUse(volumeName string) (bool, error) {
	cmd := exec.Command("docker", "ps", "-a", "--filter", fmt.Sprintf("volume=%s", volumeName), "--format", "{{.ID}}")
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}

	// If output is not empty, volume is in use
	return strings.TrimSpace(string(output)) != "", nil
}

func dockerRemoveVolume(volumeName string, force bool) error {
	args := []string{"volume", "rm"}
	if force {
		args = append(args, "-f")
	}
	args = append(args, volumeName)

	cmd := exec.Command("docker", args...)
	return cmd.Run()
}
