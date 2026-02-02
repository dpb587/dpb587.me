package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var Log = slog.Default()

func main() {
	var baseURL string
	flag.StringVar(&baseURL, "base-url", "", "Base URL for the content")
	flag.Parse()

	if baseURL == "" {
		fmt.Fprintf(os.Stderr, "Error: --base-url is required\n")
		os.Exit(1)
	}

	if flag.NArg() != 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s --base-url=BASE-URL HTML-DIR\n", os.Args[0])
		os.Exit(1)
	}

	htmlDir := flag.Arg(0)

	apiToken := os.Getenv("NG_API_TOKEN")
	if apiToken == "" {
		fmt.Fprintf(os.Stderr, "Error: NG_API_TOKEN environment variable is not set\n")
		os.Exit(1)
	}

	// Process files with worker pool
	if err := processFiles(htmlDir, baseURL, apiToken, 4); err != nil {
		fmt.Fprintf(os.Stderr, "Error processing files: %v\n", err)
		os.Exit(1)
	}
}

type job struct {
	filePath string
	baseURL  string
}

// processFiles processes HTML files in parallel using a worker pool
func processFiles(htmlDir, baseURL, apiToken string, workers int) error {
	jobs := make(chan job, workers*2)
	var workerWg sync.WaitGroup

	// Start workers
	for i := 0; i < workers; i++ {
		workerWg.Add(1)
		go func() {
			defer workerWg.Done()
			for j := range jobs {
				if err := convertHTMLToMarkdown(j.filePath, j.baseURL, apiToken); err != nil {
					Log.Error("error processing file", "path", j.filePath, "error", err)
				}
			}
		}()
	}

	// Walk directory and queue files as they are found
	walkErr := filepath.Walk(htmlDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(path, ".html") {
			return nil
		}

		// Calculate the relative path and construct the URL
		relPath, err := filepath.Rel(htmlDir, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path for %s: %w", path, err)
		}

		// Construct full URL: base URL + relative path
		fileBaseURL := strings.TrimSuffix(baseURL, "/") + "/" + relPath

		jobs <- job{
			filePath: path,
			baseURL:  fileBaseURL,
		}
		return nil
	})

	close(jobs)

	// Wait for workers to complete
	workerWg.Wait()

	// Check for walk errors
	if walkErr != nil {
		return fmt.Errorf("error walking directory: %w", walkErr)
	}

	return nil
}

// convertHTMLToMarkdown calls the API to convert HTML file to Markdown
func convertHTMLToMarkdown(filePath, baseURL, apiToken string) error {
	start := time.Now()

	// Read the HTML file
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Create multipart form
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	// Add sourceFile
	part, err := writer.CreateFormFile("sourceFile", filepath.Base(filePath))
	if err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err := io.Copy(part, file); err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	// Add format field
	if err := writer.WriteField("format", "markdown"); err != nil {
		return fmt.Errorf("failed to write format field: %w", err)
	}

	// Add baseUrl field
	if err := writer.WriteField("baseUrl", baseURL); err != nil {
		return fmt.Errorf("failed to write baseUrl field: %w", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close multipart writer: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", "https://api.namedgraph.com/toolkit.v0/textContent.export", &body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiToken)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Write response to .md file
	outputPath := filePath + ".md"
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	if _, err := io.Copy(outFile, resp.Body); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	Log.Info("wrote file", "path", outputPath, "duration_ms", time.Since(start).Milliseconds())
	return nil
}
