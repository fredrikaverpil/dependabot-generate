package generator

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

func TestGetEcosystemMap(t *testing.T) {
	t.Parallel()
	// 1. Test with no custom map
	t.Run("no custom map", func(t *testing.T) {
		t.Parallel()
		defaultMap, err := GetEcosystemMap("")
		if err != nil {
			t.Fatalf("Expected no error, but got %v", err)
		}
		if len(defaultMap) == 0 {
			t.Fatal("Expected default map, but got empty map")
		}
	})

	// 2. Test with a valid custom map
	t.Run("valid custom map", func(t *testing.T) {
		t.Parallel()
		customJSON := `[{"ecosystem": "test-eco", "patterns": ["test.file"]}]`
		mergedMap, err := GetEcosystemMap(customJSON)
		if err != nil {
			t.Fatalf("Expected no error, but got %v", err)
		}

		defaultMap, _ := GetEcosystemMap("")
		if len(mergedMap) <= len(defaultMap) {
			t.Fatal("Expected merged map to be longer than default map")
		}

		firstEntry := mergedMap[0]
		if firstEntry.Ecosystem != "test-eco" {
			t.Errorf("Expected first entry to be the custom ecosystem, but got %s", firstEntry.Ecosystem)
		}
	})

	// 3. Test with a malformed custom map
	t.Run("malformed custom map", func(t *testing.T) {
		t.Parallel()
		malformedJSON := `[{"ecosystem": "test-eco", "patterns": ["test.file"]` // Missing closing bracket
		_, err := GetEcosystemMap(malformedJSON)
		if err == nil {
			t.Fatal("Expected an error for malformed JSON, but got nil")
		}
	})
}

func TestDetectPackageEcosystems(t *testing.T) {
	t.Parallel()
	// Helper to create a temporary directory with files
	createTempDirWithFiles := func(t *testing.T, files map[string]string) string {
		t.Helper()
		dir := t.TempDir()
		for name, content := range files {
			if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
				t.Fatalf("Failed to write file %s: %v", name, err)
			}
		}
		return dir
	}

	ecosystemMap, _ := GetEcosystemMap("") // Use default map for tests

	testCases := []struct {
		name               string
		files              map[string]string
		expectedEcosystems []string
	}{
		{
			name:               "simple go.mod",
			files:              map[string]string{"go.mod": "module my-project"},
			expectedEcosystems: []string{"gomod"},
		},
		{
			name:               "poetry project",
			files:              map[string]string{"poetry.lock": "", "pyproject.toml": ""},
			expectedEcosystems: []string{"pip"},
		},
		{
			name:               "uv project",
			files:              map[string]string{"uv.lock": "", "pyproject.toml": ""},
			expectedEcosystems: []string{"uv"},
		},
		{
			name:               "multiple ecosystems",
			files:              map[string]string{"go.mod": "", "Dockerfile": ""},
			expectedEcosystems: []string{"docker", "gomod"},
		},
		{
			name:               "no match",
			files:              map[string]string{"README.md": "# My Project"},
			expectedEcosystems: []string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dir := createTempDirWithFiles(t, tc.files)
			detected, err := DetectPackageEcosystems(dir, ecosystemMap)
			if err != nil {
				t.Fatalf("Expected no error, but got %v", err)
			}

			sort.Strings(detected)
			sort.Strings(tc.expectedEcosystems)

			// Handle the nil vs. empty slice case
			if len(detected) == 0 && len(tc.expectedEcosystems) == 0 {
				return // Test passes
			}

			if !reflect.DeepEqual(detected, tc.expectedEcosystems) {
				t.Errorf("Expected ecosystems %v, but got %v", tc.expectedEcosystems, detected)
			}
		})
	}
}

func TestGenerateDependabotConfig(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name        string
		directories []string
		files       map[string]string
		goldenFile  string
	}{
		{
			name:        "single project",
			directories: []string{"."},
			files: map[string]string{
				"go.mod":     "module my-project",
				"Dockerfile": "FROM golang",
			},
			goldenFile: "single_project.golden.yml",
		},
		{
			name:        "monorepo",
			directories: []string{".", "yolo"},
			files: map[string]string{
				"go.mod":      "module my-project",
				"Dockerfile":  "FROM golang",
				"yolo/go.mod": "module yolo",
			},
			goldenFile: "monorepo.golden.yml",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ecosystemMap, _ := GetEcosystemMap("")

			// Create a temporary directory structure for the test
			rootDir := t.TempDir()
			for name, content := range tc.files {
				filePath := filepath.Join(rootDir, name)
				if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
					t.Fatal(err)
				}
			}

			config, err := GenerateDependabotConfig(rootDir, tc.directories, "daily", ecosystemMap)
			if err != nil {
				t.Fatalf("Expected no error, but got %v", err)
			}

			goldenPath := filepath.Join("testdata", tc.goldenFile)
			expected, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatalf("Failed to read golden file %s: %v", goldenPath, err)
			}

			if config != string(expected) {
				t.Errorf(
					"Generated config does not match golden file.\nGot:\n%s\n\nExpected:\n%s",
					config,
					string(expected),
				)
			}
		})
	}
}
