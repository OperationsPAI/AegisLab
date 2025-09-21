# Installation Guide

This guide provides detailed instructions for installing and configuring RCABench in various environments.

## Table of Contents

1. [System Requirements](#system-requirements)
2. [Local Development Setup](#local-development-setup)
3. [Kubernetes Deployment](#kubernetes-deployment)
4. [Configuration](#configuration)
5. [Verification](#verification)
6. [Troubleshooting Installation](#troubleshooting-installation)

## System Requirements

### Minimum Requirements

- **CPU**: 2 cores
- **Memory**: 4GB RAM
- **Storage**: 10GB available disk space
- **OS**: Linux, macOS, or Windows with WSL2

### Recommended Requirements

- **CPU**: 4+ cores
- **Memory**: 8GB+ RAM  
- **Storage**: 20GB+ available SSD storage
- **OS**: Linux (Ubuntu 20.04+ or CentOS 8+)

### Software Dependencies

#### Required
- **Docker** (>= 20.10)
- **kubectl** (compatible with your Kubernetes version)
- **Git** (for cloning the repository)

#### For Kubernetes Deployment
- **Kubernetes** (>= 1.25) or **kind/minikube** for local development
- **Helm** (>= 3.0) - optional but recommended

#### For Development
- **Go** (>= 1.23)
- **Python** (>= 3.8)
- **Make** (for using Makefile commands)

## Local Development Setup

### Option 1: Docker Compose (Recommended for Testing)

This is the easiest way to get started with RCABench for testing and development.

#### Step 1: Clone Repository

```bash
# Clone the repository
git clone <repository-url>
cd rcabench
```

#### Step 2: Start Services

```bash
# Start all services with Docker Compose
make local-debug

# This will start:
# - MySQL database
# - Redis cache
# - Jaeger tracing
# - BuildKit daemon
# - RCABench API server
```

#### Step 3: Verify Installation

```bash
# Check if all services are running
docker ps

# Test API access
curl http://localhost:8082/health

# Access Swagger documentation
open http://localhost:8082/swagger/index.html
```

### Option 2: Manual Setup

For development or when you need more control over the components.

#### Step 1: Start Infrastructure Services

```bash
# Start MySQL
docker run -d --name mysql \
  -e MYSQL_DATABASE=rcabench \
  -e MYSQL_ROOT_PASSWORD=yourpassword \
  -p 3306:3306 \
  mysql:8.0.43

# Start Redis
docker run -d --name redis \
  -p 6379:6379 \
  redis:8.0-M02-alpine3.20

# Start Jaeger
docker run -d --name jaeger \
  -p 16686:16686 \
  -p 14268:14268 \
  -p 4318:4318 \
  jaegertracing/all-in-one:latest
```

#### Step 2: Configure RCABench

```bash
# Copy configuration template
cp src/config.dev.toml src/config.toml

# Edit configuration as needed
vim src/config.toml
```

#### Step 3: Build and Run RCABench

```bash
# Build the application
cd src
go build -o rcabench main.go

# Run the application
./rcabench
```

## Kubernetes Deployment

### Prerequisites

Ensure you have a Kubernetes cluster ready:

#### Local Cluster (kind)

```bash
# Install kind
curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.20.0/kind-linux-amd64
chmod +x ./kind
sudo mv ./kind /usr/local/bin/kind

# Create cluster
kind create cluster --name rcabench

# Set kubectl context
kubectl cluster-info --context kind-rcabench
```

#### Local Cluster (minikube)

```bash
# Install minikube
curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64
sudo install minikube-linux-amd64 /usr/local/bin/minikube

# Start cluster
minikube start --memory=8192 --cpus=4

# Enable required addons
minikube addons enable ingress
```

### Deployment Steps

#### Step 1: Install Dependencies

```bash
# Check prerequisites
make check-prerequisites

# Install missing tools if needed
# For skaffold (if not already installed):
curl -Lo skaffold https://storage.googleapis.com/skaffold/releases/latest/skaffold-linux-amd64
sudo install skaffold /usr/local/bin/
```

#### Step 2: Configure Storage

For production deployments, set up persistent storage:

```bash
# Create namespace
kubectl create namespace exp

# Apply persistent volume configuration
# Edit scripts/k8s/pv.yaml to match your storage setup
kubectl apply -f scripts/k8s/pv.yaml
```

Example PV configuration for local storage:

```yaml
apiVersion: v1
kind: PersistentVolume
metadata:
  name: rcabench-dataset-pv
spec:
  capacity:
    storage: 50Gi
  accessModes:
    - ReadWriteMany
  persistentVolumeReclaimPolicy: Retain
  storageClassName: local-storage
  local:
    path: /mnt/data/rcabench_dataset
  nodeAffinity:
    required:
      nodeSelectorTerms:
      - matchExpressions:
        - key: kubernetes.io/hostname
          operator: In
          values:
          - your-node-name
```

#### Step 3: Deploy Application

```bash
# Deploy with default configuration
make run

# Or deploy with custom registry
make run DEFAULT_REPO=your-registry.com/library

# Check deployment status
make status
```

#### Step 4: Configure Ingress (Optional)

```bash
# Apply ingress configuration
kubectl apply -f - <<EOF
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: rcabench-ingress
  namespace: exp
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
spec:
  rules:
  - host: rcabench.local
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: rcabench-service
            port:
              number: 8082
EOF

# Add to /etc/hosts (for local access)
echo "127.0.0.1 rcabench.local" | sudo tee -a /etc/hosts
```

## Configuration

### Environment Configuration

RCABench uses TOML configuration files. The main configuration sections are:

#### Database Configuration

```toml
[database]
mysql_user = "root"
mysql_password = "yourpassword"
mysql_host = "mysql-service"  # Kubernetes service name
mysql_port = "3306"
mysql_db = "rcabench"
```

#### Redis Configuration

```toml
[redis]
host = "redis-service:6379"  # Kubernetes service name
```

#### Kubernetes Configuration

```toml
[k8s]
namespace = "exp"  # Target namespace for experiments

[injection]
benchmark = ["workload1", "workload2"]
target_label_key = "app"

[injection.namespace_config.exp]
count = 5
port = "80%02d"
```

#### Rate Limiting

```toml
[rate_limiting]
max_concurrent_builds = 3
max_concurrent_restarts = 2
max_concurrent_algo_execution = 5
token_wait_timeout = 10
```

### Environment Variables

Override configuration with environment variables:

```bash
export RCABENCH_DATABASE_MYSQL_PASSWORD="newpassword"
export RCABENCH_REDIS_HOST="redis.example.com:6379"
export RCABENCH_K8S_NAMESPACE="production"
```

### Secret Management

For production deployments, use Kubernetes secrets:

```bash
# Create database secret
kubectl create secret generic rcabench-db-secret \
  --from-literal=password=yourpassword \
  -n exp

# Create registry secret (if using private registry)
kubectl create secret docker-registry rcabench-registry-secret \
  --docker-server=your-registry.com \
  --docker-username=username \
  --docker-password=password \
  -n exp
```

## Verification

### Health Checks

```bash
# Check API health
curl http://localhost:8082/health

# Expected response:
# {"status": "ok", "version": "1.0.1"}
```

### Service Verification

```bash
# Check all pods are running
kubectl get pods -n exp

# Check services
kubectl get services -n exp

# Check logs
kubectl logs -f deployment/rcabench -n exp
```

### Functional Testing

```bash
# Test API endpoints
curl http://localhost:8082/api/v1/algorithms

# Test database connection
curl http://localhost:8082/api/v1/datasets

# Test fault injection capabilities
curl -X POST http://localhost:8082/api/v1/injection/test
```

### Python SDK Verification

```bash
# Install Python SDK
cd sdk/python
pip install -e .

# Test SDK
python -c "
from rcabench import RCABenchSDK
sdk = RCABenchSDK('http://localhost:8082')
print('SDK connection successful:', sdk.health_check())
"
```

## Troubleshooting Installation

### Common Issues

#### 1. Database Connection Failed

**Symptoms**: 
- API returns database connection errors
- Health check fails

**Solutions**:
```bash
# Check MySQL pod status
kubectl get pods -l app=mysql -n exp

# Check MySQL logs
kubectl logs deployment/mysql -n exp

# Verify database credentials
kubectl get secret rcabench-db-secret -o yaml -n exp

# Test database connection manually
kubectl exec -it deployment/mysql -n exp -- mysql -u root -p rcabench
```

#### 2. Pod Scheduling Issues

**Symptoms**:
- Pods stuck in Pending state
- Resource allocation errors

**Solutions**:
```bash
# Check node resources
kubectl describe nodes

# Check pod events
kubectl describe pod <pod-name> -n exp

# Check resource requests/limits
kubectl describe deployment rcabench -n exp
```

#### 3. Image Pull Errors

**Symptoms**:
- ImagePullBackOff errors
- ErrImagePull status

**Solutions**:
```bash
# Check image availability
docker pull <image-name>

# Verify registry credentials
kubectl get secret rcabench-registry-secret -n exp

# Check imagePullPolicy
kubectl get deployment rcabench -o yaml -n exp
```

#### 4. Permission Errors

**Symptoms**:
- RBAC permission denied
- Unauthorized access errors

**Solutions**:
```bash
# Check current permissions
kubectl auth can-i create pods --namespace=exp

# Apply RBAC configuration
kubectl apply -f manifests/rbac.yaml

# Check service account
kubectl get serviceaccount -n exp
```

#### 5. Network Connectivity Issues

**Symptoms**:
- Service unreachable errors
- Timeout errors

**Solutions**:
```bash
# Check service endpoints
kubectl get endpoints -n exp

# Test service connectivity
kubectl run test-pod --rm -i --tty --image=nicolaka/netshoot -- /bin/bash
nslookup rcabench-service.exp.svc.cluster.local

# Check network policies
kubectl get networkpolicies -n exp
```

### Getting Help

If you encounter issues not covered here:

1. **Check logs**: `make logs` or `kubectl logs -f deployment/rcabench -n exp`
2. **Verify configuration**: Review your `config.toml` file
3. **Resource monitoring**: Check CPU/memory usage with `kubectl top pods -n exp`
4. **Documentation**: Refer to the [Troubleshooting Guide](troubleshooting.md)

### Performance Tuning

For production deployments:

```yaml
# Recommended resource limits
resources:
  requests:
    cpu: 500m
    memory: 1Gi
  limits:
    cpu: 2000m
    memory: 4Gi

# Enable horizontal pod autoscaling
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: rcabench-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: rcabench
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
```

## Next Steps

After successful installation:

1. **Read the [User Guide](user-guide.md)** to understand how to use RCABench
2. **Review [Examples](examples.md)** for practical usage scenarios
3. **Check [API Reference](api-reference.md)** for programmatic access
4. **Explore [Algorithm Development](algorithm-development.md)** to implement custom RCA algorithms