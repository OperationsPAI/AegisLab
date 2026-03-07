#!/bin/bash
set -e
set -u

# 1. Ensure pathing for uv and go binaries (fixes "command not found" in non-interactive SSH)
export PATH=$PATH:$HOME/.local/bin:$(go env GOPATH)/bin
PROJECT_ROOT=$(pwd)

echo -e "\nCreating Command Environment..."
cd "$PROJECT_ROOT/scripts/command"
uv venv --clear
source .venv/bin/activate
uv sync --no-group test --group dev
echo "✅ Generated Command Environment"

echo -e "\nGenerating Python SDK..."
go install github.com/swaggo/swag/cmd/swag@latest
cd "$PROJECT_ROOT"
just generate-python-sdk 0.0.0
echo "✅ Generated Python SDK..."

echo -e "\nRunning regression tests..."
cd "$PROJECT_ROOT/scripts/command"
uv sync --no-group dev --group test
set +e
uv run pytest test/test_workflow.py -v -s --tb=short --color=yes --capture=no TEST_RESULT=$?
set -e
echo "✅ Tests completed"

if [ $TEST_RESULT -eq 0 ]; then
    echo "✅ Tests passed! Pushing image to registry..."
    docker push pair-diag-cn-guangzhou.cr.volces.com/pair/rcabench:latest
else
    echo "❌ Tests failed. Skipping image push."
fi

echo -e "\nCleaning up..."
kubectl delete jobs -n exp --all || true 
rm -rf .venv
cd "$PROJECT_ROOT"
rm -rf sdk/python
echo "✅ Cleaned up test server environment"

exit $TEST_RESULT