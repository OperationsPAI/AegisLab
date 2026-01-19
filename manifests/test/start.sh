#!/bin/bash
set -e  # Exit on error

# Retry function for helm install commands
# --atomic flag automatically cleans up on failure, no manual uninstall needed
# Usage: retry_helm_install <max_attempts> <command...>
retry_helm_install() {
    local max_attempts=$1
    shift 1
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        echo "Attempt $attempt/$max_attempts..."
        if "$@"; then
            return 0
        else
            if [ $attempt -lt $max_attempts ]; then
                echo "⚠️  Attempt $attempt failed (--atomic cleaned up automatically)..."
                echo "🔄 Retrying in 5 seconds..."
                sleep 5
            else
                echo "❌ All $max_attempts attempts failed"
                
                # Final cleanup
                kind delete cluster -n test
                
                return 1
            fi
        fi
        ((attempt++))
    done
}

echo ""
echo "============================================="
echo "Starting Kubernetes test cluster setup"
echo "============================================="
echo ""

echo "Creating Kind cluster..."
HTTP_PROXY=http://crash:crash@172.18.0.1:7890 \
HTTPS_PROXY=http://crash:crash@172.18.0.1:7890 \
NO_PROXY=localhost,127.0.0.1,10.96.0.0/12,172.18.0.0/16,cluster.local,svc \
kind create cluster --config=manifests/test/kind-config.yaml --name test
kubectx kind-test
echo "✅ Kind cluster created successfully"
echo ""

# Install chaos-mesh
echo "Installing Chaos Mesh..."
helm repo add chaos-mesh https://charts.chaos-mesh.org --force-update
retry_helm_install 3 helm install chaos-mesh chaos-mesh/chaos-mesh \
	--namespace chaos-mesh \
	--create-namespace \
	--version 2.8.0 \
    -f manifests/cn_mirror/chaos-mesh.yaml \
	--atomic \
	--timeout 10m
echo "✅ Chaos Mesh installed successfully"
echo ""

echo "Applying Chaos Mesh RBAC configuration..."
kubectl apply -f manifests/chaos-mesh/rbac.yaml
echo "✅ Chaos Mesh RBAC applied"
echo ""

# Install cert-manager (required by otel-kube-stack)
echo "Installing cert-manager..."
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/latest/download/cert-manager.yaml
echo "Waiting for cert-manager to be ready..."
kubectl wait --for=condition=available --timeout=5m deployment/cert-manager -n cert-manager
kubectl wait --for=condition=available --timeout=5m deployment/cert-manager-webhook -n cert-manager
echo "✅ cert-manager is ready"
echo ""

# Install clickhouse and juicefs in parallel (independent components)
echo "Installing ClickHouse and JuiceFS CSI Driver in parallel..."
(
  echo "  Installing ClickHouse stack..."
  helm repo add clickstack https://hyperdxio.github.io/helm-charts --force-update
  retry_helm_install 3 helm install clickstack clickstack/clickstack \
      --namespace monitoring \
      --create-namespace \
      -f manifests/cn_mirror/click-stack.yaml \
      --atomic \
      --timeout 5m
  echo "✅ ClickHouse stack installed"
) &
CLICKHOUSE_PID=$!

(
  echo "  Installing JuiceFS CSI Driver..."
  helm repo add juicefs https://juicedata.github.io/charts --force-update
  retry_helm_install 3 helm install juicefs-csi-driver juicefs/juicefs-csi-driver \
      --namespace kube-system \
      --atomic \
      --timeout 5m
  echo "✅ JuiceFS CSI Driver installed"
) &
JUICEFS_PID=$!

# Wait for parallel installations
wait $CLICKHOUSE_PID
wait $JUICEFS_PID
echo "✅ Parallel installations completed"
echo ""

# Install otel-kube-stack (depends on monitoring namespace from clickstack)
echo "Installing OpenTelemetry Kube Stack..."
helm repo add open-telemetry https://open-telemetry.github.io/opentelemetry-helm-charts --force-update
retry_helm_install 3 helm install opentelemetry-kube-stack open-telemetry/opentelemetry-kube-stack \
    --namespace monitoring \
    --create-namespace \
    -f manifests/cn_mirror/otel-kube-stack.yaml \
    --atomic \
    --timeout 5m
echo "✅ OpenTelemetry Kube Stack installed"
echo ""

# Install otel-demo pedestal (depends on monitoring namespace)
echo "Installing OpenTelemetry Demo application..."
helm repo add opentelemetry-demo https://lgu-se-internal.github.io/opentelemetry-demo --force-update
retry_helm_install 3 helm install otel-demo0 opentelemetry-demo/opentelemetry-demo \
    --namespace otel-demo0 \
    --create-namespace \
    -f data/initial_data/prod/otel-demo.yaml \
    --atomic \
    --timeout 10m
echo "✅ OpenTelemetry Demo installed"
echo ""

echo "============================================="
echo "✅ Test cluster setup completed successfully!"
echo "============================================="