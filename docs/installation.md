# Installation Guide

This page describes baseline setup for local development and cluster deployment.

## Local Setup

Run required infrastructure services first:

```bash
docker compose up redis mysql jaeger buildkitd -d
```

Build and run backend:

```bash
cd src
go build -o /tmp/rcabench ./main.go
/tmp/rcabench both --port 8082
```

## Cluster Setup

Use existing automation targets:

- `make check-prerequisites`
- `make run`

## Next Docs

- [User Guide](user-guide.md)
- [Troubleshooting](troubleshooting.md)
