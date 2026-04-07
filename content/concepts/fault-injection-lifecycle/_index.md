---
title: Fault Injection Lifecycle
weight: 3
---

The fault injection lifecycle in AegisLab is event-driven and stateful. After a Chaos Mesh CRD finishes, the Kubernetes callback path turns that runtime signal into consistent task state transitions and the next executable stage in the RCA pipeline. The key behavior is implemented in the CRD success handler, where the system updates both observability events and business state before enqueuing datapack construction work.

## 1. CRD Success Callback Entry

When a fault CRD succeeds, the handler reconstructs task and trace context from annotations and labels. This ensures all follow-up actions stay correlated to the original execution lineage. The callback writes a success log and span event, so operators can see exactly when the fault phase completed in tracing and logs.

## 2. Task State Update and Event Emission

The lifecycle first marks the original fault-injection task as completed and publishes a domain event for fault completion. This state transition is important: datapack build is not started until fault injection is confirmed complete. In other words, lifecycle sequencing is strict, not speculative, which prevents downstream consumers from reading partial experiment state.

## 3. BuildDatapack Child Task Submission

Inside the post-processing closure, the handler updates injection state, writes injection timestamps, and prepares benchmark environment variables (including namespace context). It then builds a payload containing benchmark and datapack metadata and submits a child task of type BuildDatapack. The child task keeps the same trace and group context while setting the parent task relationship, which preserves full workflow ancestry for debugging and reproducibility.

## 4. Hybrid Batch Completion Gate

For non-hybrid injections, post-processing is executed immediately after CRD success. For hybrid mode, however, the lifecycle uses batch accounting to defer datapack build until all required branches are complete. The batch manager increments completion count, checks whether the batch is finished, and only then triggers post-processing once. This gate avoids duplicate datapack builds and guarantees that aggregated hybrid experiments are materialized only after all sub-faults have finished.

## Why This Lifecycle Design Matters

This lifecycle combines correctness and operability: strict phase boundaries, parent-child task chaining, and hybrid completion guards make experiment outputs reproducible and auditable. It also provides clear points for error handling and metrics, which is essential for large-scale RCA benchmarking.