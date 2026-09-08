package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/dpb587/dpb587.me/tools/content"
	"github.com/dpb587/dpb587.me/tools/content/frontmatterparams"
	"github.com/dpb587/tacitkb/ext/linkmirror"
)

func main() {
	if err := mainErr(); err != nil {
		log.Fatal(err)
	}
}

func mainErr() error {
	if len(os.Args) != 3 {
		return fmt.Errorf("usage: go run ./author/cmd/linkmirrorcmd <source-hugo-directory> <target-url>")
	}

	sourceDir, err := cleanRelativeDirectory(os.Args[1])
	if err != nil {
		return err
	}
	targetURL := os.Args[2]
	repositoryDir, err := findRepositoryDir()
	if err != nil {
		return err
	}

	snapshot, err := linkmirror.Capture(context.Background(), sourceDir, targetURL)
	if err != nil {
		return err
	}
	linkmirror.MaterializeMetadata(context.Background(), snapshot, linkmirror.MirrorImageOptions{
		OutputDir: filepath.Join(repositoryDir, "tmp", "tilde", "mirror-blob-iiif-image-v3"),
		BaseURL:   linkmirror.MirrorImageServiceURL,
	})

	linkType, err := snapshotLinkType(snapshot)
	if err != nil {
		return err
	}
	document := content.Document{
		Frontmatter: &content.Content_Frontmatter{
			Params: &content.Content_Frontmatter_Params{LinkType: linkType},
		},
	}
	var output bytes.Buffer
	if _, err := document.WriteTo(&output); err != nil {
		return fmt.Errorf("write frontmatter: %w", err)
	}

	digest := sha256.Sum256([]byte(sourceDir + "\n" + targetURL))
	targetDir := filepath.Join(filepath.Dir(sourceDir), "embed")
	targetPath := filepath.Join(targetDir, "link-"+hex.EncodeToString(digest[:])[:10]+".md")
	if err := writeFileAtomically(targetPath, output.Bytes()); err != nil {
		return err
	}
	log.Printf("captured %s", targetURL)
	log.Printf("saved %s", targetPath)
	return nil
}

func cleanRelativeDirectory(raw string) (string, error) {
	if raw == "" || filepath.IsAbs(raw) {
		return "", fmt.Errorf("source Hugo directory must be a non-empty relative path")
	}
	clean := filepath.Clean(raw)
	if clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("source Hugo directory must remain within the current directory")
	}
	info, err := os.Stat(clean)
	if err != nil {
		return "", fmt.Errorf("stat source Hugo directory: %w", err)
	}
	if info.IsDir() {
		return "", fmt.Errorf("source Hugo directory is a directory: %s", clean)
	}
	return clean, nil
}

func findRepositoryDir() (string, error) {
	directory, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}
	for {
		if fileExists(filepath.Join(directory, "tools", "go.mod")) && fileExists(filepath.Join(directory, "private", "cms", "cms.v2", "go.mod")) {
			return directory, nil
		}
		parent := filepath.Dir(directory)
		if parent == directory {
			return "", fmt.Errorf("find repository directory from %s", directory)
		}
		directory = parent
	}
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func snapshotLinkType(snapshot *linkmirror.Snapshot) (*frontmatterparams.LinkType, error) {
	encoded, err := json.Marshal(snapshot)
	if err != nil {
		return nil, fmt.Errorf("marshal link snapshot: %w", err)
	}
	linkType := &frontmatterparams.LinkType{}
	if err := json.Unmarshal(encoded, linkType); err != nil {
		return nil, fmt.Errorf("unmarshal link snapshot: %w", err)
	}
	return linkType, nil
}

func writeFileAtomically(targetPath string, content []byte) error {
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}
	temporary, err := os.CreateTemp(filepath.Dir(targetPath), ".linkmirror-*")
	if err != nil {
		return fmt.Errorf("create temporary output: %w", err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if _, err := temporary.Write(content); err != nil {
		temporary.Close()
		return fmt.Errorf("write temporary output: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close temporary output: %w", err)
	}
	if err := os.Rename(temporaryPath, targetPath); err != nil {
		return fmt.Errorf("publish output: %w", err)
	}
	return nil
}
