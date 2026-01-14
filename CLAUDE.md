# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

AegisLab (RCABench) is a comprehensive Root Cause Analysis (RCA) benchmarking platform for microservices environments. It provides automated fault injection, algorithm execution, and evaluation capabilities for distributed systems research.

## Development Commands

### Environment Setup

- `make check-prerequisites` - Verify development dependencies (devbox, docker, helm, kubectl, kubectx)
- `make setup-dev-env` - Bootstrap local development environment (installs uv, applies K8s manifests, installs Lefthook)

### Local Development

- `docker compose up redis mysql jaeger buildkitd -d` - Start infrastructure services (REQUIRED before testing/building)
- `make local-debug` - Start the Go application locally (runs `src/main.go both --port 8082`)
- `cd src && go build -o /tmp/rcabench ./main.go` - Build the application binary

### Testing

- `cd src && go test ./utils/... -v` - Run unit tests (requires infrastructure services)
- `cd src && go test ./... -v` - Run all tests (some require K8s cluster access)
- **NOTE**: Tests take 15-30 seconds. Set timeouts to 60+ seconds to avoid cancellation.

### Deployment

- `make run` - Build and deploy to Kubernetes using Skaffold (requires K8s cluster)
- `make install-rcabench` - Deploy RCABench using Helm charts

### SDK Generation

- `make swag-init` - Initialize Swagger documentation
- `make generate-python-sdk` - Generate Python SDK from Swagger docs
- Manual swagger: `cd src && ~/go/bin/swag init --parseDependency --parseDepth 1 --output ./docs`

### Utilities

- `make help` - Display all available commands
- `make info` - Display project configuration information
- `make pre-commit` - Run pre-commit hooks

## Architecture

### Application Modes (Cobra CLI)

The Go application ([`src/main.go`](src/main.go)) runs in three modes:

1. **Producer Mode** (`producer`) - HTTP API server handling REST requests
2. **Consumer Mode** (`consumer`) - Background workers and Kubernetes controllers
3. **Both Mode** (`both`) - Combined producer and consumer (default for local development)

### Core Components

**Backend ([`src/`](src/))**

- [`handlers/v2/`](src/handlers/v2/) - REST API endpoints organized by domain (auth, projects, datasets, injections, executions, etc.)
- [`service/prodcuer/`](src/service/prodcuer/) - Business logic for API operations
- [`service/consumer/`](src/service/consumer/) - Asynchronous task processing
  - [`task.go`](src/service/consumer/task.go) - Task queue management with Redis
  - [`fault_injection.go`](src/service/consumer/fault_injection.go) - Chaos engineering execution
  - [`algo_execution.go`](src/service/consumer/algo_execution.go) - RCA algorithm execution
  - [`k8s_handler.go`](src/service/consumer/k8s_handler.go) - Kubernetes job management
- [`database/`](src/database/) - MySQL models and migrations
- [`client/k8s/`](src/client/k8s/) - Kubernetes client and controllers
- [`dto/`](src/dto/) - Data transfer objects
- [`middleware/`](src/middleware/) - HTTP middleware (auth, rate limiting, audit, CORS)

**Python SDK ([`sdk/python/`](sdk/python/))**

- Auto-generated from OpenAPI/Swagger specifications
- Models and REST client in [`sdk/python/src/rcabench/openapi/`](sdk/python/src/rcabench/openapi/)
- Version managed in `__init__.py`

**Command Tools ([`scripts/command/`](scripts/command/))**

- Python-based CLI for testing and workflow management
- Uses `uv` for dependency management
- Test workflows in [`test_workflow.py`](scripts/command/test/test_workflow.py)

**Deployment ([`helm/`](helm/), [`manifests/`](manifests/))**

- Helm charts for Kubernetes deployment
- Environment-specific manifests (dev, test, prod)
- Mirror configurations for different regions

### Data Flow

