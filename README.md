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

The current backend architecture is a single repository with a single `go.mod`, but it supports both:

- **local monolith-style development modes** for speed
- **split-service runtime modes** for service-boundary validation

The main service boundaries are:

- **`api-gateway`**: external HTTP/OpenAPI entrypoint
- **`iam-service`**: auth, user, RBAC, team, api-key
- **`resource-service`**: project, label, container, dataset, evaluation metadata/query
- **`orchestrator-service`**: submit, task, trace, retry, dead-letter, workflow control-plane
- **`runtime-worker-service`**: Redis async consumption, K8s/BuildKit/Helm/Chaos runtime execution
- **`system-service`**: config, audit, monitor, health, metrics

Key implementation rules:

- External APIs are HTTP/OpenAPI.
- Internal synchronous calls are gRPC via `src/internalclient/*`.
- Long-running execution stays asynchronous on Redis; it is not converted into synchronous execution RPC.
- Module-owned DB access lives in `src/module/*/repository.go`.
- Infra connectivity and low-level operations live in `src/infra/*`.

## 🧩 Runtime Modes And Injection Rules

The backend now has two categories of startup modes:

- **local integrated modes**: `producer`, `consumer`, `both`
- **dedicated service modes**: `api-gateway`, `iam-service`, `resource-service`, `orchestrator-service`, `runtime-worker-service`, `system-service`

### What `both` Actually Means

`both` is **not** the six-service topology.

It starts:

- the local HTTP stack
- the local worker/consumer stack

It is the fastest option for local end-to-end debugging such as:

- submit -> queue -> worker -> state update
- task/trace/log flow
- API + async worker integration

### Injection Matrix

| Mode / Service | What Starts | Local Owner Implementations Injected | Internal Clients Required | Best For |
| --- | --- | --- | --- | --- |
| `producer` | HTTP server only | Yes, local HTTP-facing modules | No | API, handler/service, Swagger, frontend integration |
| `consumer` | worker/controller/receiver side only | Yes, local runtime-side owners | Optional depending on config | queue/runtime/worker-only debugging |
| `both` | HTTP + worker/controller/receiver | Yes, local owners for integrated debugging | Optional depending on config | full local async loop |
| `api-gateway` | external HTTP gateway | No cross-owner local fallback as main path; service-specific remote wiring is expected | Yes | gateway boundary and remote-first debugging |
| `iam-service` | IAM gRPC service | Yes, IAM-local owners only | Only if a specific cross-service read path needs it | auth/user/rbac/team/api-key |
| `resource-service` | Resource gRPC service | Yes, resource-local owners only | Yes for orchestrator-backed queries like some statistics/evaluation views | project/container/dataset/label/evaluation |
| `orchestrator-service` | Orchestrator gRPC service | Yes, orchestrator-local owners only | Optional runtime/resource dependencies as needed | submit/task/trace/workflow |
| `runtime-worker-service` | runtime worker + runtime gRPC | Yes, runtime-side execution infrastructure only | Yes, especially orchestrator target | Redis consumer, K8s/build/helm runtime |
| `system-service` | system gRPC service | Yes, system-local owners only | Yes, especially runtime target | config/audit/monitor/metrics |

### Rule Of Thumb

- Use **`producer`** for normal API development.
- Use **`both`** when you need the local async loop.
- Use the **six dedicated services** when you need to verify service boundaries, internal gRPC, or remote-first behavior.

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

### Option 1: Local Dependencies

```bash
# Clone the repository
git clone https://github.com/OperationsPAI/AegisLab.git
cd AegisLab

# Start core dependencies
docker compose up -d redis mysql etcd jaeger buildkitd loki prometheus grafana
```

### Option 2: Fast Local API Debugging

```bash
cd src && go run . producer -conf ./config.dev.toml -port 8082

# HTTP:    http://localhost:8082
# Health:  http://localhost:8082/system/health
# Docs:    http://localhost:8082/docs/doc.json
```

### Option 3: Fast Local End-To-End Debugging

```bash
cd src && go run . both -conf ./config.dev.toml -port 8082
```

Use this mode when you need:

- HTTP + worker in one local process set
- submit -> queue -> consumer -> query loop
- task / trace / logs integration

### Option 4: Split-Service Debugging

