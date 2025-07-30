#!/bin/bash

# RCABench Release Script
# This script updates version numbers across the project and publishes the Python SDK to PyPI

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
CONTROLLER_DIR="src"
SDK_DIR="sdk/python-gen"
MAIN_GO_FILE="${CONTROLLER_DIR}/main.go"
PYPROJECT_FILE="${SDK_DIR}/pyproject.toml"

# Functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if version is provided
if [ $# -eq 0 ]; then
    log_error "Please provide a version number"
    echo "Usage: $0 <version> [--dry-run]"
    echo "Example: $0 1.0.1"
    exit 1
fi

VERSION=$1
DRY_RUN=false

if [ "$2" = "--dry-run" ]; then
    DRY_RUN=true
    log_info "Running in dry-run mode"
fi

# Validate version format (basic semver check)
if ! [[ $VERSION =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    log_error "Invalid version format. Please use semantic versioning (e.g., 1.0.1)"
    exit 1
fi

log_info "Starting release process for version $VERSION"

# Step 1: Update version in main.go
log_info "Updating version in $MAIN_GO_FILE"
if [ "$DRY_RUN" = false ]; then
    sed -i "s/@version.*/@version         $VERSION/" "$MAIN_GO_FILE"
    log_info "Updated main.go version to $VERSION"
else
    log_info "[DRY-RUN] Would update main.go version to $VERSION"
fi

# Step 2: Update version in pyproject.toml
log_info "Updating version in $PYPROJECT_FILE"
if [ "$DRY_RUN" = false ]; then
    sed -i "s/version = \".*\"/version = \"$VERSION\"/" "$PYPROJECT_FILE"
    log_info "Updated pyproject.toml version to $VERSION"
else
    log_info "[DRY-RUN] Would update pyproject.toml version to $VERSION"
fi

# Step 3: Regenerate OpenAPI documentation
log_info "Regenerating OpenAPI documentation"
if [ "$DRY_RUN" = false ]; then
    make swag-init || {
        log_error "Failed to generate Swagger documentation"
        exit 1
    }
    log_info "OpenAPI documentation regenerated"
else
    log_info "[DRY-RUN] Would regenerate OpenAPI documentation"
fi

# Step 4: Regenerate Python SDK
log_info "Regenerating Python SDK with modern packaging"
if [ "$DRY_RUN" = false ]; then
    make generate-sdk || {
        log_error "Failed to generate Python SDK"
        exit 1
    }
    log_info "Python SDK regenerated with modern packaging configuration"
else
    log_info "[DRY-RUN] Would regenerate Python SDK with modern packaging"
fi


# Step 5: Build and test the Python package
log_info "Building and testing Python package"
if [ "$DRY_RUN" = false ]; then
    cd "$SDK_DIR"
    
    # Clean previous builds
    rm -rf dist/ build/ *.egg-info/
    
    # Ensure we have the required build tools
    if ! command -v uv &> /dev/null; then
        log_error "uv is not installed. Please install uv first."
        exit 1
    fi
    
    # Install dev dependencies if not already installed
    log_info "Installing dev dependencies..."
    uv sync --dev || {
        log_error "Failed to install dev dependencies"
        exit 1
    }
    
    # Build the package using uv
    log_info "Building package..."
    uv run python -m build || {
        log_error "Failed to build Python package"
        exit 1
    }
    
    # Run basic tests
    log_info "Validating package..."
    uv run python -m twine check dist/* || {
        log_error "Package validation failed"
        exit 1
    }
    
    cd - > /dev/null
    log_info "Python package built and validated successfully"
else
    log_info "[DRY-RUN] Would build and test Python package"
fi

# Step 6: Publish to PyPI
if [ "$DRY_RUN" = false ]; then
    log_warn "Ready to publish to PyPI. Continue? (y/N)"
    read -r response
    if [[ "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
        log_info "Publishing to PyPI"
        cd "$SDK_DIR"
        uv run python -m twine upload dist/* || {
            log_error "Failed to publish to PyPI"
            exit 1
        }
        cd - > /dev/null
        log_info "Successfully published to PyPI"
    else
        log_info "Skipping PyPI publication"
    fi
else
    log_info "[DRY-RUN] Would publish to PyPI"
fi

# Step 7: Create git tag
if [ "$DRY_RUN" = false ]; then
    log_info "Creating git tag v$VERSION"
    git add .
    git commit -m "Release version $VERSION" || log_warn "No changes to commit"
    git tag -a "v$VERSION" -m "Release version $VERSION"
    log_info "Git tag v$VERSION created"
    
    log_info "Don't forget to push the changes and tag:"
    echo "  git push origin main"
    echo "  git push origin v$VERSION"
else
    log_info "[DRY-RUN] Would create git tag v$VERSION"
fi

log_info "Release process completed successfully!"
