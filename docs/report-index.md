# Report Index

> 更新时间：2026-04-19
> 目的：把当前可运行架构、保留文档、SDK/鉴权口径、调试方式和非阻塞尾项收口到一个总索引里。

## 1. 当前状态

- Fx + module + infra 主线已完成
- `producer / consumer / both` 三种模式已跑通
- 六服务入口已落地：`api-gateway / iam-service / resource-service / orchestrator-service / runtime-worker-service / system-service`
- 旧兼容层已退出主线运行态
- SDK 路由已统一收口到 `src/router/sdk.go`
- runtime 上传接口已并入 `src/router/sdk.go`，并只保留 `RequireServiceTokenAuth()`
- `portal / admin / sdk / runtime` 四类 audience 已完成一轮最终对齐补扫

## 2. 保留文档

- `docs/todo.md`
  - 最终执行状态、SDK 路由核对、验收命令、主线完成/非阻塞剩余项
- `docs/package-rename-todo.md`
  - `src/interface/grpc/*` 与包名统一记录
- `docs/api-key-auth-execution-todo.md`
  - API key / Key ID / Key Secret 执行记录
- `docs/python-runtime-wrapper-design.md`
  - Python runtime wrapper 职责边界设计
- `docs/python-runtime-wrapper-todo.md`
  - runtime wrapper 执行记录
- `docs/swagger-audience-unmarked-report.md`
  - Swagger audience 当前对齐状态、例外项与剩余 1 条空标记路由

## 3. 服务边界

### 3.1 六服务职责

- `api-gateway`
  - 对外唯一 HTTP/OpenAPI 入口；做 audience、鉴权、聚合与边缘协议适配
- `iam-service`
  - `auth / user / rbac / team / api key`
- `resource-service`
  - `project / label / container / dataset / evaluation / chaos-system`
- `orchestrator-service`
  - `execution / injection / task / trace / notification / group`
- `runtime-worker-service`
  - Redis 异步执行链、K8s / Helm / BuildKit / Chaos 运行态
- `system-service`
  - `config / audit / health / monitor / metrics`

### 3.2 基本约束

- 允许：`cmd -> app -> interface/module/infra/internalclient`
- 允许：`interface -> module/internalclient`
- 允许：`module -> infra/model/本模块 repository`
- 禁止：`gateway -> repository`
- 禁止：`interface -> repository` 直接拼业务
- 禁止：`module A -> module B repository`
- 禁止：非 owner 服务新增直接写库逻辑

## 4. SDK / 鉴权口径

### 4.1 SDK 路由口径

- 所有 `@x-api-type {"sdk":"true"}` 的运行态入口统一由 `src/router/sdk.go` 承接
- `runtime` 视为 SDK 路由中的一个专门子集，但鉴权语义单独保留
- runtime 上传接口：
  - `POST /api/v2/executions/{execution_id}/detector_results`
  - `POST /api/v2/executions/{execution_id}/granularity_results`
- 上述 runtime 路由当前仅要求：
  - `RequireServiceTokenAuth()`
- 不再叠加 `JWTAuth()`

### 4.2 API key

- 入口：`POST /api/v2/auth/api-key/token`
- 请求头：
  - `X-Key-Id`
  - `X-Timestamp`
  - `X-Nonce`
  - `X-Signature`
- canonical string：

```text
METHOD
PATH
TIMESTAMP
NONCE
SHA256(BODY)
```

- 业务 API 统一使用 `Authorization: Bearer <token>`
- `aegisctl` 与 Python SDK 已统一到这套签名换 token 流程

### 4.3 Python SDK / Runtime Client

- `RCABenchClient`
  - 公共/业务 API client；通过 `Key ID / Key Secret` 从环境变量换 token
- `RCABenchRuntimeClient`
  - runtime service-token-only client
  - 只保持 thin client，不承载 wrapper 调度语义

### 4.4 SDK generation / Apifox

- `portal`
  - 当前只生成 TypeScript SDK
- `admin`
  - 当前只生成 TypeScript SDK
- `sdk`
  - 当前只生成 Python SDK
- `runtime`
  - 当前作为 `sdk` 语义下的运行态子集保留在文档产物里，不单独作为 Apifox 上传目标
- `swagger init` 现在支持可选上传到 Apifox，但只支持三类目标：
  - `sdk`
  - `portal`
  - `admin`
- SDK 生成主入口改为显式 option 风格：
  - `sdk typescript --target portal|admin --env local|release --version <v>`
  - `sdk python --target sdk --env local|release --version <v>`
- TypeScript OpenAPI Generator 模板目录已压平为：
  - `.openapi-generator/typescript/config.json`
  - `.openapi-generator/typescript/templates/*`
- `scripts/command` 里的 Apifox / SDK / 测试安装 URL 统一改为优先从 `scripts/command/settings.toml` 读取
- `scripts/start.sh` 里的外部安装地址与测试代理也已改成顶部 env override 变量
- 不再需要单独的 `--upload-apifox`
- 只有显式传入 `--apifox-target ...` 时才会上传
- 若要一次上传全部，使用：
  - `--apifox-target all`

## 5. 启动与调试

### 5.1 什么时候用哪种模式

- `producer`
  - 调 HTTP、router、handler、Swagger、Portal/Admin API
- `consumer`
  - 调 worker / controller / receiver / runtime 执行链
- `both`
  - 调本地 submit -> queue -> worker -> query 闭环
- 六服务模式
  - 调 internal gRPC、owner 边界、remote-first 路径

说明：`both` 不是“同时启动六服务”，而是单体 HTTP + worker 组合模式。

### 5.2 六服务本地入口

| Service | Command | Default Port |
| --- | --- | --- |
| `api-gateway` | `go run ./src/cmd/api-gateway -conf ./src/config.dev.toml -port 8082` | `8082` |
| `iam-service` | `go run ./src/cmd/iam-service -conf ./src/config.dev.toml` | `9091` |
| `orchestrator-service` | `go run ./src/cmd/orchestrator-service -conf ./src/config.dev.toml` | `9092` |
| `resource-service` | `go run ./src/cmd/resource-service -conf ./src/config.dev.toml` | `9093` |
| `runtime-worker-service` | `go run ./src/cmd/runtime-worker-service -conf ./src/config.dev.toml` | `9094` |
| `system-service` | `go run ./src/cmd/system-service -conf ./src/config.dev.toml` | `9095` |

## 6. 最终结论

### 6.1 主线完成

- 单体 Fx 化完成
- 六服务边界完成
- 旧兼容层主线清理完成
- SDK / audience / API key 主线完成
- SDK 路由统一收口完成
- 基础验收与真实 K8s 集群入口完成

### 6.2 非阻塞剩余项

- 人工确认 Fx 日志输出是否要再裁剪
- 继续压 dedicated service 的少量 local fallback
- 继续深清跨 owner DB 直查
- 补发布参数、values、HPA、Ingress 等环境治理
- 补更多真实依赖集成回归
