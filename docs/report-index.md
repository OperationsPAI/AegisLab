# Report Index

> 更新时间：2026-04-18
> 目的：把本轮后端主线重构、微服务收尾、治理约定、运行口径、SDK/鉴权要点收口到少量总贴文档里。

## 1. 最终结论

- Fx + module + infra 主线已完成，`producer / consumer / both` 与六服务入口都已跑通。
- 微服务主线可视为完成，当前已形成 `api-gateway / iam-service / resource-service / orchestrator-service / runtime-worker-service / system-service` 六个明确边界。
- 旧运行态兼容面已经完成仓库级清扫：`service/producer`、`handlers/system`、`database.DB`、`GetGateway()`、`redisinfra.GetGateway()` 这批模式已退出主线生产代码。
- 当前仅剩 1 个未勾项：确认 Fx 启动日志是否可接受；这是人工验收，不阻塞代码主线收口。

## 2. 保留文档

- `docs/todo.md`
  - 主 TODO 与最终验收清单，仍作为执行源文档。
- `docs/report-index.md`
  - 当前总索引与汇总版说明。
- `docs/frontend-redesign.md`
  - 前端重设计文档，属于独立主题，未纳入本次后端文档合并。
- `docs/frontend-ui-guidelines.md`
  - 前端 UI 规范，属于独立主题，未纳入本次后端文档合并。

## 3. 服务边界与 ownership 总结

### 3.1 六服务职责

- `api-gateway`
  - 对外唯一 HTTP/OpenAPI 入口。
  - 负责 audience、鉴权、参数校验、统一错误壳、聚合响应。
  - 不直接查 DB，不直接做 K8s / Helm / BuildKit 业务判断。
- `iam-service`
  - 承接 `auth / user / rbac / team / access key`。
  - 负责 `AK/SK -> token`、token verify、permission check。
- `resource-service`
  - 承接 `project / label / container / dataset / evaluation` 元数据与查询视图。
- `orchestrator-service`
  - 承接 `execution / injection submit`、`task / trace / retry / dead-letter / cancel` 控制面。
- `runtime-worker-service`
  - 承接 Redis 异步消费、K8s/BuildKit/Helm/Chaos 执行态、limiter、namespace lock、runtime monitor。
  - 异步执行链继续保留 Redis，不改为同步执行 RPC。
- `system-service`
  - 承接 `config / audit / health / monitor / metrics` 运维控制面。

### 3.2 owner 约束

- `iam-service`
  - `users`、`roles`、`permissions`、`resources`、`teams`、`access_keys` 及其授权关系。
- `resource-service`
  - `projects`、`labels`、`containers`、`datasets`、`evaluations` 与资源元数据关系。
- `orchestrator-service`
  - `tasks`、`traces`、`fault_injections`、`executions`、重试/死信/工作流控制面。
- `system-service`
  - `dynamic_configs`、`config_histories`、`audit_logs`、`system_metrics` 等运维数据。
- `runtime-worker-service`
  - Redis runtime state、K8s/build/helm 执行态，不新增跨 owner MySQL 写入。

### 3.3 依赖规则

- 允许：`cmd -> app -> interface/module/infra/internalclient`
- 允许：`interface -> module/internalclient`
- 允许：`module -> infra/model/本模块 repository`
- 禁止：`gateway -> repository`
- 禁止：`interface -> repository` 直接拼业务
- 禁止：`module A -> module B repository`
- 禁止：非 owner 服务新增直接写库逻辑

## 4. 本地运行与发布口径

### 4.1 六服务本地入口

| Service | Command | Default Port |
| --- | --- | --- |
| `api-gateway` | `go run ./src/cmd/api-gateway -conf ./src/config.dev.toml -port 8082` | `8082` |
| `iam-service` | `go run ./src/cmd/iam-service -conf ./src/config.dev.toml` | `9091` |
| `orchestrator-service` | `go run ./src/cmd/orchestrator-service -conf ./src/config.dev.toml` | `9092` |
| `resource-service` | `go run ./src/cmd/resource-service -conf ./src/config.dev.toml` | `9093` |
| `runtime-worker-service` | `go run ./src/cmd/runtime-worker-service -conf ./src/config.dev.toml` | `9094` |
| `system-service` | `go run ./src/cmd/system-service -conf ./src/config.dev.toml` | `9095` |

