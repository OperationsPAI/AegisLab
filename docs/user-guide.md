# User Guide

This guide provides comprehensive instructions for using RCABench to conduct root cause analysis experiments and evaluations.

## Table of Contents

1. [Introduction](#introduction)
2. [Getting Started](#getting-started)
3. [Platform Overview](#platform-overview)
4. [Basic Workflow](#basic-workflow)
5. [Fault Injection](#fault-injection)
6. [Algorithm Execution](#algorithm-execution)
7. [Evaluation and Analysis](#evaluation-and-analysis)
8. [Advanced Usage](#advanced-usage)

## Introduction

RCABench is designed to facilitate research in root cause analysis for microservices. It provides a standardized platform for:

- **Reproducible Experiments**: Consistent fault injection and evaluation environments
- **Algorithm Comparison**: Fair benchmarking across different RCA approaches
- **Dataset Management**: Organized collection and sharing of experimental data
- **Automated Evaluation**: Standardized metrics for algorithm performance

## Getting Started

### System Requirements

Before using RCABench, ensure your system meets the following requirements:

- **Kubernetes cluster** (local or remote)
- **Docker** for containerized workloads
- **Python 3.8+** for SDK usage
- **8GB+ RAM** for running microservice workloads
- **20GB+ storage** for datasets and logs

### Initial Setup

1. **Install RCABench** following the [Installation Guide](installation.md)
2. **Configure your environment** by editing `src/config.toml`
3. **Deploy target workloads** (e.g., microservice applications)
4. **Verify installation** by accessing the API at `http://localhost:8082`

## Platform Overview

### Core Components

#### API Server
The central component that orchestrates all operations:
- **Endpoint**: `http://localhost:8082`
- **Documentation**: Available at `/swagger/index.html`
- **Health Check**: Available at `/health`

#### Fault Injection Engine
Chaos engineering capabilities for microservices:
- **Supported Types**: Network, Pod, Stress, Time, DNS, HTTP, JVM chaos
- **Integration**: Kubernetes-native chaos operators
- **Configuration**: YAML-based fault specifications

#### Algorithm Registry
Manages and executes RCA algorithms:
- **Algorithm Storage**: Container-based algorithm packaging
- **Execution Environment**: Isolated containers with access to data
- **Result Collection**: Standardized output formats

#### Evaluation Framework
Automated evaluation and metrics calculation:
- **Standard Metrics**: Precision, recall, F1-score, accuracy
- **Custom Metrics**: Extensible framework for domain-specific metrics
- **Comparison Tools**: Side-by-side algorithm performance analysis

## Basic Workflow

### 1. Prepare Target Workload

Deploy a microservice application to your Kubernetes cluster:

```bash
# Example: Deploy a sample workload
kubectl apply -f workloads/sample-app/

# Verify deployment
kubectl get pods -l app=sample-app
```

### 2. Configure Fault Scenarios

Define fault injection scenarios in your configuration:

```toml
[injection]
benchmark = ["sample-app"]
target_label_key = "app"

[injection.namespace_config.default]
count = 3
port = "80%02d"
```

### 3. Execute Experiments

Use the Python SDK or REST API to run experiments:

```python
from rcabench import RCABenchSDK

sdk = RCABenchSDK("http://localhost:8082")

# Step 1: Inject fault
injection_result = sdk.injection.execute([{
    "duration": 300,
    "faultType": 5,  # CPU stress
    "injectNamespace": "default",
    "injectPod": "sample-app-pod",
    "spec": {"CPULoad": 80}
}])

# Step 2: Run RCA algorithm
algorithm_result = sdk.algorithm.execute([{
    "benchmark": "sample-app",
    "algorithm": "my-rca-algorithm",
    "dataset": "experiment-dataset"
}])

# Step 3: Evaluate results
evaluation = sdk.evaluation.calculate_metrics(algorithm_result)
```

## Fault Injection

### Supported Fault Types

#### Network Chaos
Simulates network-related issues:

```python
network_fault = {
    "duration": 300,
    "faultType": 1,  # Network chaos
    "injectNamespace": "default",
    "injectPod": "target-service",
    "spec": {
        "action": "delay",
        "delay": "100ms",
        "correlation": "100"
    }
}
```

#### Pod Chaos
Simulates pod failures:

```python
pod_fault = {
    "duration": 60,
    "faultType": 2,  # Pod chaos
    "injectNamespace": "default",
    "injectPod": "target-service",
    "spec": {
        "action": "pod-failure",
        "duration": "30s"
    }
}
```

#### Stress Chaos
Simulates resource exhaustion:

```python
stress_fault = {
    "duration": 600,
    "faultType": 5,  # Stress chaos
    "injectNamespace": "default",
    "injectPod": "target-service",
    "spec": {
        "CPULoad": 80,
        "CPUWorker": 2,
        "MemoryLoad": 70
    }
}
```

### Fault Injection Best Practices

1. **Start Small**: Begin with low-impact faults to understand system behavior
2. **Monitor Impact**: Watch system metrics during fault injection
3. **Document Scenarios**: Keep detailed records of fault configurations
4. **Safety First**: Always have rollback procedures ready

## Algorithm Execution

### Algorithm Requirements

RCA algorithms must be packaged as Docker containers with:

1. **Standard Interface**: Accept input via environment variables
2. **Data Access**: Read observability data from mounted volumes
3. **Output Format**: Produce results in JSON format
4. **Error Handling**: Graceful handling of invalid inputs

### Algorithm Registration

Register your algorithm with the platform:

```python
# Register a new algorithm
sdk.algorithm.register({
    "name": "my-rca-algorithm",
    "version": "1.0.0",
    "image": "my-registry/my-rca-algorithm:1.0.0",
    "description": "Description of the algorithm",
    "parameters": {
        "threshold": {"type": "float", "default": 0.5},
        "window_size": {"type": "int", "default": 300}
    }
})
```

### Execution Parameters

Configure algorithm execution:

```python
execution_config = {
    "benchmark": "target-workload",
    "algorithm": "my-rca-algorithm",
    "dataset": "fault-scenario-data",
    "parameters": {
        "threshold": 0.7,
        "window_size": 600
    },
    "timeout": 3600  # 1 hour timeout
}
```

## Evaluation and Analysis

### Standard Metrics

RCABench provides standard evaluation metrics:

#### Accuracy Metrics
- **Precision**: Ratio of correctly identified root causes
- **Recall**: Ratio of actual root causes detected
- **F1-Score**: Harmonic mean of precision and recall
- **Accuracy**: Overall correctness of predictions

#### Performance Metrics
- **Detection Time**: Time from fault injection to detection
- **Diagnosis Time**: Time from detection to root cause identification
- **Throughput**: Number of diagnoses per unit time
- **Resource Usage**: CPU, memory, and storage consumption

### Custom Evaluation

Implement domain-specific evaluation metrics:

```python
def custom_evaluation(ground_truth, predictions):
    """Custom evaluation function"""
    # Implement your evaluation logic
    return {
        "custom_metric_1": value1,
        "custom_metric_2": value2
    }

# Register custom evaluation
sdk.evaluation.register_custom_metric("my_metric", custom_evaluation)
```

### Result Analysis

Access and analyze experimental results:

```python
# Get experiment results
results = sdk.evaluation.get_experiment_results(experiment_id)

# Generate comparison report
comparison = sdk.evaluation.compare_algorithms([
    "algorithm1", "algorithm2", "algorithm3"
], dataset="benchmark-dataset")

# Export results
sdk.evaluation.export_results(results, format="csv", output="results.csv")
```

## Advanced Usage

### Batch Experiments

Run multiple experiments in batch:

```python
experiments = [
    {"algorithm": "alg1", "dataset": "dataset1"},
    {"algorithm": "alg1", "dataset": "dataset2"},
    {"algorithm": "alg2", "dataset": "dataset1"},
]

batch_results = sdk.batch.execute_experiments(experiments)
```

### Custom Workloads

Deploy and manage custom workloads:

```python
# Deploy custom workload
workload_config = {
    "name": "my-workload",
    "namespace": "experiments",
    "replicas": 3,
    "image": "my-registry/my-app:latest"
}

sdk.workload.deploy(workload_config)
```

### Data Management

Manage datasets and experimental data:

```python
# Create dataset
dataset = sdk.dataset.create({
    "name": "experiment-20240101",
    "description": "Network latency experiment",
    "tags": ["network", "latency", "microservices"]
})

# Upload data
sdk.dataset.upload_traces(dataset_id, "traces.json")
sdk.dataset.upload_metrics(dataset_id, "metrics.json")
sdk.dataset.upload_logs(dataset_id, "logs.json")
```

### Monitoring and Observability

Access system monitoring data:

```python
# Get system metrics
metrics = sdk.monitoring.get_metrics(
    start_time="2024-01-01T00:00:00Z",
    end_time="2024-01-01T23:59:59Z"
)

# Get trace data
traces = sdk.monitoring.get_traces(service="target-service")

# Get logs
logs = sdk.monitoring.get_logs(
    namespace="default",
    pod="target-pod",
    since="1h"
)
```

## Tips and Best Practices

### Experiment Design
1. **Control Variables**: Keep most parameters constant when comparing algorithms
2. **Statistical Significance**: Run multiple iterations for reliable results
3. **Baseline Comparison**: Always include baseline/random algorithms
4. **Documentation**: Maintain detailed experiment logs

### Resource Management
1. **Resource Limits**: Set appropriate CPU/memory limits for algorithms
2. **Cleanup**: Clean up resources after experiments
3. **Storage**: Monitor disk usage for large datasets
4. **Scaling**: Use horizontal scaling for large experiments

### Troubleshooting
1. **Check Logs**: Always examine logs for error details
2. **Resource Monitoring**: Monitor system resources during experiments
3. **Network Connectivity**: Verify network connectivity between components
4. **Permissions**: Ensure proper RBAC permissions for operations

For more detailed troubleshooting, see the [Troubleshooting Guide](troubleshooting.md).