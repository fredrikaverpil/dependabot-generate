package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestE2E(t *testing.T) {
	testCases := []struct {
		name       string
		files      map[string]string
		goldenFile string
		cfg        config
	}{
		{
			name: "default",
			files: map[string]string{
				"go.mod":                          "module root-project",
				"project-a/uv.lock":               "",
				"project-a/pyproject.toml":        "",
				".venv/some-package/package.json": "{}",
			},
			goldenFile: "default.golden.yml",
			cfg: config{
				ignoreDirs: []string{".venv"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 1. Create a complex temporary directory structure
			rootDir := t.TempDir()
			t.Logf("Using temporary directory: %s", rootDir)

			for name, content := range tc.files {
				filePath := filepath.Join(rootDir, name)
				if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
					t.Fatalf("Failed to write file %s: %v", name, err)
				}
			}

			// 2. Set up arguments for the run
			outputFile := filepath.Join(rootDir, ".github", "dependabot.yml")
			tc.cfg.scanPath = rootDir
			tc.cfg.outputFilepath = outputFile
			if tc.cfg.interval == "" {
				tc.cfg.interval = "weekly"
			}
			if tc.cfg.ignoreDirs != nil {
				var ignored []string
				for _, dir := range tc.cfg.ignoreDirs {
					ignored = append(ignored, strings.TrimSpace(dir))
				}
				tc.cfg.ignoreDirs = ignored
			}

			// 3. Run the application logic
			if err := run(tc.cfg); err != nil {
				t.Fatalf("run() failed: %v", err)
			}

			// 4. Compare the output with the golden file
			generatedBytes, err := os.ReadFile(outputFile)
			if err != nil {
				t.Fatalf("Failed to read generated file: %v", err)
			}
			generatedContent := string(generatedBytes)

			goldenPath := filepath.Join("testdata", tc.goldenFile)
			expectedBytes, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatalf("Failed to read golden file: %v", err)
			}
			expectedContent := string(expectedBytes)

			if generatedContent != expectedContent {
				t.Errorf(
					"Generated config does not match golden file.\nGot:\n%s\n\nExpected:\n%s",
					generatedContent,
					expectedContent,
				)
			}
		})
	}
}
