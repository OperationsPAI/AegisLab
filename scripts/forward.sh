#!/bin/bash

# Environment configuration
# prod: 1 prefix (e.g., 18080, 19000)
# test: 2 prefix (e.g., 28080, 29000)
ENV="${1:-prod}"  # Default to prod if no argument provided
NAMESPACE="exp"

case "$ENV" in
  prod)
    PORT_PREFIX="1"
    ;;
  test)
    PORT_PREFIX="2"
    ;;
  *)
    echo "❌ Invalid environment: $ENV"
    echo "Usage: $0 {prod|test}"
    echo "  prod - Forward with 1xxxx ports (exp namespace)"
    echo "  test - Forward with 2xxxx ports (exp-dev namespace)"
    exit 1
    ;;
esac

echo "🔧 Environment: $ENV"
echo "🔧 Namespace: $NAMESPACE"
echo "🔧 Port prefix: ${PORT_PREFIX}xxxx"
echo ""

# Function to clean up old port forwards
cleanup_forwards() {
  echo "🧹 Cleaning up old port forwards..."
  pkill -9 -f "kubectl port-forward" 2>/dev/null || true

  # Clean up ports based on environment prefix
  for port in $(lsof -nP -iTCP -sTCP:LISTEN 2>/dev/null | grep -E ":${PORT_PREFIX}[0-9]{4,5}" | awk '{print $9}' | cut -d':' -f2 | sort -u); do
    pid=$(lsof -ti:$port 2>/dev/null || true)
    [ -n "$pid" ] && kill -9 $pid 2>/dev/null || true
  done

  if [ -d "$HOME/.vscode-server" ]; then
    rm -f $HOME/.vscode-server/data/Machine/port-* 2>/dev/null || true
  fi

  sleep 2
  echo "✓ Old forwards cleaned"
  echo ""
}

# Function to forward services in a namespace
forward_namespace_services() {
  local namespace=$1
  local prefix=$2
  local label=$3
  
  echo "🚀 [$label] Forwarding all services..."
  services=$(kubectl get svc -n $namespace -o json | jq -r '.items[] | "\(.metadata.name):\(.spec.ports | map(.port) | join(","))"')

  for svc_entry in $services; do
    name="${svc_entry%%:*}"
    ports="${svc_entry#*:}"
    
    IFS=',' read -ra PORT_ARRAY <<< "$ports"
    for port in "${PORT_ARRAY[@]}"; do
      # Calculate local port with overflow protection
      local_port="${prefix}${port}"
      if [ "$local_port" -gt 65535 ]; then
        # If port would overflow, map to prefix0000 + (port % 55535) range
        local_port=$((${prefix}0000 + (port % 55535)))
        echo "   [$label] $name:$port -> localhost:$local_port (remapped due to overflow)"
      else
        echo "   [$label] $name:$port -> localhost:$local_port"
      fi
      kubectl port-forward -n $namespace svc/$name --address 0.0.0.0 $local_port:$port >/dev/null 2>&1 &
      sleep 0.1
    done
  done
}

# Function to forward ClickHouse in monitoring namespace
forward_clickhouse() {
  local prefix=$1
  
  echo ""
  echo "🚀 [monitoring] Forwarding ClickHouse..."
  clickhouse_ports=$(kubectl get svc clickstack-clickhouse -n monitoring -o json | jq -r '.spec.ports | map(.port) | join(",")')

  IFS=',' read -ra PORT_ARRAY <<< "$clickhouse_ports"
  for port in "${PORT_ARRAY[@]}"; do
    # Calculate local port with overflow protection
    local_port="${prefix}${port}"
    if [ "$local_port" -gt 65535 ]; then
      # If port would overflow, map to prefix0000 + (port % 55535) range
      local_port=$((${prefix}0000 + (port % 55535)))
      echo "   [monitoring] clickstack-clickhouse:$port -> localhost:$local_port (remapped due to overflow)"
    else
      echo "   [monitoring] clickstack-clickhouse:$port -> localhost:$local_port"
    fi
    kubectl port-forward -n monitoring svc/clickstack-clickhouse --address 0.0.0.0 $local_port:$port >/dev/null 2>&1 &
    sleep 0.1
  done
}

# Main execution
cleanup_forwards
forward_namespace_services "$NAMESPACE" "$PORT_PREFIX" "$NAMESPACE"
forward_clickhouse "$PORT_PREFIX"

echo ""
echo "✅ Done! Forwarded:"
echo "   • $NAMESPACE namespace: all services (${PORT_PREFIX}xxxx ports)"
echo "   • monitoring namespace: clickstack-clickhouse (${PORT_PREFIX}8123, ${PORT_PREFIX}9000, ${PORT_PREFIX}9363)"