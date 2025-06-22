#!/bin/bash

# RCABench SDK Setup Script
# This script installs the necessary dependencies for SDK generation and publishing

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_info "Setting up RCABench SDK development environment"

# Check if Go is installed
if ! command -v go &> /dev/null; then
    log_warn "Go is not installed. Please install Go first."
    exit 1
fi

# Install swag for Swagger generation
log_info "Installing Swagger generator (swag)"
go install github.com/swaggo/swag/cmd/swag@latest

# Check if Python is installed
if ! command -v python3 &> /dev/null; then
    log_warn "Python3 is not installed. Please install Python first."
    exit 1
fi

# Install Python build tools
log_info "Installing Python build tools"
uv add build twine hatchling

# Optional: Install uv for faster package management
if command -v uv &> /dev/null; then
    log_info "uv is available for faster package management"
else
    log_info "Consider installing uv for faster Python package management:"
    echo "  curl -LsSf https://astral.sh/uv/install.sh | sh"
fi

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    log_warn "Docker is not installed. Please install Docker for SDK generation."
else
    log_info "Docker is available for SDK generation"
fi

# Make scripts executable
log_info "Making scripts executable"
chmod +x scripts/release.sh
chmod +x sdk/python-gen/publish.sh

log_info "Setup completed successfully!"
echo ""
echo "You can now use the following commands:"
echo "  make swag-init       # Generate Swagger docs"
echo "  make generate-sdk    # Generate Python SDK"
echo "  make release VERSION=1.0.1  # Full release process"
echo ""
echo "For more information, see sdk/python-gen/RELEASE.md"
