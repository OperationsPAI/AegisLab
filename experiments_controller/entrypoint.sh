#!/bin/bash
set -e
set -u
set -o pipefail

if [ -z "$(ls -A /app/algorithms)" ]; then
    echo "Algorithms directory is empty, copying legacy algorithms...";
    # Iterate through each file in the legacy_algorithms directory and only copy if it doesn't exist
    for file in /app/legacy_algorithms/*; do
        filename=$(basename "$file")
        if [ ! -e "/app/algorithms/$filename" ]; then
            cp -r "$file" "/app/algorithms/"
            echo "Copied: $filename"
        else
            echo "File exists, skipping: $filename"
        fi
    done
    echo "Legacy algorithms copied successfully.";
else
    echo "Algorithms directory is not empty, checking for missing files...";
    # Check for files that need to be copied even if the directory is not empty
    for file in /app/legacy_algorithms/*; do
        filename=$(basename "$file")
        if [ ! -e "/app/algorithms/$filename" ]; then
            cp -r "$file" "/app/algorithms/"
            echo "Copied missing file: $filename"
        fi
    done
fi

# Set default values
MODE=${1:-"producer"}
PORT=${2:-"8081"}

# Execute command with arguments
/app/exp $MODE --port $PORT