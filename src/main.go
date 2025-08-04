package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	var filePath string

	if len(os.Args) > 1 {
		filePath = os.Args[1]
	} else {
		fallbackFiles := []string{"index.md", "_index.md", "readme.md"}
		for _, fname := range fallbackFiles {
			if _, err := os.Stat(fname); err == nil {
				filePath = fname
				break
			}
		}
		if filePath == "" {
			log.Fatalf("No markdown file specified or found in current directory.")
		}
	}

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		log.Fatalf("Failed to resolve absolute path: %v.", err)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		log.Fatalf("Failed to stat file: %v.", err)
	}

	if info.IsDir() {
		log.Fatalf("'%s' is a directory.", absPath)
	}

	if !strings.HasSuffix(strings.ToLower(absPath), ".md") {
		log.Fatalf("'%s' is not a markdown file.", absPath)
	}

	LaunchViewer(absPath)
}
