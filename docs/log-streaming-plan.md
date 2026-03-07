# 实时 K8s Job 日志流架构方案

> 创建日期：2026-02-17  
> 状态：Draft  
> 范围：仅 K8s Job 日志（后端自身日志后续迭代）

## TL;DR

实现 K8s Job 日志的**生产级**实时流式传输。利用已部署的 Alloy DaemonSet 采集 Job 日志，新增 OTLP HTTP 输出到后端；后端实现 OTLP HTTP 日志接收器，提取 `task_id` 后通过 Redis Pub/Sub 分发到 WebSocket 端点；WebSocket 端点先查询 Loki 获取历史日志，再切换到实时推送。按 `task_id` 维度查询。

## 架构总览

```
  K8s Job Pods (stdout/stderr)
        │
        ▼
  ┌──────────────────────────────┐
  │     Alloy DaemonSet          │
  │  /var/log/pods/*.log         │
  │  (已部署, 按 rcabench labels │
  │   过滤 Job pods)             │
  └─────────┬────────────────────┘
            │ loki.process "pipeline"
            │ forward_to (dual-write)
            │
  ┌─────────┴───────────────────────────┐
  ▼                                     ▼
┌──────────┐                   ┌─────────────────────┐
│   Loki   │                   │ 后端 OTLP HTTP      │
│  :3100   │                   │ Receiver :4319      │
│ (持久化) │                   │ /v1/logs            │
└────┬─────┘                   └──────────┬──────────┘
     │                                    │
     │ (历史查询)                          │ 解析 OTLP LogRecord
     │                                    │ 提取 task_id
     │                                    ▼
     │                         ┌─────────────────────┐
     │                         │ Redis Pub/Sub       │
     │                         │ channel: joblogs:{task_id}
     │                         └──────────┬──────────┘
     │                                    │
     ▼                                    ▼
┌────────────────────────────────────────────────────┐
│            WebSocket Handler                       │
│  GET /api/v2/tasks/{task_id}/logs/ws               │
│                                                    │
│  连接流程:                                          │
│  1. JWT 认证 (query param: ?token=xxx)             │
│  2. HTTP → WebSocket 升级                          │
│  3. 查 Loki 历史日志 → 发送 type:"history"          │
│  4. 订阅 Redis Pub/Sub → 转发 type:"realtime"      │
│  5. 监听 task 完成 → 发送 type:"end" → 关闭        │
└────────────────────┬───────────────────────────────┘
                     │ ws://
                     ▼
               ┌──────────┐
               │   前端    │
               │ LogsTab  │
               └──────────┘
```

## 为什么选择 OTLP 而非 client-go

| 对比维度   | OTLP (Alloy → 后端)                     | client-go Pod Log Stream         |
| ---------- | --------------------------------------- | -------------------------------- |
| **解耦**   | 后端不直接连 K8s API，Alloy 负责采集    | 后端直接维护 Pod log Follow 连接 |
| **可靠性** | Alloy 有重试/缓冲机制，后端重启不丢日志 | 后端重启 = 日志流中断            |
| **扩展性** | 新增日志源只需改 Alloy 配置             | 每种日志源需要新 goroutine       |
| **标准化** | OTLP 是 OpenTelemetry 标准协议          | K8s 专有 API                     |
| **运维**   | 与现有 Alloy→Loki 管道一致              | 额外的连接管理和资源清理         |
| **生产级** | 工业标准，可对接任何 OTLP 兼容后端      | 仅适合小规模/开发环境            |

**结论**：生产环境应使用 OTLP。Alloy 已在采集 Job 日志，只需加一路 OTLP 输出，后端作为标准 OTLP 接收器处理实时分发。

## 现有基础设施

### 已有

- **Alloy DaemonSet**：已部署在 `exp` namespace，通过 `rcabench_app_id` + `job_name` label 过滤 Job pods
- **Loki**：`http://10.10.10.161:3100`，已接收 Alloy 推送的日志，支持 LogQL 查询
- **Redis**：已有 `client.RedisPublish()` / `client.GetRedisClient().Subscribe()` 方法
- **Redis Stream**：已用于 SSE 事件推送（`StreamLogKey = "trace:%s:log"`）
- **gorilla/websocket**：`v1.5.4` 已在 go.mod（间接依赖）
- **OTel proto**：`go.opentelemetry.io/proto/otlp v1.5.0` 已在 go.mod（间接依赖）
- **前端 LogsTab**：已有基础 UI 组件

