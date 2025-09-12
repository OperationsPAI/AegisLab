# Examples and Tutorials

This document provides practical examples for using RCABench in various scenarios.

## Table of Contents

1. [Basic Usage Examples](#basic-usage-examples)
2. [Fault Injection Scenarios](#fault-injection-scenarios)
3. [Algorithm Development Examples](#algorithm-development-examples)
4. [Evaluation Workflows](#evaluation-workflows)
5. [Advanced Use Cases](#advanced-use-cases)

## Basic Usage Examples

### Example 1: Simple Fault Injection and Detection

This example demonstrates a complete workflow from fault injection to algorithm execution.

#### Prerequisites
- RCABench installed and running
- A target microservice application deployed

#### Step 1: Setup Python Environment

```bash
# Install RCABench SDK
pip install rcabench

# Or install from source
cd sdk/python
pip install -e .
```

#### Step 2: Connect to RCABench

```python
from rcabench import RCABenchSDK
import time

# Initialize SDK
sdk = RCABenchSDK("http://localhost:8082")

# Verify connection
health = sdk.health_check()
print(f"RCABench status: {health}")
```

#### Step 3: List Available Resources

```python
# List available algorithms
algorithms = sdk.algorithm.list()
print("Available algorithms:")
for algo in algorithms:
    print(f"  - {algo['name']}: {algo['description']}")

# List available datasets
datasets = sdk.dataset.list()
print("Available datasets:")
for dataset in datasets:
    print(f"  - {dataset['name']}: {dataset['description']}")
```

#### Step 4: Inject a CPU Stress Fault

```python
# Define fault injection request
fault_request = [{
    "duration": 300,  # 5 minutes
    "faultType": 5,   # CPU stress fault
    "injectNamespace": "default",
    "injectPod": "target-service-pod",
    "spec": {
        "CPULoad": 80,    # 80% CPU load
        "CPUWorker": 2    # 2 worker threads
    },
    "benchmark": "my-app"
}]

# Execute fault injection
print("Injecting CPU stress fault...")
injection_result = sdk.injection.execute(fault_request)
print(f"Fault injection started: {injection_result}")

# Wait for fault to take effect
time.sleep(60)
```

#### Step 5: Run RCA Algorithm

```python
# Define algorithm execution request
algorithm_request = [{
    "benchmark": "my-app",
    "algorithm": "random-walk",  # Example algorithm
    "dataset": "live-data",
    "parameters": {
        "threshold": 0.7,
        "window_size": 300
    }
}]

# Execute algorithm
print("Running RCA algorithm...")
algorithm_result = sdk.algorithm.execute(algorithm_request)
print(f"Algorithm execution started: {algorithm_result}")

# Wait for algorithm to complete
time.sleep(120)

# Get results
results = sdk.algorithm.get_results(algorithm_result['task_id'])
print(f"Root cause analysis results: {results}")
```

## Fault Injection Scenarios

### Scenario 1: Network Latency Injection

Simulate network latency between microservices:

```python
network_latency_fault = [{
    "duration": 600,  # 10 minutes
    "faultType": 1,   # Network chaos
    "injectNamespace": "production",
    "injectPod": "api-gateway",
    "spec": {
        "action": "delay",
        "delay": "200ms",
        "correlation": "100",
        "jitter": "10ms"
    },
    "benchmark": "e-commerce-app"
}]

result = sdk.injection.execute(network_latency_fault)
```

### Scenario 2: Memory Pressure

Simulate memory pressure on a specific service:

```python
memory_pressure_fault = [{
    "duration": 900,  # 15 minutes
    "faultType": 5,   # Stress chaos
    "injectNamespace": "production",
    "injectPod": "database-service",
    "spec": {
        "MemoryLoad": 85,     # 85% memory usage
        "MemoryWorker": 1     # 1 worker thread
    },
    "benchmark": "database-workload"
}]

result = sdk.injection.execute(memory_pressure_fault)
```

### Scenario 3: Pod Failure

Simulate random pod failures:

```python
pod_failure_fault = [{
    "duration": 300,  # 5 minutes
    "faultType": 2,   # Pod chaos
    "injectNamespace": "production",
    "injectPod": "user-service",
    "spec": {
        "action": "pod-kill",
        "mode": "one"  # Kill one random pod
    },
    "benchmark": "user-management"
}]

result = sdk.injection.execute(pod_failure_fault)
```

### Scenario 4: DNS Resolution Failure

Simulate DNS resolution issues:

```python
dns_failure_fault = [{
    "duration": 400,  # ~7 minutes
    "faultType": 3,   # DNS chaos
    "injectNamespace": "production",
    "injectPod": "order-service",
    "spec": {
        "action": "error",
        "patterns": ["database.internal"],
        "mode": "all"
    },
    "benchmark": "order-processing"
}]

result = sdk.injection.execute(dns_failure_fault)
```

## Algorithm Development Examples

### Example 1: Simple Statistical Algorithm

Create a basic RCA algorithm using statistical methods:

```python
# algorithm.py
import json
import os
import numpy as np
from typing import Dict, List, Any

class SimpleStatisticalRCA:
    def __init__(self, threshold: float = 0.8):
        self.threshold = threshold
    
    def analyze(self, metrics: Dict[str, List[float]], 
                traces: List[Dict], logs: List[Dict]) -> Dict[str, Any]:
        """
        Simple statistical analysis for root cause detection
        """
        # Analyze metrics for anomalies
        anomalous_services = []
        
        for service, values in metrics.items():
            if not values:
                continue
                
            # Calculate z-score
            mean = np.mean(values)
            std = np.std(values)
            
            if std > 0:
                latest_values = values[-10:]  # Last 10 measurements
                z_scores = [(v - mean) / std for v in latest_values]
                
                # Check if any recent values are anomalous
                if any(abs(z) > 2.0 for z in z_scores):
                    anomalous_services.append({
                        "service": service,
                        "anomaly_score": max(abs(z) for z in z_scores),
                        "confidence": min(1.0, max(abs(z) for z in z_scores) / 3.0)
                    })
        
        # Sort by anomaly score
        anomalous_services.sort(key=lambda x: x["anomaly_score"], reverse=True)
        
        # Determine root cause
        if anomalous_services:
            root_cause = anomalous_services[0]
            if root_cause["confidence"] >= self.threshold:
                return {
                    "root_cause": root_cause["service"],
                    "confidence": root_cause["confidence"],
                    "anomaly_score": root_cause["anomaly_score"],
                    "all_anomalies": anomalous_services
                }
        
        return {
            "root_cause": "unknown",
            "confidence": 0.0,
            "anomaly_score": 0.0,
            "all_anomalies": anomalous_services
        }

def main():
    # Read input data from environment
    dataset_path = os.environ.get("DATASET_PATH", "/data")
    threshold = float(os.environ.get("THRESHOLD", "0.8"))
    
    # Load data
    with open(f"{dataset_path}/metrics.json", "r") as f:
        metrics = json.load(f)
    
    with open(f"{dataset_path}/traces.json", "r") as f:
        traces = json.load(f)
    
    with open(f"{dataset_path}/logs.json", "r") as f:
        logs = json.load(f)
    
    # Run analysis
    rca = SimpleStatisticalRCA(threshold=threshold)
    result = rca.analyze(metrics, traces, logs)
    
    # Output result
    with open("/output/result.json", "w") as f:
        json.dump(result, f, indent=2)
    
    print(f"Analysis complete. Root cause: {result['root_cause']}")

if __name__ == "__main__":
    main()
```

### Example 2: Dockerfile for Algorithm

```dockerfile
# Dockerfile
FROM python:3.9-slim

# Install dependencies
RUN pip install numpy pandas scikit-learn

# Copy algorithm code
COPY algorithm.py /app/algorithm.py

# Set working directory
WORKDIR /app

# Create output directory
RUN mkdir -p /output

# Run algorithm
CMD ["python", "algorithm.py"]
```

### Example 3: Algorithm Registration

```python
# Register the algorithm with RCABench
algorithm_config = {
    "name": "simple-statistical-rca",
    "version": "1.0.0",
    "image": "my-registry/simple-statistical-rca:1.0.0",
    "description": "Simple statistical RCA using z-score analysis",
    "parameters": {
        "threshold": {
            "type": "float",
            "default": 0.8,
            "description": "Confidence threshold for root cause detection"
        }
    },
    "input_format": {
        "metrics": "JSON file with service metrics",
        "traces": "JSON file with distributed traces",
        "logs": "JSON file with application logs"
    },
    "output_format": {
        "root_cause": "Identified root cause service",
        "confidence": "Confidence score (0-1)",
        "anomaly_score": "Statistical anomaly score"
    }
}

# Register with RCABench
result = sdk.algorithm.register(algorithm_config)
print(f"Algorithm registered: {result}")
```

## Evaluation Workflows

### Example 1: Single Algorithm Evaluation

```python
def evaluate_single_algorithm():
    """Evaluate a single algorithm across multiple datasets"""
    
    algorithm_name = "simple-statistical-rca"
    datasets = ["dataset1", "dataset2", "dataset3"]
    results = []
    
    for dataset in datasets:
        print(f"Evaluating {algorithm_name} on {dataset}")
        
        # Run algorithm
        execution_request = [{
            "benchmark": "evaluation-workload",
            "algorithm": algorithm_name,
            "dataset": dataset,
            "parameters": {"threshold": 0.8}
        }]
        
        exec_result = sdk.algorithm.execute(execution_request)
        task_id = exec_result['task_id']
        
        # Wait for completion
        status = sdk.algorithm.wait_for_completion(task_id, timeout=3600)
        
        if status['state'] == 'completed':
            # Get algorithm results
            algo_results = sdk.algorithm.get_results(task_id)
            
            # Get ground truth
            ground_truth = sdk.dataset.get_ground_truth(dataset)
            
            # Calculate evaluation metrics
            metrics = sdk.evaluation.calculate_metrics(
                predictions=algo_results,
                ground_truth=ground_truth
            )
            
            results.append({
                "dataset": dataset,
                "algorithm": algorithm_name,
                "metrics": metrics,
                "execution_time": status['execution_time']
            })
            
            print(f"Results for {dataset}: {metrics}")
        else:
            print(f"Execution failed for {dataset}: {status}")
    
    return results

# Run evaluation
evaluation_results = evaluate_single_algorithm()
```

### Example 2: Algorithm Comparison

```python
def compare_algorithms():
    """Compare multiple algorithms on the same datasets"""
    
    algorithms = [
        {"name": "simple-statistical-rca", "params": {"threshold": 0.8}},
        {"name": "random-walk", "params": {"steps": 100}},
        {"name": "correlation-analysis", "params": {"window": 300}}
    ]
    
    datasets = ["fault-scenario-1", "fault-scenario-2"]
    comparison_results = {}
    
    for dataset in datasets:
        comparison_results[dataset] = {}
        
        for algo_config in algorithms:
            algo_name = algo_config["name"]
            print(f"Running {algo_name} on {dataset}")
            
            # Execute algorithm
            execution_request = [{
                "benchmark": "comparison-workload",
                "algorithm": algo_name,
                "dataset": dataset,
                "parameters": algo_config["params"]
            }]
            
            exec_result = sdk.algorithm.execute(execution_request)
            
            # Wait and get results
            status = sdk.algorithm.wait_for_completion(
                exec_result['task_id'], timeout=3600
            )
            
            if status['state'] == 'completed':
                algo_results = sdk.algorithm.get_results(exec_result['task_id'])
                ground_truth = sdk.dataset.get_ground_truth(dataset)
                
                metrics = sdk.evaluation.calculate_metrics(
                    predictions=algo_results,
                    ground_truth=ground_truth
                )
                
                comparison_results[dataset][algo_name] = {
                    "metrics": metrics,
                    "execution_time": status['execution_time']
                }
    
    # Generate comparison report
    report = sdk.evaluation.generate_comparison_report(comparison_results)
    
    # Save results
    with open("algorithm_comparison.json", "w") as f:
        json.dump(comparison_results, f, indent=2)
    
    return comparison_results

# Run comparison
comparison_results = compare_algorithms()
```

### Example 3: Performance Benchmarking

```python
def benchmark_performance():
    """Benchmark algorithm performance under different loads"""
    
    algorithm_name = "simple-statistical-rca"
    load_scenarios = [
        {"services": 10, "duration": 300},
        {"services": 50, "duration": 300},
        {"services": 100, "duration": 300}
    ]
    
    performance_results = []
    
    for scenario in load_scenarios:
        print(f"Benchmarking with {scenario['services']} services")
        
        # Create synthetic dataset for the scenario
        dataset_name = f"perf-test-{scenario['services']}-services"
        
        # Run multiple iterations
        iteration_results = []
        for i in range(5):  # 5 iterations for statistical significance
            
            execution_request = [{
                "benchmark": "performance-test",
                "algorithm": algorithm_name,
                "dataset": dataset_name,
                "parameters": {"threshold": 0.8}
            }]
            
            start_time = time.time()
            exec_result = sdk.algorithm.execute(execution_request)
            
            status = sdk.algorithm.wait_for_completion(
                exec_result['task_id'], timeout=3600
            )
            
            end_time = time.time()
            
            if status['state'] == 'completed':
                iteration_results.append({
                    "execution_time": end_time - start_time,
                    "memory_usage": status.get('memory_usage', 0),
                    "cpu_usage": status.get('cpu_usage', 0)
                })
        
        # Calculate average performance
        avg_execution_time = np.mean([r['execution_time'] for r in iteration_results])
        avg_memory_usage = np.mean([r['memory_usage'] for r in iteration_results])
        avg_cpu_usage = np.mean([r['cpu_usage'] for r in iteration_results])
        
        performance_results.append({
            "scenario": scenario,
            "avg_execution_time": avg_execution_time,
            "avg_memory_usage": avg_memory_usage,
            "avg_cpu_usage": avg_cpu_usage,
            "iterations": iteration_results
        })
    
    return performance_results

# Run performance benchmark
perf_results = benchmark_performance()
```

## Advanced Use Cases

### Example 1: Automated Experiment Pipeline

```python
class ExperimentPipeline:
    def __init__(self, sdk):
        self.sdk = sdk
    
    def run_experiment(self, config):
        """Run a complete experiment pipeline"""
        
        # Phase 1: Setup environment
        self._setup_environment(config)
        
        # Phase 2: Inject faults
        fault_results = self._inject_faults(config['faults'])
        
        # Phase 3: Collect data
        self._wait_for_data_collection(config['collection_duration'])
        
        # Phase 4: Run algorithms
        algo_results = self._run_algorithms(config['algorithms'])
        
        # Phase 5: Evaluate results
        evaluation = self._evaluate_results(algo_results, config['ground_truth'])
        
        # Phase 6: Cleanup
        self._cleanup_environment(config)
        
        return {
            "fault_results": fault_results,
            "algorithm_results": algo_results,
            "evaluation": evaluation
        }
    
    def _setup_environment(self, config):
        """Setup experimental environment"""
        # Deploy workloads, configure monitoring, etc.
        pass
    
    def _inject_faults(self, fault_configs):
        """Inject configured faults"""
        results = []
        for fault in fault_configs:
            result = self.sdk.injection.execute([fault])
            results.append(result)
        return results
    
    def _wait_for_data_collection(self, duration):
        """Wait for observability data collection"""
        time.sleep(duration)
    
    def _run_algorithms(self, algorithm_configs):
        """Run configured algorithms"""
        results = {}
        for algo_config in algorithm_configs:
            result = self.sdk.algorithm.execute([algo_config])
            task_id = result['task_id']
            
            # Wait for completion
            status = self.sdk.algorithm.wait_for_completion(task_id)
            if status['state'] == 'completed':
                results[algo_config['algorithm']] = self.sdk.algorithm.get_results(task_id)
        
        return results
    
    def _evaluate_results(self, algo_results, ground_truth):
        """Evaluate algorithm results"""
        evaluation = {}
        for algo_name, results in algo_results.items():
            metrics = self.sdk.evaluation.calculate_metrics(
                predictions=results,
                ground_truth=ground_truth
            )
            evaluation[algo_name] = metrics
        return evaluation
    
    def _cleanup_environment(self, config):
        """Cleanup experimental environment"""
        # Remove faults, cleanup resources, etc.
        pass

# Use the pipeline
pipeline = ExperimentPipeline(sdk)

experiment_config = {
    "faults": [
        {
            "duration": 300,
            "faultType": 5,
            "injectNamespace": "test",
            "injectPod": "service-a",
            "spec": {"CPULoad": 80}
        }
    ],
    "algorithms": [
        {"algorithm": "alg1", "dataset": "test-data"},
        {"algorithm": "alg2", "dataset": "test-data"}
    ],
    "collection_duration": 600,
    "ground_truth": {"root_cause": "service-a", "type": "cpu"}
}

results = pipeline.run_experiment(experiment_config)
```

### Example 2: Continuous Evaluation

```python
import schedule
import time

class ContinuousEvaluator:
    def __init__(self, sdk):
        self.sdk = sdk
        self.results_history = []
    
    def run_daily_evaluation(self):
        """Run daily algorithm evaluation"""
        print("Starting daily evaluation...")
        
        # Get list of algorithms to evaluate
        algorithms = self.sdk.algorithm.list()
        
        # Get latest datasets
        datasets = self.sdk.dataset.list_recent(days=1)
        
        daily_results = {}
        
        for algorithm in algorithms:
            algo_name = algorithm['name']
            daily_results[algo_name] = {}
            
            for dataset in datasets:
                dataset_name = dataset['name']
                
                # Run algorithm
                execution_request = [{
                    "algorithm": algo_name,
                    "dataset": dataset_name,
                    "benchmark": "continuous-eval"
                }]
                
                exec_result = self.sdk.algorithm.execute(execution_request)
                status = self.sdk.algorithm.wait_for_completion(
                    exec_result['task_id'], timeout=1800
                )
                
                if status['state'] == 'completed':
                    results = self.sdk.algorithm.get_results(exec_result['task_id'])
                    ground_truth = self.sdk.dataset.get_ground_truth(dataset_name)
                    
                    metrics = self.sdk.evaluation.calculate_metrics(
                        predictions=results,
                        ground_truth=ground_truth
                    )
                    
                    daily_results[algo_name][dataset_name] = metrics
        
        # Store results
        self.results_history.append({
            "date": time.strftime("%Y-%m-%d"),
            "results": daily_results
        })
        
        # Generate trend report
        self._generate_trend_report()
        
        print("Daily evaluation completed")
    
    def _generate_trend_report(self):
        """Generate performance trend report"""
        if len(self.results_history) < 2:
            return
        
        # Calculate trends over time
        # Implementation depends on specific metrics
        pass

# Setup continuous evaluation
evaluator = ContinuousEvaluator(sdk)

# Schedule daily evaluation
schedule.every().day.at("02:00").do(evaluator.run_daily_evaluation)

# Run scheduler
while True:
    schedule.run_pending()
    time.sleep(3600)  # Check every hour
```

These examples demonstrate the flexibility and power of RCABench for various research and evaluation scenarios. Adapt them to your specific needs and requirements.