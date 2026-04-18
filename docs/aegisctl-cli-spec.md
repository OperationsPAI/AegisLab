# aegisctl CLI Client Specification

## Overview

`aegisctl` is a Go-based command-line client for the AegisLab (RCABench) backend API. It enables AI agents and human operators to drive the full RCA experiment lifecycle from the terminal — fault injection, progress monitoring, algorithm execution, and result inspection — without manual curl commands or browser interaction.

## Design Principles

### 1. Name-First, Semantic-Driven

All resource references use **names** instead of numeric IDs. The CLI internally resolves names to IDs via the API, keeping the interface human-readable and agent-friendly.

```bash
# Good — semantic
aegisctl inject get pod-kill-ts-order-20260413
aegisctl execute submit --project train-ticket --spec exec.yaml

# Bad — opaque IDs
aegisctl inject get 42
aegisctl execute submit --project 7 --spec exec.yaml
```

**Exception**: Resources without semantic names (task IDs, trace IDs, execution IDs) use their UUID/numeric identifiers directly.

### 2. Machine-Parseable Output

Every command supports `--output json` (alias `-o json`) for agent consumption. Default output is `table` for human readability.

### 3. Exit Code Convention

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Client error (invalid input, missing config, validation failure) |
| 2 | Server error (API returned 4xx/5xx) |
| 3 | Timeout (used by `wait` command) |

### 4. Stream Separation

- `stdout`: structured output only (data, tables, JSON)
- `stderr`: errors, warnings, progress messages

This allows agents to pipe stdout to `jq` while still seeing errors.

---

## Authentication & Configuration

### Config File

Location: `~/.aegisctl/config.yaml`

```yaml
current-context: dev

contexts:
  dev:
    server: http://localhost:8082
    token: eyJhbGci...
    default-project: train-ticket
    token-expiry: 2026-04-14T03:00:00Z
  staging:
    server: https://aegislab-staging.example.com
    token: eyJhbGci...

preferences:
  output: table           # Default output format (table|json|wide|yaml)
  request-timeout: 30s    # HTTP request timeout
```

### Token Resolution Priority (highest to lowest)

1. `--token` command-line flag
2. `AEGIS_TOKEN` environment variable
3. `token` field in the active context of `config.yaml`

### Environment Variable Overrides

| Variable | Purpose |
|----------|---------|
| `AEGIS_SERVER` | API server URL |
| `AEGIS_TOKEN` | Authentication token |
| `AEGIS_PROJECT` | Default project name |
| `AEGIS_OUTPUT` | Default output format |
| `AEGIS_TIMEOUT` | Default request timeout |

### Token Auto-Refresh

Before each request, check `token-expiry`. If the token will expire within 5 minutes, automatically call `/api/v2/auth/refresh` and update the config file.

---

## Command Tree

### Global Flags

Available on all commands:

| Flag | Short | Env Var | Description |
|------|-------|---------|-------------|
| `--server` | `-s` | `AEGIS_SERVER` | API server URL |
| `--token` | `-t` | `AEGIS_TOKEN` | Authentication token |
| `--project` | `-p` | `AEGIS_PROJECT` | Default project name |
| `--output` | `-o` | `AEGIS_OUTPUT` | Output format: `table`, `json`, `wide`, `yaml` |
| `--request-timeout` | | `AEGIS_TIMEOUT` | HTTP request timeout (default: 30s) |
| `--quiet` | `-q` | | Suppress progress/info messages on stderr |
| `--dry-run` | | | Validate input without submitting (where applicable) |

---

### `aegisctl auth` — Authentication

#### `aegisctl auth login`

Authenticate and persist token.

```bash
# Interactive (prompts for password)
aegisctl auth login --server http://localhost:8082 --username admin

# Non-interactive (for scripts/agents)
aegisctl auth login --server http://localhost:8082 --username admin --password admin

# With context name
aegisctl auth login --server http://localhost:8082 --username admin --password admin --context dev
```

**Behavior**:
- Calls `POST /api/v2/auth/login`
- Saves token + server + expiry to `~/.aegisctl/config.yaml`
- Sets as `current-context` if no context exists yet
- Prints authentication status to stdout

**Flags**:

| Flag | Required | Description |
|------|----------|-------------|
| `--server` | Yes | API server URL |
| `--username` | Yes | Username |
| `--password` | No | Password (prompts if omitted) |
| `--context` | No | Context name to save as (default: hostname-derived) |

#### `aegisctl auth status`

Show current authentication status.