```bash
# terminal 1
cd src && go run ./cmd/iam-service -conf ./config.dev.toml

# terminal 2
cd src && go run ./cmd/orchestrator-service -conf ./config.dev.toml

# terminal 3
cd src && go run ./cmd/resource-service -conf ./config.dev.toml

# terminal 4
cd src && go run ./cmd/runtime-worker-service -conf ./config.dev.toml

# terminal 5
cd src && go run ./cmd/system-service -conf ./config.dev.toml

# terminal 6
cd src && go run ./cmd/api-gateway -conf ./config.dev.toml -port 8082
```

### Option 5: Kubernetes Deployment

```bash
# Check prerequisites
just check-prerequisites

# Deploy to Kubernetes cluster
just run
```

If you use `scripts/start.sh` directly, the external install URLs can now be overridden with env vars such as:

- `CERT_MANAGER_MANIFEST_URL`
- `CHAOS_MESH_REPO_URL`
- `CLICKSTACK_REPO_URL`
- `OPEN_TELEMETRY_REPO_URL`
- `OTEL_DEMO_REPO_URL`
- `JUICEFS_REPO_URL`
- `TEST_HTTP_PROXY`
- `TEST_HTTPS_PROXY`
- `TEST_NO_PROXY`

## 📖 Documentation

- **[Report Index](docs/report-index.md)**: Consolidated backend refactor, runtime, governance, SDK/auth, and validation notes
- **[Refactor TODO](docs/todo.md)**: Source-of-truth task list and final acceptance checklist
- **[API Key Auth TODO](docs/api-key-auth-execution-todo.md)**: Key ID / Key Secret auth execution checklist and signing contract
- **[Package Rename TODO](docs/package-rename-todo.md)**: Go package naming cleanup record for `interface/module/infra/app`
- **[Frontend Redesign](docs/frontend-redesign.md)**: Frontend redesign plan and IA notes
- **[Frontend UI Guidelines](docs/frontend-ui-guidelines.md)**: Frontend visual/system guidelines

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

[clients.iam]
target = "127.0.0.1:9091"

[clients.resource]
target = "127.0.0.1:9093"

[clients.orchestrator]
target = "127.0.0.1:9092"

[clients.runtime]
target = "127.0.0.1:9094"

[clients.system]
target = "127.0.0.1:9095"

[iam.grpc]
addr = ":9091"

[resource.grpc]
addr = ":9093"

[orchestrator.grpc]
addr = ":9092"

[runtime_worker.grpc]
addr = ":9094"

[system.grpc]
addr = ":9095"

