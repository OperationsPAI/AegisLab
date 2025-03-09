# RCABench SDK

A Python SDK for interacting with RCABench services.

## Installation

### From PyPI

```bash
pip install rcabench
```

### From Source

```bash
# Clone the repository
git clone https://github.com/your-username/rcabench.git
cd rcabench/sdk/python

# Install the package
pip install -e .
```

## Building the Package

To build the package for distribution:

```bash
# Install build dependencies
pip install build

# Build the package
python -m build

# This will create distribution files in the dist/ directory
```

## Usage

```python
from rcabench import RCABenchSDK

# Initialize the SDK
sdk = RCABenchSDK("http://localhost:8082")

# Get available algorithms
algorithms = sdk.algorithm.list()
print(algorithms)

# Submit an injection task
injection_payload = [{
    "duration": 1,
    "faultType": 5,
    "injectNamespace": "ts",
    "injectPod": "ts-preserve-service",
    "spec": {"CPULoad": 1, "CPUWorker": 3},
    "benchmark": "clickhouse",
}]
response = sdk.injection.execute(injection_payload)
print(response)

# Run an algorithm
algorithm_payload = [{
    "benchmark": "clickhouse",
    "algorithm": "e-diagnose", 
    "dataset": "dataset-name",
}]
response = sdk.algorithm.execute(algorithm_payload)
print(response)
```

## API Reference

The SDK provides the following main components:

- `RCABenchSDK`: The main entry point for the SDK
  - `algorithm`: For interacting with algorithm endpoints
  - `evaluation`: For interacting with evaluation endpoints
  - `injection`: For interacting with injection endpoints

For detailed API documentation, please refer to the code docstrings.

## Requirements

- Python 3.8 or higher
- `requests` and `aiohttp` libraries
