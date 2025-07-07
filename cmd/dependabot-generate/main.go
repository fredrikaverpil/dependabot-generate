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

type config struct {
	scanPath       string
	interval       string
	outputFilepath string
	ignoreDirs     []string
	customMapJSON  string
}

func run(cfg config) error {
	log.Printf(
		"Starting dependabot generation with scan_path: '%s', interval: '%s', output_path: '%s'",
		cfg.scanPath,
		cfg.interval,
		cfg.outputFilepath,
	)

	ecosystemMap, err := generator.GetEcosystemMap(cfg.customMapJSON)
	if err != nil {
		return fmt.Errorf("error getting ecosystem map: %w", err)
	}

	log.Printf("Scanning for directories with dependency files in '%s'", cfg.scanPath)
	dirs, err := generator.RecursivelyScanDirectories(cfg.scanPath, cfg.ignoreDirs, ecosystemMap)
	if err != nil {
		return fmt.Errorf("error scanning directories: %w", err)
	}
	log.Printf("Found %d directories with dependency files: %v", len(dirs), dirs)

	log.Println("Generating dependabot configuration")
	configContent, err := generator.GenerateDependabotConfig(cfg.scanPath, dirs, cfg.interval, ecosystemMap)
	if err != nil {
		return fmt.Errorf("error generating config: %w", err)
	}

	outputDir := filepath.Dir(cfg.outputFilepath)
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("error creating output directory '%s': %w", outputDir, err)
	}

	log.Printf("Writing dependabot configuration to '%s'", cfg.outputFilepath)
	if err := os.WriteFile(cfg.outputFilepath, []byte(configContent), 0o644); err != nil {
		return fmt.Errorf("error writing output file: %w", err)
	}

	log.Printf("Dependabot configuration generated at '%s'", cfg.outputFilepath)
	return nil
}

func main() {
	scanPath := flag.String("scan-path", ".", "Recursively scan this path for dependency files")
	interval := flag.String("interval", "weekly", "Update interval for dependencies")
	outputFilepath := flag.String("output-filepath", ".github/dependabot.yml", "Output file path")
	ignoreDirsStr := flag.String("ignore-dirs", ".venv,node_modules", "Comma-separated string of directories to ignore")
	customMapJSON := flag.String("custom-map", "", "JSON string to extend the default ecosystem map")
	flag.Parse()

	var ignoreDirs []string
	if *ignoreDirsStr != "" {
		ignoreDirs = strings.Split(*ignoreDirsStr, ",")
		for i, dir := range ignoreDirs {
			ignoreDirs[i] = strings.TrimSpace(dir)
		}
	}

	cfg := config{
		scanPath:       *scanPath,
		interval:       *interval,
		outputFilepath: *outputFilepath,
		ignoreDirs:     ignoreDirs,
		customMapJSON:  *customMapJSON,
	}

	if err := run(cfg); err != nil {
		log.Fatalf("Application failed: %v", err)
	}
}
