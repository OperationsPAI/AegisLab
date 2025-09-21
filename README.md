# RCABench: A Comprehensive Root Cause Analysis Benchmarking Platform

[![License: Apache 2.0](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Version](https://img.shields.io/badge/Go-1.23+-blue.svg)](https://golang.org/)
[![Python Version](https://img.shields.io/badge/Python-3.10+-green.svg)](https://python.org/)

RCABench is a comprehensive benchmarking platform designed for evaluating root cause analysis (RCA) algorithms in microservices environments. It provides automated fault injection, algorithm execution, and evaluation capabilities for distributed systems research.

## 🎯 Overview

RCABench enables researchers and practitioners to:
- **Inject faults** into microservices using chaos engineering principles
- **Execute RCA algorithms** on collected observability data
- **Evaluate and compare** different root cause analysis approaches
- **Benchmark performance** across various microservice architectures
- **Manage datasets** of fault scenarios and observability traces

## 🏗️ Architecture

The platform consists of several key components:

- **Core API Server** (Go): REST API for managing experiments, algorithms, and evaluations
- **Python SDK**: Client library for programmatic interaction with the platform
- **Fault Injection Engine**: Kubernetes-native chaos engineering capabilities
- **Algorithm Registry**: Extensible framework for RCA algorithm integration
- **Evaluation Framework**: Automated metrics calculation and comparison tools
- **Observability Stack**: Integration with tracing, metrics, and logging systems

## 📋 Prerequisites

### Software Requirements
- **Docker** (>= 20.10)
- **Kubernetes** (>= 1.25) or **kind/minikube** for local development
- **kubectl** (compatible with your cluster version)
- **Go** (>= 1.23) for development
- **Python** (>= 3.10) for SDK usage

### Hardware Requirements
- **CPU**: 4+ cores recommended
- **Memory**: 8GB+ RAM
- **Storage**: 20GB+ available disk space
- **Network**: Stable internet connection for image pulls

## 🚀 Quick Start

### Option 1: Local Development with Docker Compose

```bash
# Clone the repository
git clone <repository-url>
cd rcabench

# Start local environment
make local-debug

# The API will be available at http://localhost:8082
# Swagger documentation at http://localhost:8082/swagger/index.html
```

### Option 2: Kubernetes Deployment

```bash
# Check prerequisites
make check-prerequisites

# Deploy to Kubernetes cluster
make run

# Check deployment status
make status

# View logs
make logs
```

## 📖 Documentation

- **[User Guide](docs/user-guide.md)**: Complete guide for using RCABench
- **[Installation Guide](docs/installation.md)**: Detailed setup instructions
- **[API Reference](docs/api-reference.md)**: Complete API documentation
- **[Algorithm Development](docs/algorithm-development.md)**: Guide for implementing RCA algorithms
- **[Examples](docs/examples.md)**: Usage examples and tutorials

## 🔧 Configuration

### Environment Configuration

Copy and modify the configuration file:

```bash
cp src/config.dev.toml src/config.toml
```

Key configuration sections:

```toml
[database]
mysql_host = "localhost"
mysql_port = "3306"
mysql_user = "root"
mysql_password = "yourpassword"
mysql_db = "rcabench"

[redis]
host = "localhost:6379"

[k8s]
namespace = "default"

[injection]
benchmark = ["workload-name"]
target_label_key = "app"
```

### Storage Configuration

For production deployment, configure persistent volumes:

```bash
# Create persistent volumes (adjust paths as needed)
kubectl apply -f scripts/k8s/pv.yaml
```

## 💻 Usage Examples

### Using the Python SDK

```python
from rcabench import RCABenchSDK

# Initialize the SDK
sdk = RCABenchSDK("http://localhost:8082")

# List available algorithms
algorithms = sdk.algorithm.list()
print(f"Available algorithms: {algorithms}")

# Submit a fault injection
injection_request = [{
    "duration": 300,  # 5 minutes
    "faultType": 5,   # CPU stress
    "injectNamespace": "default",
    "injectPod": "my-service",
    "spec": {"CPULoad": 80, "CPUWorker": 2},
    "benchmark": "my-workload"
}]
response = sdk.injection.execute(injection_request)

# Execute an RCA algorithm
algorithm_request = [{
    "benchmark": "my-workload",
    "algorithm": "rca-algorithm-name",
    "dataset": "fault-scenario-dataset"
}]
result = sdk.algorithm.execute(algorithm_request)
```

### Using the REST API

```bash
# Get algorithm list
curl -X GET http://localhost:8082/api/v1/algorithms

# Submit fault injection
curl -X POST http://localhost:8082/api/v1/injection \
  -H "Content-Type: application/json" \
  -d '[{
    "duration": 300,
    "faultType": 5,
    "injectNamespace": "default",
    "injectPod": "my-service",
    "spec": {"CPULoad": 80}
  }]'
```

## 🧪 Supported Fault Types

RCABench supports various chaos engineering patterns:

- **Network Chaos**: Latency, packet loss, bandwidth limitation
- **Pod Chaos**: Pod failure, pod kill
- **Stress Chaos**: CPU stress, memory stress
- **Time Chaos**: Clock skew
- **DNS Chaos**: DNS resolution failures
- **HTTP Chaos**: HTTP request/response manipulation
- **JVM Chaos**: JVM-specific faults (GC pressure, etc.)

## 🎯 Evaluation Metrics

The platform provides comprehensive evaluation metrics:

- **Accuracy**: Precision, recall, F1-score for root cause identification
- **Latency**: Time to detection and diagnosis
- **Scalability**: Performance across different system sizes
- **Robustness**: Performance under various fault scenarios

## 🔍 Monitoring and Observability

RCABench integrates with:

- **Jaeger**: Distributed tracing
- **Prometheus**: Metrics collection
- **Grafana**: Visualization dashboards
- **ClickHouse**: Analytics and data warehouse

Access monitoring:
- Jaeger UI: http://localhost:16686
- API Metrics: http://localhost:8082/metrics

## 🛠️ Development

### Building from Source

```bash
# Build the main application
cd src
go build -o rcabench main.go

# Generate API documentation
make swagger

# Run tests
go test ./...
```

### Python SDK Development

```bash
cd sdk/python

# Install in development mode
pip install -e .

# Run tests
python -m pytest tests/
```

## 📦 Available Make Targets

```bash
make help                    # Show all available commands
make run                     # Build and deploy application
make local-debug             # Start local debugging environment
make build                   # Build application only
make status                  # Check application status
make logs                    # View application logs
make clean-all              # Clean all resources
make swagger                # Generate API documentation
```

## 🐛 Troubleshooting

### Common Issues

1. **Database Connection Failed**
   ```bash
   # Check database status
   kubectl get pods | grep mysql
   
   # Reset database
   make reset-db
   ```

2. **Pod Scheduling Issues**
   ```bash
   # Check node resources
   kubectl describe nodes
   
   # Check pod status
   kubectl describe pod <pod-name>
   ```

3. **Permission Errors**
   ```bash
   # Check RBAC permissions
   kubectl auth can-i create pods --namespace=default
   ```

### Getting Help

- Check the [troubleshooting guide](docs/troubleshooting.md)
- Review application logs with `make logs`
- Verify configuration in `src/config.toml`

## 📊 Performance Considerations

For optimal performance:

- **Resource Allocation**: Ensure adequate CPU/memory for workloads
- **Storage**: Use SSD storage for databases
- **Network**: Stable network connectivity for distributed components
- **Scaling**: Horizontal scaling supported via Kubernetes deployments

## 🔒 Security Notes

- Default credentials should be changed in production
- API endpoints should be secured with proper authentication
- Network policies recommended for production deployments
- Regular security updates for container images

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🤝 Contributing

We welcome contributions! Please see our [contributing guidelines](docs/contributing.md) for details on:

- Code style and standards
- Pull request process
- Issue reporting
- Documentation improvements

## 📬 Contact

For questions, issues, or contributions, please use the project's issue tracker or discussion forums.