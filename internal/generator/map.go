package generator

import (
	"encoding/json"
	"fmt"
	"log" //nolint:depguard // No need for slog just yet.
)

// --- Type Definitions ---

type Heuristic struct {
	Present []string `json:"present"`
	Absent  []string `json:"absent,omitempty"`
}

type EcosystemMapEntry struct {
	Ecosystem  string      `json:"ecosystem"`
	Patterns   []string    `json:"patterns,omitempty"`
	Heuristics []Heuristic `json:"heuristics,omitempty"`
}

// --- Default Ecosystem Map ---

// getDefaultEcosystemMapJSON returns the default ecosystem map in JSON format.
//
// Each ecosystem entry can have either:
//   - `patterns`: An OR condition. Any file matching the glob pattern triggers detection.
//   - `heuristics` `present` list: An AND condition. All patterns in the list must be matched.
//
// List of ecosystems/package managers:
// https://docs.github.com/en/code-security/dependabot/working-with-dependabot/dependabot-options-reference#package-ecosystem-
func getDefaultEcosystemMapJSON() string {
	return `[
		{
			"ecosystem": "uv",
			"heuristics": [
				{"present": ["uv.lock"]}
			]
		},
		{
			"ecosystem": "pip",
			"heuristics": [
				{"present": ["pdm.lock", "pyproject.toml"]},
				{"present": ["poetry.lock", "pyproject.toml"]},
				{"present": ["Pipfile.lock"]},
				{"present": ["Pipfile"]},
				{"present": ["requirements.txt"]},
				{"present": ["requirements.in"]},
				{"present": ["pyproject.toml"], "absent": ["uv.lock"]}
			]
		},
		{"ecosystem": "bun", "patterns": ["bun.lockb"]},
		{"ecosystem": "bundler", "patterns": ["Gemfile"]},
		{"ecosystem": "cargo", "patterns": ["Cargo.toml"]},
		{"ecosystem": "composer", "patterns": ["composer.json"]},
		{"ecosystem": "devcontainers", "patterns": ["devcontainer.json"]},
		{"ecosystem": "docker-compose", "patterns": ["docker-compose.y?ml"]},
		{"ecosystem": "docker", "patterns": ["Dockerfile"]},
		{"ecosystem": "elm", "patterns": ["elm.json"]},
		{"ecosystem": "gitsubmodule", "patterns": [".gitmodules"]},
		{"ecosystem": "gomod", "patterns": ["go.mod"]},
		{"ecosystem": "gradle", "patterns": ["build.gradle", "build.gradle.kts"]},
		{"ecosystem": "helm", "patterns": ["Chart.yaml"]},
		{"ecosystem": "maven", "patterns": ["pom.xml"]},
		{"ecosystem": "mix", "patterns": ["mix.exs"]},
		{"ecosystem": "npm", "patterns": ["package.json"]},
		{"ecosystem": "nuget", "patterns": ["*.csproj", "*.vbproj", "*.fsproj", "packages.config", "global.json"]},
		{"ecosystem": "pub", "patterns": ["pubspec.yaml"]},
		{"ecosystem": "swift", "patterns": ["Package.swift"]},
		{"ecosystem": "terraform", "patterns": ["*.tf", "*.tf.json"]}
	]`
}

// --- Core Logic ---

func GetEcosystemMap(customMapJSON string) ([]EcosystemMapEntry, error) {
	var defaultMap []EcosystemMapEntry
	if err := json.Unmarshal([]byte(getDefaultEcosystemMapJSON()), &defaultMap); err != nil {
		return nil, fmt.Errorf("failed to parse default ecosystem map: %w", err)
	}

	if customMapJSON == "" {
		return defaultMap, nil
	}

	var customMap []EcosystemMapEntry
	if err := json.Unmarshal([]byte(customMapJSON), &customMap); err != nil {
		return nil, fmt.Errorf("failed to parse custom-map JSON: %w", err)
	}

	log.Printf("Successfully parsed custom ecosystem map, prepending to defaults: %+v", customMap)
	return append(customMap, defaultMap...), nil
}