[injection]
benchmark = ["workload-name"]
target_label_key = "app"
```

Important config rules:

- `producer` and `both` can use local owner implementations for fast debugging.
- dedicated services should use the appropriate `clients.*.target` values when a remote dependency is required.
- `api-gateway` validates `clients.iam.target`, `clients.resource.target`, `clients.orchestrator.target`, and `clients.system.target`.
- `runtime-worker-service` validates `clients.orchestrator.target`.
- `system-service` validates `clients.runtime.target`.
- `resource-service` validates `clients.orchestrator.target` for remote-backed query paths.

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

### Recommended Debug Flow

Choose the mode first:

- **API-only debugging** -> `producer`
- **local async loop debugging** -> `both`
- **service-boundary / gRPC debugging** -> six dedicated services

### Where To Put Breakpoints

#### HTTP issues

Start here:

- `src/router/*`
- `src/module/*/handler.go`
- `src/module/*/service.go`
- `src/module/*/repository.go`

If the problem only appears in split-service mode, then also check:

- `src/app/gateway/*`
- `src/internalclient/*`

#### gRPC / service-boundary issues

Start here:

- `src/internalclient/*`
- `src/interface/grpc/*`
- `src/app/{gateway,iam,resource,orchestrator,runtime,system}/*`

#### async runtime issues

Start here:

- `src/service/consumer/*`
- `src/interface/worker/*`
- `src/interface/controller/*`
- `src/infra/k8s/*`
- `src/infra/buildkit/*`
- `src/infra/helm/*`
- `src/infra/chaos/*`

### Module-Oriented Debug Map

#### Auth / User / RBAC / Team

Check:

- `src/module/auth/*`
- `src/module/user/*`
- `src/module/rbac/*`
- `src/module/team/*`

Split-service path:

- `src/app/gateway/{auth,user,rbac,team}_services.go`
- `src/internalclient/iamclient/*`
- `src/interface/grpc/iam/*`

#### Project / Label / Container / Dataset

Check:

- `src/module/project/*`
- `src/module/label/*`
- `src/module/container/*`
- `src/module/dataset/*`

Split-service path:

- `src/app/gateway/resource_services.go`
- `src/internalclient/resourceclient/*`
- `src/interface/grpc/resource/*`

#### Injection / Execution / Task / Trace / Group / Notification

Check:

- `src/module/injection/*`
- `src/module/execution/*`
- `src/module/task/*`
- `src/module/trace/*`
- `src/module/group/*`
- `src/module/notification/*`

Split-service path:

- `src/app/gateway/orchestrator_services.go`
- `src/internalclient/orchestratorclient/*`
- `src/interface/grpc/orchestrator/*`
- `src/service/consumer/*`

#### System / Metrics / Monitor / Config / Audit

Check:

- `src/module/system/*`
- `src/module/systemmetric/*`

Split-service path:

- `src/app/gateway/system_services.go`
- `src/internalclient/systemclient/*`
- `src/internalclient/runtimeclient/*`
- `src/interface/grpc/system/*`
- `src/interface/grpc/runtime/*`

#### Runtime / K8s / Build / Helm / Chaos

Check:

- `src/service/consumer/*`
- `src/interface/worker/*`
- `src/interface/controller/*`
- `src/infra/k8s/*`
- `src/infra/buildkit/*`
- `src/infra/helm/*`
- `src/infra/chaos/*`
- `src/infra/redis/*`

### Building from Source

```bash
# Build the main application
cd src
go build -o rcabench main.go

# Regenerate OpenAPI / Swagger artifacts
cd ..
just swagger-init 1.2.3

# Generate SDK packages
just generate-portal 1.2.3
just generate-admin 1.2.3
just generate-python-sdk 1.2.3

# Run tests
cd src
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

## 📦 Available Just Recipes

```bash
just --list                  # Show all available commands
just run                     # Deploy to the configured Kubernetes target
just local-deploy            # Boot local infra dependencies with Docker Compose
just local-debug             # Start local producer+consumer debug process
just swagger-init 1.2.3      # Regenerate OpenAPI / Swagger artifacts
just generate-portal 1.2.3   # Generate portal TypeScript SDK
just generate-admin 1.2.3    # Generate admin TypeScript SDK
just generate-python-sdk 1.2.3  # Generate Python SDK
just release-portal 1.2.3    # Generate release-ready portal TypeScript SDK
just release-admin 1.2.3     # Generate release-ready admin TypeScript SDK
just release-python-sdk 1.2.3  # Generate release-ready Python SDK
just test-regression         # Run the Python SDK regression workflow
```

## 🐛 Troubleshooting

### Common Issues

1. **Database Connection Failed**

   ```bash
   # Check database status
   kubectl get pods | grep mysql

   # Re-run the local debug stack after fixing config/env
   just local-debug
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

4. **A Request Works In `producer` But Fails In Split-Service Mode**

   Check in this order:

   - are the dedicated services actually running?
   - are the required `clients.*.target` values configured?
   - is the request going through `src/internalclient/*` as expected?
   - is the destination gRPC service registered and listening?

5. **Submit Works But Task State Does Not Move**

   Check in this order:

   - Redis queue health
   - `src/service/consumer/*`
   - runtime infra (`src/infra/k8s/*`, `src/infra/buildkit/*`, `src/infra/helm/*`)
   - orchestrator owner write-back path

### Quick Validation Commands

```bash
cd src && go test ./...
cd src && go test ./app -run 'TestProducerOptionsValidate|TestProducerOptionsStartStopSmoke|TestProducerOptionsHTTPIntegrationSmoke'
cd src && go test ./app -run 'TestConsumerOptions|TestBothOptions'
cd src && go test ./router ./docs ./interface/http
```

Real-cluster K8s validation:

```bash
cd src && RUN_K8S_INTEGRATION=1 go test ./infra/k8s -run TestK8sGatewayJobLifecycleIntegration
```

### Getting Help

- Review the consolidated notes in `docs/report-index.md`
- Run `just --list` to inspect the supported local workflows
- Verify configuration in `src/config.dev.toml`

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
