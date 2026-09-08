package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/dpb587/dpb587.me/tools/content"
)

func main() {
	if err := mainErr(); err != nil {
		log.Fatal(err)
	}
}

func mainErr() error {
	if len(os.Args) != 2 {
		return fmt.Errorf("usage: go run ./author/cmd/linkrehashcmd <target-directory>")
	}

	targetDir := os.Args[1]

	return filepath.WalkDir(targetDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		name := d.Name()
		if !strings.HasPrefix(name, "link-") || !strings.HasSuffix(name, ".md") {
			return nil
		}

		return rehashFile(path)
	})
}

func rehashFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open %s: %w", path, err)
	}
	var document content.Document
	_, err = document.ReadFrom(f)
	f.Close()
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}

	if document.Frontmatter == nil || document.Frontmatter.Params == nil || document.Frontmatter.Params.LinkType == nil {
		return fmt.Errorf("%s: missing params.linkType", path)
	}

	linkType := document.Frontmatter.Params.LinkType

	digest := sha256.Sum256([]byte(linkType.Referrer + "\n" + linkType.Target))
	newName := "link-" + hex.EncodeToString(digest[:])[:10] + ".md"
	if newName == filepath.Base(path) {
		return nil
	}

	newPath := filepath.Join(filepath.Dir(path), newName)
	if err := os.Rename(path, newPath); err != nil {
		return fmt.Errorf("rename %s: %w", path, err)
	}

	fmt.Println(newPath)

	return nil
}