### 需要新增

- 后端 OTLP HTTP 日志接收器（`/v1/logs`，端口 4319）
- Alloy 配置增加 OTLP 输出（dual-write）
- Loki 查询客户端
- WebSocket handler + 路由
- 日志相关 DTO

## 实现步骤

### Phase 1：后端 OTLP HTTP 日志接收器

**新建 `src/service/logreceiver/receiver.go`**

实现标准 OTLP HTTP 日志接收端点，接收 Alloy 推送的 Job 日志。

```go
// 核心结构
type OTLPLogReceiver struct {
    server     *http.Server
    redisClient *redis.Client
    port       int
    shutdownCh chan struct{}
}

// 接收端点: POST /v1/logs
// 请求体: protobuf (application/x-protobuf) 或 JSON (application/json)
// 响应: 200 OK / 400 Bad Request / 500 Internal Server Error
```

**关键实现细节**：

1. **OTLP 解析** — 使用 `go.opentelemetry.io/proto/otlp/logs/v1` 解析三层结构：

   ```
   ExportLogsServiceRequest
   └── ResourceLogs[]
       ├── Resource.Attributes (rcabench_app_id, namespace)
       └── ScopeLogs[]
           └── LogRecords[]
               ├── TimeUnixNano
               ├── Body.StringValue (日志行)
               └── Attributes (task_id, trace_id, job_id)
   ```

2. **元数据提取** — 从 Resource Attributes 和 Log Attributes 提取：
   - `task_id`（必须，用于路由到正确的 Redis Pub/Sub channel）
   - `trace_id`（可选，用于关联追踪）
   - `job_id`（可选，job 名称）
   - `rcabench_app_id`（已在 Alloy relabel 中设置）

3. **Redis Pub/Sub 发布** — 按 `task_id` 发布到 channel `joblogs:{task_id}`：

   ```go
   client.RedisPublish(ctx, fmt.Sprintf("joblogs:%s", taskID), logEntry)
   ```

4. **生产级要求**：
   - 请求体大小限制（默认 5MB）
   - Content-Type 校验（支持 protobuf 和 JSON 两种格式）
   - 请求超时控制
   - Prometheus metrics（接收速率、错误率、延迟）
   - 优雅关闭（`Shutdown(ctx)`）
   - 健康检查端点（`GET /health`）

**依赖提升**（go.mod indirect → direct）：

- `go.opentelemetry.io/proto/otlp v1.5.0`
- `github.com/gorilla/websocket v1.5.4`
- `google.golang.org/protobuf`（已有）

### Phase 2：修改 Alloy 配置，新增 OTLP 输出

**修改 `manifests/dev/exp-dev-setup.yaml`**

在现有 pipeline 中增加 OTLP dual-write：

```river
// ============ 新增: OTLP 日志输出到后端 ============

// 桥接: Loki 格式 → OpenTelemetry 格式
otelcol.receiver.loki "backend" {
  output {
    logs = [otelcol.exporter.otlphttp.backend.input]
  }
}

// OTLP HTTP 导出到后端接收器
otelcol.exporter.otlphttp "backend" {
  client {
    endpoint = "http://rcabench-service.exp.svc.cluster.local:4319"
    // 本地开发时用: endpoint = "http://host.k3d.internal:4319"

    // 生产级配置
    retry_on_failure {
      enabled         = true
      initial_interval = "1s"
      max_interval     = "30s"
      max_elapsed_time = "5m"
    }

    // 发送队列（缓冲 + 批量）
    sending_queue {
      enabled    = true
      num_consumers = 4
      queue_size = 1000
    }
  }
}
```

**修改 pipeline forward_to**：

```river
// 现有:
forward_to = [loki.write.default.receiver]

// 改为 dual-write:
forward_to = [loki.write.default.receiver, otelcol.receiver.loki.backend.receiver]
```

**修改 DaemonSet args**：

