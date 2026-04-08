package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	tildeexport "github.com/dpb587/dpb587.me/tools/tilde/export"
)

const (
	port        = "1313"
	hugoPort    = "1314"
	hugoBaseURL = "http://127.0.0.1:" + hugoPort
	upstream    = "https://storage.googleapis.com/dpb587-www-tilde-us-central1"
)

var tildeDir string
var publicDir string
var contentDir string

var exportHandler *tildeexport.Handler

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: command <repository-root>")
	}

	repositoryDir := os.Args[1]
	tildeDir = filepath.Join(repositoryDir, "tmp/tilde")
	contentDir = filepath.Join(repositoryDir, "content")
	hugoDir := filepath.Join(repositoryDir, "hugo")
	publicDir = filepath.Join(hugoDir, "public")

	for _, dir := range []string{repositoryDir, contentDir, hugoDir} {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			log.Fatalf("Directory does not exist: %s", dir)
		}
	}

	exportHandler = &tildeexport.Handler{
		PublicResolver: &tildeexport.DirResolver{Root: publicDir},
		ContentDir:     contentDir,
	}

	hugoCmd := exec.Command("hugo", "serve", "--buildFuture", "--port", hugoPort)
	hugoCmd.Dir = hugoDir
	hugoCmd.Stdout = os.Stdout
	hugoCmd.Stderr = os.Stderr

	if err := hugoCmd.Start(); err != nil {
		log.Fatalf("Failed to start hugo: %v", err)
	}

	log.Printf("Started hugo serve on port %s (pid %d)", hugoPort, hugoCmd.Process.Pid)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		log.Printf("Shutting down: terminating hugo process (pid %d)...", hugoCmd.Process.Pid)
		if err := hugoCmd.Process.Signal(syscall.SIGTERM); err != nil {
			log.Printf("Failed to terminate hugo: %v", err)
			_ = hugoCmd.Process.Kill()
		}
		_ = hugoCmd.Wait()
		os.Exit(0)
	}()

	http.HandleFunc("/", handleRequest)

	log.Printf("Starting dev server on port %s", port)
	log.Printf("Repository root: %s", repositoryDir)
	log.Printf("Hugo at %s", hugoBaseURL)
	log.Printf("Tilde fallback upstream: %s", upstream)

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
		proxyTo(hugoBaseURL, w, r)
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
	localPath, ok := safeLocalPath(tildeDir, relativePath)
	if !ok {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if fileInfo, err := os.Stat(localPath); err == nil && !fileInfo.IsDir() {
		http.ServeFile(w, r, localPath)
		return
	}

	proxyTo(upstream, w, r)
}

func proxyTo(baseURL string, w http.ResponseWriter, r *http.Request) {
	upstreamURL := baseURL + r.URL.Path
	if r.URL.RawQuery != "" {
		upstreamURL += "?" + r.URL.RawQuery
	}

	upstreamReq, err := http.NewRequest(http.MethodGet, upstreamURL, nil)
	if err != nil {
		http.Error(w, "Failed to create upstream request", http.StatusInternalServerError)
		return
	}

	for name, values := range r.Header {
		if name == "Host" {
			continue
		}
		for _, value := range values {
			upstreamReq.Header.Add(name, value)
		}
	}

	client := &http.Client{}
	resp, err := client.Do(upstreamReq)
	if err != nil {
		http.Error(w, "Failed to reach upstream", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	for name, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}

	w.WriteHeader(resp.StatusCode)

	_, err = io.Copy(w, resp.Body)
	if err != nil {
		log.Printf("Error copying response body: %v", err)
	}
}
