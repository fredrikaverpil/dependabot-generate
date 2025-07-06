import argparse
import logging
import sys
from pathlib import Path



# Map of file indicators to ecosystems.
# https://docs.github.com/en/code-security/dependabot/ecosystems-supported-by-dependabot/supported-ecosystems-and-repositories
FILE_ECOSYSTEM_MAP: dict[str, str] = {
    # Python ecosystem detection
    "uv.lock": "uv",
    # Go ecosystem detection
    "go.mod": "gomod",  # Only need go.mod, not go.sum
    # Node ecosystem detection
    "package.json": "npm",  # Primary specification file
    # Docker ecosystem detection
    "Dockerfile": "docker",
    "docker-compose.yml": "docker-compose",
    "docker-compose.yaml": "docker-compose",
    # Ruby ecosystem detection
    "Gemfile": "bundler",
    # PHP ecosystem detection
    "composer.json": "composer",
    # Rust ecosystem detection
    "Cargo.toml": "cargo",  # Only need Cargo.toml, not Cargo.lock
    # .NET ecosystem detection
    "packages.config": "nuget",
    "global.json": "dotnet-sdk",
    "Directory.Packages.props": "nuget",
    # Elixir ecosystem detection
    "mix.exs": "mix",
    # Elm ecosystem detection
    "elm.json": "elm",
    # Gradle ecosystem detection
    "build.gradle": "gradle",
    "build.gradle.kts": "gradle",
    # Maven ecosystem detection
    "pom.xml": "maven",
    # Dart/Flutter ecosystem detection
    "pubspec.yaml": "pub",
    # Swift ecosystem detection
    "Package.swift": "swift",
    # Terraform ecosystem detection
    "main.tf": "terraform",
    # Dev containers detection
    "devcontainer.json": "devcontainers",
    ".devcontainer.json": "devcontainers",
    # Git submodule detection
    ".gitmodules": "gitsubmodule",
}


def detect_package_ecosystems(directory: Path) -> list[str]:
    """
    Detect all package ecosystems in a directory.
    Returns a list of detected ecosystem names.
    """
    # Set to track unique ecosystems
    found_ecosystems: set[str] = set()

    # Check for each file type
    for filename, ecosystem in FILE_ECOSYSTEM_MAP.items():
        if (directory / filename).exists():
            found_ecosystems.add(ecosystem)
            logging.info(
                f"Detected {ecosystem} ecosystem in {directory} via {filename}"
            )

    return list(found_ecosystems)


def recursively_scan_directories(root_dir: Path, ignore_dirs: list[str]) -> list[Path]:
    """
    Recursively scan directories for dependency files and return directories
    that contain at least one dependency file.
    """
    directories_with_deps: set[Path] = set()
    ignored_dirs_set = set(ignore_dirs)

    for dirpath_str, _, filenames in root_dir.walk():
        dirpath = Path(dirpath_str)
        # Skip ignored directories
        if any(ignored_dir in str(dirpath) for ignored_dir in ignored_dirs_set):
            logging.info(f"Skipping ignored directory: {dirpath}")
            continue

        if any(indicator in filenames for indicator in FILE_ECOSYSTEM_MAP.keys()):
            directories_with_deps.add(dirpath)

    return list(directories_with_deps)


def generate_dependabot_config(directories: list[Path], interval: str) -> str:
    """
    Generate dependabot.yml configuration based on detected project types
    """

    # Map directories to ecosystems
    ecosystem_dirs: dict[str, list[str]] = {}
    for directory in directories:
        ecosystems = detect_package_ecosystems(directory)
        for ecosystem in ecosystems:
            if ecosystem not in ecosystem_dirs:
                ecosystem_dirs[ecosystem] = []
            ecosystem_dirs[ecosystem].append(str(directory))

    # Build dependabot.yml content
    config = f"""version: 2
updates:
  - package-ecosystem: "github-actions"
    directories: ["/", ".github/actions/*/*.yml", ".github/actions/*/*.yaml", "action.yml", "action.yaml", "actions/*/*.yml", "actions/*/*.yaml"]
    schedule:
      interval: "{interval}"
    groups:
      github-actions:
        patterns: ["*"]
        update-types: ["minor", "patch"]
    labels:
      - "dependencies"
"""

    # Add ecosystem-specific configurations
    for ecosystem, dirs in ecosystem_dirs.items():
        if dirs == ["tools"]:
            # TODO: skip this for now, figure out later.
            # This should be its own entry, separate from production dependencies.
            continue

        if "tools" in dirs:
            dirs.remove("tools")  # do not mix dev tooling into production dependencies
        dir_entries = '["' + '", "'.join(dirs) + '"]'
        config += f"""
  - package-ecosystem: "{ecosystem}"
    directories: {dir_entries}
    schedule:
      interval: "{interval}"
    groups:
      {ecosystem}:
        patterns: ["*"]
        update-types: ["minor", "patch"]
    labels:
      - "dependencies"
"""

    return config


def main(scan_path: Path, interval: str, output_path: Path, ignore_dirs: list[str]) -> int:
    """
    Main function to generate dependabot.yml configuration.
    """
    logging.info(f"Starting dependabot generation with scan_path: '{scan_path}', interval: '{interval}', output_path: '{output_path}', ignore_dirs: {ignore_dirs}")

    # Process directories based on input method
    logging.info(f"Scanning for directories with dependency files in '{scan_path}'")
    dirs = recursively_scan_directories(root_dir=scan_path, ignore_dirs=ignore_dirs)
    logging.info(f"Found {len(dirs)} directories with dependency files: {[str(d) for d in dirs]}")

    # Generate dependabot configuration
    logging.info("Generating dependabot configuration")
    config_content = generate_dependabot_config(
        directories=dirs,
        interval=interval,
    )

    # Create output directory if it doesn't exist
    output_dir = output_path.parent
    if not output_dir.exists():
        logging.info(f"Creating output directory '{output_dir}'")
        output_dir.mkdir(parents=True)

    # Write configuration to file
    logging.info(f"Writing dependabot configuration to '{output_path}'")
    output_path.write_text(config_content)

    logging.info(f"Dependabot configuration generated at '{output_path}'")
    return 0


if __name__ == "__main__":
    logging.basicConfig(
        level=logging.INFO,
        format="%(asctime)s - %(levelname)s - %(message)s",
        stream=sys.stdout,
    )
    parser = argparse.ArgumentParser(
        description="Generate dependabot.yml configuration"
    )

    _ = parser.add_argument(
        "--scan-path",
        help="Recursively scan this path for dependency files (default: .)",
        type=str,
        default=".",
    )

    _ = parser.add_argument(
        "--interval",
        help="Update interval for dependencies (default: weekly)",
        type=str,
        default="weekly",
    )

    _ = parser.add_argument(
        "--output-filepath",
        help="Output file path (default: .github/dependabot.yml)",
        type=str,
        default=".github/dependabot.yml",
    )

    _ = parser.add_argument(
        "--ignore-dirs",
        help="Comma-separated list of directories to ignore.",
        type=str,
        default="",
    )

    args: argparse.Namespace = parser.parse_args()
    ignore_dirs = [d.strip() for d in args.ignore_dirs.split(',') if d.strip()]
    sys.exit(main(
        scan_path=Path(args.scan_path),
        interval=str(args.interval),
        output_path=Path(args.output_filepath),
        ignore_dirs=ignore_dirs,
    ))