```bash
aegisctl auth status
# Output:
# Context:  dev
# Server:   http://localhost:8082
# User:     admin
# Token:    eyJh...xyz (expires: 2026-04-14T03:00:00Z)
# Status:   valid
```

**Behavior**: Calls `GET /api/v2/auth/profile` to verify token validity.

#### `aegisctl auth token`

Directly set an API token without login flow.

```bash
aegisctl auth token --set eyJhbGci...
```

**Use case**: CI/CD pipelines or agents that receive tokens from external secret managers.

---

### `aegisctl context` — Multi-Environment Management

#### `aegisctl context set`

Create or update a context.

```bash
aegisctl context set --name staging --server https://aegislab-staging.example.com
aegisctl context set --name dev --default-project train-ticket
```

#### `aegisctl context use`

Switch active context.

```bash
aegisctl context use staging
```

#### `aegisctl context list`

List all configured contexts.

```bash
aegisctl context list
# Output:
# NAME      SERVER                                    DEFAULT-PROJECT   CURRENT
# dev       http://localhost:8082                      train-ticket      *
# staging   https://aegislab-staging.example.com       -                 
```

---

### `aegisctl project` — Project Management

#### `aegisctl project list`

```bash
aegisctl project list
aegisctl project list -o json
```

**API**: `GET /api/v2/projects`

#### `aegisctl project get`

```bash
aegisctl project get train-ticket
aegisctl project get train-ticket -o json
```

**API**: `GET /api/v2/projects/:project_id` (resolved from name)

#### `aegisctl project create`

```bash
aegisctl project create --name train-ticket --description "Train ticket microservice system"
```

**API**: `POST /api/v2/projects`

---

### `aegisctl container` — Container Management

#### `aegisctl container list`

```bash
aegisctl container list
aegisctl container list --type algorithm
aegisctl container list --type pedestal
aegisctl container list --type benchmark
```

**API**: `GET /api/v2/containers`

**Flags**:

| Flag | Description |
|------|-------------|
| `--type` | Filter by container type: `algorithm`, `benchmark`, `pedestal` |

#### `aegisctl container get`

```bash
aegisctl container get train-ticket
```

**API**: `GET /api/v2/containers/:container_id`

#### `aegisctl container versions`

```bash
aegisctl container versions train-ticket
```

**API**: `GET /api/v2/containers/:container_id/versions`

#### `aegisctl container build`

```bash
aegisctl container build train-ticket --version v1.0.0
```

**API**: `POST /api/v2/containers/build`

---

### `aegisctl inject` — Fault Injection (Core)

#### `aegisctl inject submit`

Submit a fault injection experiment.

```bash
aegisctl inject submit --project train-ticket --spec injection-spec.yaml
aegisctl inject submit --project train-ticket --spec injection-spec.yaml --dry-run
aegisctl inject submit --project train-ticket --spec injection-spec.yaml -o json
```

**API**: `POST /api/v2/projects/:project_id/injections/inject`

**Spec File Format** (`injection-spec.yaml`):

```yaml
pedestal:
  name: train-ticket
  version: v1.0.0
benchmark:
  name: jaeger-collector
  version: v1.0.0
interval: 30          # Total experiment interval in minutes
pre_duration: 10      # Normal data collection duration before fault injection (minutes)
specs:
  # Each top-level element is a batch; faults within a batch run in parallel
  - - type: pod-kill
      namespace: ts
      target: ts-order-service
      duration: 60s
  - - type: cpu-stress
      namespace: ts
      target: ts-payment-service
      duration: 120s
    - type: network-delay
      namespace: ts
      target: ts-order-service
      duration: 120s
algorithms:             # Optional: RCA algorithms to execute after injection
  - name: rca-algo-1
    version: v1.0.0
    env_vars:
      - key: THRESHOLD
        value: "0.5"
labels:                 # Optional: labels to attach
  - key: experiment
    value: batch-001
  - key: scenario
    value: cascade-failure
```

**JSON Output** (on success):

```json
{
  "trace_id": "abc-123-def-456",
  "group_id": "grp-789",
  "tasks": [
    {"task_id": "task-001", "type": "RestartPedestal", "state": "Pending"}
  ]
}
```

**`--dry-run` behavior**: Validate the spec file against the server (check container names exist, spec structure valid) without submitting. Exit 0 if valid, exit 1 with validation errors.

#### `aegisctl inject list`

