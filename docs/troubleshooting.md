# Troubleshooting

Use this guide for common failures in setup, task execution, and observability checks.

## Common Checks

- Confirm Redis, MySQL, and tracing dependencies are running.
- Verify backend starts with expected config and ports.
- Inspect task state transitions and worker logs.
- Validate Kubernetes access for cluster-dependent operations.

## Triage Sequence

1. Reproduce with minimal scope.
2. Inspect related task logs and trace context.
3. Check environment and credentials.
4. Re-run isolated stage to identify failing boundary.

## Related Docs

- [Installation Guide](installation.md)
- [User Guide](user-guide.md)
