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

# Get version from main.go
MAIN_GO="../../experiments_controller/main.go"
if [ -f "$MAIN_GO" ]; then
    VERSION=$(grep -o '@version[[:space:]]*[0-9]\+\.[0-9]\+\.[0-9]\+' "$MAIN_GO" | sed 's/@version[[:space:]]*//')
    if [ -z "$VERSION" ]; then
        VERSION="1.0.0"
        log_warn "Could not extract version from main.go, using default: $VERSION"
    else
        log_info "Extracted version from main.go: $VERSION"
    fi
else
    VERSION="1.0.0"
    log_warn "main.go not found, using default version: $VERSION"
fi

# Create modern pyproject.toml
log_info "Creating modern pyproject.toml with version $VERSION"
cat > pyproject.toml << EOF
[project]
name = "rcabench"
version = "$VERSION"
description = "RCABench API Python Client - A comprehensive root cause analysis benchmarking platform"
readme = "README.md"
license = {text = "MIT"}
authors = [
    {name = "RCABench Team", email = "team@rcabench.com"}
]
maintainers = [
    {name = "RCABench Team", email = "team@rcabench.com"}
]
keywords = [
    "rcabench",
    "root-cause-analysis", 
    "microservices",
    "benchmarking",
    "monitoring",
    "observability",
    "fault-injection",
    "openapi",
    "api-client"
]
classifiers = [
    "Development Status :: 4 - Beta",
    "Intended Audience :: Developers",
    "Intended Audience :: System Administrators",
    "License :: OSI Approved :: MIT License",
    "Operating System :: OS Independent",
    "Programming Language :: Python :: 3",
    "Programming Language :: Python :: 3.9",
    "Programming Language :: Python :: 3.10",
    "Programming Language :: Python :: 3.11",
    "Programming Language :: Python :: 3.12",
    "Topic :: Software Development :: Libraries :: Python Modules",
    "Topic :: System :: Monitoring",
    "Topic :: System :: Systems Administration",
]
requires-python = ">=3.9"
dependencies = [
    "urllib3 >= 2.1.0, < 3.0.0",
    "python-dateutil >= 2.8.2",
    "pydantic >= 2",
    "typing-extensions >= 4.7.1",
]

[project.optional-dependencies]
dev = [
    "pytest >= 7.2.1",
    "pytest-cov >= 2.8.1",
    "tox >= 3.9.0",
    "flake8 >= 4.0.0",
    "types-python-dateutil >= 2.8.19.14",
    "mypy >= 1.5",
    "build",
    "twine",
]

[project.urls]
Homepage = "https://github.com/rcabench/rcabench"
Documentation = "https://rcabench.readthedocs.io/"
Repository = "https://github.com/rcabench/rcabench"
"Bug Tracker" = "https://github.com/rcabench/rcabench/issues"
Changelog = "https://github.com/rcabench/rcabench/blob/main/CHANGELOG.md"

[build-system]
requires = ["hatchling"]
build-backend = "hatchling.build"

[tool.hatch.build.targets.wheel]
packages = ["openapi"]

[tool.hatch.build.targets.sdist]
include = [
    "/openapi",
    "/README.md",
    "/LICENSE",
]
exclude = [
    "/.git",
    "/test",
    "/.github",
]

[tool.pylint.'MESSAGES CONTROL']
extension-pkg-whitelist = "pydantic"

[tool.mypy]
files = [
  "openapi",
  "tests",
]
warn_unused_configs = true
warn_redundant_casts = true
warn_unused_ignores = true
strict_equality = true
extra_checks = true
check_untyped_defs = true
disallow_subclassing_any = true
disallow_untyped_decorators = true
disallow_any_generics = true
EOF

# Fix package directory name if needed
if [ -d "rcabench" ] && [ ! -d "openapi" ]; then
    log_info "Renaming package directory from 'rcabench' to 'openapi'"
    mv rcabench openapi
fi

# Update imports in __init__.py if needed
if [ -f "openapi/__init__.py" ]; then
    log_info "Updating package imports"
    sed -i 's/from rcabench\./from openapi./g' openapi/__init__.py
fi

# Create a simple setup.py for backward compatibility (optional)
log_info "Creating setup.py for backward compatibility"
cat > setup.py << 'EOF'
# -*- coding: utf-8 -*-
# Backward compatibility setup.py - use pyproject.toml for configuration

from setuptools import setup

if __name__ == "__main__":
    setup()
EOF

# Make sure publish.sh is executable
if [ -f "publish.sh" ]; then
    chmod +x publish.sh
fi

# Create .gitignore if it doesn't exist
if [ ! -f ".gitignore" ]; then
    log_info "Creating .gitignore"
    cat > .gitignore << 'EOF'
# Byte-compiled / optimized / DLL files
__pycache__/
*.py[cod]
*$py.class

# Distribution / packaging
.Python
build/
develop-eggs/
dist/
downloads/
eggs/
.eggs/
lib/
lib64/
parts/
sdist/
var/
wheels/
*.egg-info/
.installed.cfg
*.egg
MANIFEST

# PyInstaller
*.manifest
*.spec

# Unit test / coverage reports
htmlcov/
.tox/
.coverage
.coverage.*
.cache
nosetests.xml
coverage.xml
*.cover
.hypothesis/
.pytest_cache/

# Environments
.env
.venv
env/
venv/
ENV/
env.bak/
venv.bak/

# mypy
.mypy_cache/
.dmypy.json
dmypy.json

# IDEs
.vscode/
.idea/
*.swp
*.swo
*~

# OS
.DS_Store
Thumbs.db
EOF
fi

log_info "✅ SDK post-processing completed successfully!"
log_info "📦 Package name: rcabench"
log_info "🏷️  Version: $VERSION"
log_info "🔧 Build system: hatchling"
log_info ""
log_info "Next steps:"
log_info "  cd $SDK_DIR"
log_info "  ./test-build.sh      # Test the build"
log_info "  ./publish.sh --test  # Publish to Test PyPI"
