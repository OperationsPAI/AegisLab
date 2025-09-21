# API Reference

This document provides comprehensive reference for the RCABench REST API and Python SDK.

## Table of Contents

1. [REST API Overview](#rest-api-overview)
2. [Authentication](#authentication)
3. [Core Endpoints](#core-endpoints)
4. [Python SDK Reference](#python-sdk-reference)
5. [Error Handling](#error-handling)
6. [Rate Limiting](#rate-limiting)

## REST API Overview

### Base URL
```
http://localhost:8082/api/v1
```

### Content Type
All requests and responses use JSON format:
```
Content-Type: application/json
```

### Response Format
All API responses follow this structure:
```json
{
  "success": true,
  "data": {},
  "message": "Operation completed successfully",
  "timestamp": "2024-01-01T12:00:00Z"
}
```

Error responses:
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid input parameters",
    "details": {}
  },
  "timestamp": "2024-01-01T12:00:00Z"
}
```

## Authentication

### API Key Authentication
Include API key in request headers:
```http
Authorization: Bearer <your-api-key>
```

### JWT Token Authentication
For advanced features, use JWT tokens:
```http
Authorization: Bearer <jwt-token>
```

## Core Endpoints

### Health Check

#### GET /health
Check API server health status.

**Response:**
```json
{
  "status": "ok",
  "version": "1.0.1",
  "uptime": 3600,
  "components": {
    "database": "healthy",
    "redis": "healthy",
    "kubernetes": "healthy"
  }
}
```

### Algorithms

#### GET /algorithms
List all available RCA algorithms.

**Parameters:**
- `page` (integer, optional): Page number (default: 1)
- `limit` (integer, optional): Results per page (default: 20)
- `category` (string, optional): Filter by algorithm category

**Response:**
```json
{
  "success": true,
  "data": {
    "algorithms": [
      {
        "id": "alg-001",
        "name": "random-walk",
        "version": "1.0.0",
        "description": "Random walk based RCA algorithm",
        "category": "statistical",
        "parameters": {
          "steps": {"type": "integer", "default": 100},
          "threshold": {"type": "float", "default": 0.5}
        },
        "created_at": "2024-01-01T10:00:00Z",
        "updated_at": "2024-01-01T10:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 45,
      "pages": 3
    }
  }
}
```

#### POST /algorithms
Register a new RCA algorithm.

**Request Body:**
```json
{
  "name": "my-rca-algorithm",
  "version": "1.0.0",
  "image": "my-registry/my-algorithm:1.0.0",
  "description": "My custom RCA algorithm",
  "category": "machine-learning",
  "parameters": {
    "learning_rate": {"type": "float", "default": 0.01},
    "epochs": {"type": "integer", "default": 100}
  },
  "requirements": {
    "cpu": "500m",
    "memory": "1Gi",
    "timeout": 3600
  }
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "algorithm_id": "alg-002",
    "status": "registered"
  }
}
```

#### DELETE /algorithms/{algorithm_id}
Delete an algorithm.

**Response:**
```json
{
  "success": true,
  "message": "Algorithm deleted successfully"
}
```

### Algorithm Execution

#### POST /algorithms/execute
Execute RCA algorithms on datasets.

**Request Body:**
```json
[
  {
    "algorithm": "random-walk",
    "dataset": "fault-scenario-001",
    "benchmark": "microservice-app",
    "parameters": {
      "steps": 150,
      "threshold": 0.7
    },
    "timeout": 3600
  }
]
```

**Response:**
```json
{
  "success": true,
  "data": {
    "task_id": "task-12345",
    "status": "submitted",
    "estimated_completion": "2024-01-01T12:30:00Z"
  }
}
```

#### GET /algorithms/execute/{task_id}
Get algorithm execution status and results.

**Response:**
```json
{
  "success": true,
  "data": {
    "task_id": "task-12345",
    "status": "completed",
    "started_at": "2024-01-01T12:00:00Z",
    "completed_at": "2024-01-01T12:25:00Z",
    "execution_time": 1500,
    "results": {
      "root_cause": "service-a",
      "confidence": 0.85,
      "ranking": [
        {"service": "service-a", "score": 0.85},
        {"service": "service-b", "score": 0.42}
      ]
    },
    "metrics": {
      "cpu_usage": "250m",
      "memory_usage": "512Mi",
      "execution_duration": "25m"
    }
  }
}
```

### Fault Injection

#### POST /injection
Execute fault injection scenarios.

**Request Body:**
```json
[
  {
    "duration": 300,
    "faultType": 5,
    "injectNamespace": "production",
    "injectPod": "api-service",
    "spec": {
      "CPULoad": 80,
      "CPUWorker": 2
    },
    "benchmark": "e-commerce"
  }
]
```

**Fault Types:**
- `1`: Network Chaos
- `2`: Pod Chaos  
- `3`: DNS Chaos
- `4`: HTTP Chaos
- `5`: Stress Chaos
- `6`: Time Chaos
- `7`: JVM Chaos

**Response:**
```json
{
  "success": true,
  "data": {
    "injection_id": "inj-67890",
    "status": "injecting",
    "started_at": "2024-01-01T14:00:00Z",
    "estimated_end": "2024-01-01T14:05:00Z"
  }
}
```

#### GET /injection/{injection_id}
Get fault injection status.

**Response:**
```json
{
  "success": true,
  "data": {
    "injection_id": "inj-67890",
    "status": "completed",
    "fault_details": {
      "type": "stress",
      "target": "api-service",
      "duration": 300
    },
    "started_at": "2024-01-01T14:00:00Z",
    "completed_at": "2024-01-01T14:05:00Z",
    "impact_metrics": {
      "response_time_increase": "150%",
      "error_rate_increase": "5%"
    }
  }
}
```

#### DELETE /injection/{injection_id}
Stop ongoing fault injection.

**Response:**
```json
{
  "success": true,
  "message": "Fault injection stopped successfully"
}
```

### Datasets

#### GET /datasets
List available datasets.

**Parameters:**
- `page` (integer, optional): Page number
- `limit` (integer, optional): Results per page
- `type` (string, optional): Dataset type filter

**Response:**
```json
{
  "success": true,
  "data": {
    "datasets": [
      {
        "id": "ds-001",
        "name": "fault-scenario-001",
        "description": "CPU stress fault in e-commerce app",
        "type": "fault-injection",
        "size": "2.5GB",
        "created_at": "2024-01-01T09:00:00Z",
        "metrics": {
          "services": 12,
          "duration": 1800,
          "fault_count": 3
        }
      }
    ]
  }
}
```

#### POST /datasets
Create a new dataset.

**Request Body:**
```json
{
  "name": "my-experiment-dataset",
  "description": "Dataset from my experiment",
  "type": "fault-injection",
  "metadata": {
    "services": ["service-a", "service-b"],
    "fault_type": "network-latency",
    "duration": 600
  }
}
```

#### GET /datasets/{dataset_id}/ground-truth
Get ground truth data for evaluation.

**Response:**
```json
{
  "success": true,
  "data": {
    "ground_truth": {
      "root_cause": "service-a",
      "fault_type": "cpu-stress",
      "start_time": "2024-01-01T14:00:00Z",
      "end_time": "2024-01-01T14:05:00Z",
      "affected_services": ["service-a", "service-b"]
    }
  }
}
```

### Evaluation

#### POST /evaluation/metrics
Calculate evaluation metrics.

**Request Body:**
```json
{
  "predictions": {
    "root_cause": "service-a",
    "confidence": 0.85,
    "ranking": [
      {"service": "service-a", "score": 0.85},
      {"service": "service-b", "score": 0.42}
    ]
  },
  "ground_truth": {
    "root_cause": "service-a",
    "fault_type": "cpu-stress"
  },
  "metrics": ["precision", "recall", "f1_score", "accuracy"]
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "metrics": {
      "precision": 0.85,
      "recall": 0.90,
      "f1_score": 0.87,
      "accuracy": 0.88,
      "detection_time": 45.2,
      "ranking_quality": 0.76
    }
  }
}
```

#### POST /evaluation/compare
Compare multiple algorithm results.

**Request Body:**
```json
{
  "algorithms": [
    {
      "name": "algorithm-1",
      "results": {...}
    },
    {
      "name": "algorithm-2", 
      "results": {...}
    }
  ],
  "ground_truth": {...},
  "comparison_metrics": ["precision", "recall", "execution_time"]
}
```

### Tasks

#### GET /tasks
List task execution history.

**Parameters:**
- `type` (string, optional): Task type filter
- `status` (string, optional): Task status filter
- `page` (integer, optional): Page number

**Response:**
```json
{
  "success": true,
  "data": {
    "tasks": [
      {
        "task_id": "task-12345",
        "type": "algorithm_execution",
        "status": "completed",
        "created_at": "2024-01-01T12:00:00Z",
        "completed_at": "2024-01-01T12:25:00Z",
        "duration": 1500,
        "resource_usage": {
          "cpu": "250m",
          "memory": "512Mi"
        }
      }
    ]
  }
}
```

## Python SDK Reference

### Installation

```bash
pip install rcabench
```

### Basic Usage

```python
from rcabench import RCABenchSDK

# Initialize SDK
sdk = RCABenchSDK("http://localhost:8082", api_key="your-api-key")
```

### SDK Classes

#### RCABenchSDK

Main SDK class for interacting with RCABench.

**Constructor:**
```python
RCABenchSDK(base_url: str, api_key: str = None, timeout: int = 30)
```

**Properties:**
- `algorithm`: Algorithm management interface
- `injection`: Fault injection interface  
- `dataset`: Dataset management interface
- `evaluation`: Evaluation interface
- `task`: Task management interface

**Methods:**

##### health_check()
Check API server health.

```python
health = sdk.health_check()
# Returns: {"status": "ok", "version": "1.0.1", ...}
```

#### AlgorithmInterface

Manage and execute RCA algorithms.

##### list(page=1, limit=20, category=None)
List available algorithms.

```python
algorithms = sdk.algorithm.list(category="statistical")
```

##### register(config)
Register a new algorithm.

```python
config = {
    "name": "my-algorithm",
    "version": "1.0.0",
    "image": "my-registry/my-algorithm:1.0.0",
    "description": "My RCA algorithm"
}
result = sdk.algorithm.register(config)
```

##### execute(requests)
Execute algorithms.

```python
requests = [{
    "algorithm": "random-walk",
    "dataset": "fault-scenario-001",
    "benchmark": "my-app"
}]
result = sdk.algorithm.execute(requests)
```

##### get_results(task_id)
Get algorithm execution results.

```python
results = sdk.algorithm.get_results("task-12345")
```

##### wait_for_completion(task_id, timeout=3600)
Wait for algorithm execution to complete.

```python
status = sdk.algorithm.wait_for_completion("task-12345", timeout=1800)
```

#### InjectionInterface

Manage fault injection.

##### execute(fault_configs)
Execute fault injection.

```python
faults = [{
    "duration": 300,
    "faultType": 5,
    "injectNamespace": "default",
    "injectPod": "my-service",
    "spec": {"CPULoad": 80}
}]
result = sdk.injection.execute(faults)
```

##### get_status(injection_id)
Get fault injection status.

```python
status = sdk.injection.get_status("inj-67890")
```

##### stop(injection_id)
Stop ongoing fault injection.

```python
result = sdk.injection.stop("inj-67890")
```

#### DatasetInterface

Manage datasets.

##### list(page=1, limit=20, type=None)
List datasets.

```python
datasets = sdk.dataset.list(type="fault-injection")
```

##### create(config)
Create new dataset.

```python
config = {
    "name": "my-dataset",
    "description": "My experiment dataset",
    "type": "fault-injection"
}
result = sdk.dataset.create(config)
```

##### get_ground_truth(dataset_id)
Get ground truth data.

```python
ground_truth = sdk.dataset.get_ground_truth("ds-001")
```

##### upload_data(dataset_id, data_type, file_path)
Upload data to dataset.

```python
sdk.dataset.upload_data("ds-001", "metrics", "metrics.json")
sdk.dataset.upload_data("ds-001", "traces", "traces.json")
sdk.dataset.upload_data("ds-001", "logs", "logs.json")
```

#### EvaluationInterface

Evaluate algorithm performance.

##### calculate_metrics(predictions, ground_truth, metrics=None)
Calculate evaluation metrics.

```python
metrics = sdk.evaluation.calculate_metrics(
    predictions=algo_results,
    ground_truth=ground_truth_data,
    metrics=["precision", "recall", "f1_score"]
)
```

##### compare_algorithms(algorithm_results, ground_truth)
Compare multiple algorithms.

```python
comparison = sdk.evaluation.compare_algorithms([
    {"name": "alg1", "results": results1},
    {"name": "alg2", "results": results2}
], ground_truth)
```

##### generate_report(evaluation_data, format="json")
Generate evaluation report.

```python
report = sdk.evaluation.generate_report(
    evaluation_data, 
    format="html"
)
```

### Error Handling

The SDK raises specific exceptions for different error types:

```python
from rcabench.exceptions import (
    RCABenchAPIError,
    AuthenticationError,
    ValidationError,
    ResourceNotFoundError,
    TimeoutError
)

try:
    result = sdk.algorithm.execute(requests)
except AuthenticationError:
    print("Invalid API key")
except ValidationError as e:
    print(f"Invalid request: {e}")
except ResourceNotFoundError:
    print("Algorithm not found")
except TimeoutError:
    print("Request timed out")
except RCABenchAPIError as e:
    print(f"API error: {e}")
```

### Async Support

The SDK also provides async support:

```python
from rcabench import AsyncRCABenchSDK
import asyncio

async def main():
    sdk = AsyncRCABenchSDK("http://localhost:8082")
    
    # Async operations
    algorithms = await sdk.algorithm.list()
    result = await sdk.algorithm.execute(requests)
    
asyncio.run(main())
```

## Error Handling

### Error Codes

| Code | Description |
|------|-------------|
| `VALIDATION_ERROR` | Invalid input parameters |
| `AUTHENTICATION_ERROR` | Invalid or missing authentication |
| `AUTHORIZATION_ERROR` | Insufficient permissions |
| `RESOURCE_NOT_FOUND` | Requested resource does not exist |
| `RESOURCE_CONFLICT` | Resource already exists or conflict |
| `RATE_LIMIT_EXCEEDED` | Too many requests |
| `INTERNAL_ERROR` | Internal server error |
| `SERVICE_UNAVAILABLE` | Service temporarily unavailable |

### HTTP Status Codes

| Status | Description |
|--------|-------------|
| 200 | Success |
| 201 | Created |
| 400 | Bad Request |
| 401 | Unauthorized |
| 403 | Forbidden |
| 404 | Not Found |
| 409 | Conflict |
| 429 | Too Many Requests |
| 500 | Internal Server Error |
| 503 | Service Unavailable |

## Rate Limiting

### Default Limits

- **API Requests**: 1000 requests per hour per API key
- **Algorithm Executions**: 10 concurrent executions per user
- **Fault Injections**: 5 concurrent injections per user
- **Dataset Uploads**: 1GB per hour per user

### Rate Limit Headers

Responses include rate limiting information:

```http
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1641024000
```

### Handling Rate Limits

```python
import time
from rcabench.exceptions import RateLimitExceededError

try:
    result = sdk.algorithm.execute(requests)
except RateLimitExceededError as e:
    # Wait for rate limit reset
    reset_time = e.reset_time
    wait_time = reset_time - time.time()
    if wait_time > 0:
        time.sleep(wait_time)
    # Retry request
    result = sdk.algorithm.execute(requests)
```

For complete API documentation with interactive examples, visit the Swagger UI at:
`http://localhost:8082/swagger/index.html`