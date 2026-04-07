---
title: Architecture
weight: 1
---

Detailed architecture of the AegisLab ecosystem and how components interact.

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    Kubernetes Cluster                            │
│                                                                  │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │                    AegisLab Core                            │ │
│  │                                                             │ │
│  │  ┌──────────────┐         ┌──────────────┐                │ │
│  │  │   Producer   │         │   Consumer   │                │ │
│  │  │  (API Server)│◀───────▶│  (Workers)   │                │ │
│  │  └──────┬───────┘         └──────┬───────┘                │ │
│  │         │                        │                         │ │
│  │         ▼                        ▼                         │ │
│  │  ┌──────────────┐         ┌──────────────┐                │ │
│  │  │    Redis     │         │    MySQL     │                │ │
│  │  │ (Task Queue) │         │  (Metadata)  │                │ │
│  │  └──────────────┘         └──────────────┘                │ │
│  └────────────────────────────────────────────────────────────┘ │
│                                                                  │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │                  Target System (TrainTicket)                │ │
│  │                                                             │ │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐  │ │
│  │  │ Gateway  │─▶│ Order    │─▶│ Payment  │─▶│ User     │  │ │
│  │  │ Service  │  │ Service  │  │ Service  │  │ Service  │  │ │
│  │  └──────────┘  └──────────┘  └──────────┘  └──────────┘  │ │
│  │       │              │              │              │       │ │
│  │       └──────────────┴──────────────┴──────────────┘       │ │
│  │                          │                                 │ │
│  │                          ▼                                 │ │
│  │                  ┌──────────────┐                          │ │
│  │                  │ OTEL Collector│                         │ │
│  │                  └──────────────┘                          │ │
│  └────────────────────────────────────────────────────────────┘ │
│                                                                  │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │                    Chaos Mesh                               │ │
│  │                                                             │ │
│  │  ┌──────────────┐         ┌──────────────┐                │ │
│  │  │  Controller  │         │  Dashboard   │                │ │
│  │  └──────────────┘         └──────────────┘                │ │
│  └────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
                              │
                              │ External Access
                              ▼
                    ┌──────────────────┐
                    │  LoadGenerator   │
                    │  Pandora         │
                    │  RCABench        │
                    └──────────────────┘
```

## Component Details

### AegisLab Producer

**Role**: REST API server for experiment management

**Responsibilities**:
- Accept fault injection requests via REST API
- Validate request parameters
- Create task records in MySQL
- Enqueue tasks to Redis
- Serve task status queries
- Manage dataset metadata

**Technology Stack**:
- Go 1.21+
- Gin web framework
- GORM for database access
- Redis client for task queue

**Key Endpoints**:
```
POST   /api/v1/fault-injection/submit
GET    /api/v1/tasks/:id
GET    /api/v1/tasks/:id/events
GET    /api/v1/datasets
GET    /api/v1/datasets/:id
```

### AegisLab Consumer

**Role**: Background workers for task execution

**Responsibilities**:
- Poll Redis task queue
- Execute fault injection tasks
- Generate Chaos Mesh CRDs
- Monitor chaos execution
- Collect observability data
- Convert data to parquet format
- Update task status

**Technology Stack**:
- Go 1.21+
- Kubernetes client-go
- Chaos Mesh SDK
- Polars (via CGO) for data processing

**Worker Types**:
- Fault injection workers
- Data collection workers
- Algorithm execution workers

### Redis Task Queue

**Role**: Asynchronous task queue with delayed execution

**Data Structures**:
```
# Pending tasks (sorted set by timestamp)
aegislab:tasks:pending

# Running tasks (set)
aegislab:tasks:running

# Task data (hash)
aegislab:task:{task_id}

# Task events (list)
aegislab:task:{task_id}:events
```

**Features**:
- Delayed task execution
- Priority queuing
- Retry with exponential backoff
- Task timeout handling

### MySQL Database

**Role**: Persistent metadata storage

**Schema**:
```sql
-- Tasks
CREATE TABLE tasks (
    id VARCHAR(36) PRIMARY KEY,
    status VARCHAR(20),
    benchmark VARCHAR(50),
    handler_nodes JSON,
    duration INT,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    dataset_id VARCHAR(36)
);

-- Datasets
CREATE TABLE datasets (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(100),
    benchmark VARCHAR(50),
    datapack_count INT,
    size_bytes BIGINT,
    created_at TIMESTAMP
);

-- Algorithms
CREATE TABLE algorithms (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(100),
    version VARCHAR(20),
    image VARCHAR(200),
    created_at TIMESTAMP
);
```

## Data Flow

### Fault Injection Flow

```
1. User submits injection request
   │
   ▼
2. Producer validates and creates task
   │
   ▼
3. Task enqueued to Redis
   │
   ▼
4. Consumer picks up task
   │
   ▼
5. Generate Chaos Mesh CRD
   │
   ▼