```bash
aegisctl inject list --project train-ticket
aegisctl inject list --project train-ticket --state build_success
aegisctl inject list --project train-ticket --fault-type pod-kill
aegisctl inject list --project train-ticket --labels experiment=batch-001
aegisctl inject list --project train-ticket --page 1 --size 20
```

**API**: `GET /api/v2/projects/:project_id/injections`

**Flags**:

| Flag | Description |
|------|-------------|
| `--state` | Filter by datapack state: `initial`, `inject_failed`, `inject_success`, `build_failed`, `build_success`, `detector_failed`, `detector_success` |
| `--fault-type` | Filter by chaos type |
| `--labels` | Filter by labels (comma-separated `key=value` pairs) |
| `--page` | Page number (default: 1) |
| `--size` | Page size (default: 20) |

**Table Output**:

```
NAME                           STATE           FAULT-TYPE   START-TIME            LABELS
pod-kill-ts-order-20260413     build_success   pod-kill     2026-04-13T10:00:00Z  experiment=batch-001
cpu-stress-ts-payment-20260413 inject_success  cpu-stress   2026-04-13T10:05:00Z  experiment=batch-001
```

#### `aegisctl inject get`

```bash
aegisctl inject get pod-kill-ts-order-20260413
aegisctl inject get pod-kill-ts-order-20260413 -o json
```

**API**: `GET /api/v2/injections/:id`

#### `aegisctl inject search`

Advanced search with multiple filters.

```bash
aegisctl inject search --project train-ticket --name-pattern "pod-kill-*" --labels experiment=batch-001
```

**API**: `POST /api/v2/projects/:project_id/injections/search`

#### `aegisctl inject logs`

```bash
aegisctl inject logs pod-kill-ts-order-20260413
```

**API**: `GET /api/v2/injections/:id/logs`

#### `aegisctl inject files`

```bash
aegisctl inject files pod-kill-ts-order-20260413
```

**API**: `GET /api/v2/injections/:id/files`

**Table Output**:

```
PATH                          SIZE      TYPE
traces/trace.parquet          12.3 MB   parquet
metrics/cpu.parquet            5.1 MB   parquet
logs/service.log               2.0 MB   text
groundtruth.yaml               0.1 KB   yaml
```

#### `aegisctl inject download`

```bash
aegisctl inject download pod-kill-ts-order-20260413 -o /tmp/datapack/
```

**API**: `GET /api/v2/injections/:id/download`

#### `aegisctl inject metadata`

Show available fault types, resources, and status mappings.

```bash
aegisctl inject metadata
```

**API**: `GET /api/v2/injections/metadata`

**Output**:

```
FAULT TYPES:
  pod-kill         Kill target pods
  cpu-stress       Inject CPU stress
  memory-stress    Inject memory stress
  network-delay    Add network latency
  network-loss     Inject packet loss
  ...

DATAPACK STATES:
  initial, inject_failed, inject_success, build_failed,
  build_success, detector_failed, detector_success
```

---

### `aegisctl execute` — Algorithm Execution

#### `aegisctl execute submit`

```bash
aegisctl execute submit --project train-ticket --spec execution-spec.yaml
aegisctl execute submit --project train-ticket --spec execution-spec.yaml -o json
```

**API**: `POST /api/v2/projects/:project_id/executions/execute`

**Spec File Format** (`execution-spec.yaml`):

```yaml
specs:
  - algorithm:
      name: rca-algo-1
      version: v1.0.0
    datapack: pod-kill-ts-order-20260413      # Reference by injection name
  - algorithm:
      name: rca-algo-2
      version: v2.0.0
    dataset:                                   # Or reference by dataset
      name: train-ticket-dataset
      version: v1.0.0
labels:
  - key: batch
    value: comparison-run
```

**Note**: `project_name` is automatically set from `--project` flag; do not include in spec file.

#### `aegisctl execute list`

```bash
aegisctl execute list --project train-ticket
```

**API**: `GET /api/v2/projects/:project_id/executions`

#### `aegisctl execute get`

```bash
aegisctl execute get 123
aegisctl execute get 123 -o json
```

**API**: `GET /api/v2/executions/:execution_id`

---

### `aegisctl task` — Task Monitoring

#### `aegisctl task list`

```bash
aegisctl task list
aegisctl task list --state Running
aegisctl task list --type FaultInjection
```

**API**: `GET /api/v2/tasks`

**Flags**:

| Flag | Description |
|------|-------------|
| `--state` | Filter: `Pending`, `Running`, `Completed`, `Error`, `Cancelled`, `Rescheduled` |
| `--type` | Filter: `BuildContainer`, `RestartPedestal`, `FaultInjection`, `RunAlgorithm`, `BuildDatapack`, `CollectResult`, `CronJob` |
| `--overdue` | Show only Pending tasks whose `execute_time` is already in the past (WAIT < 0) |

**Table Output**:

```
TASK-ID          TYPE              STATE      WAIT     TRACE-ID         PROJECT       CREATED
task-abc123      RestartPedestal   Running    -        trace-def456     train-ticket  2m ago
task-xyz789      FaultInjection    Pending    +01:23   trace-def456     train-ticket  1m ago
task-overdue     BuildDatapack     Pending    -00:05   trace-def456     train-ticket  3m ago
```

**`WAIT` column**: for `Pending` rows, shows the signed remaining time until
`execute_time`, rendered as `+MM:SS` (still waiting) or `-MM:SS` (overdue — the
scheduler has not picked it up yet). Non-Pending rows show `-`.

#### `aegisctl task expedite`

Force a `Pending` task to run on the next scheduler tick.

```bash
aegisctl task expedite <task-id>
```

**API**: `POST /api/v2/tasks/:task_id/expedite`

Atomically resets the task's `execute_time` to now in both the MySQL `tasks`
table and the Redis `task:delayed` sorted set. The consumer emits a
`task.scheduled` trace event with `reason=expedite`.

- Rejects with `state=<X>, cannot expedite` if the task is not in `Pending`.
- Idempotent: expediting an already-due task succeeds silently.
- The CLI never talks to Redis directly — all atomic work happens server-side.

#### `aegisctl task get`

```bash
aegisctl task get task-abc123
aegisctl task get task-abc123 -o json
```

**API**: `GET /api/v2/tasks/:task_id`

#### `aegisctl task logs`

Stream task logs in real-time.

```bash
aegisctl task logs task-abc123
aegisctl task logs task-abc123 --follow    # Continuously stream via WebSocket
```

**API**: `GET /api/v2/tasks/:task_id/logs/ws` (WebSocket)

**`--follow` behavior**: Keep WebSocket connection open, print new log lines as they arrive. Ctrl+C to stop.

**Without `--follow`**: Connect, read available logs, disconnect.

---

### `aegisctl trace` — Experiment Tracing

#### `aegisctl trace list`

```bash
aegisctl trace list
aegisctl trace list --project train-ticket
aegisctl trace list --state Running
```

**API**: `GET /api/v2/traces`

**Flags**:

| Flag | Description |
|------|-------------|
| `--project` | Filter by project name |
| `--state` | Filter: `Pending`, `Running`, `Completed`, `Failed` |
| `--group-id` | Filter by group ID |

**Table Output**:

```
TRACE-ID       TYPE            STATE       PROJECT         START-TIME            TASKS
trace-abc123   FullPipeline    Running     train-ticket    2026-04-13T10:00:00Z  3/5
trace-def456   AlgorithmRun    Completed   train-ticket    2026-04-13T09:30:00Z  2/2
```

#### `aegisctl trace get`

```bash
aegisctl trace get trace-abc123
aegisctl trace get trace-abc123 -o json
```

**API**: `GET /api/v2/traces/:trace_id`

**Detailed Output** (includes child tasks):

```
Trace: trace-abc123
Type:  FullPipeline
State: Running
Start: 2026-04-13T10:00:00Z

Tasks:
  TASK-ID        TYPE              STATE       DURATION
  task-001       RestartPedestal   Completed   45s
  task-002       FaultInjection    Completed   5m30s
  task-003       BuildDatapack     Running     2m10s (in progress)
  task-004       RunAlgorithm      Pending     -
  task-005       CollectResult     Pending     -
```

#### `aegisctl trace watch`

Real-time SSE event stream for a trace.

```bash
aegisctl trace watch trace-abc123
```

**API**: `GET /api/v2/traces/:trace_id/stream` (SSE)

**Output** (streaming):

```
[10:00:05] RestartPedestal  task-001  Running    Restarting pedestal...
[10:00:45] RestartPedestal  task-001  Completed  Pedestal restarted successfully
[10:00:46] FaultInjection   task-002  Running    Injecting pod-kill on ts-order-service
[10:06:16] FaultInjection   task-002  Completed  Fault injection completed
[10:06:17] BuildDatapack    task-003  Running    Building datapack...
...
```

**Termination**: Stream ends when trace reaches terminal state (`Completed` or `Failed`), or on Ctrl+C.