### 4.2 本地依赖与启动顺序

- 先起基础依赖：
  - `docker compose up -d redis mysql etcd jaeger buildkitd loki prometheus grafana`
- 如需本地全量六服务：
  - `docker compose -f docker-compose.yaml -f docker-compose.microservices.yaml up --build`
- 手动顺序建议：
  - `iam-service`
  - `orchestrator-service`
  - `resource-service`
  - `runtime-worker-service`
  - `system-service`
  - `api-gateway`

### 4.3 发布骨架

- 本地 compose 骨架：`docker-compose.microservices.yaml`
- Kubernetes skeleton：`manifests/microservices/aegislab-microservices.yaml`
- Helm 发布主口径：`helm/templates/{configmap,service,deployment}.yaml`
- 当前 Helm 已按六服务拓扑渲染通过：`helm template aegislab ./helm`

## 5. Health / Readiness 约定

- `api-gateway`
  - 协议：HTTP
  - 探针：`GET /system/health`
  - 默认端口：`8082`
- 内部 gRPC 服务
  - 服务：`iam-service`、`orchestrator-service`、`resource-service`、`runtime-worker-service`、`system-service`
  - 协议：gRPC health checking protocol
  - 默认端口：`9091` ~ `9095`
- 启动前 target 校验
  - `api-gateway` 校验 `clients.iam.target`、`clients.orchestrator.target`、`clients.resource.target`、`clients.system.target`
  - `runtime-worker-service` 校验 `clients.orchestrator.target`
  - `resource-service` 校验 `clients.orchestrator.target`
  - `system-service` 校验 `clients.runtime.target`
- 口径
  - 缺 target 时直接启动失败，不把配置缺失留给 readiness 长期兜底

## 6. 治理约定

### 6.1 Request ID

- 外部 HTTP：
  - 优先读取 `X-Request-Id`
  - 缺失时由 gateway 生成并回写响应头
- 内部 gRPC：
  - 统一 metadata key：`x-request-id`
- 当前已落地：
  - `src/router/router.go` 挂 request-id middleware
  - `src/internalclient/*` 统一透传
  - `src/interface/grpc*` 统一提取/补齐并回写 header

### 6.2 错误码

- HTTP：
  - `401`、`403`、`400`、`404`、`409`、`500`
- gRPC：
  - `Unauthenticated`
  - `PermissionDenied`
  - `InvalidArgument`
  - `NotFound`
  - `AlreadyExists`
  - 其他统一 `Internal`

### 6.3 观测与配置

- 基础标签：
  - `service.name`
  - `service.role`
  - `request.id`
  - `user.id`
  - `project.id`
  - `trace.id`
  - `task.id`
  - `group.id`
- 内部 client target 主键：
  - `clients.iam.target`
  - `clients.resource.target`
  - `clients.orchestrator.target`
  - `clients.runtime.target`
  - `clients.system.target`
- 服务监听主键：
  - `iam.grpc.addr`
  - `resource.grpc.addr`
  - `orchestrator.grpc.addr`
  - `runtime_worker.grpc.addr`
  - `system.grpc.addr`

## 7. SDK / 鉴权 / Swagger 总结

### 7.1 Swagger audience 现状

- 扫描总操作数：`173`
- 已标记操作：`100`
- 空 `@x-api-type {}`：`73`
- 缺失 `@x-api-type`：`0`
- 已标 audience 统计：
  - `sdk=5`
  - `portal=43`
  - `admin=58`

### 7.2 AK/SK -> token 规范

- 交换接口：
  - `POST /api/v2/auth/access-key/token`
- 只有这个接口直接使用 `secret_key`
- 业务 API 继续统一使用：
  - `Authorization: Bearer <token>`

必需请求头：

- `X-Access-Key`
- `X-Timestamp`
- `X-Nonce`
- `X-Signature`

canonical string：

```text
METHOD
PATH
ACCESS_KEY
TIMESTAMP
NONCE
```

签名算法：

```text
signature = hex(hmac_sha256(secret_key, canonical_string))
```

服务端规则：