```yaml
# 现有:
args:
  - --stability.level=generally-available

# 改为 (otelcol.* 组件需要 public-preview):
args:
  - --stability.level=public-preview
```

**注意事项**：

- `otelcol.receiver.loki` 将 Loki labels 自动映射为 OTLP Resource Attributes
- `task_id`、`trace_id`、`job_id` 在 Alloy relabel 阶段已设置为 Structured Metadata，会作为 OTLP Attributes 传递
- 本地开发环境后端不在 K8s 内，需要用 `host.k3d.internal` 或实际 IP

### Phase 3：后端新增 Loki 查询客户端

**新建 `src/client/loki.go`**

封装 Loki HTTP API，用于 WebSocket 连接时获取历史日志。

```go
type LokiClient struct {
    baseURL    string
    httpClient *http.Client
}

// QueryJobLogs 查询指定 task_id 的 Job 历史日志
// LogQL: {app="rcabench"} | task_id=`{taskID}`
func (c *LokiClient) QueryJobLogs(ctx context.Context, taskID string, opts QueryOpts) ([]LogEntry, error)

// QueryOpts 查询参数
type QueryOpts struct {
    Start     time.Time  // 默认: task 创建时间
    End       time.Time  // 默认: now
    Limit     int        // 默认: 5000
    Direction string     // "forward" (时间正序)
}
```

**Loki API 调用**：

- `GET /loki/api/v1/query_range`
- LogQL: `{app="rcabench"} | task_id="{task_id}"`（Structured Metadata 过滤）
- 分页: `limit` + `start`/`end` 时间范围

**配置** (`config.dev.toml` 新增)：

```toml
[loki]
url = "http://10.10.10.161:3100"
timeout = "10s"
max_entries = 5000
```

### Phase 4：定义日志 DTO 和 WebSocket 消息格式

**新建 `src/dto/log.go`**

```go
// LogEntry 统一日志条目（OTLP 接收和 Loki 查询共用）
type LogEntry struct {
    Timestamp time.Time `json:"timestamp"`           // 日志时间戳
    Line      string    `json:"line"`                // 日志内容
    TaskID    string    `json:"task_id"`             // 关联的 task ID
    JobID     string    `json:"job_id,omitempty"`    // K8s Job 名称
    TraceID   string    `json:"trace_id,omitempty"`  // 追踪 ID
    Level     string    `json:"level,omitempty"`     // 日志级别 (info/warn/error)
}

// WSLogMessage WebSocket 推送的消息格式
type WSLogMessage struct {
    Type    string     `json:"type"`              // "history" | "realtime" | "end" | "error"
    Logs    []LogEntry `json:"logs,omitempty"`     // 日志条目
    Message string     `json:"message,omitempty"`  // 错误信息或结束原因
    Total   int        `json:"total,omitempty"`    // 历史日志总条数
}
```

### Phase 5：实现 WebSocket Handler

**新建 `src/handlers/v2/task_logs.go`**

```go
// GetTaskLogsWS WebSocket 端点 - 实时 Job 日志流
// @Router /api/v2/tasks/{task_id}/logs/ws [get]
//
// 连接流程:
// 1. JWT 认证 (从 query param ?token=xxx 获取)
// 2. HTTP → WebSocket 升级
// 3. 查询 Loki 历史日志 → type:"history"
// 4. Redis Pub/Sub 订阅 → type:"realtime"
// 5. 监听 task 完成 → type:"end" → 关闭
func GetTaskLogsWS(c *gin.Context)
```

**生产级要求**：

1. **认证**：
   - WebSocket 不支持自定义 HTTP header
   - 从 URL query 参数 `?token=xxx` 获取 JWT
   - 验证 token 有效性后再升级连接

2. **连接管理**：
   - 设置读写超时（WriteWait: 10s, PongWait: 60s, PingPeriod: 54s）
   - Ping/Pong 心跳保活
   - 最大消息大小限制
   - 客户端断连时清理 Redis 订阅

3. **历史 + 实时日志无缝衔接**：
   - 先订阅 Redis Pub/Sub（确保不丢失订阅期间的日志）
   - 再查询 Loki 历史日志，发送给客户端
   - 然后开始转发 Redis 实时日志
   - 用时间戳去重（Loki 和 Redis 可能有短暂重叠）

