package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"go.yaml.in/yaml/v4"
)

var frontmatterPrefix = []byte("---\n")
var frontmatterSuffix = []byte("\n---\n")

func main() {
	if len(os.Args) != 4 {
		log.Fatal("Usage: go run ./author/cmd/movefiles <content-directory> <target-directory> <uri-list-file>")
	}

	contentDir := os.Args[1]
	targetDir := os.Args[2]
	uriListFile := os.Args[3]

	uris, err := readLines(uriListFile)
	if err != nil {
		log.Fatalf("read uri list: %v", err)
	}

	err = os.MkdirAll(targetDir, 0755)
	if err != nil {
		log.Fatalf("create target directory: %v", err)
	}

	for _, uri := range uris {
		err = moveFile(contentDir, targetDir, uri)
		if err != nil {
			log.Fatalf("%s: %v", uri, err)
		}

		fmt.Println(uri)
	}
}

func readLines(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var lines []string

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		lines = append(lines, line)
	}

	return lines, scanner.Err()
}

func moveFile(contentDir, targetDir, uri string) error {
	name := filepath.Base(uri) + ".md"

	path, err := findFile(contentDir, name)
	if err != nil {
		return err
	}

	buf, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read: %v", err)
	}

	rewritten, err := injectURL(buf, uri)
	if err != nil {
		return err
	}

	newPath := filepath.Join(targetDir, filepath.Base(path))

	err = os.WriteFile(newPath, rewritten, 0644)
	if err != nil {
		return fmt.Errorf("write: %v", err)
	}

	err = os.Remove(path)
	if err != nil {
		return fmt.Errorf("remove original: %v", err)
	}

	return nil
}

func findFile(contentDir, name string) (string, error) {
	var matches []string

	err := filepath.WalkDir(contentDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		} else if d.IsDir() {
			return nil
		} else if filepath.Base(path) == name {
			matches = append(matches, path)
		}

		return nil
	})
	if err != nil {
		return "", err
	}

	if len(matches) == 0 {
		return "", fmt.Errorf("no file found matching %s", name)
	} else if len(matches) > 1 {
		return "", fmt.Errorf("multiple files found matching %s: %s", name, strings.Join(matches, ", "))
	}

	return matches[0], nil
}

func injectURL(buf []byte, url string) ([]byte, error) {
	if !bytes.HasPrefix(buf, frontmatterPrefix) {
		return nil, fmt.Errorf("missing frontmatter")
	}

	end := bytes.Index(buf[len(frontmatterPrefix):], frontmatterSuffix)
	if end < 0 {
		return nil, fmt.Errorf("missing frontmatter")
	}

	frontmatterBytes := buf[len(frontmatterPrefix) : len(frontmatterPrefix)+end+1]
	body := buf[len(frontmatterPrefix)+end+len(frontmatterSuffix):]

	var frontmatter map[string]any

	err := yaml.Unmarshal(frontmatterBytes, &frontmatter)
	if err != nil {
		return nil, fmt.Errorf("unmarshal frontmatter: %v", err)
	}

	if _, ok := frontmatter["url"]; ok {
		return nil, fmt.Errorf("frontmatter already has url")
	}

	frontmatter["url"] = url

	rewritten := &bytes.Buffer{}

	_, err = rewritten.Write(frontmatterPrefix)
	if err != nil {
		return nil, fmt.Errorf("write separator: %v", err)
	}

	// match the existing frontmatter conventions to keep diffs minimal
	dumper, err := yaml.NewDumper(
		rewritten,
		yaml.WithV4Defaults(),
		yaml.WithIndent(2),
		yaml.WithLineWidth(120),
		yaml.WithCompactSeqIndent(true),
	)
	if err != nil {
		return nil, fmt.Errorf("create yaml dumper: %v", err)
	}

	err = dumper.Dump(frontmatter)
	if err != nil {
		return nil, fmt.Errorf("dump frontmatter: %v", err)
	}

	_, err = rewritten.Write(frontmatterPrefix)
	if err != nil {
		return nil, fmt.Errorf("write separator: %v", err)
	}

	_, err = rewritten.Write(body)
	if err != nil {
		return nil, fmt.Errorf("write body: %v", err)
	}

	return rewritten.Bytes(), nil
}
