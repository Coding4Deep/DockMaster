package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// DockerImage represents a Docker image
type DockerImage struct {
	Repository string `json:"Repository"`
	Tag        string `json:"Tag"`
	ImageID    string `json:"ImageID"`
	Created    string `json:"Created"`
	Size       string `json:"Size"`
}

type DockerHubSearchResult struct {
	Count    int                     `json:"count"`
	Next     string                  `json:"next"`
	Previous string                  `json:"previous"`
	Results  []DockerHubImageSummary `json:"results"`
}

type DockerHubImageSummary struct {
	Name             string `json:"name"`
	Description      string `json:"description"`
	StarCount        int    `json:"star_count"`
	PullCount        int    `json:"pull_count"`
	RepoOwner        string `json:"repo_owner"`
	ShortDescription string `json:"short_description"`
	IsOfficial       bool   `json:"is_official"`
	IsAutomated      bool   `json:"is_automated"`
}

type DockerHubImageDetails struct {
	Name             string                   `json:"name"`
	Description      string                   `json:"description"`
	StarCount        int                      `json:"star_count"`
	PullCount        int                      `json:"pull_count"`
	LastUpdated      string                   `json:"last_updated"`
	IsOfficial       bool                     `json:"is_official"`
	IsAutomated      bool                     `json:"is_automated"`
	CanEdit          bool                     `json:"can_edit"`
	User             string                   `json:"user"`
	HasStarred       bool                     `json:"has_starred"`
	FullDescription  string                   `json:"full_description"`
	Permissions      map[string]interface{}   `json:"permissions"`
	Tags             []DockerHubTag           `json:"tags"`
}

type DockerHubTag struct {
	Name                string    `json:"name"`
	FullSize            int64     `json:"full_size"`
	ID                  int       `json:"id"`
	Repository          int       `json:"repository"`
	Creator             int       `json:"creator"`
	LastUpdater         int       `json:"last_updater"`
	LastUpdated         time.Time `json:"last_updated"`
	ImageID             string    `json:"image_id"`
	V2                  bool      `json:"v2"`
	TagStatus           string    `json:"tag_status"`
	TagLastPulled       time.Time `json:"tag_last_pulled"`
	TagLastPushed       time.Time `json:"tag_last_pushed"`
}

type PullImageRequest struct {
	Image string `json:"image"`
	Tag   string `json:"tag"`
}

// convertToFrontendFormat converts raw Docker image data to frontend format
func convertToFrontendFormat(raw DockerImage) map[string]interface{} {
	// Parse created time
	created, _ := time.Parse("2006-01-02 15:04:05 -0700 MST", raw.Created)

	// Parse size
	sizeBytes := int64(0)
	if raw.Size != "" {
		// Simple size parsing - convert MB/GB to bytes
		sizeStr := strings.ToLower(raw.Size)
		if strings.Contains(sizeStr, "mb") {
			if val, err := strconv.ParseFloat(strings.Replace(sizeStr, "mb", "", -1), 64); err == nil {
				sizeBytes = int64(val * 1024 * 1024)
			}
		} else if strings.Contains(sizeStr, "gb") {
			if val, err := strconv.ParseFloat(strings.Replace(sizeStr, "gb", "", -1), 64); err == nil {
				sizeBytes = int64(val * 1024 * 1024 * 1024)
			}
		}
	}

	return map[string]interface{}{
		"Id":          raw.ImageID,
		"ParentId":    "",
		"RepoTags":    []string{fmt.Sprintf("%s:%s", raw.Repository, raw.Tag)},
		"RepoDigests": []string{},
		"Created":     created.Unix(),
		"Size":        sizeBytes,
		"VirtualSize": sizeBytes,
		"SharedSize":  0,
		"Labels":      map[string]string{},
		"Containers":  0,
	}
}

