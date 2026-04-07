---
title: Observability Stack
weight: 4
---

The observability stack in AegisLab connects experiment execution with diagnosis evidence. Instead of treating traces, metrics, and logs as independent outputs, the platform correlates them around task and trace context so operators and researchers can reconstruct what happened during each fault injection and algorithm run.

## Signal Layers

AegisLab uses three complementary signal layers:

- Traces for service-to-service causal paths and timing relationships
- Metrics for workload behavior, queue pressure, and latency distributions
- Logs for step-level execution narratives and troubleshooting detail

This layered approach improves debugging speed because each signal answers a different question: whether something is wrong, where it propagates, and why it happened.

## Collection and Ingestion Path

Runtime components emit telemetry through OpenTelemetry-compatible paths. For task-centric log streaming, the backend provides an OTLP HTTP receiver that accepts log records, extracts task metadata, and republishes entries by task channels. This enables near-real-time visibility for long-running asynchronous workflows without requiring direct pod-log polling in every client.

## Historical Query and Live Stream

AegisLab combines historical retrieval with live updates:

1. Query historical logs from Loki for context before the current moment.
2. Subscribe to task-scoped Pub/Sub channels for real-time updates.

This two-phase flow gives users a complete timeline view rather than a fragmented stream.

## Correlation and RCA Value

RCA benchmarking requires reproducible evidence, not only final labels. By preserving task identifiers and trace context across telemetry paths, AegisLab enables repeatable postmortem analysis, fair algorithm comparison, and easier validation of edge cases.

## Operational Benefits

The observability stack supports both development and production operations:

- Faster incident triage
- Clearer task lifecycle diagnostics
- Better confidence in benchmark result interpretation

## Next Steps

- [Architecture](../architecture): End-to-end component layout
- [Data Flow](../data-flow): Request and task movement
- [Fault Injection Lifecycle](../fault-injection-lifecycle): Callback-driven lifecycle details