6. Apply chaos to Kubernetes
   │
   ▼
7. Monitor execution
   │
   ▼
8. Collect observability data
   │
   ▼
9. Convert to parquet format
   │
   ▼
10. Store dataset and update task status
```

### Algorithm Evaluation Flow

```
1. User submits algorithm execution
   │
   ▼
2. Producer creates evaluation task
   │
   ▼
3. Consumer creates Kubernetes Job
   │
   ▼
4. Job pulls algorithm container
   │
   ▼
5. Mount dataset from storage
   │
   ▼
6. Execute algorithm
   │
   ▼
7. Collect results
   │
   ▼
8. Calculate metrics (MRR, Avg@k)
   │
   ▼
9. Store results and update task
```

## Integration Points

### AegisLab ↔ Chaos Experiment

```go
// AegisLab imports chaos-experiment as Go module
import "github.com/LGU-SE-Internal/chaos-experiment/handler"

// Generate chaos using handler
chaos, err := handler.GenerateNetworkDelay(params)

// Apply to Kubernetes
err = k8sClient.Apply(chaos)
```

### AegisLab ↔ TrainTicket

```
TrainTicket Services (Kubernetes)
    │
    │ Service Discovery (DNS)
    │
    ▼
Chaos Mesh targets pods by labels
    │
    │ label: app=ts-order-service
    │
    ▼
Fault injected to matching pods
```

### AegisLab ↔ Observability Stack

```
TrainTicket Services
    │
    │ OpenTelemetry SDK
    │
    ▼
OTEL Collector
    │
    ├─▶ ClickHouse (traces)
    ├─▶ Prometheus (metrics)
    └─▶ Loki (logs)
         │
         ▼
    AegisLab Consumer
         │
         ▼
    Parquet Files
```

## Scalability

### Horizontal Scaling

**Producer**:
- Multiple replicas behind load balancer
- Stateless design
- Shared MySQL and Redis

**Consumer**:
- Multiple worker replicas
- Task distribution via Redis queue
- Independent processing

**TrainTicket**:
- Each service independently scalable
- Kubernetes HPA for auto-scaling

### Vertical Scaling

**Resource Limits**:
```yaml
# Producer
resources:
  requests:
    cpu: 500m
    memory: 512Mi
  limits:
    cpu: 2000m
    memory: 2Gi

# Consumer
resources:
  requests:
    cpu: 1000m
    memory: 1Gi
  limits:
    cpu: 4000m
    memory: 4Gi
```

## High Availability

### Database Replication

```
MySQL Primary
    │
    ├─▶ Replica 1 (read)
    └─▶ Replica 2 (read)
```

### Redis Sentinel

```
Redis Master
    │
    ├─▶ Replica 1
    └─▶ Replica 2
         │
         ▼
    Sentinel (failover)
```

### Service Redundancy

- Multiple producer replicas
- Multiple consumer workers
- Kubernetes pod anti-affinity

## Security

### Authentication

- API token-based authentication
- JWT tokens with expiration
- Role-based access control (RBAC)

### Network Security

```
┌─────────────────────────────────────┐
│         Kubernetes Network          │
│                                     │
│  ┌─────────────────────────────┐   │
│  │  AegisLab Namespace         │   │
│  │  (NetworkPolicy: ingress)   │   │
│  └─────────────────────────────┘   │
│                                     │
│  ┌─────────────────────────────┐   │
│  │  TrainTicket Namespace      │   │
│  │  (NetworkPolicy: isolated)  │   │
│  └─────────────────────────────┘   │
└─────────────────────────────────────┘
```

### Data Security

- Secrets stored in Kubernetes Secrets
- Database credentials encrypted
- TLS for external communication

## Monitoring

### Metrics

```
# AegisLab metrics
aegislab_tasks_total
aegislab_tasks_duration_seconds
aegislab_queue_size
aegislab_worker_busy

# TrainTicket metrics
http_request_duration_seconds
http_requests_total
system_cpu_utilization
```

### Logging

```
# Structured logging
{
  "timestamp": "2026-01-18T10:00:00Z",
  "level": "INFO",
  "component": "consumer",
  "task_id": "task-123",
  "message": "Fault injection completed"
}
```

### Tracing

- OpenTelemetry for distributed tracing
- Trace context propagation
- Jaeger for visualization

## Performance Considerations

### Task Queue Optimization

- Batch task polling
- Connection pooling
- Delayed queue for scheduling

### Data Processing

- Lazy evaluation with Polars
- Streaming data conversion
- Parallel processing

### Storage

- JuiceFS for distributed storage
- Parquet for columnar storage
- Compression for space efficiency

## Next Steps

- [Data Flow](../data-flow): Detailed data flow diagrams
- [Fault Injection Lifecycle](../fault-injection-lifecycle): Complete lifecycle
- [Observability Stack](../observability-stack): Monitoring infrastructure