- 时间窗：`+- 5 minutes`
- nonce 单次使用
- disabled / deleted / expired access key 不能换 token
- replay 防护依赖 Redis nonce reservation

### 7.3 `aegisctl` 鉴权约定

- `aegisctl auth login --access-key ... --secret-key ...`
  - 走 AK/SK 签名换 token
- `aegisctl auth inspect`
  - 查看本地 auth context
- `aegisctl auth sign-debug`
  - 输出 canonical string、签名头、curl 示例
- `aegisctl auth sign-debug --execute`
  - 直接发起换 token 请求
- `aegisctl auth sign-debug --execute --save-context`
  - 成功后把 token 落当前 context

## 8. Model / DTO / repository 收口总结

- `src/database` 已整体迁到 `src/model`
- DB 初始化、迁移、生命周期已转到 `src/infra/db`
- 模块专用 API 契约已大量下沉到 `src/module/*/api_types.go`
- 全局 `src/dto/*` 已压缩为极薄共享层，只保留分页/搜索/统一响应/少量跨模块运行时载荷
- 已删除一批空心化旧仓储文件，模块专用 DB 访问回收到各自 `src/module/*/repository.go`
- 当前 `src/repository/*` 只保留仍有跨模块边界价值的共享查询能力

## 9. 当前验收状态

- 默认回归：
  - `cd src && go test ./...`
- Producer Fx 图校验与 HTTP 主路径：
  - `cd src && go test ./app -run 'TestProducerOptionsValidate|TestProducerOptionsStartStopSmoke|TestProducerOptionsHTTPIntegrationSmoke'`
- Consumer / Both 生命周期冒烟：
  - `cd src && go test ./app -run 'TestConsumerOptions|TestBothOptions'`
- 路由 / 文档主路径：
  - `cd src && go test ./router ./docs ./interface/http`
- 真实 K8s 集群验收：
  - `cd src && RUN_K8S_INTEGRATION=1 go test ./infra/k8s -run TestK8sGatewayJobLifecycleIntegration`

## 10. 当前唯一未完成人工项

- `docs/todo.md`
  - `确认 Fx 生成的启动日志是否可接受`
  - 性质：人工验收
  - 状态：非阻塞
  - 不影响当前“主线完成”判断

## 11. 仓库级补扫结果

### 11.1 已确认清空的旧兼容面

对 `src` 生产代码补扫后，以下模式当前为 `0` 命中：

```text
service/producer      = 0
handlers/system       = 0
database.DB           = 0
GetGateway(           = 0
redisinfra.GetGateway = 0
```

说明：

- 旧 `service/producer` 兼容层已退出运行态代码
- 旧 `handlers/system` 包级入口已退出运行态代码
- 全局 `database.DB` 已不再残留在主线包中
- 旧 infra 全局 gateway fallback 已不再残留在主线包中

### 11.2 `context.Background()` 补扫

在 `src/app src/interface src/module src/service src/router src/middleware` 范围补扫后：

- 生产代码未发现新的 `context.Background()` 残留
- 当前命中均来自测试文件

## 12. 微服务主线完成面

### 12.1 服务入口

当前统一二进制已支持：

- `producer`
- `consumer`
- `both`
- `api-gateway`
- `iam-service`
- `orchestrator-service`
- `resource-service`
- `runtime-worker-service`
- `system-service`

### 12.2 internal client 边界

当前已落地：

- gateway -> IAM
- gateway -> Resource
- gateway -> Orchestrator
- gateway -> System
- system -> Runtime
- resource/evaluation -> Orchestrator
- runtime -> Orchestrator

关键目录：

- `src/internalclient/iamclient`
- `src/internalclient/resourceclient`
- `src/internalclient/orchestratorclient`
- `src/internalclient/systemclient`
- `src/internalclient/runtimeclient`

### 12.3 dedicated service 收口状态

- dedicated `api-gateway` 已不再静默回退本地 owner service
- `resource-service` 已通过 remote query/source 收口 project statistics 与 evaluation 查询
- `system-service` 已通过 runtime RPC 收口 namespace locks / queued tasks
- `runtime-worker-service` 已通过 remote owner option 收口 orchestrator owner 操作
- `api-gateway` 的 team / auth / user / rbac / label / chaos-system / task / trace / group / notification 主路径已收口到内部服务边界

