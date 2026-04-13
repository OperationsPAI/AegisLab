# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

AegisLab (RCABench) is a comprehensive Root Cause Analysis (RCA) benchmarking platform for microservices environments. It provides automated fault injection, algorithm execution, and evaluation capabilities for distributed systems research.

## Development Commands

### Environment Setup

1. **Install Nix** (devbox 的前置依赖):
   ```bash
   curl --proto '=https' --tlsv1.2 -sSf -L https://install.determinate.systems/nix | sh -s -- install
   ```

2. **Install devbox**:
   ```bash
   curl -fsSL https://get.jetify.com/devbox | bash
   ```

3. **安装项目依赖**:
   ```bash
   devbox install
   ```

4. **激活 devbox 环境** (每次新开终端都需要):
   ```bash
   eval "$(devbox shellenv)"
   ```

- `make check-prerequisites` - Verify development dependencies (devbox, docker, helm, kubectl, kubectx)
- `make setup-dev-env` - Bootstrap local development environment (installs uv, applies K8s manifests, installs Lefthook)

### Local Development

- `docker compose up redis mysql jaeger buildkitd -d` - Start infrastructure services (REQUIRED before testing/building)
- `make local-debug` - Start the Go application locally (runs `src/main.go both --port 8082`)
- `cd src && go build -tags duckdb_arrow -o /tmp/rcabench ./main.go` - Build the application binary

> **重要**: 编译必须加 `-tags duckdb_arrow`，因为 `duckdb-go/v2` 的 Arrow 接口使用了 `//go:build duckdb_arrow` 构建标签。不加此 tag 会报 `undefined: duckdb.NewArrowFromConn`。

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

### Workflow Orchestration via CRD Callbacks

**Important**: Workflow orchestration is NOT implemented as a separate "Workflow" entity. Instead, it's implemented through **Kubernetes CRD event callbacks** that automatically trigger the next task in the pipeline.

#### How It Works

1. **CRD Lifecycle Events**: When a Chaos Mesh CRD (fault injection) completes, the K8s controller detects it
2. **Callback Trigger**: `HandleCRDSucceeded()` callback is invoked in [`src/service/consumer/k8s_handler.go:230`](src/service/consumer/k8s_handler.go:230)
3. **Next Task Submission**: The callback automatically submits the next task (e.g., BuildDatapack) using `common.SubmitTask()`
4. **Chain Reaction**: Each task completion triggers the next task in the workflow

#### Workflow Chain Example

```
FaultInjection (CRD) → HandleCRDSucceeded() → SubmitTask(BuildDatapack)
                                                    ↓
                                            Job Completes → HandleJobSucceeded()
                                                    ↓
                                            SubmitTask(RunAlgorithm)
                                                    ↓
                                            Job Completes → HandleJobSucceeded()
                                                    ↓
                                            SubmitTask(CollectResult)
```

#### Key Implementation Details

- **Callback Interface**: [`src/client/k8s/controller.go:46-54`](src/client/k8s/controller.go:46-54) defines the Callback interface
- **CRD Success Handler**: [`src/service/consumer/k8s_handler.go:230-317`](src/service/consumer/k8s_handler.go:230-317)
  - Parses K8s annotations and labels to extract task context
  - Updates injection state
  - Submits BuildDatapack task as child task
  - Handles batch injections with `batchManager`
- **Job Success Handler**: [`src/service/consumer/k8s_handler.go:517-650`](src/service/consumer/k8s_handler.go:517-650)
  - Handles different task types (BuildDatapack, RunAlgorithm, etc.)
  - Automatically submits next task based on current task type
  - Publishes events to Redis stream for real-time updates

#### Task Context Propagation

- **Annotations**: K8s annotations carry OpenTelemetry trace context (taskCarrier, traceCarrier)
- **Labels**: K8s labels carry task metadata (taskID, traceID, taskType, groupID, projectID, userID)
- **Parent-Child Relationship**: Each submitted task has `ParentTaskID` pointing to the previous task

#### Batch Injection Support

- **batchManager**: Tracks batch injection progress in [`src/service/consumer/fault_injection.go:34-89`](src/service/consumer/fault_injection.go:34-89)
- **Hybrid Mode**: When multiple faults are injected in parallel, the callback waits for all to complete before submitting BuildDatapack

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

## Testing Methodology

### Backend API Testing

After implementing new API endpoints, always test manually before committing:

1. **Start Infrastructure Services**:
   ```bash
   cd /home/nn/workspace/proj/AegisLab
   docker compose up -d redis mysql
   ```

2. **Build and Start Server**:
   ```bash
   cd src
   go build -o /tmp/aegislab-test ./main.go
   ENV_MODE=dev /tmp/aegislab-test both --port 8082
   ```

3. **Test Endpoints with curl**:
   ```bash
   # Test authentication
   curl -s -X POST http://localhost:8082/api/v2/auth/login \
     -H "Content-Type: application/json" \
     -d '{"username":"admin","password":"admin"}' | jq -r '.data.access_token'

   # Test new endpoint (should return 401 without auth)
   curl -s http://localhost:8082/api/v2/datapacks | jq '.'

   # Test with auth token
   TOKEN="your_token_here"
   curl -s -H "Authorization: Bearer $TOKEN" \
     http://localhost:8082/api/v2/datapacks?page=1&size=10 | jq '.'
   ```

