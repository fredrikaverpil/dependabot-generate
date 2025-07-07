# dependabot-generate

> [!WARNING]
>
> This project is not stable yet and may change in backwards breaking ways at
> any time!

This action generates the `dependabot.yml` file based on the detected package
ecosystems in the repository.

## On-Demand Usage

You can run this tool directly from its GitHub repository without a local clone.
This is useful for one-off generation or for trying out the tool.

```bash
go run github.com/fredrikaverpil/dependabot-generate/cmd/dependabot-generate@latest
```

This will generate a `.github/dependabot.yml` file in your current directory.

---

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
  workflow_dispatch:
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
        uses: fredrikaverpil/dependabot-generate@main # not stable yet!
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

| Input             | Description                                               | Default  | Required |
| ----------------- | --------------------------------------------------------- | -------- | -------- |
| `root-path`       | The path to scan for dependency files.                    | `.`      | No       |
| `exclude-paths`   | A comma-separated string of relative paths to ignore.     | `''`     | No       |
| `update-interval` | The update interval for dependencies.                     | `weekly` | No       |
| `custom-map`      | JSON string to extend the default ecosystem map.          | `''`     | No       |
| `additional-yaml` | YAML string to append to the generated dependabot config. | `''`     | No       |

### Custom Ecosystem Map

You can extend and override the default ecosystem detection by providing a
custom map. This is really only useful when resolving conflicts between
ecosystems that use the same file names for detection.

The action uses a "first match wins" strategy. Your custom rules are checked
first, giving them the highest priority.

The input must be a JSON string. Each entry can define an ecosystem using simple
`patterns` (glob support) or more advanced `heuristics`.

**Heuristic Rules:**

- `present`: A list of glob patterns that must all be found in a directory.
- `absent`: An optional list of glob patterns that must _not_ be found.

**Example:**

This example shows how to resolve a conflict between `uv` and `pip`, which both
use `pyproject.toml`. The rules are evaluated in order. The first rule that
matches wins, preventing the second rule from being evaluated.

- The first rule checks for a `uv.lock` file. If found, the directory is
  correctly identified as a `uv` project.
- Only if the first rule does _not_ match will the second rule be evaluated. It
  looks for a `pyproject.toml` but _only_ if a `uv.lock` is absent, correctly
  identifying it as a `pip` project.

```yaml
- name: Generate Dependabot Config
  uses: fredrikaverpil/dependabot-generate@main # not yet stable!
  with:
    custom-map: |
      [
        {
          "ecosystem": "uv",
          "heuristics": [
            {
              "present": ["uv.lock", "pyproject.toml"]
            }
          ]
        },
        {
          "ecosystem": "pip",
          "heuristics": [
            {
              "present": ["pyproject.toml"],
              "absent": ["uv.lock"]
            }
          ]
        }
      ]
```

A list of package managers and which ecosystem should be used for each can be
seen in
[the dependabot docs](https://docs.github.com/en/code-security/dependabot/working-with-dependabot/dependabot-options-reference#package-ecosystem-).

> [!NOTE]
>
> The example is taken from how `dependabot-generate` works out of the box, so
> you don't have to take this into consideration. This is just using a concrete
> and realistic example to explain how the heuristics engine works.

### Additional YAML

For complex scenarios where the auto-detection is not sufficient, you can append
raw YAML to the generated `dependabot.yml` file. This is useful for adding
configurations for ecosystems that are not supported by the generator, or for
overriding the configuration for a specific directory.

A common use case is to exclude a directory from the scan and then provide a
custom configuration for it.

**Example:**

```yaml
- name: Generate Dependabot Config
  uses: fredrikaverpil/dependabot-generate@main # not yet stable!
  with:
    exclude-paths: ".tools/"
    additional-yaml: |
      - package-ecosystem: "gomod"
        directory: "/.tools"
        allow:
          - dependency-type: indirect
        schedule:
          interval: "monthly"
```

## Local Development

You can run the generator locally for testing purposes using `go run`. This
allows you to see the generated `dependabot.yml` without running the full GitHub
Action.

**Basic command:**

```bash
go run ./cmd/dependabot-generate
```

**With custom arguments:**

```bash
go run ./cmd/dependabot-generate --root-path="/path/to/your/project" --update-interval="daily"
```

### Releasing

Move tag forward to latest commit:

```sh
git tag -f v1 && git push origin v1 -f
```
