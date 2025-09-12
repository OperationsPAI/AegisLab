# Algorithm Development Guide

This guide explains how to develop, package, and integrate custom RCA algorithms with RCABench.

## Table of Contents

1. [Algorithm Requirements](#algorithm-requirements)
2. [Development Environment](#development-environment)
3. [Algorithm Interface](#algorithm-interface)
4. [Data Access](#data-access)
5. [Packaging Guidelines](#packaging-guidelines)
6. [Testing Algorithms](#testing-algorithms)
7. [Registration Process](#registration-process)
8. [Best Practices](#best-practices)

## Algorithm Requirements

### Basic Requirements

All RCA algorithms must meet these requirements:

1. **Containerized**: Packaged as Docker containers
2. **Standard Interface**: Follow RCABench algorithm interface
3. **Data Format**: Accept standard observability data formats
4. **Output Format**: Produce standardized JSON results
5. **Error Handling**: Graceful error handling and logging
6. **Resource Limits**: Respect CPU/memory constraints

### Input Data Types

Algorithms receive three types of observability data:

- **Metrics**: Time-series performance metrics (CPU, memory, response time, etc.)
- **Traces**: Distributed tracing data showing request flows
- **Logs**: Application and infrastructure log messages

### Output Requirements

Algorithms must output results in this JSON format:

```json
{
  "root_cause": "service-name",
  "confidence": 0.85,
  "execution_time": 45.2,
  "algorithm_version": "1.0.0",
  "ranking": [
    {
      "service": "service-a",
      "score": 0.85,
      "reasoning": "High CPU utilization correlation"
    },
    {
      "service": "service-b", 
      "score": 0.42,
      "reasoning": "Elevated error rate"
    }
  ],
  "metadata": {
    "model_confidence": 0.78,
    "data_quality": 0.95,
    "anomaly_threshold": 0.7
  }
}
```

## Development Environment

### Setup Development Environment

```bash
# Create development directory
mkdir my-rca-algorithm
cd my-rca-algorithm

# Initialize Git repository
git init

# Create basic structure
mkdir -p src tests data docs
touch src/algorithm.py
touch Dockerfile
touch requirements.txt
touch README.md
```

### Required Files

Your algorithm package should include:

```
my-rca-algorithm/
├── src/
│   ├── algorithm.py          # Main algorithm implementation
│   ├── data_processor.py     # Data processing utilities
│   └── utils.py             # Helper functions
├── tests/
│   ├── test_algorithm.py    # Unit tests
│   └── test_data/           # Test datasets
├── data/
│   └── sample_data/         # Sample input data
├── docs/
│   └── algorithm_doc.md     # Algorithm documentation
├── Dockerfile               # Container definition
├── requirements.txt         # Python dependencies
├── config.yaml             # Algorithm configuration
└── README.md               # Project documentation
```

## Algorithm Interface

### Environment Variables

Algorithms receive configuration through environment variables:

```python
import os
import json

# Required environment variables
DATASET_PATH = os.environ.get("DATASET_PATH", "/data")
OUTPUT_PATH = os.environ.get("OUTPUT_PATH", "/output")
ALGORITHM_CONFIG = json.loads(os.environ.get("ALGORITHM_CONFIG", "{}"))

# Optional parameters
TIMEOUT = int(os.environ.get("TIMEOUT", "3600"))
DEBUG_MODE = os.environ.get("DEBUG", "false").lower() == "true"
```

### Main Algorithm Template

```python
#!/usr/bin/env python3
"""
RCA Algorithm Template
Implement your root cause analysis algorithm using this template.
"""

import os
import json
import logging
import time
from typing import Dict, List, Any, Optional
from pathlib import Path

class RCAAlgorithm:
    """Base class for RCA algorithms"""
    
    def __init__(self, config: Dict[str, Any]):
        self.config = config
        self.logger = self._setup_logging()
        
        # Algorithm-specific parameters
        self.threshold = config.get("threshold", 0.5)
        self.window_size = config.get("window_size", 300)
        self.debug_mode = config.get("debug", False)
    
    def _setup_logging(self) -> logging.Logger:
        """Setup logging configuration"""
        logging.basicConfig(
            level=logging.INFO,
            format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
        )
        return logging.getLogger(self.__class__.__name__)
    
    def load_data(self, dataset_path: str) -> Dict[str, Any]:
        """Load observability data from dataset"""
        data = {}
        
        # Load metrics
        metrics_file = Path(dataset_path) / "metrics.json"
        if metrics_file.exists():
            with open(metrics_file, 'r') as f:
                data['metrics'] = json.load(f)
                self.logger.info(f"Loaded metrics: {len(data['metrics'])} services")
        
        # Load traces
        traces_file = Path(dataset_path) / "traces.json"
        if traces_file.exists():
            with open(traces_file, 'r') as f:
                data['traces'] = json.load(f)
                self.logger.info(f"Loaded traces: {len(data['traces'])} traces")
        
        # Load logs
        logs_file = Path(dataset_path) / "logs.json"
        if logs_file.exists():
            with open(logs_file, 'r') as f:
                data['logs'] = json.load(f)
                self.logger.info(f"Loaded logs: {len(data['logs'])} log entries")
        
        return data
    
    def preprocess_data(self, data: Dict[str, Any]) -> Dict[str, Any]:
        """Preprocess and validate input data"""
        processed_data = {}
        
        # Process metrics
        if 'metrics' in data:
            processed_data['metrics'] = self._process_metrics(data['metrics'])
        
        # Process traces
        if 'traces' in data:
            processed_data['traces'] = self._process_traces(data['traces'])
        
        # Process logs
        if 'logs' in data:
            processed_data['logs'] = self._process_logs(data['logs'])
        
        return processed_data
    
    def _process_metrics(self, metrics: Dict) -> Dict:
        """Process metrics data"""
        # Implement metrics preprocessing
        # Example: normalize values, handle missing data, etc.
        return metrics
    
    def _process_traces(self, traces: List) -> List:
        """Process traces data"""
        # Implement trace preprocessing
        # Example: extract service dependencies, calculate latencies, etc.
        return traces
    
    def _process_logs(self, logs: List) -> List:
        """Process logs data"""
        # Implement log preprocessing  
        # Example: parse log levels, extract error patterns, etc.
        return logs
    
    def analyze(self, data: Dict[str, Any]) -> Dict[str, Any]:
        """
        Main analysis method - implement your RCA logic here
        
        Args:
            data: Preprocessed observability data
            
        Returns:
            Analysis results in standard format
        """
        raise NotImplementedError("Subclasses must implement analyze method")
    
    def validate_output(self, results: Dict[str, Any]) -> bool:
        """Validate output format"""
        required_fields = ["root_cause", "confidence", "execution_time"]
        
        for field in required_fields:
            if field not in results:
                self.logger.error(f"Missing required field: {field}")
                return False
        
        # Validate confidence score
        if not 0 <= results["confidence"] <= 1:
            self.logger.error("Confidence must be between 0 and 1")
            return False
        
        return True
    
    def save_results(self, results: Dict[str, Any], output_path: str):
        """Save results to output file"""
        if not self.validate_output(results):
            raise ValueError("Invalid output format")
        
        output_file = Path(output_path) / "result.json"
        output_file.parent.mkdir(parents=True, exist_ok=True)
        
        with open(output_file, 'w') as f:
            json.dump(results, f, indent=2)
        
        self.logger.info(f"Results saved to {output_file}")
    
    def run(self, dataset_path: str, output_path: str) -> Dict[str, Any]:
        """Main execution method"""
        start_time = time.time()
        
        try:
            # Load and preprocess data
            self.logger.info("Loading dataset...")
            raw_data = self.load_data(dataset_path)
            
            self.logger.info("Preprocessing data...")
            processed_data = self.preprocess_data(raw_data)
            
            # Run analysis
            self.logger.info("Running RCA analysis...")
            results = self.analyze(processed_data)
            
            # Add execution metadata
            execution_time = time.time() - start_time
            results["execution_time"] = execution_time
            results["algorithm_version"] = self.config.get("version", "unknown")
            
            # Save results
            self.save_results(results, output_path)
            
            self.logger.info(f"Analysis completed in {execution_time:.2f} seconds")
            return results
            
        except Exception as e:
            self.logger.error(f"Algorithm execution failed: {e}")
            
            # Save error result
            error_result = {
                "root_cause": "unknown",
                "confidence": 0.0,
                "execution_time": time.time() - start_time,
                "error": str(e),
                "status": "failed"
            }
            self.save_results(error_result, output_path)
            raise

class MyRCAAlgorithm(RCAAlgorithm):
    """Example RCA algorithm implementation"""
    
    def analyze(self, data: Dict[str, Any]) -> Dict[str, Any]:
        """Implement your specific RCA logic here"""
        
        # Example: Simple statistical analysis
        anomalous_services = []
        
        if 'metrics' in data:
            for service, metrics in data['metrics'].items():
                # Calculate anomaly score
                anomaly_score = self._calculate_anomaly_score(metrics)
                
                if anomaly_score > self.threshold:
                    anomalous_services.append({
                        "service": service,
                        "score": anomaly_score,
                        "reasoning": f"Anomaly score {anomaly_score:.2f} > threshold {self.threshold}"
                    })
        
        # Sort by anomaly score
        anomalous_services.sort(key=lambda x: x["score"], reverse=True)
        
        # Determine root cause
        if anomalous_services:
            root_cause = anomalous_services[0]["service"]
            confidence = min(anomalous_services[0]["score"], 1.0)
        else:
            root_cause = "unknown"
            confidence = 0.0
        
        return {
            "root_cause": root_cause,
            "confidence": confidence,
            "ranking": anomalous_services[:10],  # Top 10 services
            "metadata": {
                "threshold_used": self.threshold,
                "services_analyzed": len(data.get('metrics', {})),
                "anomalies_found": len(anomalous_services)
            }
        }
    
    def _calculate_anomaly_score(self, metrics: Dict) -> float:
        """Calculate anomaly score for a service"""
        # Implement your anomaly detection logic
        # This is a placeholder implementation
        
        scores = []
        
        # Check CPU utilization
        if 'cpu_usage' in metrics:
            cpu_values = metrics['cpu_usage']
            if cpu_values:
                avg_cpu = sum(cpu_values) / len(cpu_values)
                if avg_cpu > 80:  # High CPU usage
                    scores.append(avg_cpu / 100)
        
        # Check error rate
        if 'error_rate' in metrics:
            error_values = metrics['error_rate']
            if error_values:
                avg_errors = sum(error_values) / len(error_values)
                if avg_errors > 0.05:  # High error rate
                    scores.append(avg_errors * 10)
        
        # Return maximum anomaly score
        return max(scores) if scores else 0.0

def main():
    """Main entry point for the algorithm"""
    
    # Get configuration from environment
    dataset_path = os.environ.get("DATASET_PATH", "/data")
    output_path = os.environ.get("OUTPUT_PATH", "/output")
    config_str = os.environ.get("ALGORITHM_CONFIG", "{}")
    
    try:
        config = json.loads(config_str)
    except json.JSONDecodeError:
        config = {}
    
    # Initialize and run algorithm
    algorithm = MyRCAAlgorithm(config)
    results = algorithm.run(dataset_path, output_path)
    
    print(f"RCA analysis completed. Root cause: {results['root_cause']}")

if __name__ == "__main__":
    main()
```

## Data Access

### Data Directory Structure

Your algorithm will have access to data mounted at `/data`:

```
/data/
├── metrics.json          # Service metrics data
├── traces.json           # Distributed tracing data
├── logs.json            # Application logs
├── topology.json        # Service topology (optional)
└── metadata.json        # Dataset metadata
```

### Data Formats

#### Metrics Format

```json
{
  "service-a": {
    "cpu_usage": [45.2, 67.8, 89.1, 76.3],
    "memory_usage": [512, 678, 823, 756],
    "response_time": [120, 145, 189, 167],
    "error_rate": [0.01, 0.02, 0.15, 0.08],
    "throughput": [100, 95, 87, 92],
    "timestamps": ["2024-01-01T10:00:00Z", "2024-01-01T10:01:00Z", ...]
  },
  "service-b": {
    ...
  }
}
```

#### Traces Format

```json
[
  {
    "trace_id": "abc123",
    "spans": [
      {
        "span_id": "span1",
        "parent_span_id": null,
        "service": "api-gateway",
        "operation": "handle_request",
        "start_time": "2024-01-01T10:00:00.123Z",
        "duration": 245.6,
        "status": "ok",
        "tags": {
          "http.method": "GET",
          "http.status_code": 200
        }
      },
      {
        "span_id": "span2",
        "parent_span_id": "span1",
        "service": "user-service",
        "operation": "get_user",
        "start_time": "2024-01-01T10:00:00.150Z",
        "duration": 89.2,
        "status": "error",
        "tags": {
          "error": true,
          "error.message": "Database connection timeout"
        }
      }
    ]
  }
]
```

#### Logs Format

```json
[
  {
    "timestamp": "2024-01-01T10:00:00.123Z",
    "service": "user-service",
    "level": "ERROR",
    "message": "Database connection timeout after 30s",
    "logger": "database.connection",
    "thread": "worker-1",
    "trace_id": "abc123",
    "span_id": "span2"
  },
  {
    "timestamp": "2024-01-01T10:00:01.456Z",
    "service": "api-gateway",
    "level": "WARN",
    "message": "High response time detected: 1200ms",
    "logger": "performance.monitor"
  }
]
```

## Packaging Guidelines

### Dockerfile Template

```dockerfile
FROM python:3.9-slim

# Set working directory
WORKDIR /app

# Install system dependencies
RUN apt-get update && apt-get install -y \
    gcc \
    && rm -rf /var/lib/apt/lists/*

# Copy requirements and install Python dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy algorithm source code
COPY src/ ./src/
COPY config.yaml .

# Create directories for data and output
RUN mkdir -p /data /output

# Set environment variables
ENV PYTHONPATH=/app/src
ENV PYTHONUNBUFFERED=1

# Set default configuration
ENV DATASET_PATH=/data
ENV OUTPUT_PATH=/output
ENV ALGORITHM_CONFIG={}

# Run algorithm
CMD ["python", "src/algorithm.py"]
```

### Requirements.txt Example

```txt
# Core dependencies
numpy>=1.21.0
pandas>=1.3.0
scikit-learn>=1.0.0

# Data processing
scipy>=1.7.0
matplotlib>=3.4.0

# Logging and configuration
pyyaml>=5.4.0
structlog>=21.1.0

# Optional: Machine learning libraries
# tensorflow>=2.6.0
# torch>=1.9.0
# xgboost>=1.4.0
```

### Configuration File

```yaml
# config.yaml
algorithm:
  name: "my-rca-algorithm"
  version: "1.0.0"
  description: "My custom RCA algorithm"

parameters:
  threshold: 0.7
  window_size: 300
  min_confidence: 0.5

performance:
  max_execution_time: 3600  # 1 hour
  max_memory_usage: "2Gi"
  cpu_limit: "1000m"

logging:
  level: "INFO"
  format: "json"
```

## Testing Algorithms

### Unit Testing

```python
# tests/test_algorithm.py
import unittest
import json
import tempfile
from pathlib import Path
from src.algorithm import MyRCAAlgorithm

class TestMyRCAAlgorithm(unittest.TestCase):
    
    def setUp(self):
        self.config = {
            "threshold": 0.7,
            "window_size": 300
        }
        self.algorithm = MyRCAAlgorithm(self.config)
    
    def test_load_data(self):
        """Test data loading functionality"""
        with tempfile.TemporaryDirectory() as temp_dir:
            # Create test data
            test_metrics = {"service-a": {"cpu_usage": [80, 90, 85]}}
            
            metrics_file = Path(temp_dir) / "metrics.json"
            with open(metrics_file, 'w') as f:
                json.dump(test_metrics, f)
            
            # Test loading
            data = self.algorithm.load_data(temp_dir)
            self.assertIn('metrics', data)
            self.assertEqual(data['metrics'], test_metrics)
    
    def test_analyze(self):
        """Test analysis functionality"""
        test_data = {
            'metrics': {
                'service-a': {
                    'cpu_usage': [90, 95, 88],  # High CPU
                    'error_rate': [0.1, 0.08, 0.12]  # High errors
                },
                'service-b': {
                    'cpu_usage': [20, 25, 22],  # Normal CPU
                    'error_rate': [0.01, 0.02, 0.01]  # Normal errors
                }
            }
        }
        
        results = self.algorithm.analyze(test_data)
        
        # Check output format
        self.assertIn('root_cause', results)
        self.assertIn('confidence', results)
        self.assertIn('ranking', results)
        
        # Check that service-a is identified as root cause
        self.assertEqual(results['root_cause'], 'service-a')
        self.assertGreater(results['confidence'], 0.7)
    
    def test_validate_output(self):
        """Test output validation"""
        valid_output = {
            "root_cause": "service-a",
            "confidence": 0.85,
            "execution_time": 45.2
        }
        self.assertTrue(self.algorithm.validate_output(valid_output))
        
        invalid_output = {
            "root_cause": "service-a",
            "confidence": 1.5  # Invalid confidence > 1
        }
        self.assertFalse(self.algorithm.validate_output(invalid_output))

if __name__ == '__main__':
    unittest.main()
```

### Integration Testing

```python
# tests/test_integration.py
import unittest
import tempfile
import json
from pathlib import Path
from src.algorithm import MyRCAAlgorithm

class TestIntegration(unittest.TestCase):
    
    def test_full_pipeline(self):
        """Test complete algorithm pipeline"""
        with tempfile.TemporaryDirectory() as temp_dir:
            data_dir = Path(temp_dir) / "data"
            output_dir = Path(temp_dir) / "output"
            data_dir.mkdir()
            output_dir.mkdir()
            
            # Create test dataset
            test_data = self._create_test_dataset(data_dir)
            
            # Run algorithm
            config = {"threshold": 0.7}
            algorithm = MyRCAAlgorithm(config)
            results = algorithm.run(str(data_dir), str(output_dir))
            
            # Check results
            self.assertIsInstance(results, dict)
            self.assertIn('root_cause', results)
            
            # Check output file
            output_file = output_dir / "result.json"
            self.assertTrue(output_file.exists())
            
            with open(output_file, 'r') as f:
                saved_results = json.load(f)
            
            self.assertEqual(results, saved_results)
    
    def _create_test_dataset(self, data_dir):
        """Create test dataset for integration testing"""
        
        # Metrics data
        metrics = {
            "service-a": {
                "cpu_usage": [85, 90, 88, 92, 87],
                "error_rate": [0.08, 0.12, 0.09, 0.15, 0.11],
                "timestamps": [f"2024-01-01T10:0{i}:00Z" for i in range(5)]
            },
            "service-b": {
                "cpu_usage": [25, 28, 22, 30, 26],
                "error_rate": [0.01, 0.02, 0.01, 0.02, 0.01],
                "timestamps": [f"2024-01-01T10:0{i}:00Z" for i in range(5)]
            }
        }
        
        with open(data_dir / "metrics.json", 'w') as f:
            json.dump(metrics, f)
        
        # Traces data
        traces = [
            {
                "trace_id": "trace1",
                "spans": [
                    {
                        "span_id": "span1",
                        "service": "service-a",
                        "duration": 1200,
                        "status": "error"
                    }
                ]
            }
        ]
        
        with open(data_dir / "traces.json", 'w') as f:
            json.dump(traces, f)
        
        # Logs data
        logs = [
            {
                "timestamp": "2024-01-01T10:00:00Z",
                "service": "service-a",
                "level": "ERROR",
                "message": "High CPU usage detected"
            }
        ]
        
        with open(data_dir / "logs.json", 'w') as f:
            json.dump(logs, f)
```

### Running Tests

```bash
# Run unit tests
python -m pytest tests/ -v

# Run with coverage
python -m pytest tests/ --cov=src --cov-report=html

# Run specific test
python -m pytest tests/test_algorithm.py::TestMyRCAAlgorithm::test_analyze
```

## Registration Process

### 1. Build and Push Container

```bash
# Build container
docker build -t my-registry/my-rca-algorithm:1.0.0 .

# Test locally
docker run --rm \
  -v $(pwd)/test_data:/data \
  -v $(pwd)/output:/output \
  -e ALGORITHM_CONFIG='{"threshold": 0.8}' \
  my-registry/my-rca-algorithm:1.0.0

# Push to registry
docker push my-registry/my-rca-algorithm:1.0.0
```

### 2. Register with RCABench

```python
from rcabench import RCABenchSDK

sdk = RCABenchSDK("http://localhost:8082")

algorithm_config = {
    "name": "my-rca-algorithm",
    "version": "1.0.0",
    "image": "my-registry/my-rca-algorithm:1.0.0",
    "description": "My custom RCA algorithm using statistical analysis",
    "category": "statistical",
    "parameters": {
        "threshold": {
            "type": "float",
            "default": 0.7,
            "min": 0.0,
            "max": 1.0,
            "description": "Anomaly detection threshold"
        },
        "window_size": {
            "type": "integer", 
            "default": 300,
            "min": 60,
            "max": 3600,
            "description": "Analysis window size in seconds"
        }
    },
    "requirements": {
        "cpu": "500m",
        "memory": "1Gi",
        "timeout": 3600
    },
    "metadata": {
        "author": "Anonymous",
        "tags": ["statistical", "anomaly-detection"],
        "documentation_url": "https://example.com/docs"
    }
}

result = sdk.algorithm.register(algorithm_config)
print(f"Algorithm registered with ID: {result['algorithm_id']}")
```

## Best Practices

### Performance Optimization

1. **Efficient Data Processing**
   ```python
   # Use vectorized operations
   import numpy as np
   import pandas as pd
   
   # Instead of loops
   anomaly_scores = np.array([calculate_score(x) for x in data])
   
   # Use vectorized operations
   anomaly_scores = np.vectorize(calculate_score)(data)
   ```

2. **Memory Management**
   ```python
   # Process data in chunks for large datasets
   def process_large_dataset(data, chunk_size=1000):
       for i in range(0, len(data), chunk_size):
           chunk = data[i:i+chunk_size]
           yield process_chunk(chunk)
   ```

3. **Early Termination**
   ```python
   # Stop processing if confidence is high enough
   if confidence > 0.95:
       return early_result
   ```

### Error Handling

```python
def robust_analysis(self, data):
    try:
        return self.analyze(data)
    except KeyError as e:
        self.logger.warning(f"Missing data field: {e}")
        return self.fallback_analysis(data)
    except ValueError as e:
        self.logger.error(f"Invalid data format: {e}")
        return {"root_cause": "unknown", "confidence": 0.0}
    except Exception as e:
        self.logger.error(f"Unexpected error: {e}")
        return {"root_cause": "error", "confidence": 0.0}
```

### Logging Best Practices

```python
import logging
import json

# Structured logging
def log_analysis_start(self, service_count, trace_count):
    self.logger.info("Starting RCA analysis", extra={
        "service_count": service_count,
        "trace_count": trace_count,
        "algorithm": self.__class__.__name__
    })

def log_result(self, result):
    self.logger.info("Analysis completed", extra={
        "root_cause": result["root_cause"],
        "confidence": result["confidence"],
        "execution_time": result["execution_time"]
    })
```

### Documentation Requirements

Include comprehensive documentation:

```markdown
# Algorithm Documentation

## Overview
Brief description of the algorithm and its approach.

## Parameters
- `threshold`: Anomaly detection threshold (0.0-1.0)
- `window_size`: Analysis window in seconds

## Input Requirements
- Metrics: CPU, memory, response time
- Traces: Distributed tracing data
- Logs: Error and warning logs

## Algorithm Details
Detailed explanation of the algorithm logic.

## Performance Characteristics
- Time complexity: O(n log n)
- Memory usage: Linear in input size
- Typical execution time: 30-60 seconds

## Limitations
Known limitations and edge cases.
```

### Version Management

```python
# Include version in output
results["algorithm_metadata"] = {
    "name": "my-rca-algorithm",
    "version": "1.0.0",
    "commit_hash": "abc123",
    "build_date": "2024-01-01"
}
```

Following these guidelines will ensure your algorithm integrates smoothly with RCABench and provides reliable, high-quality results for root cause analysis evaluations.