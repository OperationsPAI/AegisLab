#!/bin/bash

# Exit codes
STATUS_SUCCESS=0
STATUS_FAILURE=1

# Colors for output
GREEN='\033[32m'
BLUE='\033[34m'
RED='\033[31m'
RESET='\033[0m'

# Check the ENV_MODE environment variable, defaults to dev
ENV_MODE=${ENV_MODE:-dev}

# Path variables (relative to the script execution directory, which is scripts/command)
COMMAND_BIN="command.bin"
VENV_DIR=".venv"
MAIN_SCRIPT="main.py"

# If the script is run from project_root, dirname "$0" returns "scripts".
# We need to cd into "scripts/command".
if ! cd "$(dirname "$0")/command"; then
    printf "${RED}❌ Failed to change directory to command folder.${RESET}\n"
    exit $STATUS_FAILURE
fi

# --- Development Mode (ENV_MODE!=1) ---
if [ "$ENV_MODE" != "prod" ]; then
    printf "${BLUE}📦 Checking development environment (ENV_MODE=${ENV_MODE})...${RESET}\n"
    
    # Check if virtual environment exists
    if [ -d "$VENV_DIR" ]; then
        printf "${GREEN}✅ Virtual environment found. Skipping setup.${RESET}\n"
        exit $STATUS_SUCCESS
    else
        printf "${BLUE}📦 Virtual environment not found. Setting up now...${RESET}\n"
        
        if ! uv venv; then
            printf "${RED}❌ uv venv failed${RESET}\n"
            exit $STATUS_FAILURE
        fi

        . "$VENV_DIR/bin/activate"

        if uv sync --quiet; then
            printf "${GREEN}✅ Dependencies installed successfully.${RESET}\n"
            exit $STATUS_SUCCESS
        else
            printf "${RED}❌ uv sync failed${RESET}\n"
            exit $STATUS_FAILURE
        fi
    fi
fi

# --- Production/Compilation Mode (ENV_MODE=prod) ---
# Check if the compiled binary already exists
if [ -f "$COMMAND_BIN" ]; then
    printf "${GREEN}✅ Command tool binary found. Skipping compilation.${RESET}\n"
    exit $STATUS_SUCCESS
fi

printf "${BLUE}📦 Command tool binary not found. Building now...${RESET}\n"

# Install system dependencies required for compilation (e.g., Nuitka post-processing)
if ! sudo apt install -y patchelf ccache > /dev/null; then
    printf "${RED}❌ Failed to install system dependencies (patchelf, ccache)${RESET}\n"
    exit $STATUS_FAILURE
fi

# Clear and set up the compilation environment
if ! uv venv --clear; then
    printf "${RED}❌ uv venv --clear failed${RESET}\n"
    exit $STATUS_FAILURE
fi

. "$VENV_DIR/bin/activate"

# Install Python dependencies including Nuitka build extras
if uv sync --quiet --extra nuitka-build; then
    printf "${GREEN}✅ Dependencies installed successfully for compilation.${RESET}\n"

    # Run Nuitka compilation
    if uv run python -m nuitka --standalone --onefile --lto=yes \
        --output-dir=. \
        --output-filename="$COMMAND_BIN" \
        "$MAIN_SCRIPT"; then
        
        printf "${GREEN}✅ Command tool compilation completed.${RESET}\n"
        exit $STATUS_SUCCESS
    else
        printf "${RED}❌ Nuitka compilation failed${RESET}\n"
        exit $STATUS_FAILURE
    fi
else
    printf "${RED}❌ uv sync failed${RESET}\n"
    exit $STATUS_FAILURE
fi