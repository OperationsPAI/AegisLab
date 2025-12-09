#!/usr/bin/env python3
"""
Extract Docker images from Helm charts using recursive helm show commands.

This script recursively discovers chart dependencies using 'helm show chart'
and extracts all image references from values using 'helm show values'.
It processes the entire chart hierarchy without needing to download or render templates.

Usage:
    # Extract from Helm repository chart
    python extract-images-helm-v2.py open-telemetry/opentelemetry-kube-stack

    # Extract from local chart directory
    python extract-images-helm-v2.py /path/to/chart

    # Specify custom output files
    python extract-images-helm-v2.py repo/chart -o images.txt -j metadata.json

Output:
    - images.txt: Plain text list of unique images (one per line)
    - images.json: Detailed JSON with image metadata including registry, repository, tag, chart path

Author: Copilot
Date: 2024
"""

import argparse
import json
import subprocess
import sys
import yaml
from pathlib import Path
from typing import Dict, List, Set, Any
from collections import defaultdict


def run_helm_show_chart(chart_ref: str) -> Dict[str, Any]:
    """
    Run 'helm show chart' to get Chart.yaml metadata.

    Args:
        chart_ref: Chart reference (repo/chart or local path)

    Returns:
        Parsed Chart.yaml as dictionary
    """
    cmd = ["helm", "show", "chart", chart_ref]

    try:
        result = subprocess.run(cmd, capture_output=True, text=True, check=True)
        return yaml.safe_load(result.stdout) or {}
    except subprocess.CalledProcessError as e:
        print(f"❌ Error running helm show chart for {chart_ref}:", file=sys.stderr)
        print(f"Stderr: {e.stderr}", file=sys.stderr)
        raise
    except yaml.YAMLError as e:
        print(f"❌ Error parsing Chart.yaml for {chart_ref}: {e}", file=sys.stderr)
        raise


def run_helm_show_values(chart_ref: str) -> Dict[str, Any]:
    """
    Run 'helm show values' to get default values.yaml.

    Args:
        chart_ref: Chart reference (repo/chart or local path)

    Returns:
        Parsed values.yaml as dictionary
    """
    cmd = ["helm", "show", "values", chart_ref]

    try:
        result = subprocess.run(cmd, capture_output=True, text=True, check=True)
        return yaml.safe_load(result.stdout) or {}
    except subprocess.CalledProcessError as e:
        print(
            f"⚠️  Warning: Could not get values for {chart_ref}: {e.stderr}",
            file=sys.stderr,
        )
        return {}
    except yaml.YAMLError as e:
        print(
            f"⚠️  Warning: Could not parse values for {chart_ref}: {e}", file=sys.stderr
        )
        return {}


def extract_images_from_values(
    values: Dict[str, Any], prefix: str = ""
) -> List[Dict[str, str]]:
    """
    Recursively extract image references from values dictionary.

    Common patterns:
        image: "repo/name:tag"
        image:
          repository: "repo/name"
          tag: "tag"
        images:
          - repository: "repo/name"
            tag: "tag"

    Args:
        values: Values dictionary
        prefix: Key prefix for tracking location (e.g., "mysql", "redis")

    Returns:
        List of dicts with image info: {full, registry, repository, tag, path}
    """
    images = []

    if not isinstance(values, dict):
        return images

    for key, value in values.items():
        current_path = f"{prefix}.{key}" if prefix else key

        # Pattern 1: image: "full/image:tag"
        if key == "image" and isinstance(value, str) and value:
            images.append({"full": value, "path": prefix or "root"})

        # Pattern 2: image: {repository, tag, ...} or {registry, repository, tag}
        elif key == "image" and isinstance(value, dict):
            registry = value.get("registry", "")
            repo = value.get("repository", "")
            tag = value.get("tag", "")

            if repo:
                # Build full image reference
                if registry:
                    full_repo = f"{registry}/{repo}"
                else:
                    full_repo = repo

                if tag:
                    full_image = f"{full_repo}:{tag}"
                else:
                    # No tag specified, use :latest as convention
                    full_image = f"{full_repo}:latest"

                images.append({"full": full_image, "path": prefix or "root"})

        # Pattern 3: images: [{repository, tag}, ...]
        elif key == "images" and isinstance(value, list):
            for i, img in enumerate(value):
                if isinstance(img, dict):
                    repo = img.get("repository", "")
                    tag = img.get("tag", "")
                    if repo:
                        if tag:
                            full_image = f"{repo}:{tag}"
                        else:
                            full_image = repo
                        images.append(
                            {"full": full_image, "path": f"{current_path}[{i}]"}
                        )

        # Recurse into nested dictionaries
        elif isinstance(value, dict):
            images.extend(extract_images_from_values(value, current_path))

        # Recurse into lists of dictionaries
        elif isinstance(value, list):
            for i, item in enumerate(value):
                if isinstance(item, dict):
                    images.extend(
                        extract_images_from_values(item, f"{current_path}[{i}]")
                    )

    return images