---

### `aegisctl dataset` — Dataset Management

#### `aegisctl dataset list`

```bash
aegisctl dataset list
```

**API**: `GET /api/v2/datasets`

#### `aegisctl dataset get`

```bash
aegisctl dataset get train-ticket-dataset
```

**API**: `GET /api/v2/datasets/:dataset_id`

#### `aegisctl dataset versions`

```bash
aegisctl dataset versions train-ticket-dataset
```

**API**: `GET /api/v2/datasets/:dataset_id/versions`

---

### `aegisctl eval` — Evaluation Results

#### `aegisctl eval list`

```bash
aegisctl eval list
```

**API**: `GET /api/v2/evaluations`

#### `aegisctl eval get`

```bash
aegisctl eval get 123
```

**API**: `GET /api/v2/evaluations/:id`

---

### `aegisctl wait` — Block Until Completion

Block execution until a trace or task reaches a terminal state. This is the primary synchronization primitive for agents.

```bash
aegisctl wait trace-abc123
aegisctl wait trace-abc123 --timeout 600s
aegisctl wait task-xyz789 --timeout 300s --interval 5s
aegisctl wait trace-abc123 --exit-on error
```

**Behavior**:
1. Detect whether the argument is a trace ID or task ID (by format or API probe)
2. Poll the status at `--interval` (default: 5s)
3. Print status line on each poll (to stderr, unless `--quiet`)
4. Exit when terminal state is reached or timeout

**Flags**:

| Flag | Default | Description |
|------|---------|-------------|
| `--timeout` | `600s` | Maximum wait time |
| `--interval` | `5s` | Poll interval |
| `--exit-on` | `completed,error` | Which terminal states to exit on |
| `--quiet` | `false` | Suppress polling status output |

**Exit codes**:
- `0`: Completed successfully
- `2`: Completed with error/failure
- `3`: Timeout

**JSON output** (`-o json`): On exit, prints the final resource state to stdout.

```json
{
  "id": "trace-abc123",
  "state": "Completed",
  "duration": "5m30s",
  "tasks_completed": 5,
  "tasks_total": 5
}
```

**Polling status** (stderr):

```
Waiting for trace-abc123... [Running] BuildDatapack (3/5 tasks) 2m10s elapsed
Waiting for trace-abc123... [Running] RunAlgorithm (4/5 tasks) 4m30s elapsed
Waiting for trace-abc123... [Completed] 5/5 tasks in 5m30s
```

---

### `aegisctl status` — Global Overview

Show cluster status, task summary, recent traces, and infrastructure health.

```bash
aegisctl status
aegisctl status -o json
```

**Output**:

```
Server:    http://localhost:8082 (dev)
User:      admin
Connected: yes

Active Tasks:     3
  Running:        2
  Pending:        1

Recent Traces:
Trace-ID       State       Type            Project
trace-abc123   Running     FullPipeline    train-ticket
trace-def456   Completed   AlgorithmRun    train-ticket

Infrastructure Health:
  ✓ buildkit     2ms
  ✓ database     3.5ms
  ✓ jaeger       1ms
  ✓ kubernetes   5ms
  ✓ redis        1.2ms
```

**Unhealthy service output**:

```
Infrastructure Health:
  ✓ buildkit     2ms
  ✗ database     N/A (connection refused)
  ✓ jaeger       1ms
  ✓ kubernetes   5ms
  ✓ redis        1.2ms
```

**API**: Aggregates `GET /api/v2/auth/profile`, `GET /api/v2/tasks`, `GET /api/v2/traces`, and `GET /system/health`.

**Behavior**:
- Calls `/system/health` to check Redis, MySQL/database, Kubernetes, Jaeger, and BuildKit connectivity.
- Displays green (`✓`) for healthy services and red (`✗`) for unhealthy services with ANSI color codes.
- Services are listed in alphabetical order.
- If the health endpoint is unreachable, a single `✗` line indicates the failure.
- `--output json` returns a combined JSON object with `server`, `context`, `connected`, `username`, `tasks`, `recent_traces`, and `health` fields.

---

### `aegisctl cluster` — Cluster Dependency Management

Operations that target the AegisLab cluster and its backing services
(Kubernetes, MySQL, ClickHouse, Redis, etcd).

#### `aegisctl cluster preflight`

Verify that every dependency required by AegisLab is reachable and
configured. The command prints one row per check with `[OK]` / `[FAIL]` /
`[WARN]` and a suggested fix on failure. Overall exit code is `0` when
every executed check is OK, `1` otherwise.