4. **优雅终止**：
   - 监听 task 状态变化（轮询 DB 或订阅 Redis Stream 的 task 完成事件）
   - Task 完成后等待 5s（flush 最后的日志）再发送 `type:"end"`
   - 支持客户端主动关闭

5. **并发安全**：
   - WebSocket 写操作需要互斥锁（`sync.Mutex`）
   - Redis 订阅和 Loki 查询在独立 goroutine 中

### Phase 6：路由注册

**修改 `src/router/v2.go`**

```go
// 在 tasks 路由组中添加:
tasks.GET("/:task_id/logs/ws", v2.GetTaskLogsWS)
```

- WebSocket 端点**不使用**标准 JWT middleware（因为 token 在 query param）
- 在 handler 内部手动验证 token

### Phase 7：应用启动集成

**修改 `src/main.go`**

在 `consumer` 和 `both` 模式中启动 OTLP 接收器：

```go
// consumer/both 模式
go logreceiver.Start(ctx, config.GetInt("otlp_receiver.port"))
```

### Phase 8：配置更新

**修改 `src/config.dev.toml`**

```toml
[loki]
url = "http://10.10.10.161:3100"
timeout = "10s"
max_entries = 5000

[otlp_receiver]
port = 4319
max_request_size = "5MB"

[logging.job]
dir = "jobs"
log_retention_days = 30
pubsub_channel_prefix = "joblogs"
```

## Redis Channel 设计

```
joblogs:{task_id}    # Pub/Sub channel，实时 Job 日志
                     # 每条消息: JSON 序列化的 LogEntry
                     # 生命周期: task 运行期间活跃

trace:{trace_id}:log # Redis Stream（已有），task 状态事件
                     # 用于监听 task 完成信号
```

与现有 `StreamLogKey = "trace:%s:log"` Redis Stream 独立，不影响现有 SSE 事件推送。

## 验证清单

### 开发环境验证

1. **OTLP 接收器启动**：

   ```bash
   # 构建并启动
   cd src && go build -o /tmp/rcabench ./main.go
   ENV_MODE=dev /tmp/rcabench both --port 8082
   # 检查日志: "OTLP log receiver started on :4319"
   ```

2. **OTLP 接收器功能测试**：

   ```bash
   # 发送测试 OTLP 日志（JSON 格式）
   curl -X POST http://localhost:4319/v1/logs \
     -H "Content-Type: application/json" \
     -d '{"resourceLogs":[{"resource":{"attributes":[{"key":"task_id","value":{"stringValue":"test-123"}}]},"scopeLogs":[{"logRecords":[{"timeUnixNano":"1708100000000000000","body":{"stringValue":"test log line"}}]}]}]}'
   # 响应: 200 OK
   ```

3. **Redis Pub/Sub 验证**：

   ```bash
   # 终端 A: 订阅
   redis-cli subscribe joblogs:test-123
   # 终端 B: 发送上面的 OTLP 测试请求
   # 终端 A 应收到 LogEntry JSON
   ```

4. **WebSocket 端到端测试**：

   ```bash
   # 获取 JWT token
   TOKEN=$(curl -s -X POST http://localhost:8082/api/v2/auth/login \
     -H "Content-Type: application/json" \
     -d '{"username":"admin","password":"admin"}' | jq -r '.data.access_token')

   # 连接 WebSocket
   websocat "ws://localhost:8082/api/v2/tasks/{task_id}/logs/ws?token=$TOKEN"
   ```

### K8s 集群验证

5. **Alloy 配置更新**：

   ```bash
   kubectl apply -f manifests/dev/exp-dev-setup.yaml
   # 验证 Alloy pod 重启成功
   kubectl get pods -n exp -l app=alloy
   ```

6. **端到端 Job 日志流**：

   ```bash
   # 创建一个测试 fault injection → 触发 Job
   # WebSocket 客户端应实时收到 Job 日志
   ```

