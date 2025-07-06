# dependabot-generate

This action generates the `dependabot.yml` file based on the detected package
ecosystems in the repository.

Using
[`peter-evans/crate-pull-request`](https://github.com/peter-evans/create-pull-request),
you can have a PR created once the `dependabot.yml` changes.

## Features

The action comes with some sane defaults, like:

- Grouping of directories, for the same eco-system.
- Grouping of minor and patch level version bumping.
- A label `dependencies` is added to dependabot PRs.

## Setup

Place this workflow in e.g. `.github/workflows/dependabot-generate.yml`.

```yaml
name: Generate Dependabot Config

on:
  push:
    branches:
      - main
  schedule:
    - cron: "0 9 * * 1"

jobs:
  generate-dependabot:
    runs-on: ubuntu-latest
    permissions:
      contents: write
      pull-requests: write
    steps:
      - name: Check out repo
        uses: actions/checkout@v4
      - name: Generate Dependabot Config
        uses: fredrikaverpil/dependabot-generate@v1
      - name: Create Pull Request
        uses: peter-evans/create-pull-request@v7
        with:
          token: ${{ secrets.GITHUB_TOKEN }}
          commit-message: "chore(dependabot): generate dependabot.yml"
          title: "chore(dependabot): generate dependabot.yml"
          body: "This PR adds the generated dependabot.yml file."
          branch: "chore/dependabot-generate"
          delete-branch: true
```

## Inputs

You can customize the following inputs.

| Input       | Description                            | Default  | Required |
| ----------- | -------------------------------------- | -------- | -------- |
| `scan-path` | The path to scan for dependency files. | `.`      | No       |
| `interval`  | The update interval for dependencies.  | `weekly` | No       |
| `ignore-dirs` | A comma-separated string of relative paths to ignore. | `''` | No |
| `custom-map` | JSON string to extend the default ecosystem map. | `''` | No |

### Custom Ecosystem Map

You can extend and override the default ecosystem detection by providing a custom map. This is useful for proprietary package managers or for resolving conflicts between tools that use the same file names (like `pyproject.toml`).

The action uses a "first match wins" strategy. Your custom rules are checked first, giving them the highest priority.

The input must be a JSON string. Each entry can define an ecosystem using simple `patterns` (glob support) or more advanced `heuristics`.

**Heuristic Rules:**
- `present`: A list of glob patterns that must all be found in a directory.
- `absent`: An optional list of glob patterns that must *not* be found.

**Example:**

This example adds a new rule for a custom build system and overrides the default `pip` behavior to prefer a `requirements.in` file.

```yaml
- name: Generate Dependabot Config
  uses: fredrikaverpil/dependabot-generate@v1
  with:
    custom-map: |
      [
        {
          "ecosystem": "my-custom-build",
          "patterns": ["build.special", "*.custom"]
        },
        {
          "ecosystem": "pip",
          "heuristics": [
            {
              "present": ["requirements.in"]
            }
          ]
        }
      ]
```
