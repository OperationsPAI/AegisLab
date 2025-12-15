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