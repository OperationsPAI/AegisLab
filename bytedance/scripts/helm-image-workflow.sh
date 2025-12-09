#!/usr/bin/env bash

# Complete workflow: Extract, check, and push Helm chart images
# This script combines extraction, remote checking, and pushing

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
Usage: $0 [OPTIONS] <chart-directory>

Complete workflow to extract images from Helm charts and push to registry.

OPTIONS:
    -r, --registry <registry>    Target registry (required)
    -o, --output <dir>           Output directory (default: current directory)
    -p, --push                   Push images after extraction
    -s, --skip-existing          Skip images that already exist in registry
    -f, --force-pull             Force pull images before pushing
    -c, --continue               Continue on error when pushing
    -h, --help                   Show this help message

EXAMPLES:
    # Extract only
    $0 -r myregistry.com/project /path/to/chart

    # Extract and push
    $0 -p -r myregistry.com/project /path/to/chart

    # Extract, push, skip existing
    $0 -p -s -r myregistry.com/project /path/to/chart

    # Full workflow: extract, force pull, skip existing, continue on error
    $0 -p -s -f -c -r myregistry.com/project /path/to/chart

OUTPUT FILES:
    - images.txt                    Simple list of images
    - images.json                   Detailed JSON with paths and metadata
    - images_override.yaml           Helm values override for new registry

EOF
}

# Default values
REGISTRY=""
OUTPUT_DIR="."
PUSH_IMAGES=false
SKIP_EXISTING=false
FORCE_PULL=false
CONTINUE_ON_ERROR=false
CHART_DIR=""

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
        -s|--skip-existing)
            SKIP_EXISTING=true
            shift
            ;;
        -f|--force-pull)
            FORCE_PULL=true
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
            CHART_DIR="$1"
            shift
            ;;
    esac
done

# Validate inputs
if [[ -z "$CHART_DIR" ]]; then
    print_error "Chart directory is required"
    usage
    exit 1
fi

if [[ ! -d "$CHART_DIR" ]]; then
    print_error "Chart directory not found: $CHART_DIR"
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

# Resolve chart directory to absolute path
CHART_DIR=$(cd "$CHART_DIR" && pwd)

# Output files
IMAGES_TXT="${OUTPUT_DIR}/images.txt"
IMAGES_JSON="${OUTPUT_DIR}/images.json"
VALUES_OVERRIDE="${OUTPUT_DIR}/values_image_override.yaml"

echo ""
echo "════════════════════════════════════════════════════════"
echo "  Helm Chart Image Management Workflow"
echo "════════════════════════════════════════════════════════"
echo ""
print_info "Chart directory: $CHART_DIR"
print_info "Target registry: $REGISTRY"
print_info "Output directory: $OUTPUT_DIR"
echo ""

# Step 1: Extract images
print_info "Step 1: Extracting images from Helm charts..."
echo "────────────────────────────────────────────────────────"
if python3 "$SCRIPT_DIR/extract-images-v3.py" "$CHART_DIR" "$IMAGES_TXT" "$REGISTRY"; then
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

# Count images
IMAGE_COUNT=$(grep -v '^#' "$IMAGES_TXT" | grep -v '^[[:space:]]*$' | wc -l)
print_info "Found $IMAGE_COUNT images"
echo ""

# Display generated files
print_success "Generated files:"
echo "  📄 Image list:      $IMAGES_TXT"
if [[ -f "$IMAGES_JSON" ]]; then
    echo "  📋 Detailed JSON:   $IMAGES_JSON"
fi
if [[ -f "$VALUES_OVERRIDE" ]]; then
    echo "  ⚙️  Values override: $VALUES_OVERRIDE"
fi
echo ""

# Step 2: Push images (if requested)
if [[ "$PUSH_IMAGES" == "true" ]]; then
    echo "════════════════════════════════════════════════════════"
    print_info "Step 2: Pushing images to registry..."
    echo "────────────────────────────────────────────────────────"
    
    # Build push command
    PUSH_CMD="$SCRIPT_DIR/../../scripts/push-image-v2.sh -r $REGISTRY -f $IMAGES_TXT"
    
    if [[ "$SKIP_EXISTING" == "true" ]]; then
        PUSH_CMD="$PUSH_CMD -s"
        print_info "Remote check enabled: will skip existing images"
    fi
    
    if [[ "$FORCE_PULL" == "true" ]]; then
        PUSH_CMD="$PUSH_CMD -p"
        print_info "Force pull enabled: will pull latest versions"
    fi
    
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
    print_info "Step 2: Skipping image push (use -p to push)"
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
echo "   helm install myrelease ./chart \\"
echo "     -f $(basename $VALUES_OVERRIDE)"
echo ""
echo "3️⃣  Or manually push images:"
echo "   $SCRIPT_DIR/../../scripts/push-image-v2.sh -s -r $REGISTRY -f $IMAGES_TXT"
echo ""