```bash
aegisctl cluster preflight
aegisctl cluster preflight --check k8s.rcabench-sa
aegisctl cluster preflight --fix
aegisctl cluster preflight --config /path/to/config.dev.toml
```

**Flags**:

| Flag | Description |
|------|-------------|
| `--check <id>` | Run only the named check |
| `--fix` | Apply idempotent remediation for failing checks that support it |
| `--config <path>` | Path to a specific config TOML (defaults to `config.$ENV_MODE.toml` in cwd) |
| `--check-timeout <sec>` | Per-check timeout (default: 10s) |

**Check catalog**:

| ID | Description | `--fix` support |
|----|-------------|-----------------|
| `k8s.exp-namespace` | namespace `exp` exists | — |
| `k8s.rcabench-sa` | ServiceAccount `rcabench-sa` in `exp` exists | yes (kubectl create sa) |
| `k8s.dataset-pvc` | PVC `rcabench-juicefs-dataset` in `exp` exists & Bound | — (storage-class decision required) |
| `k8s.chaosmesh-crds` | `chaos-mesh.org` CRDs present | — |
| `db.mysql` | TCP reachable using `database.mysql.host:port` | — |
| `db.clickhouse` | TCP reachable using `database.clickhouse.host:port` | — |
| `db.redis` | TCP reachable using `redis.host` | — |
| `db.etcd` | TCP reachable using `etcd.endpoints[0]` | — |
| `clickhouse.otel-tables` | `otel_traces`, `otel_metrics_gauge`, `otel_metrics_sum`, `otel_metrics_histogram`, `otel_logs` tables exist in the `otel` db | — |
| `redis.token-bucket-leaks` | no terminal tasks leaking slots in `token_bucket:restart_service` | yes (SREM leaked task_ids) |

**Output** (truncated):

```
CHECK                     STATUS  DETAIL
------------------------  ------  --------------------
k8s.exp-namespace         [OK]    namespace "exp" present
k8s.rcabench-sa           [FAIL]  ServiceAccount exp/rcabench-sa missing
                                    fix: kubectl -n exp create serviceaccount rcabench-sa (or rerun with --fix)
k8s.dataset-pvc           [OK]    exp/rcabench-juicefs-dataset Bound
...
```

**Config resolution**: The command reads `config.$ENV_MODE.toml` (default
`ENV_MODE=dev`) from the current working directory. Required keys:
`[database.mysql] host/port`, `[database.clickhouse] host/port/database`,
`redis.host`, `etcd.endpoints`, `k8s.namespace`, and the JuiceFS PVC +
service-account names under `[k8s.job.*]`.

**Not yet implemented** (intentionally, to keep preflight fast):

- container_versions registry pullability — too slow for synchronous run.
- `helm_configs.repo_url` reachability — too slow.

Both are tracked as TODO comments in `src/cmd/aegisctl/cluster/checks.go`.

---

### `aegisctl completion` — Shell Completion

```bash
aegisctl completion bash > /etc/bash_completion.d/aegisctl
aegisctl completion zsh > "${fpath[1]}/_aegisctl"
aegisctl completion fish > ~/.config/fish/completions/aegisctl.fish
```

---

## Internal Architecture

### Directory Structure

```
src/cmd/aegisctl/
├── main.go                    # Entry point
├── cmd/
│   ├── root.go                # Cobra root command + global flags
│   ├── auth.go                # auth login, status, token
│   ├── context.go             # context set, use, list
│   ├── project.go             # project list, get, create
│   ├── container.go           # container list, get, versions, build
│   ├── inject.go              # inject submit, list, get, search, logs, files, download, metadata
│   ├── execute.go             # execute submit, list, get
│   ├── task.go                # task list, get, logs
│   ├── trace.go               # trace list, get, watch
│   ├── dataset.go             # dataset list, get, versions
│   ├── eval.go                # eval list, get
│   ├── wait.go                # wait (poll trace/task state)
│   ├── status.go              # status overview
│   └── completion.go          # shell completion generation
├── client/
│   ├── client.go              # Core HTTP client (request/response/error handling)
│   ├── auth.go                # Token management + auto-refresh
│   ├── sse.go                 # SSE streaming (trace watch, group stream)
│   ├── ws.go                  # WebSocket (task logs --follow)
│   └── resolver.go            # Name-to-ID resolution + cache
├── config/
│   └── config.go              # ~/.aegisctl/config.yaml read/write
└── output/
    ├── format.go              # Output dispatcher (table/json/wide/yaml)
    ├── table.go               # Table formatting with column alignment
    └── printer.go             # stdout/stderr stream separation
```

