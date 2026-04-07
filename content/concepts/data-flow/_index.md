---
title: Data Flow
weight: 2
---

Data flow in AegisLab describes how a request is transformed from an API call into a reproducible experiment artifact and then into analysis-ready outputs. The design separates synchronous request handling from asynchronous execution so that the platform remains responsive under load while preserving complete execution context for RCA benchmarking.

## Request to Task

The flow starts at the Producer API layer. After validating the request payload, the platform persists metadata in MySQL and submits executable intent to Redis-backed task queues. This split keeps user-facing APIs fast and ensures task execution is decoupled from HTTP lifecycle constraints. At this stage, identity and trace context are attached to the task so downstream workers can preserve ownership, auditing, and correlation.

## Queue and Scheduling Lifecycle

AegisLab uses explicit queue roles to control execution state:

- Delayed queue for scheduled execution windows
- Ready queue for immediate consumption by workers
- Dead-letter queue for failure isolation and diagnostics

This queue design enables retry policies, bounded concurrency, and clear backpressure behavior. It also supports operational visibility because each transition reflects a meaningful state change in the pipeline.

## Fault Injection to Datapack

When Chaos Mesh CRD execution completes successfully, Kubernetes callback handlers update task and injection states, then automatically enqueue the next child task for datapack building. Parent-child task relationships preserve workflow lineage, which is critical for reproducibility and postmortem analysis. In hybrid or batch scenarios, the callback waits for all required branches to finish before continuing to the datapack stage.

## Data Collection and Analysis Readiness

During and after fault injection, logs, traces, and metrics are collected and converted into benchmark-consumable artifacts. Task events and timing metrics are emitted throughout execution, making it possible to reconstruct timelines, inspect bottlenecks, and compare algorithm behavior consistently across runs.

## Why This Matters

The data-flow model in AegisLab is built for repeatable RCA evaluation: every stage records enough context to replay, compare, and explain outcomes. This is the foundation for fair algorithm benchmarking in distributed microservice systems.

## Next Steps

- [Architecture](../architecture): Detailed system architecture
- [Fault Injection Lifecycle](../fault-injection-lifecycle): End-to-end fault injection process
- [Observability Stack](../observability-stack): Monitoring and telemetry foundations