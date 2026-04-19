#!/bin/bash
set -e  # Exit on error

# Get ENV_MODE parameter (default: test)
ENV_MODE=${1:-test}

CERT_MANAGER_MANIFEST_URL=${CERT_MANAGER_MANIFEST_URL:-"https://github.com/cert-manager/cert-manager/releases/latest/download/cert-manager.yaml"}
CHAOS_MESH_REPO_URL=${CHAOS_MESH_REPO_URL:-"https://charts.chaos-mesh.org"}
CLICKSTACK_REPO_URL=${CLICKSTACK_REPO_URL:-"https://hyperdxio.github.io/helm-charts"}
OPEN_TELEMETRY_REPO_URL=${OPEN_TELEMETRY_REPO_URL:-"https://open-telemetry.github.io/opentelemetry-helm-charts"}
OTEL_DEMO_REPO_URL=${OTEL_DEMO_REPO_URL:-"https://operationspai.github.io/opentelemetry-demo"}
JUICEFS_REPO_URL=${JUICEFS_REPO_URL:-"https://juicedata.github.io/charts"}
TEST_HTTP_PROXY=${TEST_HTTP_PROXY:-"http://crash:crash@172.18.0.1:7890"}
TEST_HTTPS_PROXY=${TEST_HTTPS_PROXY:-"http://crash:crash@172.18.0.1:7890"}
TEST_NO_PROXY=${TEST_NO_PROXY:-"localhost,127.0.0.1,10.96.0.0/12,172.18.0.0/16,cluster.local,svc"}

echo "Running in $ENV_MODE mode"
echo ""

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
echo "Starting Kubernetes cluster setup ($ENV_MODE mode)"
echo "============================================="
echo ""

if [ "$ENV_MODE" = "prod" ]; then
    # ========================================
    # Production Mode Workflow
    # ========================================
    echo "🚀 Running production workflow..."
    echo ""
    
    # Install chaos-mesh
    echo "Installing Chaos Mesh..."
    helm repo add chaos-mesh "$CHAOS_MESH_REPO_URL" --force-update
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
    
    # Install cert-manager
    echo "Installing cert-manager..."
    kubectl apply -f "$CERT_MANAGER_MANIFEST_URL"
    echo "Waiting for cert-manager to be ready..."
    kubectl wait --for=condition=available --timeout=5m deployment/cert-manager -n cert-manager
    kubectl wait --for=condition=available --timeout=5m deployment/cert-manager-webhook -n cert-manager
    echo "✅ cert-manager is ready"
    echo ""
    
    # Install ClickHouse only (no JuiceFS in prod)
    echo "Installing ClickHouse stack..."
    helm repo add clickstack "$CLICKSTACK_REPO_URL" --force-update
    retry_helm_install 3 helm install clickstack clickstack/clickstack \
        --namespace monitoring \
        --create-namespace \
        -f manifests/cn_mirror/click-stack.yaml \
        --set global.storageClassName=rcabench \
        --atomic \
        --timeout 5m
    echo "✅ ClickHouse stack installed"
    echo ""
    
    # Install otel-kube-stack
    echo "Installing OpenTelemetry Kube Stack..."
    helm repo add open-telemetry "$OPEN_TELEMETRY_REPO_URL" --force-update
    retry_helm_install 3 helm install opentelemetry-kube-stack open-telemetry/opentelemetry-kube-stack \
        --namespace monitoring \
        --create-namespace \
        -f manifests/cn_mirror/otel-kube-stack.yaml \
        --atomic \
        --timeout 5m
    echo "✅ OpenTelemetry Kube Stack installed"
    echo ""
    
    # Install otel-demo
    echo "Installing OpenTelemetry Demo application..."
    helm repo add opentelemetry-demo "$OTEL_DEMO_REPO_URL" --force-update
    retry_helm_install 3 helm install otel-demo0 opentelemetry-demo/opentelemetry-demo \
        --namespace otel-demo0 \
        --create-namespace \
        -f data/initial_data/prod/otel-demo.yaml \
        --atomic \
        --timeout 10m
    echo "✅ OpenTelemetry Demo installed"
    echo ""
    
else
    # ========================================
    # Test Mode Workflow (default)
    # ========================================
    echo "🧪 Running test workflow..."
    echo ""
    
    # Create Kind cluster
    echo "Creating Kind cluster..."
    HTTP_PROXY="$TEST_HTTP_PROXY" \
    HTTPS_PROXY="$TEST_HTTPS_PROXY" \
    NO_PROXY="$TEST_NO_PROXY" \
    kind create cluster --config=manifests/test/kind-config.yaml --name test
    kubectx kind-test
    echo "✅ Kind cluster created successfully"
    echo ""
    
    # Install chaos-mesh
    echo "Installing Chaos Mesh..."
    helm repo add chaos-mesh "$CHAOS_MESH_REPO_URL" --force-update
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
    
    # Install cert-manager
    echo "Installing cert-manager..."
    kubectl apply -f "$CERT_MANAGER_MANIFEST_URL"
    echo "Waiting for cert-manager to be ready..."
    kubectl wait --for=condition=available --timeout=5m deployment/cert-manager -n cert-manager
    kubectl wait --for=condition=available --timeout=5m deployment/cert-manager-webhook -n cert-manager
    echo "✅ cert-manager is ready"
    echo ""
    
    # Install ClickHouse and JuiceFS in parallel
    echo "Installing ClickHouse and JuiceFS CSI Driver in parallel..."
    (
      echo "  Installing ClickHouse stack..."
      helm repo add clickstack "$CLICKSTACK_REPO_URL" --force-update
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
      helm repo add juicefs "$JUICEFS_REPO_URL" --force-update
      retry_helm_install 3 helm install juicefs-csi-driver juicefs/juicefs-csi-driver \
          --namespace kube-system \
          -f manifests/cn_mirror/juicefs-csi-driver.yaml \
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
    
    # Install otel-kube-stack
    echo "Installing OpenTelemetry Kube Stack..."
    helm repo add open-telemetry "$OPEN_TELEMETRY_REPO_URL" --force-update
    retry_helm_install 3 helm install opentelemetry-kube-stack open-telemetry/opentelemetry-kube-stack \
        --namespace monitoring \
        --create-namespace \
        -f manifests/cn_mirror/otel-kube-stack.yaml \
        --atomic \
        --timeout 5m
    echo "✅ OpenTelemetry Kube Stack installed"
    echo ""
    
    # Install otel-demo
    echo "Installing OpenTelemetry Demo application..."
    helm repo add opentelemetry-demo "$OTEL_DEMO_REPO_URL" --force-update
    retry_helm_install 3 helm install otel-demo0 opentelemetry-demo/opentelemetry-demo \
        --namespace otel-demo0 \
        --create-namespace \
        -f data/initial_data/prod/otel-demo.yaml \
        --atomic \
        --timeout 10m
    echo "✅ OpenTelemetry Demo installed"
    echo ""
fi

echo "============================================="
echo "✅ Cluster setup completed successfully!"
echo "============================================="
