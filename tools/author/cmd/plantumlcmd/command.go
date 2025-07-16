package main

// Hello, agent. Review this comment block and implement a concise solution. Retain all original comment instructions in the final code.
//
// This command line tool must traverse a source directory to find any directories which contain an `index.*` file.
// When a matching file is found, iterate through the index file's sibling directories for any matches to `*-plantuml`.
// For each matching plantuml subdirectory:
// 1. read `input.txt` file
// 2. find the `@startuml` tag, and insert the following lines after it:
//    []string{
//    	"skinparam backgroundColor transparent",
//    	"skinparam hyperlinkColor #0c4a6e",
//    	"skinparam hyperlinkUnderline false", // doesn't affect footer links; bug?
//    }
// 2. use the `EncodePlantUML` function from `encode.go`
// 3. perform an HTTP GET request to `{PLANTUML_SERVER}/svg/{encodedInput}`
// 4. save the response to a file named `output.svg` in the same directory
// 3. perform another HTTP GET request to `{PLANTUML_SERVER}/png/{encodedInput}`
// 4. save the response to a file named `output.png` in the same directory
// 5. log the output file directory
//
// The PLANTUML_SERVER should be provided via command line argument.
// The source directory should also be provided via command line argument.

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) < 3 {
		log.Fatal("Usage: plantumlcmd <PLANTUML_SERVER> <SOURCE_DIRECTORY>")
	}

	plantumlServer := os.Args[1]
	sourceDir := os.Args[2]

	err := processDirectory(plantumlServer, sourceDir)
	if err != nil {
		log.Fatal(err)
	}
}

// processDirectory traverses the source directory to find directories with index.* files
func processDirectory(plantumlServer, sourceDir string) error {
	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Check if this is an index.* file
		if !info.IsDir() && strings.HasPrefix(info.Name(), "index.") {
			// Get the parent directory of the index file
			parentDir := filepath.Dir(path)

			// Look for *-plantuml subdirectories in the parent directory
			entries, err := os.ReadDir(parentDir)
			if err != nil {
				return err
			}

			for _, entry := range entries {
				if entry.IsDir() && strings.HasSuffix(entry.Name(), "-plantuml") {
					plantumlDir := filepath.Join(parentDir, entry.Name())
					err := processPlantumlDirectory(plantumlServer, plantumlDir)
					if err != nil {
						log.Printf("Error processing %s: %v", plantumlDir, err)
					}
				}
			}
		}
		return nil
	})
}

// processPlantumlDirectory processes a single *-plantuml directory
func processPlantumlDirectory(plantumlServer, plantumlDir string) error {
	inputFile := filepath.Join(plantumlDir, "input.txt")

	// Read the input.txt file
	content, err := os.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("failed to read %s: %v", inputFile, err)
	}

	// Modify the content by inserting skinparam lines after @startuml
	modifiedContent := insertSkinparams(string(content))

	// Encode the modified content
	encodedInput := EncodePlantUML(modifiedContent)

	// Generate SVG
	svgURL := fmt.Sprintf("%s/svg/%s", plantumlServer, encodedInput)
	err = downloadAndSave(svgURL, filepath.Join(plantumlDir, "output.svg"))
	if err != nil {
		return fmt.Errorf("failed to download SVG: %v", err)
	}

	// Generate PNG
	pngURL := fmt.Sprintf("%s/png/%s", plantumlServer, encodedInput)
	err = downloadAndSave(pngURL, filepath.Join(plantumlDir, "output.png"))
	if err != nil {
		return fmt.Errorf("failed to download PNG: %v", err)
	}

	// Log the output file directory
	log.Printf("Generated PlantUML outputs in: %s", plantumlDir)

	return nil
}

// insertSkinparams inserts skinparam lines after @startuml tag
func insertSkinparams(content string) string {
	skinparams := []string{
		"skinparam backgroundColor transparent",
		"skinparam hyperlinkColor #0c4a6e",
		"skinparam hyperlinkUnderline false", // doesn't affect footer links; bug?
	}

	lines := strings.Split(content, "\n")
	var result []string
	for _, line := range lines {
		result = append(result, line)

		// Check if this line contains @startuml
		if strings.Contains(line, "@startuml") {
			// Insert skinparam lines after @startuml
			for _, skinparam := range skinparams {
				result = append(result, skinparam)
			}
		}
	}

	return strings.Join(result, "\n")
}

// downloadAndSave downloads content from a URL and saves it to a file
func downloadAndSave(url, filePath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP request failed with status: %s", resp.Status)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	return err
}