7. **Loki 历史查询验证**：
   ```bash
   curl "http://10.10.10.161:3100/loki/api/v1/query_range" \
     --data-urlencode 'query={app="rcabench"} | task_id="xxx"' \
     --data-urlencode 'start=2026-02-17T00:00:00Z' \
     --data-urlencode 'end=2026-02-17T23:59:59Z' \
     --data-urlencode 'limit=100'
   ```

### 构建和测试

8. **Go 构建**：`cd src && go build -o /tmp/rcabench ./main.go`
9. **单元测试**：`cd src && go test ./utils/... -v`
10. **OTLP 接收器单元测试**：`cd src && go test ./service/logreceiver/... -v`

## 设计决策

| 决策项    | 选择                  | 原因                                                            |
| --------- | --------------------- | --------------------------------------------------------------- |
| 日志采集  | Alloy OTLP dual-write | 已有 Alloy pipeline，标准化 OTLP 协议，生产级可靠性             |
| 传输协议  | WebSocket             | 双向通信，未来可扩展暂停/过滤控制                               |
| 实时中转  | Redis Pub/Sub         | 多客户端广播，轻量级，与现有 Redis 基础设施复用                 |
| 历史日志  | Loki 查询             | 已有完整 Alloy → Loki 管道，LogQL 支持 Structured Metadata 过滤 |
| 日志维度  | 按 task_id            | 匹配前端 LogsTab 在任务详情页的展示场景                         |
| OTLP 格式 | HTTP (非 gRPC)        | 更简单调试，curl 测试友好，防火墙友好                           |
| 日志范围  | 仅 K8s Job            | 后端自身日志后续迭代添加                                        |

## 风险和缓解

| 风险                           | 影响                                         | 缓解措施                                                                         |
| ------------------------------ | -------------------------------------------- | -------------------------------------------------------------------------------- |
| Alloy `--stability.level` 升级 | `public-preview` 组件可能有 breaking changes | 锁定 Alloy 镜像版本 (v1.13.1)，升级前测试                                        |
| Redis Pub/Sub 无持久化         | 连接前的实时日志丢失                         | 先订阅再查 Loki 历史，时间戳去重覆盖间隙 (Loki ~1-5s 延迟)                       |
| OTLP 接收器宕机                | 实时日志丢失                                 | Alloy `retry_on_failure` 重试 + `sending_queue` 缓冲；历史日志仍走 Loki 不受影响 |
| WebSocket 连接泄漏             | 资源耗尽                                     | Ping/Pong 心跳 + 读写超时 + task 完成自动关闭                                    |
| Loki 查询慢                    | WebSocket 连接等待时间长                     | 分页查询 + 超时控制 + 先发送部分历史再分批补充                                   |
| OTLP protobuf 解析复杂         | 开发周期长                                   | 同时支持 JSON 格式，优先用 JSON 开发调试                                         |

## 实现优先级

```
Phase 1 (P0): OTLP 接收器 + JSON 格式支持          → 可独立验证
Phase 2 (P0): Alloy 配置 dual-write                → 打通采集链路
Phase 3 (P1): Loki 查询客户端                       → 历史日志
Phase 4 (P1): DTO + WebSocket Handler              → 前端可用
Phase 5 (P1): 路由 + 启动集成 + 配置                → 完整功能
Phase 6 (P2): protobuf 格式支持                     → 性能优化
Phase 7 (P2): Prometheus metrics + 监控仪表盘        → 可观测性
Phase 8 (P3): 后端自身日志采集                       → 扩展范围
```

## 文件清单（预期产出）

```
src/service/logreceiver/
├── receiver.go          # OTLP HTTP 接收器（核心）
├── parser.go            # OTLP LogRecord 解析 + 元数据提取
├── receiver_test.go     # 单元测试
└── metrics.go           # Prometheus 指标

src/client/
└── loki.go              # Loki HTTP 查询客户端

src/dto/
└── log.go               # LogEntry, WSLogMessage DTO

src/handlers/v2/
└── task_logs.go         # WebSocket handler

src/router/v2.go         # 路由注册（修改）
src/main.go              # 启动集成（修改）
src/config.dev.toml      # 配置新增（修改）

manifests/dev/
└── exp-dev-setup.yaml   # Alloy 配置 dual-write（修改）
```