## 13. 当前仍剩余，但不阻塞主线

- 兼容入口 `producer / consumer / both` 内部仍可继续压缩本地 owner 组合
- 少量跨服务 DB 直查/直写仍可继续按 owner 深清
- 发布层后续仍可继续细化环境参数、镜像策略、HPA、Ingress 与 values 编排

## 14. 建议下一阶段顺序

1. 继续清兼容入口里的本地 owner 组合面
2. 继续清跨服务 DB 直查/直写
3. 做版本级环境参数与发布编排抛光

## 15. 开发与调试说明

### 15.1 先选调试模式

日常开发现在建议按下面三种模式选：

- `producer`
  - 适合只调 HTTP/API、Swagger、handler/service 主链
  - 不需要 runtime worker 异步消费时优先用它
- `both`
  - 适合本地联调 submit -> queue -> worker -> query 的完整闭环
  - 一次起 HTTP + worker，最省事
  - 注意：`both` 不是六服务模式，不会同时起 `api-gateway / iam-service / resource-service / orchestrator-service / runtime-worker-service / system-service`
- 六服务模式
  - 适合调试微服务边界、internal client、remote-first 路径、服务 ownership
  - 需要确认 gateway 是否真的走 gRPC、某个 dedicated service 是否不再回退本地实现时，用这一套

简单建议：

- 改接口/页面联调：先用 `producer`
- 改异步执行链：先用 `both`
- 改 internal client / gRPC / 服务边界：直接用六服务模式

### 15.2 本地基础依赖

先起基础依赖：

```bash
docker compose up -d redis mysql etcd jaeger buildkitd loki prometheus grafana
```

配置主文件：

- `src/config.dev.toml`

重点配置：

- MySQL / Redis / Etcd / Loki / BuildKit 连接
- `clients.iam.target`
- `clients.resource.target`
- `clients.orchestrator.target`
- `clients.runtime.target`
- `clients.system.target`
- `iam.grpc.addr`
- `resource.grpc.addr`
- `orchestrator.grpc.addr`
- `runtime_worker.grpc.addr`
- `system.grpc.addr`

### 15.3 最常用启动方式

#### A. 只调 HTTP

```bash
cd src && go run . producer -conf ./config.dev.toml -port 8082
```

适合：

- router / handler / module service
- Swagger / OpenAPI
- Portal / Admin / SDK HTTP 联调

#### B. 调完整单机闭环

```bash
cd src && go run . both -conf ./config.dev.toml -port 8082
```

适合：

- execution / injection submit
- queue 消费
- task / trace / logs 主链

#### C. 调微服务边界

建议顺序：

```bash
# terminal 1
cd src && go run ./cmd/iam-service -conf ./config.dev.toml

# terminal 2
cd src && go run ./cmd/orchestrator-service -conf ./config.dev.toml

# terminal 3
cd src && go run ./cmd/resource-service -conf ./config.dev.toml

# terminal 4
cd src && go run ./cmd/runtime-worker-service -conf ./config.dev.toml

# terminal 5
cd src && go run ./cmd/system-service -conf ./config.dev.toml

# terminal 6
cd src && go run ./cmd/api-gateway -conf ./config.dev.toml -port 8082
```

如果只想调某一条边界，不需要六个都起：

- 调 auth/user/rbac/access key：起 `iam-service` + `api-gateway`
- 调 project/container/dataset/evaluation：起 `resource-service` + `api-gateway`
- 调 submit/task/trace：起 `orchestrator-service` + `api-gateway`
- 调 monitor/config/audit：起 `system-service` + `api-gateway`
- 调 worker/runtime：起 `runtime-worker-service`，必要时再带 `orchestrator-service`

### 15.4 如何判断现在该打在哪一层断点

#### HTTP 问题

优先看：

- `src/router/*`
- `src/module/*/handler.go`
- `src/module/*/service.go`

如果是 dedicated gateway 路径，再看：

- `src/app/gateway/*`
- `src/internalclient/*`

判断原则：