### Name-to-ID Resolver

The resolver is the core abstraction that makes name-based references work. It maintains a short-lived cache to avoid redundant API calls within a single command session.

```go
type Resolver struct {
    client *Client
    cache  map[string]int  // key format: "resource_type:name" -> ID
    ttl    time.Duration   // Cache TTL (default: 5 minutes)
}

// Core resolution methods
func (r *Resolver) ProjectID(name string) (int, error)
func (r *Resolver) ContainerID(name string) (int, error)
func (r *Resolver) InjectionID(name string) (int, error)
func (r *Resolver) DatasetID(name string) (int, error)
```

**Resolution strategy**:
1. Check local cache
2. Call list API with name filter (e.g., `GET /api/v2/projects?name=train-ticket`)
3. If exactly one match, cache and return ID
4. If zero matches, return error: `project "train-ticket" not found`
5. If multiple matches, return error with disambiguation hint

### HTTP Client

```go
type Client struct {
    baseURL    string
    token      string
    httpClient *http.Client
    resolver   *Resolver
}

// APIResponse is the standard response envelope
type APIResponse[T any] struct {
    Code      int    `json:"code"`
    Message   string `json:"message"`
    Data      T      `json:"data"`
    Timestamp string `json:"timestamp"`
    Errors    []string `json:"errors,omitempty"`
}

// PaginatedData wraps list responses
type PaginatedData[T any] struct {
    Items      []T        `json:"items"`
    Pagination Pagination `json:"pagination"`
}

type Pagination struct {
    Page  int `json:"page"`
    Size  int `json:"size"`
    Total int `json:"total"`
    Pages int `json:"pages"`
}
```

### SSE Reader

```go
type SSEReader struct {
    url     string
    client  *http.Client
    token   string
    lastID  string
}

func (r *SSEReader) Stream(ctx context.Context) (<-chan SSEEvent, error)
```

### WebSocket Reader

```go
type WSReader struct {
    url   string
    token string
}

func (r *WSReader) Stream(ctx context.Context) (<-chan string, error)
```

---

## Agent Workflow Examples

### Example 1: Full Pipeline Experiment

```bash
#!/bin/bash
set -e

# Setup
aegisctl auth login --server http://aegislab:8082 --username agent --password secret

# Discover resources
ALGORITHMS=$(aegisctl container list --type algorithm -o json)
PEDESTALS=$(aegisctl container list --type pedestal -o json)

# Generate spec file (agent generates this programmatically)
cat > /tmp/inject-spec.yaml <<EOF
pedestal:
  name: train-ticket
  version: v1.0.0
benchmark:
  name: jaeger-collector
  version: v1.0.0
interval: 30
pre_duration: 10
specs:
  - - type: pod-kill
      namespace: ts
      target: ts-order-service
      duration: 60s
algorithms:
  - name: rca-algo-1
    version: v1.0.0
labels:
  - key: agent-run
    value: "001"
EOF

# Submit and capture trace ID
RESULT=$(aegisctl inject submit --project train-ticket --spec /tmp/inject-spec.yaml -o json)
TRACE_ID=$(echo "$RESULT" | jq -r '.trace_id')

# Wait for completion
aegisctl wait "$TRACE_ID" --timeout 900s
EXIT_CODE=$?

if [ $EXIT_CODE -eq 0 ]; then
    echo "Experiment completed successfully"
    # Inspect results
    aegisctl inject list --project train-ticket --labels agent-run=001 -o json
elif [ $EXIT_CODE -eq 2 ]; then
    echo "Experiment failed"
    # Check logs
    TASKS=$(aegisctl trace get "$TRACE_ID" -o json | jq -r '.tasks[] | select(.state=="Error") | .task_id')
    for TASK in $TASKS; do
        aegisctl task logs "$TASK"
    done
elif [ $EXIT_CODE -eq 3 ]; then
    echo "Experiment timed out"
fi
```

### Example 2: Batch Algorithm Comparison

