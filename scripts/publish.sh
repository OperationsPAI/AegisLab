#!/bin/bash
set -e

CHART_DIR="helm"
REPO_NAME="rcabench"
REPO_URL="https://lgu-se-internal.github.io/rcabench"

helm dependency update $CHART_DIR

# 打包 Chart
mkdir -p .deploy
helm package $CHART_DIR -d .deploy

# 生成或更新 index.yaml
cd .deploy
if [ -f index.yaml ]; then
    helm repo index . --url $REPO_URL --merge index.yaml
else
    helm repo index . --url $REPO_URL
fi
cd ..