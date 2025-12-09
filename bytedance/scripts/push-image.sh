#!/usr/bin/env bash

# Color definitions
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default target registry
DEFAULT_REGISTRY="pair-diag-cn-guangzhou.cr.volces.com/pair"

# Function to print colored messages
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_skip() {
    echo -e "${YELLOW}[SKIP]${NC} $1"
}

# Function to check if image exists in remote registry
check_remote_image() {
    local image="$1"
    
    # Try to inspect the image using docker manifest
    if docker manifest inspect "$image" >/dev/null 2>&1; then
        return 0  # Image exists
    fi
    
    # Fallback: try skopeo if available
    if command -v skopeo &> /dev/null; then
        if skopeo inspect "docker://$image" >/dev/null 2>&1; then
            return 0  # Image exists
        fi
    fi
    
    return 1  # Image doesn't exist or can't be checked
}

# Function to show usage
usage() {
    cat << EOF
Usage: $0 [OPTIONS] <source-image | -f images-file>

Push Docker images to a target registry. If the image exists locally, skip pulling.

OPTIONS:
    -r, --registry <registry>    Target registry (default: ${DEFAULT_REGISTRY})
    -t, --tag <tag>              Override the image tag (optional, only for single image)
    -p, --pull                   Force pull even if image exists locally
    -f, --file <file>            Read image list from file (one image per line)
    -c, --continue               Continue on error when processing multiple images
    -h, --help                   Show this help message

EXAMPLES:
    # Push a single existing local image
    $0 redis:7.0

    # Push to a custom registry
    $0 -r myregistry.com/myproject redis:7.0

    # Force pull and push
    $0 -p bitnami/redis:latest

    # Override tag for single image
    $0 -t v1.0.0 myapp:latest

    # Batch push from file
    $0 -f images.txt

    # Batch push with force pull and continue on error
    $0 -p -c -f images.txt

    # Batch push to custom registry
    $0 -r myregistry.com -f images.txt

ARGUMENTS:
    source-image                 Source image name (e.g., redis:7.0, bitnami/redis:latest)

FILE FORMAT:
    The images file should contain one image per line:
        redis:7.0
        bitnami/mysql:8.0
        nginx:alpine
    Empty lines and lines starting with # are ignored.

EOF
}

# Parse command line arguments
FORCE_PULL=false
TARGET_REGISTRY="$DEFAULT_REGISTRY"
OVERRIDE_TAG=""
IMAGES_FILE=""
CONTINUE_ON_ERROR=false

while [[ $# -gt 0 ]]; do
    case $1 in
        -r|--registry)
            TARGET_REGISTRY="$2"
            shift 2
            ;;
        -t|--tag)
            OVERRIDE_TAG="$2"
            shift 2
            ;;
        -p|--pull)
            FORCE_PULL=true
            shift
            ;;
        -f|--file)
            IMAGES_FILE="$2"
            shift 2
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
            SOURCE_IMAGE="$1"
            shift
            ;;
    esac
done

# Function to process a single image
process_image() {
    local SOURCE_IMAGE="$1"
    local OVERRIDE_TAG="$2"
    
    # Parse image name and tag
    if [[ "$SOURCE_IMAGE" =~ (.+):(.+) ]]; then
        IMAGE_NAME="${BASH_REMATCH[1]}"
        IMAGE_TAG="${BASH_REMATCH[2]}"
    else
        IMAGE_NAME="$SOURCE_IMAGE"
        IMAGE_TAG="latest"
    fi

    # Use override tag if provided
    if [[ -n "$OVERRIDE_TAG" ]]; then
        IMAGE_TAG="$OVERRIDE_TAG"
    fi

    # Extract short name (last part after /)
    SHORT_NAME="${IMAGE_NAME##*/}"

    # Build target image name
    TARGET_IMAGE="${TARGET_REGISTRY}/${SHORT_NAME}:${IMAGE_TAG}"

    print_info "Source Image: ${SOURCE_IMAGE}"
    print_info "Target Image: ${TARGET_IMAGE}"
    echo ""

    # STEP 1: Check if target image already exists in remote registry
    print_info "Checking if target image exists in registry..."
    if check_remote_image "$TARGET_IMAGE"; then
        print_skip "Target image already exists in registry, skipping all operations"
        echo ""
        return 2  # Special return code for skipped
    fi
    print_info "Target image not found in registry, proceeding with pull+tag+push"
    echo ""

    # STEP 2: Check if source image exists locally and pull if needed
    if docker image inspect "${SOURCE_IMAGE}" >/dev/null 2>&1; then
        print_success "Source image ${SOURCE_IMAGE} found locally"
        
        if [[ "$FORCE_PULL" == "true" ]]; then
            print_warning "Force pull enabled, pulling latest version..."
            if docker pull "${SOURCE_IMAGE}"; then
                print_success "Image pulled successfully"
            else
                print_error "Failed to pull image"
                return 1
            fi
        else
            print_info "Skipping pull (use -p to force pull)"
        fi
    else
        print_warning "Source image ${SOURCE_IMAGE} not found locally, pulling..."
        if docker pull "${SOURCE_IMAGE}"; then
            print_success "Image pulled successfully"
        else
            print_error "Failed to pull image"
            return 1
        fi
    fi

    echo ""

    # Tag the image
    print_info "Tagging image as ${TARGET_IMAGE}..."
    if docker tag "${SOURCE_IMAGE}" "${TARGET_IMAGE}"; then
        print_success "Image tagged successfully"
    else
        print_error "Failed to tag image"
        return 1
    fi

    echo ""

    # Push the image
    print_info "Pushing image to ${TARGET_REGISTRY}..."
    if docker push "${TARGET_IMAGE}"; then
        print_success "Image pushed successfully to ${TARGET_IMAGE}"
    else
        print_error "Failed to push image"
        return 1
    fi

    echo ""
    return 0
}

