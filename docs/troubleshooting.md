# Troubleshooting Guide

This guide helps diagnose and resolve common issues when using RCABench.

## Table of Contents

1. [General Troubleshooting](#general-troubleshooting)
2. [Installation Issues](#installation-issues)
3. [API and Connectivity Issues](#api-and-connectivity-issues)
4. [Fault Injection Problems](#fault-injection-problems)
5. [Algorithm Execution Issues](#algorithm-execution-issues)
6. [Performance Problems](#performance-problems)
7. [Data and Storage Issues](#data-and-storage-issues)
8. [Kubernetes-Specific Issues](#kubernetes-specific-issues)

## General Troubleshooting

### Step 1: Check System Health

```bash
# Check API health
curl http://localhost:8082/health

# Expected response:
# {"status": "ok", "version": "1.0.1", "uptime": 3600}

# Check system resources
kubectl top pods -n exp
kubectl top nodes

# Check application logs
make logs
# or
kubectl logs -f deployment/rcabench -n exp
```

### Step 2: Verify Configuration

```bash
# Check configuration file
cat src/config.toml

# Verify environment variables
kubectl get configmap rcabench-config -o yaml -n exp

# Check secrets
kubectl get secrets -n exp
```

### Step 3: Network Connectivity

```bash
# Test internal service connectivity
kubectl run test-pod --rm -i --tty --image=nicolaka/netshoot -- /bin/bash

# Inside the test pod:
nslookup rcabench-service.exp.svc.cluster.local
curl http://rcabench-service.exp.svc.cluster.local:8082/health
```

## Installation Issues

### Issue: Docker Compose Fails to Start

**Symptoms:**
- Services fail to start with docker-compose
- Port binding errors
- Volume mount failures

**Diagnosis:**
```bash
# Check Docker daemon status
docker info

# Check for port conflicts
netstat -tulpn | grep :8082
netstat -tulpn | grep :3306
netstat -tulpn | grep :6379

# Check Docker logs
docker-compose logs mysql
docker-compose logs redis
```

**Solutions:**

1. **Port Conflicts:**
```bash
# Stop conflicting services
sudo systemctl stop mysql
sudo systemctl stop redis

# Or change ports in docker-compose.yaml
```

2. **Permission Issues:**
```bash
# Fix Docker permissions
sudo usermod -aG docker $USER
newgrp docker

# Fix volume permissions
sudo chown -R $USER:$USER ./data
```

3. **Insufficient Resources:**
```bash
# Increase Docker resources
# Edit Docker Desktop settings: Memory > 8GB, Swap > 2GB
```

### Issue: Kubernetes Deployment Fails

**Symptoms:**
- Pods stuck in Pending/ImagePullBackOff state
- RBAC permission errors
- Resource allocation failures

**Diagnosis:**
```bash
# Check pod status
kubectl get pods -n exp
kubectl describe pod <pod-name> -n exp

# Check events
kubectl get events -n exp --sort-by='.lastTimestamp'

# Check resource quotas
kubectl describe resourcequota -n exp
```

**Solutions:**

1. **Image Pull Issues:**
```bash
# Check image registry access
docker pull <image-name>

# Create registry secret
kubectl create secret docker-registry registry-secret \
  --docker-server=<registry> \
  --docker-username=<username> \
  --docker-password=<password> \
  -n exp

# Update deployment to use secret
kubectl patch deployment rcabench -p '{"spec":{"template":{"spec":{"imagePullSecrets":[{"name":"registry-secret"}]}}}}' -n exp
```

2. **RBAC Issues:**
```bash
# Check current permissions
kubectl auth can-i create pods --namespace=exp

# Apply RBAC configuration
kubectl apply -f manifests/rbac.yaml

# Check service account
kubectl get serviceaccount -n exp
kubectl describe serviceaccount rcabench -n exp
```

3. **Resource Issues:**
```bash
# Check node resources
kubectl describe nodes

# Reduce resource requests
kubectl patch deployment rcabench -p '{"spec":{"template":{"spec":{"containers":[{"name":"rcabench","resources":{"requests":{"memory":"512Mi","cpu":"250m"}}}]}}}}' -n exp
```

## API and Connectivity Issues

### Issue: API Returns 500 Internal Server Error

**Symptoms:**
- HTTP 500 responses from API endpoints
- Database connection errors in logs
- Service unavailable errors

**Diagnosis:**
```bash
# Check API logs
kubectl logs deployment/rcabench -n exp | grep ERROR

# Check database connectivity
kubectl exec -it deployment/mysql -n exp -- mysql -u root -p rcabench

# Check Redis connectivity  
kubectl exec -it deployment/redis -n exp -- redis-cli ping
```

**Solutions:**

1. **Database Connection Issues:**
```bash
# Check database credentials
kubectl get secret mysql-secret -o yaml -n exp

# Verify database service
kubectl get service mysql -n exp
kubectl describe service mysql -n exp

# Test connection manually
kubectl run mysql-client --rm -i --tty --image=mysql:8.0 -- mysql -h mysql.exp.svc.cluster.local -u root -p
```

2. **Configuration Issues:**
```bash
# Check configuration
kubectl get configmap rcabench-config -o yaml -n exp

# Update configuration
kubectl create configmap rcabench-config --from-file=config.toml -n exp --dry-run=client -o yaml | kubectl apply -f -

# Restart deployment
kubectl rollout restart deployment/rcabench -n exp
```

### Issue: SDK Connection Timeout

**Symptoms:**
- Python SDK raises timeout errors
- Connection refused errors
- Slow API response times

**Diagnosis:**
```python
from rcabench import RCABenchSDK
import requests

# Test direct HTTP connection
response = requests.get("http://localhost:8082/health", timeout=30)
print(response.status_code, response.text)

# Test SDK with debug
sdk = RCABenchSDK("http://localhost:8082", timeout=60)
```

**Solutions:**

1. **Increase Timeouts:**
```python
# Increase SDK timeout
sdk = RCABenchSDK("http://localhost:8082", timeout=120)

# Or set per-operation timeout
result = sdk.algorithm.execute(requests, timeout=300)
```

2. **Check Network Latency:**
```bash
# Test latency to API
ping rcabench.local
curl -w "@curl-format.txt" -o /dev/null -s http://localhost:8082/health

# Create curl-format.txt:
echo "time_namelookup: %{time_namelookup}\ntime_connect: %{time_connect}\ntime_total: %{time_total}" > curl-format.txt
```

## Fault Injection Problems

### Issue: Fault Injection Fails to Start

**Symptoms:**
- Injection requests return errors
- Target pods not affected by faults
- Chaos resources not created

**Diagnosis:**
```bash
# Check chaos resources
kubectl get podchaos -A
kubectl get networkchaos -A
kubectl get stresschaos -A

# Check chaos-mesh operator
kubectl get pods -n chaos-engineering

# Check injection logs
kubectl logs -f -l app=rcabench -n exp | grep injection
```

**Solutions:**

1. **Install Chaos Mesh:**
```bash
# Install chaos-mesh operator
curl -sSL https://mirrors.chaos-mesh.org/v2.6.2/install.sh | bash

# Verify installation
kubectl get pods -n chaos-engineering
```

2. **RBAC Permissions:**
```bash
# Grant chaos permissions to service account
kubectl apply -f - <<EOF
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: rcabench-chaos
rules:
- apiGroups: ["chaos-mesh.org"]
  resources: ["*"]
  verbs: ["*"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: rcabench-chaos
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: rcabench-chaos
subjects:
- kind: ServiceAccount
  name: rcabench
  namespace: exp
EOF
```

3. **Target Pod Selection:**
```bash
# Check target pods exist
kubectl get pods -l app=target-service -n production

# Verify label selectors
kubectl get pods --show-labels -n production

# Test chaos resource manually
kubectl apply -f - <<EOF
apiVersion: chaos-mesh.org/v1alpha1
kind: PodChaos
metadata:
  name: test-chaos
  namespace: exp
spec:
  action: pod-kill
  mode: one
  selector:
    namespaces:
      - production
    labelSelectors:
      app: target-service
  duration: "30s"
EOF
```

### Issue: Fault Injection Affects Wrong Services

**Symptoms:**
- Unintended services affected by faults
- Fault isolation failures
- Cascading failures

**Solutions:**

1. **Improve Target Selection:**
```python
# Use more specific selectors
fault_request = {
    "duration": 300,
    "faultType": 5,
    "injectNamespace": "production",
    "injectPod": "api-service-v2",  # Specific pod name
    "spec": {
        "CPULoad": 50,  # Reduced load
        "CPUWorker": 1  # Fewer workers
    },
    "selector": {
        "labelSelectors": {
            "app": "api-service",
            "version": "v2"
        }
    }
}
```

2. **Use Namespace Isolation:**
```bash
# Create isolated namespace for experiments
kubectl create namespace experiment-1
kubectl label namespace experiment-1 chaos.alpha.kubernetes.io/inject=enabled
```

## Algorithm Execution Issues

### Issue: Algorithm Execution Hangs

**Symptoms:**
- Algorithm tasks stuck in "running" state
- No output produced after timeout
- High resource usage

**Diagnosis:**
```bash
# Check algorithm pod status
kubectl get pods -l job-name=algorithm-task-12345

# Check pod logs
kubectl logs -f algorithm-task-12345-xxxx

# Check resource usage
kubectl top pod algorithm-task-12345-xxxx
```

**Solutions:**

1. **Increase Timeout:**
```python
# Increase execution timeout
algorithm_request = [{
    "algorithm": "my-algorithm",
    "dataset": "large-dataset", 
    "benchmark": "app",
    "timeout": 7200  # 2 hours
}]
```

2. **Optimize Algorithm:**
```python
# Add progress logging to algorithm
import logging

class MyAlgorithm:
    def analyze(self, data):
        logging.info("Starting analysis...")
        
        # Process in chunks
        chunk_size = 1000
        total_items = len(data)
        
        for i in range(0, total_items, chunk_size):
            chunk = data[i:i+chunk_size]
            self.process_chunk(chunk)
            
            # Log progress
            progress = (i + chunk_size) / total_items * 100
            logging.info(f"Progress: {progress:.1f}%")
```

### Issue: Algorithm Produces Invalid Output

**Symptoms:**
- Output validation failures
- Missing required fields in results
- Incorrect confidence scores

**Solutions:**

1. **Add Output Validation:**
```python
def validate_output(self, results):
    required_fields = ["root_cause", "confidence", "execution_time"]
    
    for field in required_fields:
        if field not in results:
            raise ValueError(f"Missing required field: {field}")
    
    if not 0 <= results["confidence"] <= 1:
        raise ValueError("Confidence must be between 0 and 1")
    
    return True
```

2. **Debug Algorithm Logic:**
```python
# Add debug logging
def analyze(self, data):
    self.logger.debug(f"Input data keys: {list(data.keys())}")
    self.logger.debug(f"Metrics count: {len(data.get('metrics', {}))}")
    
    # Your analysis logic here
    results = self.run_analysis(data)
    
    self.logger.debug(f"Analysis results: {results}")
    return results
```

## Performance Problems

### Issue: Slow API Response Times

**Symptoms:**
- API requests take > 30 seconds
- Timeout errors from clients
- High CPU/memory usage

**Diagnosis:**
```bash
# Check resource usage
kubectl top pods -n exp

# Check API performance
curl -w "@curl-format.txt" -o /dev/null -s http://localhost:8082/api/v1/algorithms

# Check database performance
kubectl exec -it deployment/mysql -n exp -- mysql -u root -p -e "SHOW PROCESSLIST;"
```

**Solutions:**

1. **Scale Resources:**
```bash
# Increase pod resources
kubectl patch deployment rcabench -p '{"spec":{"template":{"spec":{"containers":[{"name":"rcabench","resources":{"requests":{"memory":"2Gi","cpu":"1000m"},"limits":{"memory":"4Gi","cpu":"2000m"}}}]}}}}' -n exp

# Scale replicas
kubectl scale deployment rcabench --replicas=3 -n exp
```

2. **Database Optimization:**
```sql
-- Add database indexes
CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_created_at ON tasks(created_at);

-- Check slow queries
SHOW FULL PROCESSLIST;
```

3. **Enable Caching:**
```toml
# In config.toml
[redis]
host = "redis:6379"
cache_ttl = 3600

[performance]
enable_caching = true
cache_algorithm_results = true
```

### Issue: High Memory Usage

**Symptoms:**
- Pods killed with OOMKilled status
- Memory usage constantly increasing
- System becomes unresponsive

**Solutions:**

1. **Optimize Data Processing:**
```python
# Process data in smaller chunks
def process_large_dataset(data, chunk_size=1000):
    for i in range(0, len(data), chunk_size):
        chunk = data[i:i+chunk_size]
        result = process_chunk(chunk)
        # Clear chunk from memory
        del chunk
        yield result
```

2. **Memory Profiling:**
```python
import psutil
import gc

def monitor_memory():
    process = psutil.Process()
    memory_info = process.memory_info()
    print(f"Memory usage: {memory_info.rss / 1024 / 1024:.1f} MB")
    
    # Force garbage collection
    gc.collect()
```

## Data and Storage Issues

### Issue: Dataset Upload Failures

**Symptoms:**
- Upload requests timeout
- Data corruption during upload
- Storage quota exceeded

**Solutions:**

1. **Check Storage Permissions:**
```bash
# Check PV/PVC status
kubectl get pv,pvc -n exp

# Check storage permissions
kubectl exec -it deployment/rcabench -n exp -- ls -la /data

# Fix permissions
kubectl exec -it deployment/rcabench -n exp -- chown -R app:app /data
```

2. **Increase Upload Limits:**
```toml
# In config.toml
[api]
max_upload_size = "10GB"
upload_timeout = 3600

[storage]
data_retention_days = 30
cleanup_old_datasets = true
```

### Issue: Data Format Errors

**Symptoms:**
- Algorithms fail with data parsing errors
- Validation errors during data loading
- Inconsistent data formats

**Solutions:**

1. **Data Validation:**
```python
import jsonschema

# Define data schema
metrics_schema = {
    "type": "object",
    "patternProperties": {
        ".*": {  # Service name
            "type": "object",
            "properties": {
                "cpu_usage": {"type": "array", "items": {"type": "number"}},
                "memory_usage": {"type": "array", "items": {"type": "number"}},
                "timestamps": {"type": "array", "items": {"type": "string"}}
            },
            "required": ["cpu_usage", "timestamps"]
        }
    }
}

# Validate data
def validate_metrics(data):
    try:
        jsonschema.validate(data, metrics_schema)
        return True
    except jsonschema.ValidationError as e:
        print(f"Validation error: {e}")
        return False
```

2. **Data Conversion:**
```python
# Convert data to standard format
def standardize_metrics(raw_metrics):
    standardized = {}
    
    for service, metrics in raw_metrics.items():
        # Ensure all metrics are lists
        for metric_name, values in metrics.items():
            if not isinstance(values, list):
                metrics[metric_name] = [values]
        
        # Ensure timestamps are strings
        if 'timestamps' in metrics:
            metrics['timestamps'] = [str(ts) for ts in metrics['timestamps']]
        
        standardized[service] = metrics
    
    return standardized
```

## Kubernetes-Specific Issues

### Issue: Pod Scheduling Problems

**Symptoms:**
- Pods stuck in Pending state
- Node affinity conflicts
- Resource constraints

**Diagnosis:**
```bash
# Check pod scheduling events
kubectl describe pod <pod-name> -n exp

# Check node resources
kubectl describe nodes

# Check taints and tolerations
kubectl get nodes -o custom-columns=NAME:.metadata.name,TAINTS:.spec.taints
```

**Solutions:**

1. **Node Selector Issues:**
```yaml
# Update deployment with correct node selector
apiVersion: apps/v1
kind: Deployment
metadata:
  name: rcabench
spec:
  template:
    spec:
      nodeSelector:
        kubernetes.io/os: linux
        node-role.kubernetes.io/worker: "true"
```

2. **Resource Requests:**
```yaml
# Reduce resource requests
resources:
  requests:
    memory: "512Mi"
    cpu: "250m"
  limits:
    memory: "2Gi"
    cpu: "1000m"
```

### Issue: Service Discovery Problems

**Symptoms:**
- Services cannot connect to each other
- DNS resolution failures
- Network connectivity issues

**Solutions:**

1. **Check DNS Resolution:**
```bash
# Test DNS from inside a pod
kubectl run test-dns --rm -i --tty --image=busybox -- nslookup kubernetes.default.svc.cluster.local

# Check CoreDNS
kubectl get pods -n kube-system -l k8s-app=kube-dns
kubectl logs -f deployment/coredns -n kube-system
```

2. **Network Policies:**
```bash
# Check network policies
kubectl get networkpolicies -A

# Allow all traffic (for debugging)
kubectl apply -f - <<EOF
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-all
  namespace: exp
spec:
  podSelector: {}
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - {}
  egress:
  - {}
EOF
```

## Getting Additional Help

### Enable Debug Logging

```toml
# In config.toml
[debugging]
enable = true
log_level = "DEBUG"
log_format = "json"

[logging]
level = "DEBUG"
output = "stdout"
```

### Collect Diagnostic Information

```bash
#!/bin/bash
# diagnostic-script.sh

echo "=== RCABench Diagnostic Information ==="
echo "Date: $(date)"
echo "Kubernetes version: $(kubectl version --short)"
echo

echo "=== Pod Status ==="
kubectl get pods -n exp -o wide

echo "=== Service Status ==="
kubectl get services -n exp

echo "=== ConfigMaps ==="
kubectl get configmaps -n exp

echo "=== Secrets ==="
kubectl get secrets -n exp

echo "=== Events ==="
kubectl get events -n exp --sort-by='.lastTimestamp' | tail -20

echo "=== Resource Usage ==="
kubectl top pods -n exp
kubectl top nodes

echo "=== Logs (last 100 lines) ==="
kubectl logs deployment/rcabench -n exp --tail=100
```

### Contact Support

When reporting issues, please include:

1. **System Information**: Kubernetes version, node specifications
2. **Configuration**: Sanitized config.toml and environment variables
3. **Error Messages**: Complete error messages and stack traces
4. **Logs**: Relevant application and system logs
5. **Reproduction Steps**: Detailed steps to reproduce the issue
6. **Expected vs Actual**: What you expected vs what actually happened

For additional help:
- Check the [User Guide](user-guide.md) for usage instructions
- Review [API Reference](api-reference.md) for API details
- See [Examples](examples.md) for working code samples