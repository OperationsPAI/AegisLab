#!/bin/bash

# 用法:
# ./policy.sh apply     # 应用策略
# ./policy.sh delete    # 删除策略

set -e

if [[ "$#" -ne 1 ]]; then
  echo "Usage: $0 apply|delete"
  exit 1
fi

ACTION="$1"
NAMESPACES=("ts0" "ts1" "ts2" "ts3" "ts4" "ts5")
POLICY_FILE="L7-policy.yaml"
POLICY_NAME="allow-all-with-l7-observability"

for ns in "${NAMESPACES[@]}"; do
  echo "==> Processing namespace: $ns"
  if [[ "$ACTION" == "apply" ]]; then
    NAMESPACE="$ns" envsubst < "$POLICY_FILE" | kubectl apply -f -
  elif [[ "$ACTION" == "delete" ]]; then
    kubectl delete cnp "$POLICY_NAME" -n "$ns" --ignore-not-found
  else
    echo "Unknown action: $ACTION"
    exit 1
  fi
done