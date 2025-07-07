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
	rootPath       string
	updateInterval string
	outputPath     string
	excludePaths   []string
	customMapJSON  string
}

func run(cfg config) error {
	log.Printf(
		"Starting dependabot generation with root_path: '%s', update_interval: '%s', output_path: '%s'",
		cfg.rootPath,
		cfg.updateInterval,
		cfg.outputPath,
	)

	ecosystemMap, err := generator.GetEcosystemMap(cfg.customMapJSON)
	if err != nil {
		return fmt.Errorf("error getting ecosystem map: %w", err)
	}

	log.Printf("Scanning for directories with dependency files in '%s'", cfg.rootPath)
	dirs, err := generator.RecursivelyScanDirectories(cfg.rootPath, cfg.excludePaths, ecosystemMap)
	if err != nil {
		return fmt.Errorf("error scanning directories: %w", err)
	}
	log.Printf("Found %d directories with dependency files: %v", len(dirs), dirs)

	log.Println("Generating dependabot configuration")
	configContent, err := generator.GenerateDependabotConfig(cfg.rootPath, dirs, cfg.updateInterval, ecosystemMap)
	if err != nil {
		return fmt.Errorf("error generating config: %w", err)
	}

	outputDir := filepath.Dir(cfg.outputPath)
	//nolint:gosec // The permissions 0o755 are standard for directories and necessary for CI/CD environments.
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("error creating output directory '%s': %w", outputDir, err)
	}

	log.Printf("Writing dependabot configuration to '%s'", cfg.outputPath)
	//nolint:gosec // The permissions 0o644 are standard for non-executable files and necessary for CI/CD environments.
	if err := os.WriteFile(cfg.outputPath, []byte(configContent), 0o644); err != nil {
		return fmt.Errorf("error writing output file: %w", err)
	}

	log.Printf("Dependabot configuration generated at '%s'", cfg.outputPath)
	return nil
}

func main() {
	rootPath := flag.String("root-path", ".", "Recursively scan this path for dependency files")
	updateInterval := flag.String("update-interval", "weekly", "Update interval for dependencies")
	outputPath := flag.String("output-path", ".github/dependabot.yml", "Output file path")
	excludePathsStr := flag.String(
		"exclude-paths",
		".venv,node_modules",
		"Comma-separated string of directories to ignore",
	)
	customMapJSON := flag.String("custom-map", "", "JSON string to extend the default ecosystem map")
	flag.Parse()

	var excludePaths []string
	if *excludePathsStr != "" {
		excludePaths = strings.Split(*excludePathsStr, ",")
		for i, dir := range excludePaths {
			excludePaths[i] = strings.TrimSpace(dir)
		}
	}

	cfg := config{
		rootPath:       *rootPath,
		updateInterval: *updateInterval,
		outputPath:     *outputPath,
		excludePaths:   excludePaths,
		customMapJSON:  *customMapJSON,
	}

	if err := run(cfg); err != nil {
		log.Fatalf("Application failed: %v", err)
	}
}
