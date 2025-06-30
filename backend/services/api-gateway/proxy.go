package main

import (
	"fmt"
	"io"
	"net/http"
	"os/exec"

	"github.com/sirupsen/logrus"
)

func proxyToService(serviceURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Build target URL
		targetURL := fmt.Sprintf("http://%s%s", serviceURL, r.URL.Path)
		if r.URL.RawQuery != "" {
			targetURL += "?" + r.URL.RawQuery
		}

		// Create new request
		proxyReq, err := http.NewRequest(r.Method, targetURL, r.Body)
		if err != nil {
			logrus.WithError(err).Error("Failed to create proxy request")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Copy headers
		for key, values := range r.Header {
			for _, value := range values {
				proxyReq.Header.Add(key, value)
			}
		}

		// Make request
		client := &http.Client{}
		resp, err := client.Do(proxyReq)
		if err != nil {
			logrus.WithError(err).WithField("service", serviceURL).Error("Failed to proxy request")
			http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
			return
		}
		defer resp.Body.Close()

		// Copy response headers
		for key, values := range resp.Header {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}

		// Set status code
		w.WriteHeader(resp.StatusCode)

		// Copy response body
		io.Copy(w, resp.Body)
	}
}

func getSystemInfo(w http.ResponseWriter, r *http.Request) {
	// Get Docker system info
	cmd := exec.Command("docker", "system", "info", "--format", "json")
	output, err := cmd.Output()
	if err != nil {
		logrus.WithError(err).Error("Failed to get system info")
		http.Error(w, "Failed to get system info: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Parse and return the JSON response
	w.Header().Set("Content-Type", "application/json")
	w.Write(output)
}
