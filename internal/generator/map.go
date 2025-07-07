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

var defaultEcosystemMapJSON = `[
		{
			"ecosystem": "uv",
			"heuristics": [
				{"present": ["uv.lock"]}
			]
		},
		{
			"ecosystem": "pip",
			"heuristics": [
				{"present": ["poetry.lock", "pyproject.toml"]},
				{"present": ["Pipfile.lock"]},
				{"present": ["Pipfile"]},
				{"present": ["requirements.txt"]},
				{"present": ["pyproject.toml"], "absent": ["uv.lock"]}
			]
		},
		{"ecosystem": "gomod", "patterns": ["go.mod"]},
		{"ecosystem": "npm", "patterns": ["package.json"]},
		{"ecosystem": "docker", "patterns": ["Dockerfile"]},
		{"ecosystem": "bundler", "patterns": ["Gemfile"]},
		{"ecosystem": "composer", "patterns": ["composer.json"]},
		{"ecosystem": "cargo", "patterns": ["Cargo.toml"]},
		{"ecosystem": "nuget", "patterns": ["*.csproj", "packages.config"]},
		{"ecosystem": "mix", "patterns": ["mix.exs"]},
		{"ecosystem": "elm", "patterns": ["elm.json"]},
		{"ecosystem": "gradle", "patterns": ["build.gradle", "build.gradle.kts"]},
		{"ecosystem": "maven", "patterns": ["pom.xml"]},
		{"ecosystem": "pub", "patterns": ["pubspec.yaml"]},
		{"ecosystem": "swift", "patterns": ["Package.swift"]},
		{"ecosystem": "terraform", "patterns": ["*.tf", "*.tf.json"]},
		{"ecosystem": "devcontainers", "patterns": ["devcontainer.json"]},
		{"ecosystem": "gitsubmodule", "patterns": [".gitmodules"]}
	]`

// --- Core Logic ---

func GetEcosystemMap(customMapJSON string) ([]EcosystemMapEntry, error) {
	var defaultMap []EcosystemMapEntry
	if err := json.Unmarshal([]byte(defaultEcosystemMapJSON), &defaultMap); err != nil {
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
