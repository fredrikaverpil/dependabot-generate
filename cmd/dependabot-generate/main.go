package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/fredrikaverpil/dependabot-generate/internal/generator"
)

func run() error {
	scanPath := flag.String("scan-path", ".", "Recursively scan this path for dependency files")
	interval := flag.String("interval", "weekly", "Update interval for dependencies")
	outputFilepath := flag.String("output-filepath", ".github/dependabot.yml", "Output file path")
	ignoreDirsStr := flag.String("ignore-dirs", ".venv,node_modules", "Comma-separated string of directories to ignore")
	customMapJSON := flag.String("custom-map", "", "JSON string to extend the default ecosystem map")
	flag.Parse()

	log.Printf("Starting dependabot generation with scan_path: '%s', interval: '%s', output_path: '%s'", *scanPath, *interval, *outputFilepath)

	var ignoreDirs []string
	if *ignoreDirsStr != "" {
		ignoreDirs = strings.Split(*ignoreDirsStr, ",")
		for i, dir := range ignoreDirs {
			ignoreDirs[i] = strings.TrimSpace(dir)
		}
	}

	ecosystemMap, err := generator.GetEcosystemMap(*customMapJSON)
	if err != nil {
		return fmt.Errorf("error getting ecosystem map: %w", err)
	}

	log.Printf("Scanning for directories with dependency files in '%s'", *scanPath)
	dirs, err := generator.RecursivelyScanDirectories(*scanPath, ignoreDirs, ecosystemMap)
	if err != nil {
		return fmt.Errorf("error scanning directories: %w", err)
	}
	log.Printf("Found %d directories with dependency files: %v", len(dirs), dirs)

	log.Println("Generating dependabot configuration")
	configContent, err := generator.GenerateDependabotConfig(*scanPath, dirs, *interval, ecosystemMap)
	if err != nil {
		return fmt.Errorf("error generating config: %w", err)
	}

	outputDir := filepath.Dir(*outputFilepath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("error creating output directory '%s': %w", outputDir, err)
	}

	log.Printf("Writing dependabot configuration to '%s'", *outputFilepath)
	if err := os.WriteFile(*outputFilepath, []byte(configContent), 0644); err != nil {
		return fmt.Errorf("error writing output file: %w", err)
	}

	log.Printf("Dependabot configuration generated at '%s'", *outputFilepath)
	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("Application failed: %v", err)
	}
}