package main

// Hello, agent. Review this comment block and implement a concise solution. Retain this comment block, unmodified, in any solution.
//
// Input Requirements:
//
// * document root is for a web server which contains HTML files
// * static site generator creates an index.html file in most directories. For example, `/contact/index.html` might be a contact page.
// * HTML files will typically link to directories on the assumption that a directory path will resolve to the index.html file. For example, an HTML file may include a link to `/contact/`.
//
// Output Requirements:
//
// * all HTML links to local files must be rewritten without a trailing slash. For example, `/contact/` should be `/contact`.
// * assumption: a web server will ensure that `/contact` continues to serve the file located at `/contact/index.html`.
// * only relative links which resolve to a local file should be rewritten.
// * if an HTML file does not contain any valid links, it should be left unchanged.
// * the effective destination of a link must not change. For example, a link to `/` must not be rewritten to an empty string or entirely different path.
// * the `sitemap.xml` file must have <loc> entries rewritten to remove trailing slashes, when possible.
//
// Interface:
//
// * a command line tool which accepts a single argument to the document root directory
// * log every link that is rewritten (including the original and rewritten link)
// * log every file path when it is written to disk
// * use the std slog library for logging

import (
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <document-root>\n", os.Args[0])
		os.Exit(1)
	}

	documentRoot := os.Args[1]

	// Verify document root exists and is a directory
	info, err := os.Stat(documentRoot)
	if err != nil {
		slog.Error("failed to access document root", "path", documentRoot, "error", err)
		os.Exit(1)
	}
	if !info.IsDir() {
		slog.Error("document root is not a directory", "path", documentRoot)
		os.Exit(1)
	}

	err = filepath.WalkDir(documentRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		filename := d.Name()
		if strings.HasSuffix(filename, ".html") || filename == "sitemap.xml" {
			return processFile(path, documentRoot)
		}

		return nil
	})

	if err != nil {
		slog.Error("failed to walk directory", "path", documentRoot, "error", err)
		os.Exit(1)
	}
}

func processFile(filePath, documentRoot string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	originalContent := string(content)
	var modifiedContent string
	var hasChanges bool

	if filepath.Base(filePath) == "sitemap.xml" {
		modifiedContent, hasChanges = processSitemap(originalContent, filePath)
	} else {
		modifiedContent, hasChanges = processHTML(originalContent, filePath, documentRoot)
	}

	if hasChanges {
		err = os.WriteFile(filePath, []byte(modifiedContent), 0644)
		if err != nil {
			return fmt.Errorf("failed to write file %s: %w", filePath, err)
		}
		slog.Info("file written to disk", "path", filePath)
	}

	return nil
}

func processSitemap(content, filePath string) (string, bool) {
	// Match <loc>...</loc> entries and remove trailing slashes
	locRegex := regexp.MustCompile(`<loc>([^<]+)</loc>`)
	hasChanges := false

	result := locRegex.ReplaceAllStringFunc(content, func(match string) string {
		// Extract the URL from <loc>URL</loc>
		url := match[5 : len(match)-6] // Remove <loc> and </loc>

		// Don't modify root path "/" or URLs that don't end with "/"
		if url == "/" || !strings.HasSuffix(url, "/") {
			return match
		}

		// Remove trailing slash
		newURL := strings.TrimSuffix(url, "/")
		newMatch := "<loc>" + newURL + "</loc>"

		if newMatch != match {
			slog.Info("rewritten sitemap link", "file", filePath, "original", url, "rewritten", newURL)
			hasChanges = true
		}

		return newMatch
	})

	return result, hasChanges
}

func processHTML(content, filePath, documentRoot string) (string, bool) {
	// Match href attributes in anchor tags
	hrefRegex := regexp.MustCompile(`href=["']([^"']+)["']`)
	hasChanges := false

	result := hrefRegex.ReplaceAllStringFunc(content, func(match string) string {
		// Extract the href value
		quote := match[5]               // Either " or '
		href := match[6 : len(match)-1] // Remove href=" and "

		// Skip external URLs, anchors, mailto, etc.
		if strings.Contains(href, "://") ||
			strings.HasPrefix(href, "#") ||
			strings.HasPrefix(href, "mailto:") ||
			strings.HasPrefix(href, "tel:") ||
			strings.HasPrefix(href, "javascript:") {
			return match
		}

		// Don't modify root path "/" or URLs that don't end with "/"
		if href == "/" || !strings.HasSuffix(href, "/") {
			return match
		}

		// Check if this is a relative link that resolves to a local file
		var fullPath string
		if strings.HasPrefix(href, "/") {
			// Absolute path relative to document root
			fullPath = filepath.Join(documentRoot, href[1:])
		} else {
			// Relative path relative to current file's directory
			currentDir := filepath.Dir(filePath)
			fullPath = filepath.Join(currentDir, href)
		}

		// Check if there's an index.html file at this location
		indexPath := filepath.Join(fullPath, "index.html")
		if _, err := os.Stat(indexPath); err == nil {
			// Remove trailing slash
			newHref := strings.TrimSuffix(href, "/")
			newMatch := fmt.Sprintf("href=%c%s%c", quote, newHref, quote)

			if newMatch != match {
				slog.Info("rewritten link", "file", filePath, "original", href, "rewritten", newHref)
				hasChanges = true
			}

			return newMatch
		}

		return match
	})

	return result, hasChanges
}
