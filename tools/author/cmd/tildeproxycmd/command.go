package main

// Hello, agent. Review this comment block and implement a concise solution.
//
// Core Requirements:
//
// * this is a development service
// * run a basic HTTP server on port 1314
// * only GET requests are allowed
// * any request received with the /~/ prefix must be handled as follows:
//   1. if the file exists from a virtual OS FS based on the first command line argument, return the file content
//   2. otherwise, forward to the upstream of https://storage.googleapis.com/dpb587-www-tilde-us-central1/ (appending the request path)
// * CORS headers must be set to allow any origin
//
// For example, a request to:
//
// http://localhost:1314/~/blob-geojson/00e238839079063e5ffba27ff65f4ac221e1e5ae17b46fc4686fc59ffc7ff807/geojson.json
//
// the file system must be checked for the path of:
//
// os.Args[1] + "/blob-geojson/00e238839079063e5ffba27ff65f4ac221e1e5ae17b46fc4686fc59ffc7ff807/geojson.json"
//
// must return the upstream response from:
//
// https://storage.googleapis.com/dpb587-www-tilde-us-central1/~/blob-geojson/00e238839079063e5ffba27ff65f4ac221e1e5ae17b46fc4686fc59ffc7ff807/geojson.json
//

import (
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	port     = "1314"
	upstream = "https://storage.googleapis.com/dpb587-www-tilde-us-central1"
)

var localRoot string

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: command <local-root-directory>")
	}

	localRoot = os.Args[1]

	// Verify the local root directory exists
	if _, err := os.Stat(localRoot); os.IsNotExist(err) {
		log.Fatalf("Local root directory does not exist: %s", localRoot)
	}

	http.HandleFunc("/", handleRequest)

	log.Printf("Starting proxy server on port %s", port)
	log.Printf("Local filesystem root: %s", localRoot)
	log.Printf("Proxying /~/ requests to %s", upstream)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers to allow any origin
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET")
	w.Header().Set("Access-Control-Allow-Headers", "*")

	// Only allow GET requests
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Only handle requests that start with /~/
	if !strings.HasPrefix(r.URL.Path, "/~/") {
		http.NotFound(w, r)
		return
	}

	// Extract the path after /~/ for local filesystem check
	relativePath := strings.TrimPrefix(r.URL.Path, "/~/")
	localPath := filepath.Join(localRoot, relativePath)

	// Check if file exists locally
	if fileInfo, err := os.Stat(localPath); err == nil && !fileInfo.IsDir() {
		// File exists locally, serve it
		http.ServeFile(w, r, localPath)
		return
	}

	// File doesn't exist locally, forward to upstream
	forwardToUpstream(w, r)
}

func forwardToUpstream(w http.ResponseWriter, r *http.Request) {
	// Build upstream URL
	upstreamURL := upstream + r.URL.Path
	if r.URL.RawQuery != "" {
		upstreamURL += "?" + r.URL.RawQuery
	}

	// Create request to upstream
	upstreamReq, err := http.NewRequest(http.MethodGet, upstreamURL, nil)
	if err != nil {
		http.Error(w, "Failed to create upstream request", http.StatusInternalServerError)
		return
	}

	// Copy relevant headers from original request
	for name, values := range r.Header {
		if name == "Host" {
			continue // Don't copy Host header
		}
		for _, value := range values {
			upstreamReq.Header.Add(name, value)
		}
	}

	// Make request to upstream
	client := &http.Client{}
	resp, err := client.Do(upstreamReq)
	if err != nil {
		http.Error(w, "Failed to reach upstream", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for name, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}

	// Set status code
	w.WriteHeader(resp.StatusCode)

	// Copy response body
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		log.Printf("Error copying response body: %v", err)
	}
}
