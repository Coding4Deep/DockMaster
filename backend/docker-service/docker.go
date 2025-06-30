package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// DockerContainer represents a Docker container
type DockerContainer struct {
	ID      string       `json:"ID"`
	Names   string       `json:"Names"`
	Image   string       `json:"Image"`
	Command string       `json:"Command"`
	Created string       `json:"CreatedAt"`
	Ports   string       `json:"Ports"`
	Labels  string       `json:"Labels"`
	State   string       `json:"State"`
	Status  string       `json:"Status"`
	Mounts  string       `json:"Mounts"`
	Size    string       `json:"Size"`
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
		// Simple port parsing - in production you'd want more robust parsing
		portParts := strings.Split(raw.Ports, ",")
		for _, portStr := range portParts {
			portStr = strings.TrimSpace(portStr)
			if strings.Contains(portStr, "->") {
				// Format: "0.0.0.0:8080->8080/tcp"
				parts := strings.Split(portStr, "->")
				if len(parts) == 2 {
					publicPart := strings.TrimSpace(parts[0])
					privatePart := strings.TrimSpace(parts[1])
					
					// Extract public port
					if colonIdx := strings.LastIndex(publicPart, ":"); colonIdx != -1 {
						publicPortStr := publicPart[colonIdx+1:]
						if publicPort, err := strconv.Atoi(publicPortStr); err == nil {
							// Extract private port and type
							if slashIdx := strings.Index(privatePart, "/"); slashIdx != -1 {
								privatePortStr := privatePart[:slashIdx]
								portType := privatePart[slashIdx+1:]
								if privatePort, err := strconv.Atoi(privatePortStr); err == nil {
									ports = append(ports, map[string]interface{}{
										"PublicPort":  publicPort,
										"PrivatePort": privatePort,
										"Type":        portType,
									})
								}
							}
						}
					}
				}
			} else if strings.Contains(portStr, "/") {
				// Format: "80/tcp"
				parts := strings.Split(portStr, "/")
				if len(parts) == 2 {
					if privatePort, err := strconv.Atoi(parts[0]); err == nil {
						ports = append(ports, map[string]interface{}{
							"PrivatePort": privatePort,
							"Type":        parts[1],
						})
					}
				}
			}
		}
	}

	return map[string]interface{}{
		"Id":     raw.ID,
		"Names":  names,
		"Image":  raw.Image,
		"State":  raw.State,
		"Status": raw.Status,
		"Ports":  ports,
	}
}

// DockerImage represents a Docker image
type DockerImage struct {
	ID         string `json:"ID"`
	Repository string `json:"Repository"`
	Tag        string `json:"Tag"`
	Created    string `json:"CreatedAt"`
	Size       string `json:"Size"`
}

// DockerVolume represents a Docker volume
type DockerVolume struct {
	Driver     string `json:"Driver"`
	Name       string `json:"Name"`
	Size       string `json:"Size"`
	CreatedAt  string `json:"CreatedAt"`
}

// DockerNetwork represents a Docker network
type DockerNetwork struct {
	ID      string `json:"ID"`
	Name    string `json:"Name"`
	Driver  string `json:"Driver"`
	Scope   string `json:"Scope"`
	Created string `json:"CreatedAt"`
}

// ContainerStats represents container statistics
type ContainerStats struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	CPUPerc     float64 `json:"cpuPerc"`
	MemUsage    int64   `json:"memUsage"`
	MemLimit    int64   `json:"memLimit"`
	MemPerc     float64 `json:"memPerc"`
	NetRx       int64   `json:"netRx"`
	NetTx       int64   `json:"netTx"`
	BlockRead   int64   `json:"blockRead"`
	BlockWrite  int64   `json:"blockWrite"`
	PIDs        int64   `json:"pids"`
}

// SystemInfo represents Docker system information
type SystemInfo struct {
	Containers map[string]interface{} `json:"containers"`
	Images     int                    `json:"images"`
	Version    map[string]interface{} `json:"version"`
	System     map[string]interface{} `json:"system"`
}

// executeDockerCommand executes a docker command and returns the output
func executeDockerCommand(args ...string) ([]byte, error) {
	cmd := exec.Command("docker", args...)
	output, err := cmd.Output()
	if err != nil {
		logrus.WithError(err).WithField("command", "docker "+strings.Join(args, " ")).Error("Docker command failed")
		return nil, fmt.Errorf("docker command failed: %v", err)
	}
	return output, nil
}

