---
title: Core Concepts
weight: 3
---

Understanding the fundamental concepts behind the AegisLab ecosystem.

## Topics

{{< cards >}}
  {{< card link="architecture" title="Architecture" icon="document" >}}
  {{< card link="data-flow" title="Data Flow" icon="arrow-circle-right" >}}
  {{< card link="fault-injection-lifecycle" title="Fault Injection Lifecycle" icon="beaker" >}}
  {{< card link="observability-stack" title="Observability Stack" icon="document" >}}
{{< /cards >}}

## Overview

The AegisLab ecosystem is built on several key concepts that work together to enable fault injection and RCA evaluation:

### Fault Injection

The process of deliberately introducing failures into a system to study its behavior and resilience. AegisLab supports:

- **Chaos Engineering**: Systematic experimentation to build confidence in system resilience
- **Intelligent Scheduling**: Using genetic algorithms to evolve fault scenarios
- **Automated Collection**: Capturing observability data during fault injection

### Root Cause Analysis (RCA)

The process of identifying the underlying cause of system failures. AegisLab provides:

- **Standardized Interface**: Common API for RCA algorithms
- **Evaluation Metrics**: MRR, Avg@k, Top-k accuracy for comparing algorithms
- **Reproducible Datasets**: Consistent data format for fair comparison

### Observability

The ability to understand system internal state from external outputs. AegisLab collects:

- **Traces**: Distributed traces showing request flow through services
- **Metrics**: Time-series data on resource usage and performance
- **Logs**: Structured logs with context and correlation

### Microservices Architecture

Distributed systems composed of loosely coupled services. AegisLab targets:

- **Service Mesh**: Network of microservices with observability
- **Kubernetes**: Container orchestration platform
- **OpenTelemetry**: Unified observability framework

## Key Principles

### Reproducibility

All experiments are reproducible through:
- Standardized data formats (parquet files)
- Ground truth metadata (fault injection details)
- Consistent evaluation metrics
- Version-controlled configurations

### Scalability

The system scales through:
- Kubernetes-based deployment
- Asynchronous task processing
- Lazy data evaluation
- Distributed tracing

### Extensibility

The ecosystem is extensible via:
- Plugin-based algorithm registry
- Modular fault injection handlers
- Configurable benchmarks
- Open APIs and SDKs

### Automation

Automation is achieved through:
- REST APIs for programmatic access
- Background workers for async processing
- Genetic algorithms for intelligent scheduling
- Automated data collection and conversion

## Next Steps

- [Architecture](architecture): Detailed system architecture
- [Data Flow](data-flow): How data flows through the system
- [Fault Injection Lifecycle](fault-injection-lifecycle): End-to-end fault injection process
- [Observability Stack](observability-stack): Understanding the observability infrastructure
