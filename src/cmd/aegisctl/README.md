# aegisctl

Command-line client for the AegisLab (RCABench) platform. Designed for both human operators and AI agents to drive the full RCA experiment lifecycle from the terminal.

## Build

```bash
just build-aegisctl
# or manually:
cd src && go build -o /tmp/aegisctl ./cmd/aegisctl
```

Note: aegisctl does NOT require `-tags duckdb_arrow`.

## Quick Start

```bash
# Login
aegisctl auth login --server http://HOST:8082 --username admin --password admin123

# Set default project
aegisctl context set --name default --default-project pair_diagnosis

# Browse resources
aegisctl project list
aegisctl container list
aegisctl dataset list
```

## Subcommands

| Command | Description |
|---------|-------------|
| `auth` | Login, token management, whoami |
| `context` | Manage named configuration contexts |
| `project` | List, get, create projects |
| `container` | List, get, build containers |
| `inject` | Submit, list, get fault injections |
| `execute` | Submit, list algorithm executions |
| `task` | List, inspect background tasks |
| `trace` | List, get, watch execution traces |
| `dataset` | List, get, manage datasets |
| `eval` | List, get evaluation results |
| `wait` | Block until a resource reaches terminal state |
| `status` | View system status |
| `completion` | Generate shell completions |

## Environment Variables

| Variable | Description |
|----------|-------------|
| `AEGIS_SERVER` | Server URL (overridden by `--server`) |
| `AEGIS_TOKEN` | Auth token (overridden by `--token`) |
| `AEGIS_PROJECT` | Default project name (overridden by `--project`) |
| `AEGIS_OUTPUT` | Output format: `table` or `json` (overridden by `-o`) |
| `AEGIS_TIMEOUT` | Request timeout in seconds (overridden by `--request-timeout`) |

## Full Specification

See [docs/aegisctl-cli-spec.md](../../../docs/aegisctl-cli-spec.md) for the complete design specification.