// getRealContainers gets actual containers from Docker
func getRealContainers(all bool) ([]map[string]interface{}, error) {
	args := []string{"ps", "--format", "json", "--no-trunc"}
	if all {
		args = append(args, "-a")
	}

	output, err := executeDockerCommand(args...)
	if err != nil {
		return nil, err
	}

	var containers []map[string]interface{}
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		
		var container DockerContainer
		if err := json.Unmarshal([]byte(line), &container); err != nil {
			logrus.WithError(err).WithField("line", line).Error("Failed to parse container JSON")
			continue
		}
		
		// Convert to frontend format
		frontendContainer := convertToFrontendFormat(container)
		containers = append(containers, frontendContainer)
	}

	return containers, nil
}

// getRealImages gets actual images from Docker
func getRealImages() ([]map[string]interface{}, error) {
	output, err := executeDockerCommand("images", "--format", "json", "--no-trunc")
	if err != nil {
		return nil, err
	}

	var images []map[string]interface{}
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		
		var image DockerImage
		if err := json.Unmarshal([]byte(line), &image); err != nil {
			logrus.WithError(err).WithField("line", line).Error("Failed to parse image JSON")
			continue
		}
		
		// Convert to frontend format
		repoTag := image.Repository + ":" + image.Tag
		if image.Tag == "<none>" {
			repoTag = "<none>:<none>"
		}
		
		frontendImage := map[string]interface{}{
			"Id":       image.ID,
			"RepoTags": []string{repoTag},
			"Created":  parseDockerTime(image.Created),
			"Size":     parseDockerSize(image.Size),
		}
		images = append(images, frontendImage)
	}

	return images, nil
}

// getRealVolumes gets actual volumes from Docker
func getRealVolumes() ([]map[string]interface{}, error) {
	output, err := executeDockerCommand("volume", "ls", "--format", "json")
	if err != nil {
		return nil, err
	}

	var volumes []map[string]interface{}
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		
		var volume DockerVolume
		if err := json.Unmarshal([]byte(line), &volume); err != nil {
			logrus.WithError(err).WithField("line", line).Error("Failed to parse volume JSON")
			continue
		}
		
		// Convert to frontend format
		frontendVolume := map[string]interface{}{
			"Name":       volume.Name,
			"Driver":     volume.Driver,
			"Mountpoint": "/var/lib/docker/volumes/" + volume.Name + "/_data",
			"CreatedAt":  volume.CreatedAt,
			"Scope":      "local",
			"Labels":     map[string]string{},
			"Options":    map[string]string{},
		}
		volumes = append(volumes, frontendVolume)
	}

	return volumes, nil
}

// getRealNetworks gets actual networks from Docker
func getRealNetworks() ([]map[string]interface{}, error) {
	output, err := executeDockerCommand("network", "ls", "--format", "json")
	if err != nil {
		return nil, err
	}

	var networks []map[string]interface{}
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		
		var network DockerNetwork
		if err := json.Unmarshal([]byte(line), &network); err != nil {
			logrus.WithError(err).WithField("line", line).Error("Failed to parse network JSON")
			continue
		}
		
		// Convert to frontend format
		frontendNetwork := map[string]interface{}{
			"Id":         network.ID,
			"Name":       network.Name,
			"Created":    network.Created,
			"Scope":      network.Scope,
			"Driver":     network.Driver,
			"EnableIPv6": false,
			"Internal":   false,
			"Attachable": false,
			"Ingress":    false,
			"ConfigOnly": false,
			"Containers": map[string]interface{}{},
			"Options":    map[string]string{},
			"Labels":     map[string]string{},
		}
		networks = append(networks, frontendNetwork)
	}

	return networks, nil
}

