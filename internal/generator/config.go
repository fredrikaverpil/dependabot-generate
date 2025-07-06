package generator

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

func GenerateDependabotConfig(
	scanPath string,
	directories []string,
	interval string,
	ecosystemMap []EcosystemMapEntry,
) (string, error) {
	ecosystemDirs := make(map[string][]string)

	for _, dir := range directories {
		// The 'dir' is relative to the original scanPath. We need to join them
		// to get the correct absolute path for os.ReadDir to work.
		absoluteDir := filepath.Join(scanPath, dir)

		detected, err := DetectPackageEcosystems(absoluteDir, ecosystemMap)
		if err != nil {
			return "", err
		}
		for _, eco := range detected {
			ecosystemDirs[eco] = append(ecosystemDirs[eco], dir)
		}
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(`version: 2
updates:
  - package-ecosystem: "github-actions"
    directories: ["/", ".github/actions/*/*.yml", ".github/actions/*/*.yaml", "action.yml", "action.yaml", "actions/*/*.yml", "actions/*/*.yaml"]
    schedule:
      interval: "%s"
    groups:
      github-actions:
        patterns: ["*"]
        update-types: ["minor", "patch"]
    labels:
      - "dependencies"
`, interval))

	// Sort ecosystems for deterministic output
	var sortedEcosystems []string
	for eco := range ecosystemDirs {
		sortedEcosystems = append(sortedEcosystems, eco)
	}
	sort.Strings(sortedEcosystems)

	for _, eco := range sortedEcosystems {
		dirs := ecosystemDirs[eco]
		// Deduplicate and sort
		dirSet := make(map[string]struct{})
		for _, d := range dirs {
			dirSet[d] = struct{}{}
		}
		uniqueDirs := make([]string, 0, len(dirSet))
		for d := range dirSet {
			uniqueDirs = append(uniqueDirs, d)
		}
		sort.Strings(uniqueDirs)

		sb.WriteString(fmt.Sprintf(`
  - package-ecosystem: "%s"
    directories: ["%s"]
    schedule:
      interval: "%s"
    allow:
      - dependency-type: all
    groups:
      %s:
        patterns: ["*"]
        update-types: ["minor", "patch"]
    labels:
      - "dependencies"
`, eco, strings.Join(uniqueDirs, `", "`), interval, eco))
	}

	return sb.String(), nil
}
