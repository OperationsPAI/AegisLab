#!/bin/bash

echo "🧹 Cleaning up old port forwards..."
pkill -9 -f "kubectl port-forward" 2>/dev/null || true

for port in $(lsof -nP -iTCP -sTCP:LISTEN 2>/dev/null | grep -E ":1[0-9]{4,5}" | awk '{print $9}' | cut -d':' -f2 | sort -u); do
  pid=$(lsof -ti:$port 2>/dev/null || true)
  [ -n "$pid" ] && kill -9 $pid 2>/dev/null || true
done

if [ -d "$HOME/.vscode-server" ]; then
  rm -f $HOME/.vscode-server/data/Machine/port-* 2>/dev/null || true
fi

sleep 2
echo "✓ Old forwards cleaned"

echo ""
echo "🚀 [exp] Forwarding all services..."
services=$(kubectl get svc -n exp -o json | jq -r '.items[] | "\(.metadata.name):\(.spec.ports | map(.port) | join(","))"')

for svc_entry in $services; do
  name="${svc_entry%%:*}"
  ports="${svc_entry#*:}"
  
  IFS=',' read -ra PORT_ARRAY <<< "$ports"
  for port in "${PORT_ARRAY[@]}"; do
    # Calculate local port with overflow protection
    local_port="1${port}"
    if [ "$local_port" -gt 65535 ]; then
      # If port would overflow, map to 10000 + (port % 55535) range
      local_port=$((10000 + (port % 55535)))
      echo "   [exp] $name:$port -> localhost:$local_port (remapped due to overflow)"
    else
      echo "   [exp] $name:$port -> localhost:$local_port"
    fi
    kubectl port-forward -n exp svc/$name --address 0.0.0.0 $local_port:$port >/dev/null 2>&1 &
    sleep 0.1
  done
done

echo ""
echo "🚀 [monitoring] Forwarding ClickHouse..."
clickhouse_ports=$(kubectl get svc clickstack-clickhouse -n monitoring -o json | jq -r '.spec.ports | map(.port) | join(",")')

IFS=',' read -ra PORT_ARRAY <<< "$clickhouse_ports"
for port in "${PORT_ARRAY[@]}"; do
  # Calculate local port with overflow protection
  local_port="1${port}"
  if [ "$local_port" -gt 65535 ]; then
    # If port would overflow, map to 10000 + (port % 55535) range
    local_port=$((10000 + (port % 55535)))
    echo "   [monitoring] clickstack-clickhouse:$port -> localhost:$local_port (remapped due to overflow)"
  else
    echo "   [monitoring] clickstack-clickhouse:$port -> localhost:$local_port"
  fi
  kubectl port-forward -n monitoring svc/clickstack-clickhouse --address 0.0.0.0 $local_port:$port >/dev/null 2>&1 &
  sleep 0.1
done

echo ""
echo "✅ Done! Forwarded:"
echo "   • exp namespace: all services"
echo "   • monitoring namespace: clickstack-clickhouse (ports: 18123, 19000, 19363)"