// getRealContainerStats gets actual container statistics
func getRealContainerStats(containerID string) (*ContainerStats, error) {
	// Get container stats using docker stats command
	output, err := executeDockerCommand("stats", "--no-stream", "--format", "json", containerID)
	if err != nil {
		return nil, err
	}

	// Parse the stats JSON
	var rawStats map[string]interface{}
	if err := json.Unmarshal(output, &rawStats); err != nil {
		return nil, fmt.Errorf("failed to parse stats JSON: %v", err)
	}

	// Extract and convert stats
	stats := &ContainerStats{
		ID:   containerID,
		Name: getStringValue(rawStats, "Name"),
	}

	// Parse CPU percentage
	if cpuStr := getStringValue(rawStats, "CPUPerc"); cpuStr != "" {
		cpuStr = strings.TrimSuffix(cpuStr, "%")
		if cpu, err := strconv.ParseFloat(cpuStr, 64); err == nil {
			stats.CPUPerc = cpu
		}
	}

	// Parse memory usage and percentage
	if memStr := getStringValue(rawStats, "MemUsage"); memStr != "" {
		parts := strings.Split(memStr, " / ")
		if len(parts) == 2 {
			if usage := parseMemoryString(parts[0]); usage > 0 {
				stats.MemUsage = usage
			}
			if limit := parseMemoryString(parts[1]); limit > 0 {
				stats.MemLimit = limit
				if stats.MemUsage > 0 {
					stats.MemPerc = float64(stats.MemUsage) / float64(stats.MemLimit) * 100
				}
			}
		}
	}

	// Parse network I/O
	if netStr := getStringValue(rawStats, "NetIO"); netStr != "" {
		parts := strings.Split(netStr, " / ")
		if len(parts) == 2 {
			if rx := parseMemoryString(parts[0]); rx > 0 {
				stats.NetRx = rx
			}
			if tx := parseMemoryString(parts[1]); tx > 0 {
				stats.NetTx = tx
			}
		}
	}

	// Parse block I/O
	if blockStr := getStringValue(rawStats, "BlockIO"); blockStr != "" {
		parts := strings.Split(blockStr, " / ")
		if len(parts) == 2 {
			if read := parseMemoryString(parts[0]); read > 0 {
				stats.BlockRead = read
			}
			if write := parseMemoryString(parts[1]); write > 0 {
				stats.BlockWrite = write
			}
		}
	}

	// Parse PIDs
	if pidsStr := getStringValue(rawStats, "PIDs"); pidsStr != "" {
		if pids, err := strconv.ParseInt(pidsStr, 10, 64); err == nil {
			stats.PIDs = pids
		}
	}

	return stats, nil
}

// getRealSystemInfo gets actual Docker system information
func getRealSystemInfo() (*SystemInfo, error) {
	// Get system info
	output, err := executeDockerCommand("system", "info", "--format", "json")
	if err != nil {
		return nil, err
	}

	var rawInfo map[string]interface{}
	if err := json.Unmarshal(output, &rawInfo); err != nil {
		return nil, fmt.Errorf("failed to parse system info: %v", err)
	}

	// Get version info
	versionOutput, err := executeDockerCommand("version", "--format", "json")
	if err != nil {
		logrus.WithError(err).Warn("Failed to get Docker version")
	}

	var versionInfo map[string]interface{}
	if versionOutput != nil {
		json.Unmarshal(versionOutput, &versionInfo)
	}

	// Build system info response
	info := &SystemInfo{
		Containers: map[string]interface{}{
			"total":   getIntValue(rawInfo, "Containers"),
			"running": getIntValue(rawInfo, "ContainersRunning"),
			"paused":  getIntValue(rawInfo, "ContainersPaused"),
			"stopped": getIntValue(rawInfo, "ContainersStopped"),
		},
		Images: getIntValue(rawInfo, "Images"),
		System: map[string]interface{}{
			"totalMemory":  getIntValue(rawInfo, "MemTotal"),
			"cpus":         getIntValue(rawInfo, "NCPU"),
			"osType":       getStringValue(rawInfo, "OSType"),
			"architecture": getStringValue(rawInfo, "Architecture"),
		},
	}

	// Add version info if available
	if versionInfo != nil {
		if server, ok := versionInfo["Server"].(map[string]interface{}); ok {
			info.Version = map[string]interface{}{
				"version":    getStringValue(server, "Version"),
				"apiVersion": getStringValue(server, "ApiVersion"),
				"goVersion":  getStringValue(server, "GoVersion"),
			}
		}
	}

	return info, nil
}