- 请求没进业务：看 router / middleware / handler
- 请求进了业务但结果不对：看 module service / repository
- dedicated gateway 下结果和单体模式不同：看 `app/gateway` remote-aware 装配和 `internalclient/*`

#### gRPC / 微服务边界问题

优先看：

- `src/internalclient/*`
- `src/interface/grpc*/*`
- 对应 `src/app/{gateway,iam,resource,orchestrator,runtime,system}/*`

判断原则：

- 调用没发出去：看 internal client target、dial、interceptor
- 服务收不到：看 grpc service registration / lifecycle
- dedicated service 启动就失败：先查 target 配置是否缺失

#### 异步执行链问题

优先看：

- `src/interface/worker/*`
- `src/interface/controller/*`
- `src/service/consumer/*`
- `src/module/task/*`
- `src/module/execution/*`
- `src/module/injection/*`
- `src/infra/k8s/*`

判断原则：

- submit 成功但没消费：先看 Redis / consumer
- 消费了但没执行：看 runtime owner、k8s/build/helm gateway
- 执行了但状态没回写：看 orchestrator owner facade / consumer owner adapter

### 15.5 快速验活命令

HTTP：

```bash
curl -I http://127.0.0.1:8082/docs/doc.json
curl -i http://127.0.0.1:8082/system/health
```

gRPC：

```bash
grpcurl -plaintext 127.0.0.1:9091 list
grpcurl -plaintext 127.0.0.1:9092 list
grpcurl -plaintext 127.0.0.1:9093 list
grpcurl -plaintext 127.0.0.1:9094 list
grpcurl -plaintext 127.0.0.1:9095 list
```

### 15.6 推荐调试顺序

遇到问题时建议固定按这条顺序排：

1. 服务有没有起来
2. 配置 target/addr 对不对
3. 请求到底走的是本地还是 remote
4. request-id 是否贯通
5. 业务 service 是否收到正确参数
6. infra gateway / DB / Redis / K8s 是否返回异常

### 15.7 现在最重要的几个判断点

- 调 dedicated `api-gateway` 时，不要默认它会静默回退本地 owner service
- 调 `system-service` / `runtime-worker-service` 时，先确认对应 `clients.*.target` 已配
- 调 submit / task / trace 闭环时，优先用 `both`
- 调 ownership / internal RPC 时，优先用六服务模式
- 调 repository 逻辑时，优先从各模块 `src/module/*/repository.go` 看，不要再去旧 compat 层找

### 15.8 常用回归命令

```bash
cd src && go test ./...
cd src && go test ./app -run 'TestProducerOptionsValidate|TestProducerOptionsStartStopSmoke|TestProducerOptionsHTTPIntegrationSmoke'
cd src && go test ./app -run 'TestConsumerOptions|TestBothOptions'
cd src && go test ./router ./docs ./interface/http
```

真实 K8s 集群验收：

```bash
cd src && RUN_K8S_INTEGRATION=1 go test ./infra/k8s -run TestK8sGatewayJobLifecycleIntegration
```

### 15.9 一句话建议

- 大多数日常功能开发：先 `producer`
- 需要异步闭环：用 `both`
- 需要查微服务边界：直接六服务
- 查不清时先看 `app/*` 装配，再看 `module/*/service.go`，最后看 `infra/*`

### 15.10 按模块分类的 debug 路线图

#### Auth / User / RBAC / Team

先看：

- `src/module/auth/handler.go`
- `src/module/auth/service.go`
- `src/module/auth/repository.go`
- `src/module/user/handler.go`
- `src/module/user/service.go`
- `src/module/user/repository.go`
- `src/module/rbac/handler.go`
- `src/module/rbac/service.go`
- `src/module/rbac/repository.go`
- `src/module/team/handler.go`
- `src/module/team/service.go`
- `src/module/team/repository.go`

如果是 dedicated gateway 下的认证/权限问题，再看：

- `src/app/gateway/auth_services.go`
- `src/app/gateway/user_services.go`
- `src/app/gateway/rbac_services.go`
- `src/app/gateway/team_services.go`
- `src/internalclient/iamclient/*`
- `src/interface/grpciam/*`

常见问题先查：

- 登录/换 token：`auth/service.go` + `iamclient`
- 权限不对：`middleware/*` + `rbac/service.go`
- team project/member 视图不对：`team/service.go` + remote project reader

