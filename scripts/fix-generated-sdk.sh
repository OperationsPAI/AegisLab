#!/bin/bash

# Fix Generated SDK Script
# This script fixes the generated Python SDK to use modern packaging

set -e

SDK_DIR="sdk/python-gen"
TEMPLATE_DIR=".openapi-generator"

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_info "🔧 Post-processing generated Python SDK..."

# Check if SDK directory exists
if [ ! -d "$SDK_DIR" ]; then
    log_warn "SDK directory not found: $SDK_DIR"
    exit 1
fi

cd "$SDK_DIR"

# Backup original pyproject.toml if it exists
if [ -f "pyproject.toml" ]; then
    log_info "Backing up original pyproject.toml"
    cp pyproject.toml pyproject.toml.backup
fi


# Fix package directory name if needed
if [ -d "rcabench" ] && [ ! -d "openapi" ]; then
    log_info "Renaming package directory from 'rcabench' to 'openapi'"
    mv rcabench openapi
fi

# Update imports in __init__.py if needed
if [ -f "openapi/__init__.py" ]; then
    log_info "Updating package imports"
    sed -i 's/from rcabench\./from rcabench.openapi./g' openapi/__init__.py
fi