4. **Verify in Logs**:
   ```bash
   tail -f /tmp/aegislab-server.log
   ```

### Frontend Testing

1. **Start Frontend Dev Server**:
   ```bash
   cd frontend
   npm run dev
   ```

2. **Manual Browser Testing**:
   - Navigate to http://localhost:3000
   - Test new pages and features
   - Check browser console for errors
   - Verify API calls in Network tab

3. **Build Verification**:
   ```bash
   npm run build
   npm run preview
   ```

### Integration Testing

1. **End-to-End Workflow**:
   - Create injection → verify datapack created → run algorithm → view results
   - Test real-time updates (SSE)
   - Verify navigation and routing

2. **Common Issues**:
   - **CORS errors**: Check API proxy in `vite.config.ts`
   - **Auth failures**: Verify token storage in localStorage
   - **404 errors**: Check route definitions in `App.tsx`

## Troubleshooting

- **Database connection issues**: Ensure MySQL container is running and accessible
- **Kubernetes errors**: Expected in non-K8s environments; application requires K8s for full functionality
- **BuildKit failures**: May occasionally fail to start; doesn't affect core development
- **Import errors**: Run `go mod tidy` in `src/` directory

## SDK Generation

### Adding APIs to SDK

The SDK generator only includes APIs that are explicitly marked for SDK inclusion. To add an API to the generated SDK:

1. Add `@x-api-type {"sdk":"true"}` annotation to the handler function's Swagger comments:

```go
// @Router  /api/v2/your-endpoint [get]
// @x-api-type {"sdk":"true"}
func YourHandler(c *gin.Context) {
```

2. Regenerate the SDK:
```bash
make generate-typescript-sdk SDK_VERSION=0.0.0
```

3. Install the updated SDK in frontend:
```bash
cd frontend && npm install
```

### SDK Files Location
- **TypeScript SDK**: `sdk/typescript/` - Used by frontend via `@rcabench/client`
- **Python SDK**: `sdk/python/` - Used by CLI tools

### API Filtering Logic
The SDK generation script (`scripts/command/src/swagger.py`) filters APIs based on the `x-api-type.sdk` field. Only APIs with `{"sdk":"true"}` are included in the generated SDK.

## Reference Documentation

For comprehensive factual documentation about existing implementations, data models, and API endpoints, refer to:

- **[Codebase Facts & Relationships](/.claude/plans/codebase-facts-and-relationships.md)** - Complete inventory of:
  - Data model relationships (FaultInjection, Task, Execution, DatasetVersion, ContainerVersion)
  - All 58+ v2 API endpoints organized by domain
  - 30+ existing frontend pages and components
  - Key facts: FaultInjection = Datapack (same entity), workflow orchestration via CRD callbacks
  - Database relationships and label system

- **[Implementation Status](/.claude/plans/implementation-status-updated.md)** - Current progress:
  - Phase 1 completion status (75%)
  - Verified working components (backend API, frontend services)
  - Remaining tasks for Phase 1 completion
  - Phase 2-4 roadmap

These documents prevent duplicate implementation and provide quick reference for existing functionality.

<!-- auto-harness:begin -->
## North-Star Targets

1. **Full-Stack Spec Alignment** — 每条 requirement 在后端+前端+文档三处都有实现 (currently: 待测量)
   Measure: agent audit against `project-index.yaml`

2. **Zero Mock Code** — 前后端均无 mock 替代真实逻辑 (currently: 待测量)
   Measure: `grep` patterns in both repos (excluding `_test.go`)

3. **End-to-End Acceptance** — UI requirement 必须经用户浏览器验收才能标记 tested (currently: 0%)
   Measure: `acceptance.status=passed` / total UI requirements in `project-index.yaml`

Secondary: 合约优先于实现细节 — 前后端一致性 > 单端代码整洁

## Unified Spec

- `project-index.yaml` 是跨前后端的统一需求索引
- 每条 requirement 同时包含 `code`（后端）和 `frontend`（前端）字段
- 前端仓库通过 symlink 引用此文件
- **所有变更必须追溯到 spec 中的某条 requirement**

## Active Skills

- dev-loop — 完整开发循环: implement → test → vibe-check → AI-review → measure
- north-star — 量化优化目标与观测机制
- long-horizon — 自主决策与升级阶梯 (L1–L5)
- existing-project — 存量代码库需求恢复
- aegislab-dev-loop-profile — AegisLab 全栈定制 dev-loop (项目具体命令和门禁)
- aegislab-north-star — AegisLab 全栈定制 north-star (3 个核心目标和观测优先级)

## Full-Stack Development Rules

- **Spec 是唯一入口**: 改代码前先确认对应的 requirement
- **后端先行**: 先实现 API，再生成 SDK，再实现前端
- **零 Mock**: 前端必须调用真实 API，后端非测试代码不允许 mock
- **SDK 同步**: API 变更后必须 `just swag-init && just generate-typescript-client`
- **用户验收**: UI requirement 标记 tested 前必须请求用户浏览器验收
- **编译必带 tag**: `go build -tags duckdb_arrow`

## Observation Gates (P0 — 每次变更必须通过)

```bash
# Backend
cd src && go build -tags duckdb_arrow -o /dev/null ./main.go
cd src && golangci-lint run
cd src && go test ./utils/... -v

# Frontend
cd ../AegisLab-frontend && pnpm type-check && pnpm lint && pnpm build
```
<!-- auto-harness:end -->