# Process batch mode (from file)
if [[ -n "$IMAGES_FILE" ]]; then
    # Validate file exists
    if [[ ! -f "$IMAGES_FILE" ]]; then
        print_error "Images file not found: $IMAGES_FILE"
        exit 1
    fi

    # Check for tag override with batch mode
    if [[ -n "$OVERRIDE_TAG" ]]; then
        print_warning "Tag override (-t) is ignored in batch mode"
    fi

    print_info "Processing images from file: $IMAGES_FILE"
    print_info "Target registry: $TARGET_REGISTRY"
    echo ""

    # Read images from file
    TOTAL_IMAGES=0
    SUCCESS_COUNT=0
    SKIPPED_COUNT=0
    FAILED_COUNT=0
    declare -a FAILED_IMAGES
    declare -a SKIPPED_IMAGES

    while IFS= read -r line || [[ -n "$line" ]]; do
        # Skip empty lines and comments
        [[ -z "$line" || "$line" =~ ^[[:space:]]*# ]] && continue
        
        # Trim whitespace
        IMAGE=$(echo "$line" | xargs)
        [[ -z "$IMAGE" ]] && continue

        TOTAL_IMAGES=$((TOTAL_IMAGES + 1))
        
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        print_info "Processing image $TOTAL_IMAGES: $IMAGE"
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        
        process_image "$IMAGE" ""
        result=$?
        
        if [[ $result -eq 0 ]]; then
            SUCCESS_COUNT=$((SUCCESS_COUNT + 1))
            print_success "✅ Image $TOTAL_IMAGES completed - pushed successfully"
        elif [[ $result -eq 2 ]]; then
            SKIPPED_COUNT=$((SKIPPED_COUNT + 1))
            SKIPPED_IMAGES+=("$IMAGE")
            print_skip "⏭️  Image $TOTAL_IMAGES completed - already exists, skipped"
        else
            FAILED_COUNT=$((FAILED_COUNT + 1))
            FAILED_IMAGES+=("$IMAGE")
            print_error "❌ Image $TOTAL_IMAGES failed"
            
            if [[ "$CONTINUE_ON_ERROR" == "false" ]]; then
                print_error "Stopping due to error (use -c to continue on error)"
                exit 1
            fi
        fi
        
        echo ""
    done < "$IMAGES_FILE"

    # Print summary
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    print_info "SUMMARY"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    print_info "Total images: $TOTAL_IMAGES"
    print_success "Pushed: $SUCCESS_COUNT"
    print_skip "Skipped (already exists): $SKIPPED_COUNT"
    
    if [[ $FAILED_COUNT -gt 0 ]]; then
        print_error "Failed: $FAILED_COUNT"
        echo ""
    fi
    
    if [[ $SKIPPED_COUNT -gt 0 ]]; then
        echo ""
        print_skip "Skipped images (already exist in target registry):"
        for img in "${SKIPPED_IMAGES[@]}"; do
            echo "  - $img"
        done
    fi
    
    if [[ $FAILED_COUNT -gt 0 ]]; then
        echo ""
        print_error "Failed images:"
        for img in "${FAILED_IMAGES[@]}"; do
            echo "  - $img"
        done
        exit 1
    else
        print_success "✅ All images processed successfully!"
    fi
    
    exit 0
fi

# Process single image mode
if [[ -z "$SOURCE_IMAGE" ]]; then
    print_error "Source image is required (or use -f for batch mode)"
    usage
    exit 1
fi

if process_image "$SOURCE_IMAGE" "$OVERRIDE_TAG"; then
    print_success "✅ Complete! Image available at: ${TARGET_IMAGE}"
    exit 0
else
    print_error "Failed to process image"
    exit 1
fi
