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

// DockerNetwork represents a Docker network
type DockerNetwork struct {
	NetworkID string `json:"NetworkID"`
	Name      string `json:"Name"`
	Driver    string `json:"Driver"`
	Scope     string `json:"Scope"`
	IPv6      string `json:"IPv6"`
	Internal  string `json:"Internal"`
	Labels    string `json:"Labels"`
	CreatedAt string `json:"CreatedAt"`
}

type CreateNetworkRequest struct {
	Name       string            `json:"name"`
	Driver     string            `json:"driver"`
	Options    map[string]string `json:"options"`
	Labels     map[string]string `json:"labels"`
	Internal   bool              `json:"internal"`
	EnableIPv6 bool              `json:"enable_ipv6"`
	IPAM       *IPAMConfig       `json:"ipam"`
}

type IPAMConfig struct {
	Driver  string       `json:"driver"`
	Config  []IPAMSubnet `json:"config"`
	Options map[string]string `json:"options"`
}

type IPAMSubnet struct {
	Subnet  string `json:"subnet"`
	Gateway string `json:"gateway"`
}

type NetworkInspectResult struct {
	Name       string                 `json:"Name"`
	ID         string                 `json:"Id"`
	Created    string                 `json:"Created"`
	Scope      string                 `json:"Scope"`
	Driver     string                 `json:"Driver"`
	EnableIPv6 bool                   `json:"EnableIPv6"`
	IPAM       map[string]interface{} `json:"IPAM"`
	Internal   bool                   `json:"Internal"`
	Attachable bool                   `json:"Attachable"`
	Ingress    bool                   `json:"Ingress"`
	ConfigFrom map[string]interface{} `json:"ConfigFrom"`
	ConfigOnly bool                   `json:"ConfigOnly"`
	Containers map[string]interface{} `json:"Containers"`
	Options    map[string]string      `json:"Options"`
	Labels     map[string]string      `json:"Labels"`
}

// convertToFrontendFormat converts raw Docker network data to frontend format
func convertToFrontendFormat(raw DockerNetwork) map[string]interface{} {
	// Parse labels
	labels := make(map[string]string)
	if raw.Labels != "" && raw.Labels != "{}" {
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

	// Parse created time
	created, _ := time.Parse("2006-01-02 15:04:05 -0700 MST", raw.CreatedAt)

	return map[string]interface{}{
		"Name":       raw.Name,
		"Id":         raw.NetworkID,
		"Created":    created.Format(time.RFC3339),
		"Scope":      raw.Scope,
		"Driver":     raw.Driver,
		"EnableIPv6": raw.IPv6 == "true",
		"Internal":   raw.Internal == "true",
		"Attachable": true,
		"Ingress":    false,
		"ConfigOnly": false,
		"Containers": map[string]interface{}{},
		"Options":    map[string]string{},
		"Labels":     labels,
	}
}

func getRealNetworks() ([]map[string]interface{}, error) {
	cmd := exec.Command("docker", "network", "ls", "--format", "json")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute docker network ls: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var networks []map[string]interface{}
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var networkJSON map[string]interface{}
		if err := json.Unmarshal([]byte(line), &networkJSON); err != nil {
			logrus.WithError(err).WithField("line", line).Warn("Failed to parse network JSON")
			continue
		}

		// Convert to our expected format
		network := map[string]interface{}{
			"id":       networkJSON["ID"],
			"name":     networkJSON["Name"],
			"driver":   networkJSON["Driver"],
			"scope":    networkJSON["Scope"],
			"ipv6":     networkJSON["IPv6"],
			"internal": networkJSON["Internal"],
			"labels":   networkJSON["Labels"],
			"created":  networkJSON["CreatedAt"],
		}

		networks = append(networks, network)
	}

	return networks, nil
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

func createNetwork(w http.ResponseWriter, r *http.Request) {
	var req CreateNetworkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Build docker network create command
	args := []string{"network", "create"}

	// Add driver if specified
	if req.Driver != "" {
		args = append(args, "--driver", req.Driver)
	}

	// Add options
	for key, value := range req.Options {
		args = append(args, "--opt", fmt.Sprintf("%s=%s", key, value))
	}

	// Add labels
	for key, value := range req.Labels {
		args = append(args, "--label", fmt.Sprintf("%s=%s", key, value))
	}

	// Add internal flag
	if req.Internal {
		args = append(args, "--internal")
	}

	// Add IPv6 flag
	if req.EnableIPv6 {
		args = append(args, "--ipv6")
	}

	// Add IPAM configuration
	if req.IPAM != nil {
		if req.IPAM.Driver != "" {
			args = append(args, "--ipam-driver", req.IPAM.Driver)
		}
		for _, config := range req.IPAM.Config {
			if config.Subnet != "" {
				args = append(args, "--subnet", config.Subnet)
			}
			if config.Gateway != "" {
				args = append(args, "--gateway", config.Gateway)
			}
		}
		for key, value := range req.IPAM.Options {
			args = append(args, "--ipam-opt", fmt.Sprintf("%s=%s", key, value))
		}
	}

	// Add network name
	args = append(args, req.Name)

	cmd := exec.Command("docker", args...)
	output, err := cmd.Output()
	if err != nil {
		logrus.WithError(err).WithField("args", args).Error("Failed to create network")
		http.Error(w, "Failed to create network: "+err.Error(), http.StatusInternalServerError)
		return
	}

	networkID := strings.TrimSpace(string(output))
	logrus.WithFields(logrus.Fields{
		"network_id":   networkID,
		"network_name": req.Name,
		"driver":       req.Driver,
	}).Info("Network created successfully")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message":    "Network created successfully",
		"network_id": networkID,
		"name":       req.Name,
	})
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

func inspectNetwork(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	cmd := exec.Command("docker", "network", "inspect", id)
	output, err := cmd.Output()
	if err != nil {
		logrus.WithError(err).WithField("network", id).Error("Failed to inspect network")
		http.Error(w, "Failed to inspect network: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var inspectResult []NetworkInspectResult
	if err := json.Unmarshal(output, &inspectResult); err != nil {
		logrus.WithError(err).Error("Failed to parse network inspect output")
		http.Error(w, "Failed to parse network inspect output", http.StatusInternalServerError)
		return
	}

	if len(inspectResult) == 0 {
		http.Error(w, "Network not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(inspectResult[0])
}

func dockerRemoveNetwork(networkID string) error {
	cmd := exec.Command("docker", "network", "rm", networkID)
	return cmd.Run()
}
