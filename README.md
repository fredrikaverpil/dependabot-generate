# dependabot-generate

Generate the `dependabot.yml` based on files detected in your GitHub repository
([example here](.github/dependabot.yml)).

> [!WARNING]
>
> This project is not stable yet and may change in backwards breaking ways at
> any time!

## CLI

You can run this tool directly from your GitHub repository. This is useful for
one-off generation or for trying out the tool.

```bash
go run github.com/fredrikaverpil/dependabot-generate/cmd/dependabot-generate@latest
```

This will generate a `.github/dependabot.yml` file in your current directory.

---

## Composite Action

Place this workflow in e.g. `.github/workflows/dependabot-generate.yml`, and a
PR will automatically be opened in your GitHub repository if new lockfiles have
been added.

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

## Sane defaults

- For each directory of the same eco-system, group their minor and patch level
  dependencies together into one PR.
- Each major-bumped dependency update will become its own PR.
- A label `dependencies` is added to dependabot PRs.

## Customizations

| Input             | Description                                               | Default  | Required |
| ----------------- | --------------------------------------------------------- | -------- | -------- |
| `root-path`       | The path to scan for dependency files.                    | `.`      | No       |
| `exclude-paths`   | A comma-separated string of relative paths to ignore.     | `''`     | No       |
| `update-interval` | The update interval for dependencies.                     | `weekly` | No       |
| `custom-map`      | JSON string to extend the default ecosystem map.          | `''`     | No       |
| `additional-yaml` | YAML string to append to the generated dependabot config. | `''`     | No       |

### Custom ecosystem logic

You can extend and override the default ecosystem detection by providing a
custom map. This is useful for resolving conflicts between package managers that
might use the same file names (e.g., `pyproject.toml`).

The action processes rules in a specific order, and this order is important.
Your custom rules are always checked first, giving them the highest priority.
The detection logic iterates through all rules, from your custom ones to the
built-in defaults. It does not stop after the first ecosystem is found. This
means a single directory can be associated with multiple ecosystems if it meets
the criteria for each (e.g., containing both a `Dockerfile` and a `go.mod`).

The "first match wins" principle applies when you have conflicting rules. If a
rule matches a set of files, it "claims" them, and subsequent rules can be set
up to ignore those files, preventing multiple ecosystems from being incorrectly
assigned to the same dependency definition file.

The input must be a JSON string. Each entry can define an ecosystem using simple
`patterns` (glob support) or more advanced `heuristics`.

**Heuristic Rules:**

- `present`: A list of glob patterns that must all be found in a directory.
- `absent`: An optional list of glob patterns that must _not_ be found.

**Example:**

This example shows how `generate-dependabot` resolves a conflict between `uv`
and `pip`, which both use `pyproject.toml`. The conflict lies within that we
don't want two ecosystem entries for one `pyproject.toml` file. The rules are
evaluated in order. The first rule that matches wins, preventing the second rule
from being evaluated.

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
      # custom updates
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