// Helper functions
// Helper functions
func getStringValue(data map[string]interface{}, key string) string {
	if val, ok := data[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getIntValue(data map[string]interface{}, key string) int {
	if val, ok := data[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case float64:
			return int(v)
		case string:
			if i, err := strconv.Atoi(v); err == nil {
				return i
			}
		}
	}
	return 0
}

func parseDockerTime(timeStr string) int64 {
	// Try to parse Docker time format
	if t, err := time.Parse("2006-01-02 15:04:05 -0700 MST", timeStr); err == nil {
		return t.Unix()
	}
	return time.Now().Unix()
}

func parseDockerSize(sizeStr string) int64 {
	// Simple size parsing - convert MB/GB to bytes
	sizeStr = strings.TrimSpace(sizeStr)
	if sizeStr == "" {
		return 0
	}
	
	multiplier := int64(1)
	if strings.HasSuffix(sizeStr, "MB") {
		multiplier = 1024 * 1024
		sizeStr = strings.TrimSuffix(sizeStr, "MB")
	} else if strings.HasSuffix(sizeStr, "GB") {
		multiplier = 1024 * 1024 * 1024
		sizeStr = strings.TrimSuffix(sizeStr, "GB")
	} else if strings.HasSuffix(sizeStr, "kB") {
		multiplier = 1024
		sizeStr = strings.TrimSuffix(sizeStr, "kB")
	}
	
	if val, err := strconv.ParseFloat(sizeStr, 64); err == nil {
		return int64(val * float64(multiplier))
	}
	
	return 0
}

func parseMemoryString(memStr string) int64 {
	memStr = strings.TrimSpace(memStr)
	if memStr == "" {
		return 0
	}

	// Handle different units
	multiplier := int64(1)
	if strings.HasSuffix(memStr, "KiB") || strings.HasSuffix(memStr, "kB") || strings.HasSuffix(memStr, "K") {
		multiplier = 1024
		memStr = strings.TrimSuffix(strings.TrimSuffix(strings.TrimSuffix(memStr, "KiB"), "kB"), "K")
	} else if strings.HasSuffix(memStr, "MiB") || strings.HasSuffix(memStr, "MB") || strings.HasSuffix(memStr, "M") {
		multiplier = 1024 * 1024
		memStr = strings.TrimSuffix(strings.TrimSuffix(strings.TrimSuffix(memStr, "MiB"), "MB"), "M")
	} else if strings.HasSuffix(memStr, "GiB") || strings.HasSuffix(memStr, "GB") || strings.HasSuffix(memStr, "G") {
		multiplier = 1024 * 1024 * 1024
		memStr = strings.TrimSuffix(strings.TrimSuffix(strings.TrimSuffix(memStr, "GiB"), "GB"), "G")
	} else if strings.HasSuffix(memStr, "B") {
		memStr = strings.TrimSuffix(memStr, "B")
	}

	if val, err := strconv.ParseFloat(memStr, 64); err == nil {
		return int64(val * float64(multiplier))
	}

	return 0
}

// Docker operations
func dockerStart(containerID string) error {
	_, err := executeDockerCommand("start", containerID)
	return err
}

func dockerStop(containerID string) error {
	_, err := executeDockerCommand("stop", containerID)
	return err
}

func dockerRestart(containerID string) error {
	_, err := executeDockerCommand("restart", containerID)
	return err
}

func dockerRemove(containerID string, force bool) error {
	args := []string{"rm", containerID}
	if force {
		args = []string{"rm", "-f", containerID}
	}
	_, err := executeDockerCommand(args...)
	return err
}

func dockerRemoveImage(imageID string, force bool) error {
	args := []string{"rmi", imageID}
	if force {
		args = []string{"rmi", "-f", imageID}
	}
	_, err := executeDockerCommand(args...)
	return err
}

func dockerRemoveVolume(volumeName string, force bool) error {
	args := []string{"volume", "rm", volumeName}
	if force {
		args = []string{"volume", "rm", "-f", volumeName}
	}
	_, err := executeDockerCommand(args...)
	return err
}

func dockerRemoveNetwork(networkID string) error {
	_, err := executeDockerCommand("network", "rm", networkID)
	return err
}

func dockerLogs(containerID string, tail string) ([]map[string]interface{}, error) {
	args := []string{"logs", "--timestamps"}
	if tail != "" {
		args = append(args, "--tail", tail)
	}
	args = append(args, containerID)

	output, err := executeDockerCommand(args...)
	if err != nil {
		return nil, err
	}

	var logs []map[string]interface{}
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		// Parse timestamp and log message
		parts := strings.SplitN(line, " ", 2)
		if len(parts) >= 2 {
			timestamp := parts[0]
			message := parts[1]
			
			// Try to parse timestamp
			var parsedTime time.Time
			if t, err := time.Parse(time.RFC3339Nano, timestamp); err == nil {
				parsedTime = t
			} else {
				parsedTime = time.Now()
			}

			logs = append(logs, map[string]interface{}{
				"timestamp": parsedTime,
				"stream":    "stdout",
				"log":       message,
			})
		} else {
			logs = append(logs, map[string]interface{}{
				"timestamp": time.Now(),
				"stream":    "stdout",
				"log":       line,
			})
		}
	}

	return logs, nil
}
