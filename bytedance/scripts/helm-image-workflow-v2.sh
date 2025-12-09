#!/usr/bin/env bash

# Complete workflow: Extract images from Helm charts, generate override, and push
# Supports both Helm repository charts (repo/chart) and local chart directories

set -euo pipefail

# Color definitions
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

print_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
print_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
print_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
print_error() { echo -e "${RED}[ERROR]${NC} $1"; }

usage() {
    cat << EOF
Usage: $0 [OPTIONS] <chart-reference>

Complete workflow to extract images from Helm charts, generate override, and push to registry.
Supports both Helm repository charts (repo/chart) and local chart directories.

OPTIONS:
    -r, --registry <registry>    Target registry (required)
    -o, --output <dir>           Output directory (default: current directory)
    -p, --push                   Push images after extraction
    -c, --continue               Continue on error when pushing
    -h, --help                   Show this help message

EXAMPLES:
    # Extract from Helm repository chart (extract only)
    $0 -r myregistry.com/project open-telemetry/opentelemetry-kube-stack

    # Extract from local chart directory
    $0 -r myregistry.com/project /path/to/local/chart

    # Extract and push with auto-skip existing images
    $0 -p -r myregistry.com/project open-telemetry/opentelemetry-kube-stack

    # Extract and push, continue on error
    $0 -p -c -r myregistry.com/project bitnami/redis

OUTPUT FILES:
    - images.txt                   Simple list of images (one per line)
    - images.json                  Detailed JSON with registry, repository, tag, chart path
    - values_image_override.yaml   Helm values override file for new registry

FEATURES:
    ✓ Recursively extracts images from all subcharts using 'helm show' commands
    ✓ Automatically checks target registry before pulling (saves bandwidth)
    ✓ Skips images that already exist in target registry
    ✓ Generates Helm values override file with new registry paths
    ✓ Supports both Helm repository (repo/chart) and local chart references
    ✓ No need to download charts locally

REQUIREMENTS:
    - helm command must be installed
    - docker command must be installed (for push)
    - python3 must be installed
    - Required Helm repositories must be added (e.g., helm repo add open-telemetry ...)

NOTES:
    - Target registry check is always enabled (no -s flag needed)
    - Force pull is not needed as we always check target first

EOF
}

# Default values
REGISTRY=""
OUTPUT_DIR="."
PUSH_IMAGES=false
CONTINUE_ON_ERROR=false
CHART_REF=""

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -r|--registry)
            REGISTRY="$2"
            shift 2
            ;;
        -o|--output)
            OUTPUT_DIR="$2"
            shift 2
            ;;
        -p|--push)
            PUSH_IMAGES=true
            shift
            ;;
        -c|--continue)
            CONTINUE_ON_ERROR=true
            shift
            ;;
        -h|--help)
            usage
            exit 0
            ;;
        -*)
            print_error "Unknown option: $1"
            usage
            exit 1
            ;;
        *)
            CHART_REF="$1"
            shift
            ;;
    esac
done

# Validate inputs
if [[ -z "$CHART_REF" ]]; then
    print_error "Chart reference is required"
    usage
    exit 1
fi

if [[ -z "$REGISTRY" ]]; then
    print_error "Target registry is required (use -r)"
    usage
    exit 1
fi

# Create output directory
mkdir -p "$OUTPUT_DIR"
OUTPUT_DIR=$(cd "$OUTPUT_DIR" && pwd)

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Output files
IMAGES_TXT="${OUTPUT_DIR}/images.txt"
IMAGES_JSON="${OUTPUT_DIR}/images.json"
VALUES_OVERRIDE="${OUTPUT_DIR}/values_image_override.yaml"

echo ""
echo "════════════════════════════════════════════════════════"
echo "  Helm Chart Image Management Workflow"
echo "════════════════════════════════════════════════════════"
echo ""
print_info "Chart reference: $CHART_REF"
print_info "Target registry: $REGISTRY"
print_info "Output directory: $OUTPUT_DIR"
echo ""

# Step 1: Extract images using helm show commands
print_info "Step 1: Extracting images from Helm chart (recursive)..."
echo "────────────────────────────────────────────────────────"

if python3 "$SCRIPT_DIR/extract-images-helm-v2.py" \
    "$CHART_REF" \
    -o "$IMAGES_TXT" \
    -j "$IMAGES_JSON"; then
    print_success "Images extracted successfully"
else
    print_error "Failed to extract images"
    exit 1
fi

echo ""

# Check if files were created
if [[ ! -f "$IMAGES_TXT" ]]; then
    print_error "Images list file not created"
    exit 1
fi

if [[ ! -f "$IMAGES_JSON" ]]; then
    print_error "Images JSON file not created"
    exit 1
fi

# Count images
IMAGE_COUNT=$(grep -v '^#' "$IMAGES_TXT" | grep -v '^[[:space:]]*$' | wc -l)
print_info "Found $IMAGE_COUNT unique images"
echo ""