func getRealImages() ([]map[string]interface{}, error) {
	cmd := exec.Command("docker", "images", "--format", "json")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute docker images: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var images []map[string]interface{}
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var imageJSON map[string]interface{}
		if err := json.Unmarshal([]byte(line), &imageJSON); err != nil {
			logrus.WithError(err).WithField("line", line).Warn("Failed to parse image JSON")
			continue
		}

		// Convert to our expected format
		image := map[string]interface{}{
			"id":         imageJSON["ID"],
			"repository": imageJSON["Repository"],
			"tag":        imageJSON["Tag"],
			"created":    imageJSON["CreatedAt"],
			"size":       imageJSON["Size"],
		}

		images = append(images, image)
	}

	return images, nil
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

func searchImages(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Query parameter 'q' is required", http.StatusBadRequest)
		return
	}

	// First search local images
	localImages, err := searchLocalImages(query)
	if err != nil {
		logrus.WithError(err).Warn("Failed to search local images")
	}

	// Then search Docker Hub
	hubImages, err := searchDockerHub(query)
	if err != nil {
		logrus.WithError(err).Warn("Failed to search Docker Hub")
	}

	result := map[string]interface{}{
		"local":      localImages,
		"docker_hub": hubImages,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func searchLocalImages(query string) ([]map[string]interface{}, error) {
	images, err := getRealImages()
	if err != nil {
		return nil, err
	}

	var filtered []map[string]interface{}
	queryLower := strings.ToLower(query)

	for _, image := range images {
		if repoTags, ok := image["RepoTags"].([]string); ok {
			for _, tag := range repoTags {
				if strings.Contains(strings.ToLower(tag), queryLower) {
					filtered = append(filtered, image)
					break
				}
			}
		}
	}

	return filtered, nil
}

func searchDockerHub(query string) ([]DockerHubImageSummary, error) {
	encodedQuery := url.QueryEscape(query)
	url := fmt.Sprintf("https://hub.docker.com/v2/search/repositories/?query=%s&page_size=25", encodedQuery)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var searchResult DockerHubSearchResult
	if err := json.Unmarshal(body, &searchResult); err != nil {
		return nil, err
	}

	return searchResult.Results, nil
}

func getImageDetails(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	imageID := vars["id"]

	// Check if it's a local image first
	localImages, err := getRealImages()
	if err == nil {
		for _, image := range localImages {
			if image["Id"] == imageID {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(image)
				return
			}
		}
	}

	// If not found locally, try to get details from Docker Hub
	// This assumes the imageID is in format "repo/name" or "name"
	hubDetails, err := getDockerHubImageDetails(imageID)
	if err != nil {
		http.Error(w, "Image not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(hubDetails)
}

func getDockerHubImageDetails(imageName string) (*DockerHubImageDetails, error) {
	// Handle official images (no namespace)
	if !strings.Contains(imageName, "/") {
		imageName = "library/" + imageName
	}

	url := fmt.Sprintf("https://hub.docker.com/v2/repositories/%s/", imageName)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("image not found on Docker Hub")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var details DockerHubImageDetails
	if err := json.Unmarshal(body, &details); err != nil {
		return nil, err
	}

	return &details, nil
}

func pullImage(w http.ResponseWriter, r *http.Request) {
	var req PullImageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	imageName := req.Image
	if req.Tag != "" && req.Tag != "latest" {
		imageName = fmt.Sprintf("%s:%s", req.Image, req.Tag)
	}

	logrus.WithField("image", imageName).Info("Starting image pull")

	cmd := exec.Command("docker", "pull", imageName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		logrus.WithError(err).WithField("image", imageName).Error("Failed to pull image")
		http.Error(w, fmt.Sprintf("Failed to pull image: %s", string(output)), http.StatusInternalServerError)
		return
	}

	logrus.WithField("image", imageName).Info("Image pulled successfully")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Image pulled successfully",
		"image":   imageName,
		"output":  string(output),
	})
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

func dockerRemoveImage(imageID string, force bool) error {
	args := []string{"rmi"}
	if force {
		args = append(args, "-f")
	}
	args = append(args, imageID)

	cmd := exec.Command("docker", args...)
	return cmd.Run()
}