def get_chart_dependencies(chart_metadata: Dict[str, Any]) -> List[Dict[str, Any]]:
    """
    Extract dependencies from Chart.yaml metadata.

    Args:
        chart_metadata: Parsed Chart.yaml

    Returns:
        List of dependency dictionaries with name, repository, version, condition, etc.
    """
    return chart_metadata.get("dependencies", [])


def get_helm_repo_mapping() -> Dict[str, str]:
    """
    Get mapping from repository URL to repository name.

    Returns:
        Dict mapping URL to repo name
    """
    try:
        result = subprocess.run(
            ["helm", "repo", "list", "-o", "json"],
            capture_output=True,
            text=True,
            check=True,
        )
        repos = json.loads(result.stdout)
        return {repo["url"].rstrip("/"): repo["name"] for repo in repos}
    except Exception as e:
        print(f"⚠️  Warning: Could not get helm repo list: {e}", file=sys.stderr)
        return {}


def process_chart_recursive(
    chart_ref: str,
    chart_name: str | None = None,
    processed: Set[str] | None = None,
    depth: int = 0,
    repo_mapping: Dict[str, str] | None = None,
) -> List[Dict[str, Any]]:
    """
    Recursively process a chart and all its subcharts.

    Args:
        chart_ref: Chart reference (repo/chart or local path)
        chart_name: Name for tracking (e.g., "kube-state-metrics")
        processed: Set of already processed chart names (to avoid cycles)
        depth: Current recursion depth

    Returns:
        List of dicts: [{chart_name, chart_ref, values, dependencies}, ...]
    """
    if processed is None:
        processed = set()

    indent = "  " * depth
    chart_display_name = chart_name or chart_ref

    # Avoid infinite recursion
    if chart_display_name in processed:
        print(f"{indent}⏭️  Skipping already processed: {chart_display_name}")
        return []

    processed.add(chart_display_name)

    print(f"{indent}🔍 Processing chart: {chart_display_name}")

    results = []

    try:
        # Get chart metadata
        chart_metadata = run_helm_show_chart(chart_ref)
        chart_actual_name = chart_metadata.get("name", chart_name or "unknown")

        # Get values
        values = run_helm_show_values(chart_ref)

        # Get dependencies
        dependencies = get_chart_dependencies(chart_metadata)

        results.append(
            {
                "chart_name": chart_actual_name,
                "chart_ref": chart_ref,
                "values": values,
                "dependencies": dependencies,
            }
        )

        print(f"{indent}  ✅ Found {len(dependencies)} dependencies")

        # Initialize repo_mapping if not provided
        if repo_mapping is None:
            repo_mapping = get_helm_repo_mapping()

        # Process each dependency recursively
        for dep in dependencies:
            dep_name = dep.get("name")
            dep_repo = dep.get("repository")

            if not dep_name:
                continue

            # NOTE: We process ALL dependencies regardless of their enabled status
            # This is because users might enable them later with custom values

            # Construct subchart reference
            if dep_repo and dep_repo.startswith("http"):
                # External repository - look up repo name from mapping
                dep_repo_normalized = dep_repo.rstrip("/")
                repo_name = repo_mapping.get(dep_repo_normalized)

                if repo_name:
                    subchart_ref = f"{repo_name}/{dep_name}"
                    print(
                        f"{indent}  📦 Processing external subchart: {dep_name} (from {repo_name})"
                    )

                    # Recursively process this subchart
                    subchart_results = process_chart_recursive(
                        subchart_ref,
                        chart_name=dep_name,
                        processed=processed,
                        depth=depth + 1,
                        repo_mapping=repo_mapping,
                    )
                    results.extend(subchart_results)
                else:
                    print(
                        f"{indent}  ⚠️  Repository not found in helm repos: {dep_repo}"
                    )
                    print(f"{indent}     Add it with: helm repo add <name> {dep_repo}")
                continue

            elif not dep_repo or dep_repo == "":
                # Local subchart (in charts/ directory)
                print(f"{indent}  📦 Found local subchart: {dep_name}")
                print(
                    f"{indent}     Extracting values from parent chart's {dep_name} section"
                )

                # Try to get subchart values from parent values under subchart name
                subchart_values = values.get(dep_name, {})
                if subchart_values and isinstance(subchart_values, dict):
                    # Create a pseudo-result for this subchart
                    results.append(
                        {
                            "chart_name": dep_name,
                            "chart_ref": f"{chart_ref}/charts/{dep_name}",
                            "values": subchart_values,
                            "dependencies": [],
                        }
                    )
                continue
            else:
                # Repository alias or other format
                print(
                    f"{indent}  ⚠️  Unknown repository format for {dep_name}: {dep_repo}"
                )
                continue

    except Exception as e:
        print(
            f"{indent}❌ Error processing chart {chart_display_name}: {e}",
            file=sys.stderr,
        )
        import traceback

        traceback.print_exc()

    return results