```bash
#!/bin/bash
set -e

# Run multiple algorithms on the same datapack
DATAPACK="pod-kill-ts-order-20260413"

for ALGO in rca-algo-1 rca-algo-2 rca-algo-3; do
    cat > /tmp/exec-spec.yaml <<EOF
specs:
  - algorithm:
      name: $ALGO
      version: v1.0.0
    datapack: $DATAPACK
labels:
  - key: comparison
    value: batch-001
EOF

    RESULT=$(aegisctl execute submit --project train-ticket --spec /tmp/exec-spec.yaml -o json)
    TRACE_ID=$(echo "$RESULT" | jq -r '.trace_id')
    echo "Submitted $ALGO: trace=$TRACE_ID"
done

# Wait for all (simplified — real agent would track all trace IDs)
sleep 300
aegisctl eval list -o json | jq '.items[] | select(.labels | any(.key=="comparison" and .value=="batch-001"))'
```

### Example 3: Agent Discovery and Exploration

```bash
# Agent's first interaction — explore what's available
aegisctl status

# What projects exist?
aegisctl project list -o json

# What containers are available?
aegisctl container list -o json

# What fault types can I inject?
aegisctl inject metadata -o json

# What injections have been done?
aegisctl inject list --project train-ticket -o json

# What's the status of a specific injection?
aegisctl inject get pod-kill-ts-order-20260413 -o json

# What files are in the datapack?
aegisctl inject files pod-kill-ts-order-20260413 -o json
```

---

## Implementation Priority

### P0 — Minimal Viable Agent Workflow

Core capabilities needed for an agent to run a complete experiment cycle.

| # | Command | API Endpoint | Description |
|---|---------|-------------|-------------|
| 1 | `auth login` | `POST /api/v2/auth/login` | Authenticate and persist token |
| 2 | `auth token --set` | (local) | Set token directly |
| 3 | Config file read/write | (local) | `~/.aegisctl/config.yaml` management |
| 4 | `project list` | `GET /api/v2/projects` | List projects |
| 5 | `project get` | `GET /api/v2/projects/:id` | Get project details |
| 6 | `inject submit` | `POST /api/v2/projects/:pid/injections/inject` | Submit fault injection |
| 7 | `inject list` | `GET /api/v2/projects/:pid/injections` | List injections |
| 8 | `inject get` | `GET /api/v2/injections/:id` | Get injection details |
| 9 | `inject metadata` | `GET /api/v2/injections/metadata` | Get available fault types |
| 10 | `execute submit` | `POST /api/v2/projects/:pid/executions/execute` | Submit algorithm execution |
| 11 | `execute list` | `GET /api/v2/projects/:pid/executions` | List executions |
| 12 | `execute get` | `GET /api/v2/executions/:id` | Get execution details |
| 13 | `task list` | `GET /api/v2/tasks` | List tasks |
| 14 | `task get` | `GET /api/v2/tasks/:id` | Get task details |
| 15 | `wait` | poll `GET /api/v2/traces/:id` or `GET /api/v2/tasks/:id` | Block until completion |
| 16 | Name-to-ID resolver | various list APIs | Resolve names to IDs |
| 17 | `--output json` | (local) | JSON output on all commands |

### P1 — Complete Experience

| # | Command | Description |
|---|---------|-------------|
| 18 | `trace list/get` | Trace management |
| 19 | `trace watch` | SSE real-time streaming |
| 20 | `task logs --follow` | WebSocket log streaming |
| 21 | `inject logs` | Injection execution logs |
| 22 | `inject files` | Datapack file listing |
| 23 | `inject download` | Datapack download |
| 24 | `inject search` | Advanced injection search |
| 25 | `container list/get/versions` | Container management |
| 26 | `container build` | Trigger container build |
| 27 | `context set/use/list` | Multi-environment |
| 28 | `status` | Global overview |
| 29 | `--dry-run` | Validation without submission |
| 30 | Token auto-refresh | Automatic token renewal |

### P2 — Polish

| # | Command | Description |
|---|---------|-------------|
| 31 | `dataset list/get/versions` | Dataset management |
| 32 | `eval list/get` | Evaluation results |
| 33 | `project create` | Project creation |
| 34 | `completion` | Shell auto-completion |
| 35 | `auth status` | Authentication status check |

---

## Build & Install

```makefile
# Add to Makefile
.PHONY: aegisctl
aegisctl:
	cd src && go build -tags duckdb_arrow -o $(GOPATH)/bin/aegisctl ./cmd/aegisctl/main.go

# Or install directly
install-aegisctl:
	cd src && go install -tags duckdb_arrow ./cmd/aegisctl/
```

**Binary name**: `aegisctl`

**Dependencies**: Only Go standard library + `github.com/spf13/cobra` (CLI framework). Shares `aegis/dto` and `aegis/consts` types from the main codebase — no duplication.
