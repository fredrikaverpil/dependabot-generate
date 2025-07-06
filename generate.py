import argparse
import fnmatch
import json
import logging
import sys
from pathlib import Path
from typing import TypedDict, cast

# --- Type Definitions ---


class Heuristic(TypedDict, total=False):
    """A rule for detecting an ecosystem."""

    present: list[str]  # Glob patterns that must be present
    absent: list[str]  # Glob patterns that must be absent


class SimpleEcosystem(TypedDict):
    """A simple ecosystem definition using basic patterns."""

    ecosystem: str
    patterns: list[str]


class HeuristicEcosystem(TypedDict):
    """An ecosystem definition using advanced heuristics."""

    ecosystem: str
    heuristics: list[Heuristic]


EcosystemMap = SimpleEcosystem | HeuristicEcosystem

DEFAULT_ECOSYSTEM_MAP: list[EcosystemMap] = [
    HeuristicEcosystem(
        ecosystem="uv",
        heuristics=[
            {"present": ["uv.lock"]},
        ],
    ),
    HeuristicEcosystem(
        ecosystem="pip",
        heuristics=[
            {"present": ["poetry.lock", "pyproject.toml"]},
            {"present": ["Pipfile.lock"]},
            {"present": ["Pipfile"]},
            {"present": ["requirements.txt"]},
            {"present": ["requirements.in"]},
            {"present": ["pyproject.toml"], "absent": ["uv.lock"]},
        ],
    ),
    SimpleEcosystem(ecosystem="gomod", patterns=["go.mod"]),
    SimpleEcosystem(ecosystem="npm", patterns=["package.json"]),
    SimpleEcosystem(ecosystem="docker", patterns=["Dockerfile"]),
    SimpleEcosystem(ecosystem="bundler", patterns=["Gemfile"]),
    SimpleEcosystem(ecosystem="composer", patterns=["composer.json"]),
    SimpleEcosystem(ecosystem="cargo", patterns=["Cargo.toml"]),
    SimpleEcosystem(ecosystem="nuget", patterns=["*.csproj", "packages.config"]),
    SimpleEcosystem(ecosystem="mix", patterns=["mix.exs"]),
    SimpleEcosystem(ecosystem="elm", patterns=["elm.json"]),
    SimpleEcosystem(ecosystem="gradle", patterns=["build.gradle", "build.gradle.kts"]),
    SimpleEcosystem(ecosystem="maven", patterns=["pom.xml"]),
    SimpleEcosystem(ecosystem="pub", patterns=["pubspec.yaml"]),
    SimpleEcosystem(ecosystem="swift", patterns=["Package.swift"]),
    SimpleEcosystem(ecosystem="terraform", patterns=["*.tf", "*.tf.json"]),
    SimpleEcosystem(ecosystem="devcontainers", patterns=["devcontainer.json"]),
    SimpleEcosystem(ecosystem="gitsubmodule", patterns=[".gitmodules"]),
]


def get_ecosystem_map(custom_map_json: str) -> list[EcosystemMap]:
    """
    Merges the default ecosystem map with a user-provided custom map.
    User-provided rules are given higher precedence.
    """
    if not custom_map_json:
        return DEFAULT_ECOSYSTEM_MAP
    try:
        custom_map = cast(list[EcosystemMap], json.loads(custom_map_json))
        logging.info(
            "Successfully parsed custom ecosystem map, prepending to defaults: %s",
            custom_map,
        )
        return custom_map + DEFAULT_ECOSYSTEM_MAP
    except json.JSONDecodeError:
        logging.error("Failed to parse custom-map JSON. Using default map only.")
        return DEFAULT_ECOSYSTEM_MAP


def detect_package_ecosystems(
    directory: Path, ecosystem_map: list[EcosystemMap]
) -> list[str]:
    """
    Detect all package ecosystems in a directory based on the provided ecosystem map.
    """
    found_ecosystems: set[str] = set()
    try:
        files_in_dir = [f.name for f in directory.iterdir() if f.is_file()]
    except FileNotFoundError:
        logging.warning(f"Directory not found, cannot detect ecosystems: {directory}")
        return []

    for entry in ecosystem_map:
        if "heuristics" in entry:
            heuristic_entry = cast(HeuristicEcosystem, entry)  # pyright: ignore[reportUnnecessaryCast]
            ecosystem = heuristic_entry["ecosystem"]
            # Advanced heuristic-based matching
            for rule in heuristic_entry["heuristics"]:
                present_patterns = rule.get("present", [])
                absent_patterns = rule.get("absent", [])

                present_match = all(
                    any(fnmatch.fnmatch(f, p) for f in files_in_dir)
                    for p in present_patterns
                )
                absent_match = not any(
                    any(fnmatch.fnmatch(f, p) for f in files_in_dir)
                    for p in absent_patterns
                )

                if present_match and absent_match:
                    logging.info(
                        f"Detected {ecosystem} in {directory} via heuristic: {rule}"
                    )
                    found_ecosystems.add(ecosystem)
        elif "patterns" in entry:
            ecosystem = entry["ecosystem"]
            # Simple pattern matching
            for pattern in entry["patterns"]:
                if any(fnmatch.fnmatch(f, pattern) for f in files_in_dir):
                    logging.info(
                        f"Detected {ecosystem} in {directory} via pattern '{pattern}'"
                    )
                    found_ecosystems.add(ecosystem)
    return list(found_ecosystems)