1. **HTTP Requests** → Producer (Gin router) → Handlers → Service layer → Database
2. **Background Tasks** → Redis queues → Consumer workers → Kubernetes jobs
3. **Tracing** → OpenTelemetry → Jaeger
4. **Storage** → MySQL (persistent) + Redis (cache/queues) + JuiceFS (shared files)

### Task Processing System

The consumer uses a sophisticated Redis-based task queue system:

- **Delayed Queue** - Tasks scheduled for future execution (sorted set by timestamp)
- **Ready Queue** - Tasks ready for immediate processing (list)
- **Dead Letter Queue** - Failed tasks for inspection (sorted set)
- **Concurrency Control** - Maximum 20 concurrent tasks
- **Task States**: Pending → Running → Completed/Error
- **Retry Policy**: Configurable max attempts with backoff

See [`src/service/consumer/task.go`](src/service/consumer/task.go:36-43) for queue constants and task processing logic.

### Key Dependencies

- **Chaos Engineering**: `github.com/LGU-SE-Internal/chaos-experiment` for fault injection
- **Kubernetes**: `controller-runtime` for K8s controllers
- **Tracing**: OpenTelemetry for distributed tracing
- **Router**: Gin for HTTP routing
- **Database**: GORM for MySQL ORM

## Configuration

### Local Development

- Configuration: [`src/config.dev.toml`](src/config.dev.toml)
- Default ports: API (8082), MySQL (3306), Redis (6379), Jaeger (16686)
- Namespace: `exp` (Kubernetes)

### Environment Variables

- `ENV_MODE` - Environment mode: `dev`, `test`, or `prod`
- Default configs are in `src/config.*.toml` files

## Testing Guidelines

### Infrastructure Requirements

Tests require these services to be running:

```bash
docker compose up redis mysql jaeger buildkitd -d
```

### Expected Timings

- Go build: ~13 seconds (first time with dependencies)
- Go tests: ~15 seconds (with infrastructure running)
- Docker Compose startup: ~20 seconds (including image pulls)
- **CRITICAL**: Set timeouts to 60+ seconds to avoid premature cancellation

### Known Limitations

- Some tests require Kubernetes cluster access and will fail in environments without K8s
- Full application functionality requires K8s cluster (`stat /home/runner/.kube/config: no such file or directory` error is expected in non-K8s environments)

## Project-Specific Patterns

### Tracing Context

Tasks use hierarchical tracing contexts:

- **Group context** - Top-level trace (grandfather span)
- **Trace context** - Task type-specific span (father span)
- **Task context** - Individual task execution span

See [`extractContext()`](src/service/consumer/task.go:275-306) in task.go.

### Error Handling

- Tasks use retry policies with configurable max attempts and backoff
- Failed tasks move to dead letter queue with automatic retry
- Cancellation supported via context cancellation registry

### Database Models

- GORM-based models in [`src/database/`](src/database/)
- Use repository pattern in [`src/repository/`](src/repository/) for data access
- Soft deletes supported via `gorm.DeletedAt`

### API Versioning

- V2 API endpoints in [`src/handlers/v2/`](src/handlers/v2/)
- Swagger annotations on handlers for auto-documentation
- JWT authentication required for most endpoints

## Development Workflow

1. Start infrastructure: `docker compose up redis mysql jaeger buildkitd -d`
2. Build application: `cd src && go build -o /tmp/rcabench ./main.go`
3. Run tests: `cd src && go test ./utils/... -v`
4. For API changes: Regenerate swagger (`make swag-init`) and SDK (`make generate-python-sdk`)
5. Use `make local-debug` for interactive development with live reload

## Troubleshooting

- **Database connection issues**: Ensure MySQL container is running and accessible
- **Kubernetes errors**: Expected in non-K8s environments; application requires K8s for full functionality
- **BuildKit failures**: May occasionally fail to start; doesn't affect core development
- **Import errors**: Run `go mod tidy` in `src/` directory
