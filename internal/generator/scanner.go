// Package generator provides the core logic for scanning directories and generating Dependabot configuration.
package generator

import (
	"fmt"
	"io/fs"
	"log" //nolint:depguard // No need for slog just yet.
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// RecursivelyScanDirectories walks a directory tree from a given root path,
// identifies directories containing recognizable package ecosystems, and returns
// a sorted list of their relative paths. It skips any directories specified
// in ignoreDirs.
func RecursivelyScanDirectories(root string, ignoreDirs []string, ecosystemMap []EcosystemMapEntry) ([]string, error) {
	directoriesWithDeps := make(map[string]struct{})

	walkFunc := func(path string, d fs.DirEntry, _ error) error {
		return processDirectoryEntry(path, d, root, ignoreDirs, ecosystemMap, directoriesWithDeps)
	}

	if err := filepath.WalkDir(root, walkFunc); err != nil {
		return nil, fmt.Errorf("error walking directories: %w", err)
	}

	var result []string
	for dir := range directoriesWithDeps {
		result = append(result, dir)
	}
	sort.Strings(result)
	return result, nil
}

// processDirectoryEntry is a helper function for filepath.WalkDir. It processes
// a single directory entry, checking for package ecosystems and adding them to
// the directoriesWithDeps map if found.
func processDirectoryEntry(
	path string,
	d fs.DirEntry,
	root string,
	ignoreDirs []string,
	ecosystemMap []EcosystemMapEntry,
	directoriesWithDeps map[string]struct{},
) error {
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
}

// DetectPackageEcosystems scans a single directory to identify all package
// ecosystems present, based on a provided map of detection rules. It returns a
// sorted list of all unique ecosystems found.
func DetectPackageEcosystems(directory string, ecosystemMap []EcosystemMapEntry) ([]string, error) {
	filesInDir, err := getFilesInDir(directory)
	if err != nil {
		return nil, fmt.Errorf("could not read directory %s: %w", directory, err)
	}

	foundEcosystems := make(map[string]struct{})
	for _, entry := range ecosystemMap {
		var matched bool
		var err error

		if len(entry.Heuristics) > 0 {
			matched, err = checkHeuristics(filesInDir, entry.Heuristics)
			if err != nil {
				return nil, err
			}
			if matched {
				log.Printf("Detected %s in %s via heuristic", entry.Ecosystem, directory)
				foundEcosystems[entry.Ecosystem] = struct{}{}
				continue // First match wins
			}
		}

		if len(entry.Patterns) > 0 {
			matched, err = anyFileMatches(filesInDir, entry.Patterns...)
			if err != nil {
				return nil, err
			}
			if matched {
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

// getFilesInDir reads a directory and returns a slice of the names of the files it contains.
func getFilesInDir(directory string) ([]string, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, err
	}

	filesInDir := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			filesInDir = append(filesInDir, entry.Name())
		}
	}
	return filesInDir, nil
}

// checkHeuristics evaluates a set of heuristic rules against the files in a
// directory. A rule matches if all `Present` patterns are found and no `Absent`
// patterns are found. It returns true on the first rule that matches.
func checkHeuristics(filesInDir []string, rules []Heuristic) (bool, error) {
	for _, rule := range rules {
		presentMatch := true
		for _, p := range rule.Present {
			match, err := anyFileMatches(filesInDir, p)
			if err != nil {
				return false, err
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
				return false, err
			}
			if match {
				absentMatch = false
			}
		}

		if presentMatch && absentMatch {
			return true, nil
		}
	}
	return false, nil
}

// anyFileMatches checks if any of the provided files match any of the given glob patterns.
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
