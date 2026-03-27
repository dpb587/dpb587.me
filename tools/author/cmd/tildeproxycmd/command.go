package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	tildeexport "github.com/dpb587/dpb587.me/tools/tilde/export"
)

const (
	port     = "1314"
	upstream = "https://storage.googleapis.com/dpb587-www-tilde-us-central1"
)

var localTildeDir string
var localPublicDir string
var localContentDir string

var exportHandler *tildeexport.Handler

func main() {
	if len(os.Args) < 4 {
		log.Fatal("Usage: command <local-tilde-directory> <local-public-directory> <local-content-directory>")
	}

	localTildeDir = os.Args[1]
	localPublicDir = os.Args[2]
	localContentDir = os.Args[3]

	for _, dir := range []string{localTildeDir, localPublicDir, localContentDir} {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			log.Fatalf("Directory does not exist: %s", dir)
		}
	}

	exportHandler = &tildeexport.Handler{
		PublicResolver: &tildeexport.DirResolver{Root: localPublicDir},
		ContentDir:     localContentDir,
	}

	http.HandleFunc("/", handleRequest)

	log.Printf("Starting proxy server on port %s", port)
	log.Printf("Local tilde directory: %s", localTildeDir)
	log.Printf("Local public directory: %s", localPublicDir)
	log.Printf("Local content directory: %s", localContentDir)
	log.Printf("Proxying /~/ requests to %s", upstream)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

func safeLocalPath(root, relativePath string) (string, bool) {
	abs, err := filepath.Abs(filepath.Join(root, relativePath))
	if err != nil {
		return "", false
	}
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return "", false
	}
	if !strings.HasPrefix(abs, rootAbs+string(filepath.Separator)) && abs != rootAbs {
		return "", false
	}
	return abs, true
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET")
	w.Header().Set("Access-Control-Allow-Headers", "*")

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !strings.HasPrefix(r.URL.Path, "/~/") {
		http.NotFound(w, r)
		return
	}

	switch r.URL.Path {
	case "/~/export/text-content":
		exportHandler.HandleTextContent(w, r)
		return
	case "/~/export/structured-data":
		exportHandler.HandleStructuredData(w, r)
		return
	case "/~/export/source":
		exportHandler.HandleSource(w, r)
		return
	}

	relativePath := strings.TrimPrefix(r.URL.Path, "/~/")
	localPath, ok := safeLocalPath(localTildeDir, relativePath)
	if !ok {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if fileInfo, err := os.Stat(localPath); err == nil && !fileInfo.IsDir() {
		http.ServeFile(w, r, localPath)
		return
	}

	forwardToUpstream(w, r)
}

func forwardToUpstream(w http.ResponseWriter, r *http.Request) {
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