#### Project / Label / Container / Dataset

先看：

- `src/module/project/handler.go`
- `src/module/project/service.go`
- `src/module/project/repository.go`
- `src/module/label/handler.go`
- `src/module/label/service.go`
- `src/module/label/repository.go`
- `src/module/container/handler.go`
- `src/module/container/service.go`
- `src/module/container/repository.go`
- `src/module/dataset/handler.go`
- `src/module/dataset/service.go`
- `src/module/dataset/repository.go`

如果是 dedicated gateway / resource-service 边界问题，再看：

- `src/app/gateway/resource_services.go`
- `src/internalclient/resourceclient/*`
- `src/interface/grpcresource/*`

常见问题先查：

- list/detail 不对：各模块 `repository.go` 查询条件
- label 关系不对：`project/container/dataset` service 里的 label 管理逻辑
- 统计字段不对：`project` statistics source 与 orchestrator/resource 边界

#### Injection / Execution / Task / Trace / Group / Notification

先看：

- `src/module/injection/handler.go`
- `src/module/injection/service.go`
- `src/module/injection/repository.go`
- `src/module/execution/handler.go`
- `src/module/execution/service.go`
- `src/module/execution/repository.go`
- `src/module/task/handler.go`
- `src/module/task/service.go`
- `src/module/task/repository.go`
- `src/module/trace/handler.go`
- `src/module/trace/service.go`
- `src/module/group/handler.go`
- `src/module/group/service.go`
- `src/module/notification/handler.go`
- `src/module/notification/service.go`

如果是 submit / task / trace / stream 走向问题，再看：

- `src/app/gateway/orchestrator_services.go`
- `src/internalclient/orchestratorclient/*`
- `src/interface/grpcorchestrator/*`

如果是异步执行闭环问题，再补看：

- `src/service/consumer/*`
- `src/interface/worker/*`
- `src/interface/controller/*`

常见问题先查：

- submit 成功但 task 不生成：`injection/execution service` -> orchestrator facade
- task 有了但状态不推进：`service/consumer` + owner adapter
- trace/group/notification stream 不对：`orchestrator_services.go` + stream read RPC
- task logs WebSocket 不对：`task/service.go` + orchestrator log poll

#### System / SystemMetric / Monitor / Config / Audit

先看：

- `src/module/system/handler.go`
- `src/module/system/service.go`
- `src/module/system/repository.go`
- `src/module/systemmetric/handler.go`
- `src/module/systemmetric/service.go`

如果是 dedicated system-service / gateway 边界问题，再看：

- `src/app/gateway/system_services.go`
- `src/internalclient/systemclient/*`
- `src/internalclient/runtimeclient/*`
- `src/interface/grpcsystem/*`
- `src/interface/grpcruntime/*`

常见问题先查：

- config/audit 查询不对：`system/repository.go`
- monitor / queue / lock 不对：`system/service.go` 是否走 runtime RPC
- `/system/health` 异常：`system/handler.go` + service health 依赖

#### Runtime / K8s / Build / Helm / Chaos

先看：

- `src/service/consumer/*`
- `src/interface/worker/*`
- `src/interface/controller/*`
- `src/infra/k8s/*`
- `src/infra/buildkit/*`
- `src/infra/helm/*`
- `src/infra/chaos/*`
- `src/infra/redis/*`

如果是 dedicated runtime-worker-service 问题，再看：

- `src/app/runtime/*`
- `src/internalclient/orchestratorclient/*`
- `src/interface/grpcruntime/*`

常见问题先查：

- queue 不消费：`consumer` + Redis
- k8s job 不创建：`infra/k8s`
- build / helm 失败：对应 `infra/buildkit` / `infra/helm`
- 状态回写不到 orchestrator：`consumer owner` + orchestrator client

#### 看文件顺序的偷懒法

如果你一时不确定从哪进，统一按这个顺序看：

1. `src/router/*` 或 `src/interface/grpc*/*`
2. `src/module/*/handler.go`
3. `src/module/*/service.go`
4. `src/module/*/repository.go`
5. `src/app/*` 装配
6. `src/internalclient/*`
7. `src/infra/*`