def parse_image(image: str) -> Dict[str, str]:
    """
    Parse Docker image reference into components.

    Args:
        image: Docker image reference string

    Returns:
        Dictionary with parsed components
    """
    result = {
        "full": image,
        "registry": "docker.io",
        "repository": "",
        "tag": "latest",
        "digest": "",
    }

    # Handle digest
    if "@" in image:
        image_part, digest = image.split("@", 1)
        result["digest"] = digest
    else:
        image_part = image

    # Split into registry/repo and tag
    if ":" in image_part and "/" in image_part:
        first_slash_idx = image_part.index("/")
        first_colon_idx = image_part.index(":")

        if first_colon_idx > first_slash_idx:
            repo_part, tag = image_part.rsplit(":", 1)
            result["tag"] = tag
        else:
            repo_part = image_part
    elif ":" in image_part:
        repo_part, tag = image_part.rsplit(":", 1)
        result["tag"] = tag
    else:
        repo_part = image_part

    # Parse registry and repository
    parts = repo_part.split("/")

    if len(parts) == 1:
        result["repository"] = f"library/{parts[0]}"
    elif len(parts) == 2:
        if "." in parts[0] or ":" in parts[0] or parts[0] == "localhost":
            result["registry"] = parts[0]
            result["repository"] = parts[1]
        else:
            result["repository"] = repo_part
    else:
        result["registry"] = parts[0]
        result["repository"] = "/".join(parts[1:])

    return result