def recursively_scan_directories(
    root_dir: Path, ignore_dirs: list[str], ecosystem_map: list[EcosystemMap]
) -> list[Path]:
    """
    Recursively scan for directories that contain at least one dependency file.
    """
    directories_with_deps: set[Path] = set()
    ignored_dirs_set = set(ignore_dirs)

    # Combine root_dir with all its subdirectories for a complete scan
    all_dirs_to_scan = [root_dir] + list(root_dir.rglob("*"))

    for item in all_dirs_to_scan:
        if not item.is_dir():
            continue

        dirpath = item
        if dirpath in directories_with_deps:
            continue  # Already processed

        if any(ignored_dir in str(dirpath) for ignored_dir in ignored_dirs_set):
            continue

        if detect_package_ecosystems(dirpath, ecosystem_map):
            directories_with_deps.add(dirpath)

    return sorted(list(directories_with_deps))


def generate_dependabot_config(
    directories: list[Path], interval: str, ecosystem_map: list[EcosystemMap]
) -> str:
    """
    Generate dependabot.yml configuration based on detected project types.
    """
    ecosystem_dirs: dict[str, list[str]] = {}
    for directory in directories:
        detected_ecosystems = detect_package_ecosystems(directory, ecosystem_map)
        for ecosystem in detected_ecosystems:
            if ecosystem not in ecosystem_dirs:
                ecosystem_dirs[ecosystem] = []
            ecosystem_dirs[ecosystem].append(str(directory))

    config = f"""version: 2
updates:
  - package-ecosystem: "github-actions"
    directory: "/"
    schedule:
      interval: "{interval}"
    groups:
      github-actions:
        patterns: ["*"]
        update-types: ["minor", "patch"]
    labels:
      - "dependencies"
"""

    for ecosystem, dirs in sorted(ecosystem_dirs.items()):
        unique_dirs = sorted(list(set(dirs)))
        dir_entries = (
            '["' + '", "'.join(d if d != "." else "/" for d in unique_dirs) + '"]'
        )
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


def main(
    scan_path: Path,
    interval: str,
    output_path: Path,
    ignore_dirs: list[str],
    custom_map_json: str,
) -> int:
    """
    Main function to generate dependabot.yml configuration.
    """
    logging.info(
        (
            "Starting dependabot generation with scan_path: '%s', interval: '%s', "
            "output_path: '%s', ignore_dirs: %s"
        ),
        scan_path,
        interval,
        output_path,
        ignore_dirs,
    )

    ecosystem_map = get_ecosystem_map(custom_map_json)

    logging.info(f"Scanning for directories with dependency files in '{scan_path}'")
    dirs = recursively_scan_directories(
        root_dir=scan_path, ignore_dirs=ignore_dirs, ecosystem_map=ecosystem_map
    )
    logging.info(
        f"Found {len(dirs)} directories with dependency files: {[str(d) for d in dirs]}"
    )

    logging.info("Generating dependabot configuration")
    config_content = generate_dependabot_config(
        directories=dirs, interval=interval, ecosystem_map=ecosystem_map
    )

    output_dir = output_path.parent
    if not output_dir.exists():
        logging.info(f"Creating output directory '{output_dir}'")
        output_dir.mkdir(parents=True, exist_ok=True)

    logging.info(f"Writing dependabot configuration to '{output_path}'")
    _ = output_path.write_text(config_content)

    logging.info(f"Dependabot configuration generated at '{output_path}'")
    return 0


if __name__ == "__main__":
    logging.basicConfig(
        level=logging.INFO,
        format="%(asctime)s - %(levelname)s - %(message)s",
        stream=sys.stdout,
    )

    parser = argparse.ArgumentParser(
        description="Generate dependabot.yml configuration."
    )
    _ = parser.add_argument(
        "--scan-path",
        type=str,
        default=".",
        help="Recursively scan this path for dependency files (default: .)",
    )
    _ = parser.add_argument(
        "--interval",
        type=str,
        default="weekly",
        help="Update interval for dependencies (default: weekly)",
    )
    _ = parser.add_argument(
        "--output-filepath",
        type=str,
        default=".github/dependabot.yml",
        help="Output file path (default: .github/dependabot.yml)",
    )
    _ = parser.add_argument(
        "--ignore-dirs",
        type=str,
        default=".venv,node_modules",
        help="Comma-separated string of relative paths to ignore.",
    )
    _ = parser.add_argument(
        "--custom-map",
        type=str,
        default="",
        help="JSON string to extend the default ecosystem map.",
    )

    args = parser.parse_args()
    ignore_dirs_str = cast(str, args.ignore_dirs)
    ignore_dirs = [d.strip() for d in ignore_dirs_str.split(",") if d.strip()]

    sys.exit(
        main(
            scan_path=Path(cast(str, args.scan_path)),
            interval=cast(str, args.interval),
            output_path=Path(cast(str, args.output_filepath)),
            ignore_dirs=ignore_dirs,
            custom_map_json=cast(str, args.custom_map),
        )
    )