# Step 2: Generate values override file
print_info "Step 2: Generating Helm values override file..."
echo "────────────────────────────────────────────────────────"

# Use Python to generate the override file from JSON
python3 << 'PYTHON_SCRIPT' "$IMAGES_JSON" "$VALUES_OVERRIDE" "$REGISTRY"
import sys
import json
import yaml

images_json_file = sys.argv[1]
output_file = sys.argv[2]
target_registry = sys.argv[3]

# Load images metadata
with open(images_json_file, 'r') as f:
    data = json.load(f)

# Build override structure
override = {}

for img_data in data['images']:
    chart_path = img_data.get('chart_path', '')
    full_image = img_data['full']
    registry = img_data['registry']
    repository = img_data['repository']
    tag = img_data['tag']
    
    # Parse chart path (e.g., "kube-state-metrics" or "opentelemetry-operator.manager")
    path_parts = chart_path.split('.')
    
    # Build target image
    target_repo = repository.split('/')[-1] if '/' in repository else repository
    target_image = f"{target_registry}/{target_repo}:{tag}"
    
    # Navigate/create nested dict structure
    current = override
    for i, part in enumerate(path_parts):
        if i == len(path_parts) - 1:
            # Last part - this is where we set the image
            if part == path_parts[0] and len(path_parts) == 1:
                # Root level chart (e.g., "kube-state-metrics")
                if 'image' not in current:
                    current['image'] = {}
                if isinstance(current.get('image'), dict):
                    current['image']['repository'] = f"{target_registry}/{target_repo}"
                    current['image']['tag'] = tag
            else:
                # Nested path (e.g., "manager" in "opentelemetry-operator.manager")
                if part not in current:
                    current[part] = {}
                if 'image' not in current[part]:
                    current[part]['image'] = {}
                if isinstance(current[part].get('image'), dict):
                    current[part]['image']['repository'] = f"{target_registry}/{target_repo}"
                    current[part]['image']['tag'] = tag
        else:
            # Intermediate parts - create subchart sections
            if part not in current:
                current[part] = {}
            current = current[part]

# Write YAML
with open(output_file, 'w') as f:
    f.write('# Helm values override for custom registry\n')
    f.write(f'# Generated from: {images_json_file}\n')
    f.write(f'# Target registry: {target_registry}\n')
    f.write('#\n')
    f.write('# Usage: helm install myrelease ./chart -f values_image_override.yaml\n')
    f.write('\n')
    yaml.dump(override, f, default_flow_style=False, sort_keys=False)

print(f"✓ Generated values override: {output_file}")
PYTHON_SCRIPT

if [[ $? -eq 0 ]]; then
    print_success "Values override file generated"
else
    print_error "Failed to generate values override file"
    exit 1
fi

echo ""

# Display generated files
print_success "Generated files:"
echo "  📄 Image list:      $IMAGES_TXT"
echo "  📋 Detailed JSON:   $IMAGES_JSON"
echo "  ⚙️  Values override: $VALUES_OVERRIDE"
echo ""

# Step 3: Push images (if requested)
if [[ "$PUSH_IMAGES" == "true" ]]; then
    echo "════════════════════════════════════════════════════════"
    print_info "Step 3: Pushing images to registry..."
    echo "────────────────────────────────────────────────────────"
    print_info "Target check is enabled: will skip existing images"
    echo ""
    
    # Build push command (target check is always enabled in new push-image.sh)
    PUSH_CMD="$SCRIPT_DIR/push-image.sh -r $REGISTRY -f $IMAGES_TXT"
    
    if [[ "$CONTINUE_ON_ERROR" == "true" ]]; then
        PUSH_CMD="$PUSH_CMD -c"
        print_info "Continue on error enabled"
    fi
    
    echo ""
    
    if eval "$PUSH_CMD"; then
        print_success "Images pushed successfully"
    else
        print_error "Some images failed to push"
        exit 1
    fi
else
    echo "════════════════════════════════════════════════════════"
    print_info "Step 3: Skipping image push (use -p to push)"
    echo "════════════════════════════════════════════════════════"
fi

echo ""
echo "════════════════════════════════════════════════════════"
print_success "✅ Workflow completed successfully!"
echo "════════════════════════════════════════════════════════"
echo ""

# Show next steps
print_info "Next steps:"
echo ""
echo "1️⃣  Review extracted images:"
echo "   cat $IMAGES_TXT"
echo ""
echo "2️⃣  Use the values override with Helm:"
echo "   helm install myrelease $CHART_REF \\"
echo "     -f $VALUES_OVERRIDE"
echo ""
if [[ "$PUSH_IMAGES" != "true" ]]; then
    echo "3️⃣  Push images to registry:"
    echo "   $SCRIPT_DIR/push-image.sh -r $REGISTRY -f $IMAGES_TXT"
    echo ""
fi
