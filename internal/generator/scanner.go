package generator

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func DetectPackageEcosystems(directory string, ecosystemMap []EcosystemMapEntry) ([]string, error) {
	foundEcosystems := make(map[string]struct{})

	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, fmt.Errorf("could not read directory %s: %w", directory, err)
	}

	filesInDir := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			filesInDir = append(filesInDir, entry.Name())
		}
	}

	for _, entry := range ecosystemMap {
		if len(entry.Heuristics) > 0 {
			for _, rule := range entry.Heuristics {
				presentMatch := true
				for _, p := range rule.Present {
					match, err := anyFileMatches(filesInDir, p)
					if err != nil {
						return nil, err
					}
					if !match {
						presentMatch = false
						break
					}
				}

				if !presentMatch {
					continue
				}

				absentMatch := true
				if len(rule.Absent) > 0 {
					match, err := anyFileMatches(filesInDir, rule.Absent...)
					if err != nil {
						return nil, err
					}
					if match {
						absentMatch = false
					}
				}

				if presentMatch && absentMatch {
					log.Printf("Detected %s in %s via heuristic: %+v", entry.Ecosystem, directory, rule)
					foundEcosystems[entry.Ecosystem] = struct{}{}
				}
			}
		} else if len(entry.Patterns) > 0 {
			match, err := anyFileMatches(filesInDir, entry.Patterns...)
			if err != nil {
				return nil, err
			}
			if match {
				log.Printf("Detected %s in %s via patterns", entry.Ecosystem, directory)
				foundEcosystems[entry.Ecosystem] = struct{}{}
			}
		}
	}

	var result []string
	for eco := range foundEcosystems {
		result = append(result, eco)
	}
	sort.Strings(result)
	return result, nil
}

func anyFileMatches(files []string, patterns ...string) (bool, error) {
	for _, pattern := range patterns {
		for _, file := range files {
			match, err := filepath.Match(pattern, file)
			if err != nil {
				return false, fmt.Errorf("invalid glob pattern '%s': %w", pattern, err)
			}
			if match {
				return true, nil
			}
		}
	}
	return false, nil
}

func RecursivelyScanDirectories(root string, ignoreDirs []string, ecosystemMap []EcosystemMapEntry) ([]string, error) {
	directoriesWithDeps := make(map[string]struct{})

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}

		for _, ignored := range ignoreDirs {
			if strings.Contains(path, ignored) {
				log.Printf("Skipping ignored directory: %s", path)
				return filepath.SkipDir
			}
		}

		ecosystems, err := DetectPackageEcosystems(path, ecosystemMap)
		if err != nil {
			log.Printf("Warning: could not detect ecosystems in %s: %v", path, err)
			return nil
		}

		if len(ecosystems) > 0 {
			relPath, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			if relPath == "." {
				relPath = "/"
			}
			directoriesWithDeps[relPath] = struct{}{}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking directories: %w", err)
	}

	var result []string
	for dir := range directoriesWithDeps {
		result = append(result, dir)
	}
	sort.Strings(result)
	return result, nil
}
