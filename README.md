# dependabot-generate

This action generates the `dependabot.yml` file based on the detected package
ecosystems in the repository.

Using
[`peter-evans/crate-pull-request`](https://github.com/peter-evans/create-pull-request),
you can have a PR created once the `dependabot.yml` changes.

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
          branch: "chore/dependabot"
          delete-branch: true
```

## Inputs

| Input       | Description                            | Default  | Required |
| ----------- | -------------------------------------- | -------- | -------- |
| `scan-path` | The path to scan for dependency files. | `.`      | No       |
| `interval`  | The update interval for dependencies.  | `weekly` | No       |
