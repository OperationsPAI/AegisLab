#!/bin/bash
set -e
set -u

# 1. Ensure pathing for uv and go binaries (fixes "command not found" in non-interactive SSH)
export PATH=$PATH:$HOME/.local/bin:$(go env GOPATH)/bin
PROJECT_ROOT=$(pwd)

DEFAULT_VERSION=$(git rev-parse --short HEAD 2>/dev/null || echo "0.0.0")
VERSION=${SDK_VERSION:-${1:-$DEFAULT_VERSION}}

echo -e "\nCreating Command Environment..."
cd "$PROJECT_ROOT/scripts/command"
uv venv --clear
source .venv/bin/activate
uv sync --no-group test --group dev
echo "✅ Generated Command Environment"

echo -e "\nGenerating Python SDK..."
go install github.com/swaggo/swag/cmd/swag@latest
cd "$PROJECT_ROOT"
just generate-python-sdk "$VERSION"
echo "✅ Generated Python SDK..."

echo -e "\nRunning regression tests..."
cd "$PROJECT_ROOT/scripts/command"
uv sync --no-group dev --group test
set +e
uv run pytest test/test_workflow.py -v -s --tb=short --color=yes --capture=no
TEST_RESULT=$?
set -e
echo "✅ Tests completed"

echo -e "\nCleaning up..."
kubectl delete jobs -n exp --all || true 
rm -rf .venv
cd "$PROJECT_ROOT"
rm -rf sdk/python/src/rcabench/openapi
echo "✅ Cleaned up test server environment"

exit $TEST_RESULT
