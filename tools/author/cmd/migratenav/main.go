package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"go.yaml.in/yaml/v4"
)

var frontmatterPrefix = []byte("---\n")
var frontmatterSuffix = []byte("\n---\n")

// navKeyPrefix maps a legacy `params.nav.*` key to the topic key prefix it
// migrates to. Keys absent from this map are dropped.
var navKeyPrefix = map[string]string{
	"place":     "places/",
	"placePark": "places/",
	"tag":       "",
	"topic":     "",
}

var navKeyDrop = map[string]bool{
	"type": true,
}

func main() {
	if len(os.Args) != 2 {
		log.Fatal("Usage: go run ./author/cmd/migratenav <content-directory>")
	}

	var migrated int

	err := filepath.WalkDir(os.Args[1], func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		} else if d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		changed, err := migrateFile(path)
		if err != nil {
			return fmt.Errorf("%s: %v", path, err)
		} else if changed {
			migrated++

			fmt.Println(path)
		}

		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("migrated %d file(s)", migrated)
}

func migrateFile(path string) (bool, error) {
	buf, err := os.ReadFile(path)
	if err != nil {
		return false, fmt.Errorf("read: %v", err)
	}

	if !bytes.HasPrefix(buf, frontmatterPrefix) {
		return false, nil
	}

	end := bytes.Index(buf[len(frontmatterPrefix):], frontmatterSuffix)
	if end < 0 {
		return false, nil
	}

	frontmatterBytes := buf[len(frontmatterPrefix) : len(frontmatterPrefix)+end+1]
	body := buf[len(frontmatterPrefix)+end+len(frontmatterSuffix):]

	var frontmatter map[string]any

	err = yaml.Unmarshal(frontmatterBytes, &frontmatter)
	if err != nil {
		return false, fmt.Errorf("unmarshal frontmatter: %v", err)
	}

	changed, err := migrateFrontmatter(frontmatter)
	if err != nil {
		return false, err
	} else if !changed {
		// return false, nil
	}

	rewritten := &bytes.Buffer{}

	_, err = rewritten.Write(frontmatterPrefix)
	if err != nil {
		return false, fmt.Errorf("write separator: %v", err)
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
		return false, fmt.Errorf("create yaml dumper: %v", err)
	}

	err = dumper.Dump(frontmatter)
	if err != nil {
		return false, fmt.Errorf("dump frontmatter: %v", err)
	}

	_, err = rewritten.Write(frontmatterPrefix)
	if err != nil {
		return false, fmt.Errorf("write separator: %v", err)
	}

	_, err = rewritten.Write(body)
	if err != nil {
		return false, fmt.Errorf("write body: %v", err)
	}

	err = os.WriteFile(path, rewritten.Bytes(), 0644)
	if err != nil {
		return false, fmt.Errorf("write: %v", err)
	}

	return true, nil
}

func migrateFrontmatter(frontmatter map[string]any) (bool, error) {
	params, ok := frontmatter["params"].(map[string]any)
	if !ok {
		return false, nil
	}

	navRaw, ok := params["nav"]
	if !ok {
		return false, nil
	}

	delete(params, "nav")

	topics := map[string]any{}

	if existing, ok := params["topics"].(map[string]any); ok {
		for k, v := range existing {
			topics[k] = v
		}
	}

	if nav, ok := navRaw.(map[string]any); ok {
		for navKey, navValue := range nav {
			if navKeyDrop[navKey] {
				continue
			}

			prefix, known := navKeyPrefix[navKey]
			if !known {
				return false, fmt.Errorf("unsupported nav key: %s", navKey)
			}

			navValueMap, ok := navValue.(map[string]any)
			if !ok {
				return false, fmt.Errorf("unsupported nav value for key: %s", navKey)
			}

			for _, topic := range sortedKeys(navValueMap) {
				topicKey := prefix + topic

				if _, ok := topics[topicKey]; !ok {
					topics[topicKey] = map[string]any{}
				}
			}
		}
	} else if navRaw != nil {
		return false, fmt.Errorf("unsupported nav value")
	}

	if len(topics) > 0 {
		params["topics"] = topics
	} else {
		delete(params, "topics")
	}

	if len(params) == 0 {
		delete(frontmatter, "params")
	}

	return true, nil
}

func sortedKeys(v map[string]any) []string {
	keys := make([]string, 0, len(v))

	for k := range v {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	return keys
}