def generate_outputs(
    images_with_paths: List[Dict[str, str]], output_txt: Path, output_json: Path
):
    """
    Generate output files: text list and JSON metadata.

    Args:
        images_with_paths: List of dicts with {full, path} keys
        output_txt: Path for text output file
        output_json: Path for JSON output file
    """
    # Get unique images
    unique_images = {}
    for img_info in images_with_paths:
        full = img_info["full"]
        if full not in unique_images:
            unique_images[full] = img_info["path"]

    sorted_images = sorted(unique_images.keys())

    # Generate text file
    with open(output_txt, "w") as f:
        for image in sorted_images:
            f.write(f"{image}\n")

    print(f"✅ Saved {len(sorted_images)} images to: {output_txt}")

    # Generate JSON metadata
    metadata = {"total_images": len(sorted_images), "images": []}

    for image in sorted_images:
        parsed = parse_image(image)
        parsed["chart_path"] = unique_images[image]
        metadata["images"].append(parsed)

    with open(output_json, "w") as f:
        json.dump(metadata, f, indent=2)

    print(f"✅ Saved metadata to: {output_json}")

    # Print summary
    registry_count = defaultdict(int)
    path_count = defaultdict(int)

    for image in sorted_images:
        parsed = parse_image(image)
        registry_count[parsed["registry"]] += 1
        path_count[unique_images[image]] += 1

    print("\n📊 Summary by registry:")
    for registry, count in sorted(registry_count.items(), key=lambda x: -x[1]):
        print(f"  {registry}: {count} images")

    print("\n📊 Summary by chart:")
    for path, count in sorted(path_count.items(), key=lambda x: -x[1]):
        print(f"  {path}: {count} images")


def main():
    parser = argparse.ArgumentParser(
        description="Extract Docker images from Helm charts using recursive helm show commands",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  # Extract from Helm repository
  %(prog)s open-telemetry/opentelemetry-kube-stack

  # Extract from local chart
  %(prog)s ./my-chart

  # Custom output location
  %(prog)s repo/chart -o /tmp/images.txt -j /tmp/metadata.json
        """,
    )

    parser.add_argument(
        "chart", type=str, help="Helm chart reference (repo/chart or local path)"
    )

    parser.add_argument(
        "-o",
        "--output",
        type=Path,
        default=Path("images.txt"),
        help="Output text file for image list (default: images.txt)",
    )

    parser.add_argument(
        "-j",
        "--json",
        type=Path,
        default=Path("images.json"),
        help="Output JSON file for metadata (default: images.json)",
    )

    args = parser.parse_args()

    print(f"🎯 Extracting images from chart: {args.chart}")
    print("=" * 70)

    try:
        # Process chart recursively
        all_chart_data = process_chart_recursive(args.chart)

        if not all_chart_data:
            print("\n⚠️  Warning: No chart data collected", file=sys.stderr)
            sys.exit(1)

        print("\n" + "=" * 70)
        print(f"✅ Processed {len(all_chart_data)} chart(s)")

        # Extract images from all charts
        print("\n🔍 Extracting images from values...")
        all_images = []

        for chart_data in all_chart_data:
            chart_name = chart_data["chart_name"]
            values = chart_data["values"]

            images = extract_images_from_values(values, prefix=chart_name)

            if images:
                print(f"  {chart_name}: {len(images)} image(s)")
                all_images.extend(images)

        if not all_images:
            print("\n⚠️  Warning: No images found in chart values", file=sys.stderr)
            sys.exit(0)

        print(f"\n✅ Found {len(all_images)} total image references")
        unique_count = len(set(img["full"] for img in all_images))
        print(f"✅ Found {unique_count} unique images")

        # Generate outputs
        print("\n📝 Generating output files...")
        generate_outputs(all_images, args.output, args.json)

        print("\n✅ Done!")

    except subprocess.CalledProcessError:
        print("\n❌ Failed to run helm command", file=sys.stderr)
        print(
            "Make sure helm is installed and the chart reference is correct",
            file=sys.stderr,
        )
        sys.exit(1)
    except Exception as e:
        print(f"\n❌ Error: {e}", file=sys.stderr)
        import traceback

        traceback.print_exc()
        sys.exit(1)


if __name__ == "__main__":
    main()
