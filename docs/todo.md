# Backend Fx Refactor TODO

> 创建日期：2026-04-15  
> 目标：把后端从全局初始化 + 包级函数，逐步迁移到 Fx app + 明确模块边界 + 生命周期管理。

## 使用方式

- 先画模块边界，再做 DI。
- 每次只迁移一个入口或一个模块。
- 第一阶段不改 URL，不搬大目录，不重写业务逻辑。
- Fx 先管理启动和生命周期，再逐步替换 handler / service / repository 的包级函数。
- 当前已有旧 DI 骨架视为临时试验，后续由 Fx 替换。

## 0. 准备阶段

- [x] 确认项目更适合 Fx，而不是继续扩大旧 DI 方案
- [x] 确认后端有多入口：producer / consumer / both
- [x] 确认有多基础设施资源：DB / Redis / Etcd / K8s / Loki / tracing / receiver
- [x] 确认第一阶段不改 URL
- [x] 确认第一阶段不大规模搬目录
- [x] 决定是否立即删除当前旧 DI 骨架
- [ ] 确认 Fx 生成的启动日志是否可接受
  - 属于人工验收项，不阻塞当前代码主线收口。

验证：

- [x] 阅读汇总文档与主线设计说明
  - 当前总入口已收口到 [report-index.md](./report-index.md)
- [x] `cd src && go test ./app ./router ./handlers/v2`
  - 实际执行：`cd src && go test ./app ./interface/http ./router ./handlers/v2`

## 1. 停止继续旧 DI 扩张

- [x] 不再新增旧 DI provider set
- [x] 不再继续按旧 DI TODO 迁移 Project service / repository
- [x] 决定旧 DI 文件处理方式
  - [x] 方案 A：立即删除 app 下旧 DI 生成文件与相关依赖
  - 方案 B 未采用：不再保留旧 DI 骨架等待后删。
- [x] 文档和 TODO 全部切换到 Fx 方案

验证：

- [x] 检查 app 与依赖中旧 DI 痕迹
  - 代码和依赖已删除；文档中的历史说明也已切成中性表述。

## 2. 引入 Fx 基础设施

- [x] 在 `src/go.mod` 增加 `go.uber.org/fx`
- [x] 新建或调整 `src/app` 为 Fx app 入口
- [x] 新建 `src/app/options.go`
- [x] 新建 `src/app/producer.go`
- [x] 新建 `src/app/consumer.go`
- [x] 新建 `src/app/both.go`
- [x] 定义 `CommonOptions()`
- [x] 定义 `ProducerOptions()`
- [x] 定义 `ConsumerOptions()`
- [x] 定义 `BothOptions()`

目标：

```go
func ProducerOptions() fx.Option
func ConsumerOptions() fx.Option
func BothOptions() fx.Option
```

验证：

- [x] `cd src && go test ./app`

## 3. Config / Logger Module

- [x] 新建 `src/infra/config/module.go`
- [x] 包装现有 `config.Init`
- [x] 让配置路径从 app 参数传入，而不是各处自行读取
- [x] 新建 `src/infra/logger/module.go`
- [x] 把 logrus 初始化从 `main.go` 收进 logger module
- [x] 确认 logger 初始化只执行一次

验收：

- [x] `main.go` 不再直接配置 logger
- [x] `main.go` 不再直接调用 `config.Init`
  - 当前 `config.Init` 仅保留在 `infra/config` module 与少量测试中，producer / consumer / both 主启动链均已通过 Fx 配置模块进入。
- [x] `cd src && go test ./infra/config ./infra/logger`

## 4. DB Module

- [x] 新建 `src/infra/db/module.go`
- [x] 新建 `NewGormDB`
- [x] 将现有 `database.InitDB()` 包装进 Fx provider 或重构为返回 `*gorm.DB`
- [x] DB module 提供 `*gorm.DB`
- [x] 使用 `fx.Lifecycle` 注册 DB close
- [x] 过渡期继续同步 `database.DB = db`，避免一次性修改旧代码

目标：

```go
var Module = fx.Module("db",
    fx.Provide(NewGormDB),
)
```

验收：

- [x] producer 可通过 Fx 初始化 DB
- [x] `database.DB` 兼容旧代码
- [x] DB 关闭逻辑在 `OnStop`

## 5. Redis / Etcd / Tracing Module

Redis：

- [x] 新建 `src/infra/redis/module.go`
- [x] 提供 Redis client
- [x] `OnStop` 关闭 Redis
- [x] Redis 实现已从 `src/client/redis_client.go` 并入 `src/infra/redis/*`
- [x] `src/infra/redis/client.go` 已删除，连接创建已继续并入 `src/infra/redis/gateway.go` 私有方法
- [x] 过渡期兼容 `client.GetRedisClient()`
  - 兼容期已结束；`module/system` / `module/trace` / `module/group` / `module/notification` / `service/common` / `service/consumer` / `service/logreceiver` 等调用点已切到 `redisinfra`。

Etcd：

- [x] 新建 `src/infra/etcd/module.go`
- [x] 提供 Etcd client 或 gateway
- [x] 收口 Etcd watch / get / put 的初始化
- [x] Etcd 实现已从 `src/client/etcd_client.go` 并入 `src/infra/etcd/*`
- [x] `src/infra/etcd/client.go` 已删除，连接创建已继续并入 `src/infra/etcd/gateway.go` 私有方法

Tracing：

- [x] 新建 `src/infra/tracing/module.go`
- [x] 包装 `client.InitTraceProvider()`
- [x] 如支持 shutdown，则注册 `OnStop`
- [x] tracing provider 实现已从 `src/client/jaeger.go` 并入 `src/infra/tracing/*`

Loki：

- [x] 新建 `src/infra/loki/module.go`
- [x] Fx graph 提供 `*client.LokiClient`
- [x] `module/task.LokiGateway` 注入 `*client.LokiClient`，不再自行 `client.NewLokiClient()`
- [x] Loki 实现已从 `src/client/loki.go` 并入 `src/infra/loki/*`
- [x] `module/task` / `module/injection` / `app.CommonResources` 已切到 `lokiinfra.Client`

验收：

- [x] 基础设施资源由 Fx module 创建
- [x] 旧代码仍可运行
- [x] `cd src && go test ./infra/...`
  - 实际执行：`cd src && go test ./app ./infra/config ./infra/logger ./infra/db ./infra/redis ./infra/etcd ./infra/tracing ./interface/http ./router ./handlers/v2`

## 6. HTTP Interface Module

- [x] 新建 `src/interface/http/module.go`
- [x] 新建 `src/interface/http/server.go`
- [x] 新建 `src/interface/http/router.go`
- [x] 将现有 `router.New(...)` 包装为 Fx provider
- [x] HTTP server 使用 `http.Server`
- [x] `OnStart` 启动 server goroutine
- [x] `OnStop` graceful shutdown
- [x] producer 模式通过 Fx 启动 HTTP server

目标：

```go
var Module = fx.Module("http",
    fx.Provide(NewGinEngine, NewHTTPServer),
    fx.Invoke(RegisterHTTPServerLifecycle),
)
```

验收：

- [x] `main.go producer` 不再直接 `engine.Run`
- [x] HTTP server 可优雅停止
- [x] API URL 不变
- [x] `cd src && go test ./interface/http ./router`
  - 实际执行：`cd src && go test ./app ./interface/http ./router ./handlers/v2`

## 7. main.go 收口

- [x] `main.go` 只保留 cobra mode 解析
- [x] producer mode 调用 `fx.New(app.ProducerOptions(...)).Run()`
- [x] consumer mode 调用 `fx.New(app.ConsumerOptions(...)).Run()`
- [x] both mode 调用 `fx.New(app.BothOptions(...)).Run()`
- [x] 删除 `main.go` 中直接 DB 初始化
- [x] 删除 `main.go` 中直接 trace 初始化
- [x] 删除 `main.go` 中直接 HTTP server 启动

验收：

- [x] `main.go` 明显变薄
- [x] producer 可启动
- [x] consumer 暂时可保留旧逻辑或已接入 Fx
- [x] both 可启动

## 8. Consumer / Scheduler Module

- [x] 新建 `src/interface/worker/module.go`
- [x] 包装 `consumer.StartScheduler`
- [x] 包装 `consumer.ConsumeTasks`
- [x] 使用 Fx lifecycle 管理 context cancel
- [x] `OnStart` 启动 scheduler goroutine
- [x] `OnStart` 启动 consumer goroutine
- [x] `OnStop` cancel context
- [x] 避免 consumer 阻塞 Fx 启动流程

验收：

- [x] consumer mode 通过 Fx 启动
- [x] both mode 通过 Fx 同时启动 HTTP 和 consumer
- [x] 停止时能 cancel worker context

## 9. K8s Controller / Chaos Module

- [x] 新建 `src/infra/k8s/module.go`
- [x] 包装 `k8s.GetK8sController()`
- [x] 包装 K8s rest config
- [x] 新建 `src/infra/chaos/module.go`
- [x] 包装 `chaosCli.InitWithConfig`
- [x] 新建 `src/interface/controller/module.go`
- [x] 用 lifecycle 启动 K8s controller
- [x] 用 lifecycle 停止 controller context

验收：

- [x] consumer / both mode 中 K8s controller 由 Fx 启动
- [x] 初始化顺序由 Fx 表达

## 10. OTLP Receiver Module

- [x] 新建 `src/interface/receiver/module.go`
- [x] 包装 `logreceiver.NewOTLPLogReceiver`
- [x] receiver port 从 config module 注入
- [x] `OnStart` 启动 receiver
- [x] `OnStop` shutdown receiver

验收：

- [x] consumer / both mode 中 receiver 由 Fx 启动
- [x] 停止时 receiver 正常关闭

## 11. HTTP Routes 按受众拆分

先拆注册函数，不改 URL。

- [x] 新建或迁移 public routes
- [x] 新建或迁移 sdk routes
- [x] 新建或迁移 portal routes
- [x] 新建或迁移 admin routes
- [x] 整理 system routes
- [x] route 注册依赖 handler 容器

建议：

```go
func RegisterPublicRoutes(...)
func RegisterSDKRoutes(...)
func RegisterPortalRoutes(...)
func RegisterAdminRoutes(...)
func RegisterSystemRoutes(...)
```

验收：

- [x] URL 不变
- [x] `router/v2.go` 变薄
  - 从 647 行降到 365 行；核心业务路由仍保留在 `v2.go`，后续随业务 module 迁移继续拆。
- [x] Admin / SDK / Portal 边界在代码上可见

## 12. 业务 Module 壳

先建立壳，不急着重写内部逻辑。

- [x] `module/project`
- [x] `module/auth`
- [x] `module/task`
- [x] `module/injection`
- [x] `module/execution`
- [x] `module/container`
- [x] `module/dataset`
- [x] `module/rbac`
- [x] `module/user`

每个模块先暴露：

```go
var Module = fx.Module("project",
    fx.Provide(NewHandler),
)
```

过渡期 handler 可以 wrapper 旧函数。

验收：

- [x] app 通过业务 module 收集 handler
  - 已新增 `app.ProducerHTTPModules()` 与 `router.Module`，Producer/Both 由 app 统一收集业务 module，再向 HTTP interface 提供 `router.Handlers`。
- [x] router 不直接散装引用所有裸函数
  - 业务路由已统一经 `router.Handlers` 聚合，`interface/http` 不再散装依赖各模块构造；剩余主要是旧兼容层清理与少量 middleware/初始化收尾。

## 13. Project 模块正式迁移

Project 作为第一个完整业务样板。

- [x] 新建 `module/project/repository.go`
- [x] 新建 `module/project/service.go`
- [x] 新建 `module/project/handler.go`
- [x] 新建 `module/project/module.go`
- [x] Repository 注入 `*gorm.DB`
- [x] Service 注入 Repository
  - RBAC 独立接口化与 Label 接口进一步抽象保留为后续优化项，不阻塞当前主线。
- [x] Handler 注入 Service
- [x] Handler method 使用 `c.Request.Context()`
- [x] 移除 Project handler wrapper 对旧包级函数的依赖
  - Project CRUD/labels 已移除旧 wrapper；project 下 injection/execution routes 已切到 `module/injection` 和 `module/execution`。
- [x] Project routes 使用新 handler

验收：

- [x] Project handler 不直接 import `database`
- [x] Project handler 不直接 import repository implementation
- [x] Project service 不直接使用全局 `database.DB`
- [x] Project CRUD 行为不变

## 14. Auth / Task 模块迁移

Auth：

- [x] 新建 Auth module
- [x] Token blacklist 从 repository 迁移到 store
  - 新路由已走 `module/auth.TokenStore`；旧 `repository/token.go` 与 `service/producer/auth.go` 已删除。
- [x] Auth service 注入 UserRepository / RoleRepository / TokenStore
- [x] Auth handler method 化

Task：

- [x] 新建 Task module
- [x] Task queue Redis 访问收口到 store
- [x] WebSocket handler 只做认证和连接升级
  - 日志推送已走 `module/task.TaskLogService`。
- [x] 日志历史查询走 Loki gateway
- [x] 订阅逻辑走 service / store
  - Redis Pub/Sub 已收口到 `module/task.TaskQueueStore`；task state polling 已收口到 `module/task.TaskLogService`。

验收：

- [x] `handlers/v2/tasks.go` 不再 direct import `database`
- [x] `handlers/v2/tasks.go` 不再 direct import `repository`
  - 旧文件已删除，Task 路由切到 `module/task.Handler`。
- [x] `repository/token.go` 能力迁移出去
  - Auth 黑名单能力已统一收口到 `module/auth.TokenStore`。

## 15. 核心业务模块逐个迁移

按顺序推进：

- [x] Injection
  - 已建立 module/handler/service 壳并切 Project 子路由；深层 producer/repository 逻辑后续继续下沉。
- [x] Execution
  - 已建立 module/handler/service 壳并切 Project 子路由；深层 producer/repository 逻辑后续继续下沉。
- [x] Container
- [x] Dataset
- [x] Evaluation
- [x] Trace
- [x] Metrics
- [x] Group
- [x] Notification
- [x] SDK Evaluation
- [x] Chaos System

主线完成项：

- [x] Module / Handler / Service / Repository 壳
- [x] Fx providers
- [x] route 切换
- [x] 主路径所需 Store / Gateway 收口
- [x] 关键模块级测试
  - Container / Dataset / Evaluation / Trace / Group / Metrics / Notification / SDK Evaluation / Chaos System 已完成 module/handler/service/repository 壳、Fx providers 和 route 切换；其中 Metrics / SDK Evaluation / Chaos System 已进一步切离 `service/producer` 包级入口。测试层面已覆盖 auth / project / execution / injection / task / user / sdk / docs / app 等主路径，剩余测试补强属于后续质量项，不阻塞当前主线。

## 16. Store / Gateway 拆分

- [x] Redis token blacklist -> TokenStore
  - 旧 `repository/token.go` 已删除，认证退出逻辑统一走 `module/auth/token_store.go`。
- [x] Redis task queue -> TaskQueueStore
  - 已新增 `src/infra/redis/task_queue.go`，consumer / scheduler / system monitor 队列读写全部从 `repository/task.go` 迁出。
- [x] Loki -> LokiGateway
  - 当前完成 Task 日志查询侧，并将 Loki client 纳入 Fx graph；Injection 日志查询仍待深层 service 迁移。
- [x] K8s -> K8sGateway
  - 已新增 `src/infra/k8s/gateway.go`，统一收口 controller / create job / volume mount / job logs / health check 访问；并已把 `src/client/k8s/*` 的真实实现整体迁入 `src/infra/k8s/*`。其中 `RestConfig / Client / DynamicClient` 这类简单转发已继续收口，改由 `infra/k8s` 内部私有 getter 与 Fx provider 使用。
- [x] Etcd -> EtcdGateway
  - 已新增 `src/infra/etcd/gateway.go`，配置监听与动态配置发布已切到 gateway；`Put/Get/Delete/Watch` 逻辑也已收回 gateway，`client.go` 仅保留底层连接创建与关闭。
- [x] Harbor -> HarborGateway
  - 已新增 `src/infra/harbor/gateway.go`，并把 Harbor client 实现从 `src/client/harbor_client.go` 并入 `src/infra/harbor/*`；对外不再保留空转发 `Client` 抽象，逻辑已直接内聚到 gateway。
- [x] Helm -> HelmGateway
  - 已新增 `src/infra/helm/gateway.go`，consumer pedestal 安装侧已改由 gateway 直接承担 repo/install 逻辑；Helm 实现已从 `src/client/helm.go` 并入 `src/infra/helm/*`，不再额外保留对外 `Client` 层。
- [x] BuildKit -> BuildKitGateway
  - 已新增 `src/infra/buildkit/gateway.go`，BuildKit 健康检查与构建 client 创建已开始从业务层抽离。

验收：

- [x] repository 只负责 DB
- [x] service 依赖接口而不是全局 client
  - Redis / Loki / Etcd / Harbor / Helm / K8s 这批调用点已不再依赖 `aegis/client` 包级入口；root `client` 目录现仅剩 debug 与 aegisctl 客户端侧代码。
- [x] 外部系统资源按需纳入 Fx graph / lifecycle 管理
  - Redis / Etcd / tracing 已在 Fx lifecycle 中管理；Harbor / Helm / Loki 当前以无状态或按需 client 为主，不额外引入 shutdown 生命周期也不阻塞主线。

## 17. SDK / Portal / Admin 标记治理

当前口径：

- 不再在代码里 hardcode audience allowlist。
- audience 归属统一以 `src/docs/openapi3/openapi.json` 里的 `x-api-type` 扩展为准。
- 只要某个 operation 带有对应 key 且值为 `"true"`，就会被提取进对应产物。
- 同一个 operation 可以同时落入多个 audience。
- Python SDK 只消费 `sdk.json`；TypeScript SDK 不再消费共享并集视图，而是分别按 `portal.json` / `admin.json` 生成独立 Portal SDK 与 Admin SDK。
- SDK audience 改为显式白名单；默认不再把通用登录、Portal/Admin 控制面接口顺手放进 `sdk.json`。
- SDK / CLI 认证主线改为 `Key ID / Key Secret -> access token`；`username/password login` 仅保留给 Portal / Admin 等人类交互入口。
- `Key ID / Key Secret -> token` 进一步改为 header 签名模式：`X-Key-Id`、`X-Timestamp`、`X-Nonce`、`X-Signature`，业务接口仍继续走 Bearer token。

本轮执行清单：

- [x] 收缩 `sdk` audience 到最小可维护白名单
  - 已继续把剩余误标的 `sdk` audience 收回，只保留 `POST /api/v2/auth/api-key/token` 与 `src/router/sdk.go` 下 4 个 SDK 样例接口；当前 `sdk.json` 已收缩到 `5 paths / 5 operations`。
- [x] 从 Swagger audience 中移除 `POST /api/v2/auth/register` 的 `sdk`
- [x] 从 Swagger audience 中移除 `POST /api/v2/auth/login` 的 `sdk`
- [x] 盘点并设计 API key / Key ID / Key Secret 数据模型
  - 已新增 `model.APIKey`，并把物理 schema 一并改到 `api_keys` / `key_id` / `key_secret_hash` / `key_secret_ciphertext` / `active_key_id`，覆盖 `owner`、`enabled/disabled/deleted`、`expires_at`、`last_used_at`、`name/description` 等字段，并纳入 `AutoMigrate`。
- [x] 增加 API key 管理接口
  - 已补 `portal` 路由：`GET/POST /api/v2/api-keys`、`GET/DELETE /api/v2/api-keys/{id}`、`POST /api/v2/api-keys/{id}/rotate|disable|enable`。
- [x] 增加 `Key ID / Key Secret -> token` 接口并标为 `sdk`
  - 已补 `POST /api/v2/auth/api-key/token`，返回 Bearer token，并在 JWT claims 中标记 `auth_type=api_key` 与 `api_key_id`；当前入口统一使用 `X-Key-Id` / `X-Timestamp` / `X-Nonce` / `X-Signature` 头签名校验，canonical string 已收敛为 `METHOD\\nPATH\\nTIMESTAMP\\nNONCE\\nSHA256(BODY)`，服务端会校验 5 分钟时间窗并用 Redis 做 nonce 防重放；`iam.proto` / `src/interface/grpciam` / `src/internalclient/iamclient` 这条内部链也已统一到 `key_id` 字段名。
- [x] 将 Python SDK 的鉴权入口切到 Key ID / Key Secret
  - `sdk/python/src/rcabench/client/http_client.py` 已改为优先使用 `token` 或 `key_id + key_secret`；SDK 不再依赖 username/password login，环境变量主入口切到 `RCABENCH_KEY_ID` / `RCABENCH_KEY_SECRET`，并按 `METHOD\\nPATH\\nTIMESTAMP\\nNONCE\\nSHA256(BODY)` 规范计算 HMAC-SHA256 签名头；重生成后的 `sdk/python/src/rcabench/openapi/*` 也已切到 `X-Key-Id`、`key_id`、`key_secret` schema，不再暴露旧 `X-Access-Key` / `access_key` / `secret_key` 鉴权字段。
- [x] 将 `aegisctl` 的鉴权入口切到 Key ID / Key Secret
  - `src/cmd/aegisctl/cmd/auth.go` / `src/cmd/aegisctl/client/auth.go` 已切到 `--key-id` + `--key-secret` 签名换取 `POST /api/v2/auth/api-key/token`；环境变量主入口改为 `AEGIS_KEY_ID` / `AEGIS_KEY_SECRET`，登录结果继续只落盘 Bearer token 与 `key_id`，不保存 `key_secret`；旧 flag/env 与响应字段兼容读取已删除。
- [x] 给 `aegisctl` 增加本地签名排障命令
  - 已补 `aegisctl auth inspect` 与 `aegisctl auth sign-debug`；前者可检查当前 context 的 token / auth_type / key_id / expiry，后者可直接打印 canonical string、签名头与 curl 样例，并可通过 `--execute` 直接发起换 token 请求回显响应，或通过 `--save-context` 直接把成功返回的 Bearer token 落盘到当前 CLI context，便于排查 SDK / CLI / 服务端签名不一致问题。
- [x] 补充 Key ID / Key Secret 头签名规范文档
  - 相关说明现已并入 `docs/report-index.md`：明确 canonical string、Header 约定、HMAC 规则、时间窗与 nonce 防重放语义，并补了 Portal 上 API key 的使用说明与 `aegisctl` 排障命令说明；Swagger 注释与生成文档也已统一到 `X-Key-Id`、`key_id`、`key_secret` 口径，可直接供文档站与 SDK 生成消费。
- [x] 收口 API key 命名与样例前缀
  - auth handler / service / gRPC / internal client / CLI / Swagger / Python SDK 现已统一使用 `API key`、`key_id`、`key_secret`；公开样例前缀也统一为 `pk_...` / `ks_...`。Go 存储模型与物理 schema 现都已统一到 `APIKey` / `api_keys` / `key_id` 口径，不再保留旧 `user_access_keys` / `access_key` 兼容层。
- [x] 补 API key 的 `scopes` / `revoked_at` 语义
  - `model.APIKey` 已新增 `scopes` 与 `revoked_at`；创建接口支持提交 `scopes`，默认会归一化为 `["*"]`；列表/详情/创建/轮换响应会返回 scopes 与 revoked_at；同时新增 `POST /api/v2/api-keys/{id}/revoke`，被 revoke 的 API key 会被永久拒绝换 token，且不能再 re-enable / rotate。
- [x] 把 API key scopes 继续推进到 bearer token / gRPC verify / middleware 上下文
  - `utils.Claims` 已补 `api_key_scopes`，`Key ID / Key Secret -> token` 成功后签发的 JWT 会携带 scopes；`iam.proto` / `src/interface/grpciam` / `src/internalclient/iamclient` 的 verify 响应链也已同步透传，HTTP middleware 现会把 `auth_type` / `api_key_id` / `api_key_scopes` 一并放入请求上下文，后续做 scope enforcement 不用再回查 API key 表。
- [x] 给 API key scopes 接上首版运行时拦截
  - `src/middleware/permission.go` 现已在 permission middleware 里先按 `api_key_scopes` 做匹配，再落 DB 权限校验；当前支持 `*`、`resource`、`resource:action`、`resource:action:scope` 以及各段 `*` 通配，先覆盖所有基于 `RequirePermission/RequireAnyPermission/RequireAllPermissions` 的路由。
- [x] 把 team/project 成员关系型中间件也接上 API key scopes 预过滤
  - `RequireTeamMemberAccess` / `RequireTeamAdminAccess` / `RequireProjectAccess(...)` 现在会先按 API key scope 做 read/manage 级别过滤，再执行成员/管理员关系判断，避免 API key bearer token 绕过非 permission 型访问守卫。
- [x] 再扫一轮 JWTAuth-only 路由，把明显漏掉的敏感守卫补齐
  - 已补上 `team list/create`、`/api/v2/resources*`、`/api/v2/systems*`、`/api/v2/system/metrics*` 这批原先只有 `JWTAuth()` 的敏感入口；当前剩余仅 `auth profile/logout/change-password`、`/api/v2/api-keys/*` 自助凭证管理，以及 `sdk` 样例查询这几类刻意保留的 JWTAuth-only 路由，后两者若要继续收紧可再单独引入更明确的 API key/self-service scope 语义。
- [x] 给 `sdk/*` 和 `/api/v2/api-keys/*` 落一版明确语义
  - `src/router/sdk.go` 已引入显式 API key scope gate：`/api/v2/sdk/evaluations*` 需要 `sdk:*` / `sdk:evaluations:*` / `sdk:evaluations:read`，`/api/v2/sdk/datasets` 需要 `sdk:*` / `sdk:datasets:*` / `sdk:datasets:read`；同时 `src/router/portal.go` 的 `/api/v2/api-keys/*` 已统一挂 `RequireHumanUserAuth()`，明确只允许人类用户 session 管理 API key，禁止“API key 再管理 API key”。
- [x] 把 auth 自助接口也限制为 human session
  - `src/router/public.go` 的 `/api/v2/auth/profile`、`/logout`、`/change-password` 现已统一挂 `RequireHumanUserAuth()`；当前 API key bearer token 只保留给显式允许的 SDK/业务 API，不再可进入用户账号自助管理接口。
- [x] 启动 Python SDK / runtime wrapper 分层主线第一批落地
  - 已新增 `docs/python-runtime-wrapper-design.md` 与 `docs/python-runtime-wrapper-todo.md`；Swagger `x-api-type` 现支持 `runtime` audience，并新增 `src/docs/converted/runtime.json`；`module/execution` 的 detector/granularity upload 已标为 `runtime:"true"`，`src/router/runtime.go` 也已把这两条 `/api/v2/executions/{execution_id}/*_results` 路由挂到 `JWTAuth() + RequireServiceTokenAuth()`；同时 `sdk/python/src/rcabench/client/runtime_client.py` 已新增 `RCABenchRuntimeClient`，并从 Python 包根导出，作为后续 `rcabench-platform` wrapper 的 service-token-only 基础客户端。
- [x] 收紧 hand-written Python client 边界：public client 只保留 API key，runtime client 只保留 service token
  - `sdk/python/src/rcabench/client/base.py` 已新增 `BaseRCABenchClient` 抽出共享 session/api-client 生命周期；`RCABenchClient` 已删除直接 bearer token 模式，仅保留 `key_id + key_secret`；`RCABenchRuntimeClient` 现改成与 public client 同结构的 service-token-only connector，不再承载 detector/granularity upload 调度语义，后续 heartbeat / status / artifact/result 的调用时机统一留给外仓 `rcabench-platform` wrapper 控制。
- [x] 重新生成 `openapi3` / `sdk.json` 并回填最新统计
  - 当前生成结果为：`openapi3/openapi.json` `138 paths / 173 operations`，`sdk.json` `5 / 5`，`portal.json` `31 / 43`，`admin.json` `48 / 58`；Python SDK 已按最新 `sdk.json` 重新生成，TypeScript 侧改为分别消费 `portal.json` 与 `admin.json`。

完成项：

- [x] 统计 OpenAPI3 中的 `x-api-type` audience 标记
  - 当前已按 Go Swagger 注释补齐一批 `portal:"true"` / `admin:"true"` 标记；`converted/portal.json` 与 `converted/admin.json` 会分别作为独立 TypeScript SDK 的输入。
- [x] Python SDK 只提取 `x-api-type.sdk == "true"` 的接口
- [x] Portal 产物提取 `x-api-type.portal == "true"` 的接口
- [x] Admin 产物提取 `x-api-type.admin == "true"` 的接口
- [x] TypeScript Portal SDK 仅提取 `x-api-type.portal == "true"` 的接口
- [x] TypeScript Admin SDK 仅提取 `x-api-type.admin == "true"` 的接口
- [x] 更新 SDK 生成脚本
  - `scripts/command/src/swagger/init.py` 现会先完整重跑 `swag init`，再把 `openapi2/swagger.json` 本地转换成 `openapi3/openapi.json`，并继续产出 `client.json`、`sdk.json`、`portal.json`、`admin.json`；不再依赖 Docker 生成 OpenAPI3，也不再产生 root-owned 文档目录。
- [x] 修回 `swagger init` 主链可完整再生
  - 通过恢复 `src/handlers/debug.go`、`src/handlers/system/*`、`src/handlers/v2/*` 这批仅用于 Swagger 注释扫描的 build-ignored 文档桩，`swag init` 已重新稳定产出全量接口；当前 `openapi2/swagger.json`、`openapi3/openapi.json`、`converted/client.json` 均为 `132 paths / 165 operations`。
- [x] 校正 Python SDK 生成链
  - `scripts/command/src/formatter/python.py` 现优先使用本地 `scripts/command/.venv/bin/ruff`，缺失时也不会再因为 formatter 中断；`scripts/command/src/swagger/python.py` 的 Docker 生成步骤继续显式使用当前用户 UID:GID 运行，避免再次产出 root-owned 文件。
- [x] 重新生成 TypeScript SDK
  - 已执行 `cd scripts/command && ./.venv/bin/python main.py swagger generate-sdk -l typescript -v 1.2.1`
- [x] 检查 SDK diff
  - TypeScript 不再输出共享 `typescript.json` / `sdk/typescript`；当前口径改为 `sdk/typescript/portal` 与 `sdk/typescript/admin` 两套独立产物，分别只消费 `portal.json` 与 `admin.json`，避免 Portal/Admin 共用同一份 TS SDK。
- [x] 验证 audience 文档产物
  - 已执行 `cd scripts/command && ./.venv/bin/python main.py swagger init -v 1.2.1` 与 `cd src && go test ./docs`

## 18. 删除旧 DI 骨架和旧兼容层

等 Fx producer / consumer / both 跑通后执行。

- [x] 删除 app 下旧 DI 生成文件
- [x] 删除旧 DI 依赖
- [x] 删除过渡 handler wrapper
  - `src/handlers/v2` 现仅保留空的 `doc.go` 占位包以兼容既有测试命令，已不再承担任何运行态 wrapper 职责；`src/handlers/debug.go`、`src/handlers/system/*` 与 `src/handlers/v2/*` 旧兼容入口均已清空或删除。最近一轮又把 `src/app/producer_init.go`、`src/interface/{worker,controller,receiver}/module.go`、`src/interface/http/server.go` 中仅供 Fx 编排使用的注册 helper 全部缩成包内私有实现，启动链公开暴露面继续收口。
- [x] 删除旧包级 service 函数
  - `module/user` CRUD / 资源授权、`module/systemmetric` 指标查询、`module/rbac` 已基本切离 `service/producer`；`handlers/system/monitor.go`、`configs.go`、`audit.go` 主路由入口也已并入 `module/system`。此前已删除旧 `service/producer` 中的 system / metrics / sdk / chaos-system / permission / audit / evaluation / notification / team / trace / group 兼容入口；middleware 也不再直接依赖旧 producer。`module/container` 与 `module/dataset` 现已进一步把 CRUD / detail / list / labels / version 元数据、container build / helm upload、dataset filename / download / version injection 路径下沉到模块 service/repository，并把直接碰 `config` / git / 文件系统的部分收成模块内 gateway/store。旧 `service/producer/container.go` / `dataset.go` 已删除；初始化已改走 `module/container` / `module/dataset` 暴露的 core helper。最近几轮里，`module/injection` 已先后接管 datapack download / files / file query / upload / build 提交流程，以及 injection list / project list / detail / labels / logs / submit fault injection / search / no-issues / with-issues / clone / batch delete 主路径；`src/service/producer/injection.go` 已整体删除。随后又继续按“模块语义留在模块 repo、纯转发尽量删除”的口径收缩：`module/injection` 把 search / list / labels / batch label 管理，以及 project injection list 的标签装配收进 `repository.go`，并继续把 `LoadInjection` / `FindInjectionByName` / `CreateInjectionRecord` / `LoadTask` / `LoadPedestalHelmConfig` / label/execution 删除辅助等一批原子转发写实到模块仓储；最近三轮又把 project resolve、detail with labels、existing injection map、label 条件聚合、project injection list、issue/no-issue 视图、label id by key、fault injection 批量 with labels 这批组合查询继续收成模块内实现。`module/user` 这一轮又把 `CreateUser + EnsureUserUnique`、`Get/Update` 这批基础 CRUD 空包装进一步折成 `CreateUserIfUnique`、`GetUserDetailBase`、`UpdateMutableUser`、`ListUserViews`，并把 `DeleteUserCascade`、global/container/dataset/project 的 assign/remove、permission batch create/delete 这批 relation 逻辑也直接写进模块 repo；随后又把 user detail 关系装配，以及 role/container/dataset/project 的加载 helper 继续改为模块内直接查库；最近又把 permission id 批量校验也直接内聚到模块仓储，并把纯存在性校验提升成公开 `EnsureUserExists(...)` 供 service 组合点复用。`module/rbac` 把 role 详情装配、权限批量校验、角色删除级联、resource/permission 关系查询收进模块 repo，并继续把 role / permission / resource 的基础 list/load/create 查询直接内聚到模块仓储；最近又把 role detail、role->user、permission->role、resource->permission 这批组合视图改成模块内直查；上一轮再把 role delete cascade、mutable update、permission id 批量加载也进一步改成模块仓储自管；这一轮继续把“可写 role”校验收口成模块内 `loadWritableRole(...)`，同时把通用 `LoadPermission` / `LoadResource` 改成更贴业务语义的 `GetPermissionDetail(...)` / `GetResourceDetail(...)`。`module/project` 现已把 create-with-owner、delete cascade、detail/list 视图装配、mutable update、label reload 与按 key 移除标签收进自身 repo，这几轮继续把 project owner role 查询、project statistics 聚合、label 批量装配 / project label id 查找 / usage decrease 一并写实；这一轮再把内部 helper 命名继续往语义侧收紧成 `loadProjectRecord(...)` / `listProjectStatistics(...)`。`module/team` 也把 create-with-creator、detail 聚合、visible list、team project list、member add/remove/update role、team visibility 读取等操作收进 repo，并把 team project statistics 聚合也留在模块内；这一轮又把 team 加载进一步收成 `loadTeam(...)`，用于 detail / mutable update / ensure exists / visibility 读取，同时把 project statistics helper 明确成 `listTeamProjectStatistics(...)`。`module/execution` 现已接管 project list / global list / detail / labels / batch delete / detector result / granularity result / submit execution 全链路，新增自身 `repository.go` 并删除旧 `src/service/producer/execution.go`。由于 project 主路径此前早已由 `module/project` 承接，本轮也同步删除了已空心化的 `src/service/producer/project.go`；同时 `service/producer/label.go` 也已删除，初始化阶段改走 `module/label.CreateLabelCore`。`service/producer/relation.go`、`user.go`、`role.go`、`resource.go`、`auth_helpers.go`、`permission_helpers.go`、`datapack_archive.go` 同样已清掉，producer 侧残余重点进一步收敛到更少的共享逻辑；当前 `src/service/producer` 已无 Go 源文件残留。与此同时，旧 `src/client/loki.go` / `jaeger.go` / `redis_client.go` / `etcd_client.go` / `harbor_client.go` / `helm.go` / `client/k8s/*` 及 Helm 对应测试也已从 root `client` 包清走，真实实现统一并入 `src/infra/*`；上一轮已把 `src/infra/k8s/client.go` 删除，rest/client/dynamic/controller 的单例初始化直接吸回 `src/infra/k8s/gateway.go`；这一轮继续把 `service/consumer` / `service/initialization` 中的 `CurrentK8sController()` fallback 干掉，改成由 Fx 注入 `*k8sinfra.Controller`，同时 `service/common` 的 etcd fallback 改为回落到 `infra/etcd.GetGateway()` 单点入口，并进一步删掉 `service/consumer/deps.go` / `service/common/deps.go` 这类旧全局依赖注册文件。`service/consumer` 中剩余的 K8s / BuildKit / Helm 访问也继续改为直接走 `infra/*` 单点入口：新增 `buildkitinfra.GetGateway()`、`helminfra.GetGateway()`，`CurrentK8sGateway()` / `currentBuildkitGateway()` / `currentHelmGateway()` 已全部清掉；这轮又把 `app/startup.go` 删除，并进一步引入 `app.RegisterProducerInitialization`，把 producer 初始化从 `context.Background()` 改成走 Fx `OnStart` 生命周期上下文。随后又继续把 `interface/controller` / `interface/receiver` / `interface/worker` 的生命周期上下文改成从 Fx `OnStart` 派生，不再在模块注册期直接构造 `context.Background()`；再往下一轮又把 `service/consumer/task.go` / `trace.go` / `jvm_runtime_mutator.go` / `k8s_handler.go` 里残余 `context.Background()` 全部清成 consumer 内部 detached context helper。初始化侧原先带 callback 的 `registerHandlers(...)` 旧 helper 也已改成更窄职责的 `activateConfigScope(...)`，consumer / producer 各自显式注册所需 handlers，再统一激活 listener scope；这一轮再把 `GetConfigUpdateListener(...)` 单例 helper 从启动链收掉，改为在 producer / worker Fx `OnStart` 生命周期里显式创建 `ConfigUpdateListener` 后传给 initialization。`service/consumer` 的 Redis 直连也开始往更窄语义收：新增内部 `currentRedisGateway` / `currentRedisClient` / `publishRedisStreamEvent` / `publishTraceStreamEvent` / `loadCachedInjectionAlgorithms` helper，先把 trace/group stream 发布、detector cache 读取，以及 `monitor` / `rate_limiter` 对 Redis gateway 的获取收进更窄入口；随后又把 monitor 的上下文来源收回 worker lifecycle，并把 namespace SMembers/HGet/HSet/Pipeline 这批读取/写入改为统一走 consumer 内部 Redis helper 取 client，同时 `rate_limiter` 也不再自持 Redis client，而是统一经由 consumer Redis helper 获取连接；最近一轮再把 namespace key / exists / field read / seed / lock write 继续折成 `monitor` 内部更窄 helper，减少 monitor 主流程里散落的 Redis 原语；上一轮则继续把 rate limiter Redis 操作下沉成独立 `tokenBucketStore`，把 token acquire/release 的 Redis 细节与 limiter 配置/调度逻辑分开；这一轮再正式把 monitor 按同一路径拆出独立 `namespaceStore`，把 namespace key/list/exists/read/write/watch/status 这批 Redis 操作从 monitor 主流程里抽走；紧接着又继续深拆成 `namespaceCatalogStore` / `namespaceLockStore` / `namespaceStatusStore` 三个更窄 store，把锁读取/抢占/释放、namespace 注册、status 读写彻底从 `monitor.go` 抽开，并删除已空心化的 `src/service/consumer/namespace_store.go`。这一轮再把 startup / interface 链路里对 monitor 的旧包级获取收一批：`consumer.NewMonitor(...)` 作为 Fx provider 现在直接吃 `*redisinfra.Gateway` 并在内部自取 client，monitor 构造期不再向启动链暴露裸 `*redis.Client`，`initialization.InitializeConsumer(...)`、`RegisterConsumerHandlers(...)`、`interface/controller` 的 K8s callback 构造均改为显式注入 monitor，而不再自己碰 `GetMonitor()`；紧接着又继续把运行时执行主流程里的 monitor 单例拿掉，新增 `consumer.RuntimeDeps` 由 worker lifecycle 显式传入，`dispatchTask(...)` / `executeTaskWithRetry(...)` / `executeFaultInjection(...)` / `executeRestartPedestal(...)` 已不再自己碰 `GetMonitor()`。这一轮继续顺着同一主线把 rate limiter 也从进程级单例收成纯 Fx provider：`NewRestartPedestalRateLimiter(...)` / `NewBuildContainerRateLimiter(...)` / `NewAlgoExecutionRateLimiter(...)` 现在直接吃 `*redisinfra.Gateway` 构造 limiter，不再经过 `Get*RateLimiter()` / `sync.Once`；`executeBuildContainer(...)`、`executeAlgorithm(...)`、`executeRestartPedestal(...)` 与 K8s job 回调里的 algorithm token release 也都改为走显式传入 limiter，不再直接碰旧包级 getter。与此同时，`service/common/config_registry.go` / `config_listener.go` 把配置元数据读取继续收成 `service/common/config_store.go` 本地语义 store，不再穿过公共 `repository` 包；随后又把 producer/worker/controller/receiver 的启动执行体再收成显式可替换的 `ProducerInitializer` / `LifecycleRunner` 依赖，避免 lifecycle 本身直接抱一大串底层依赖，主路径更贴近 Fx；在此基础上，`src/app/startup_validate_test.go` 与 `src/app/startup_smoke_test.go` 现在已经补上 producer / consumer / both 三种 app option 的 Fx 图校验与 start/stop smoke（通过替换重型初始化依赖，验证 HTTP/worker/controller/receiver/producer lifecycle 编排本身可启动可停止）。这一轮继续顺着同一条线，把 `service/common/config_registry.go` 里的 `sync.Once` / `globalHandlersOnce` 再压掉，改成常驻 registry + 幂等注册逻辑，并补上 `config_registry_test.go` 锁住“全局 handlers 多次注册不重复”行为，进一步减少 config startup 主路径上的一次性单例状态；紧接着又继续把 listener / publish 周边的剩余全局依赖再收一层：`ConfigUpdateListener` 现在显式携带 `*gorm.DB`，不再在读取配置元数据和处理变更时回落到 `database.DB`；`RegisterGlobalHandlers(...)` / `RegisterConsumerHandlers(...)` 也开始显式接收 `ConfigPublisher`，`PublishWrapper(...)` 改成走传入 publisher，而不再自己碰 `redisinfra.GetGateway()`。对应地 producer / consumer 初始化与 worker lifecycle 现已把 Redis gateway / DB 一路显式传进 config listener 与 handler 注册主链。顺手也暴露并修复了 producer 模式此前缺少 `k8sinfra.Module`、导致 `chaosinfra.Module` 无法解析 `*rest.Config` 的问题。当前 producer / consumer / both 三种 app options 都已能通过 `go test ./app` 的图校验和启动链 smoke。这一轮继续把 `service/common` 热路径往显式 DB 收：`DBMetadataStore` 改成由 initialization 注入 `*gorm.DB` 创建，`container` / `dataset` / `task` 公共能力补上 `WithDB` 变体，`module/execution` / `module/injection` / `module/container` 的提交与 ref 解析主路径已改用模块 repo 自带 DB，不再回落到 `database.DB`。这一轮又继续把 consumer 运行态主链的 DB 依赖显式化：`consumer.RuntimeDeps` 开始携带 DB，worker/controller 生命周期分别把 DB 显式注入 task runtime 与 K8s handler，build/restart/algo reschedule、fault injection 落库、collect result 查询、K8s job/CRD 回调里的 execution/injection 状态推进与后续 task submit 也都改成优先走注入 DB，而不再默认抓全局 `database.DB`。这一轮继续把状态同步链也收进显式 DB：`taskStateUpdate` 新增 DB 上下文，`updateTaskState(...)` / `updateTraceState(...)` / trace optimistic lock 更新现在优先沿调用链携带的 DB 执行；K8s error context 也开始透传 handler 注入 DB，因此 consumer 主链里剩余 `database.DB` 基本只落在少量兼容 fallback 和 `service/common` 默认 wrapper。这一轮顺手再把 `module/evaluation` -> `service/analyzer` 这条链也切到显式 DB：evaluation service 改用 repo 持有 DB 调 analyzer 的 `WithDB` 版本，container/dataset ref 解析与 evaluation 持久化不再依赖 analyzer 内部全局 DB；同时 `module/injection.ExtractDatapacksWithDB(...)` 解析 dataset 时也已改走传入 DB 的 `MapRefsToDatasetVersionsWithDB(...)`。再往下一步，consumer 里 `collect_result` / `fault_injection` / `createExecution` 这类原先“nil 就回落全局 DB”的点也开始直接要求 runtime DB 存在，进一步缩小 fallback 面积。这一轮再继续把兼容层直接砍掉：`service/common` 里默认版 `MapRefsToContainerVersions` / `MapRefsToDatasetVersions` / `ListContainerVersionEnvVars` / `ListHelmConfigValues` / `SubmitTask` / `ProduceFaultInjectionTasks` 已删除，`service/analyzer` 里的默认版 evaluation 入口也删掉，只保留显式 `WithDB` 路径；同时 `consumer/task.go` / `trace.go` / `k8s_handler.go` 里的 DB fallback 也改成显式报错，不再默默回落全局 `database.DB`。紧接着又把 `module/system` 里最后一处直接碰 `database.DB` 的 health check 改成走 `repo.DB()`；目前 `module/*`、`service/common`、`service/consumer`、`service/analyzer` 这批主线包内已无 `database.DB` 残留。顺手又把 repository 层里残留的统计/搜索/资源/注入查询改成统一吃显式 `db` 参数，`repository/task.go` 的 `ListTasksByTimeRange(...)` 也不再偷偷回落全局 DB；现在全仓库只剩 `src/infra/db/module.go` 这一处集中持有 `database.DB`，作为 Fx 提供与关闭数据库连接的基础设施边界。最近两轮又继续把 consumer 外部依赖收窄到 Fx 注入：`interface/worker` 把 `*k8sinfra.Gateway` / `*buildkitinfra.Gateway` / `*helminfra.Gateway` / `*consumer.FaultBatchManager` / `*redisinfra.Gateway` 显式塞进 `consumer.RuntimeDeps`，`build container` / `build datapack` / `algo execution` / `restart pedestal` / `collect result` / task retry / trace state update / K8s callback 已不再直接碰 `GetGateway()` 与 fault batch `sync.Once` 单例；`interface/controller` 同步把 K8s gateway、Redis gateway 和 batch manager 显式交给 `consumer.NewHandler(...)`；`service/logreceiver` 也开始由 `interface/receiver` 注入 Redis publisher，OTLP receiver 不再自己抓 `redisinfra.GetGateway()`。这一轮又继续把 HTTP 链路里的 middleware 全局态收掉：`src/middleware/deps.go` 现在提供 `middleware.Service` 与 `InjectService(...)`，`src/router/router.go` 在根路由中显式注入 middleware service，`src/middleware/permission.go` / `audit.go` 改为按请求从 Gin context 读取 checker/logger，不再持有 `currentPermissionChecker` / `currentAuditLogger` 这类包级默认服务；`src/interface/http/module.go` 也不再用 `fx.Invoke(middleware.RegisterDeps)` 做全局注册。最近这一轮再把 startup 初始化链里的隐藏 fatal 收掉：`newConfigDataWithDB(...)`、`activateConfigScope(...)`、`InitializeProducer(...)`、`InitializeConsumer(...)` 全部改成显式返回 `error`，producer/worker 的 Fx `OnStart` 现在会把初始化失败直接上抛，而不再在 helper 内部 `logrus.Fatalf(...)` 提前退出进程。紧接着这一轮又继续把 consumer startup 链里的 Redis 裸 client 收口到 gateway：worker 初始化改为走 `RedisGateway.InitConcurrencyLock(...)`，`monitor` / `rate limiter` provider 也改成只依赖 `*redisinfra.Gateway`。再下一轮又把模块侧剩余 Redis 全局入口清掉：`module/group` / `module/notification` / `module/trace` / `module/injection` / `module/systemmetric` / `module/system` 现在都改为通过构造注入 `*redisinfra.Gateway`，trace/group/notification stream 读取、injection algorithm cache、system config response subscribe、system metric Redis 查询不再直接碰 `redisinfra.GetGateway()`。这一轮继续把任务队列 helper 也收回 gateway：`infra/redis/task_queue.go` 里的 submit/get/reschedule/dead-letter/queue index/concurrency lock/list/remove 操作全部改成 `Gateway` 方法，`service/common.SubmitTaskWithDB(...)`、`service/consumer` 调度与取消链路、`module/systemmetric` 排队任务查询都已改走显式 Redis gateway。紧接着又把 `infra/redis` / `infra/etcd` / `infra/buildkit` / `infra/helm` / `infra/k8s` 里已经没有调用方的 `GetGateway()` 单例 fallback 全部删除，主线现在只剩少量 lifecycle/startup 组织层 wrapper 需要再压。当前 `src/service/consumer` / `src/middleware` 里残余重点已从“全局 gateway fallback / 全局 default service”收缩到更少的流程组织 helper 与 initialization 邻近收尾。
- [x] 删除旧包级 repository wrapper
  - 已移除 `repository/task.go` 中 Redis 队列职责与 `repository/token.go` 黑名单兼容层。
- [x] 删除全局 default service
- [x] 清理未使用 imports

验收：

- [x] 检查源码与文档中旧 DI 文案残留
- [x] `cd src && go test ./...`
  - 已补齐 `src/cmd/aegisctl/output/output.go`，修复 `cmd/aegisctl` 缺失输出包导致的全量测试阻塞；同时把 `infra/k8s` 的集成 Job 用例改为 `RUN_K8S_INTEGRATION=1` 显式开启，避免默认 `go test ./...` 卡在真实集群状态。

## 19. 最终验收

功能验收：

- [x] producer 模式可启动
- [x] consumer 模式可启动
- [x] both 模式可启动
- [x] login / register / refresh 正常
- [x] Project CRUD 正常
- [x] Injection 提交流程正常
- [x] Execution 提交流程正常
- [x] Task 状态和日志正常
- [x] Admin 用户管理正常
- [x] SDK 生成正常
- [x] Swagger 文档正常

架构验收：

- [x] `main.go` 只负责 mode 和 Fx 启动
- [x] DB / Redis / HTTP / worker / receiver / controller 都有 lifecycle
- [x] handler 不直接 import `database`
- [x] handler 不直接 import repository implementation
- [x] service 不直接使用全局 `database.DB`
- [x] repository 不直接访问 Redis / K8s / Loki / Etcd
- [x] middleware 不直接依赖具体 producer package
- [x] Public / SDK / Portal / Admin / System 路由分离
- [x] 业务模块通过 `Module` 暴露

说明：

- `src/app/startup_validate_test.go` 已覆盖 producer / consumer / both 三种 Fx 图校验。
- `src/app/startup_smoke_test.go` 已覆盖 producer / consumer / both 三种 start/stop smoke，并继续补上 consumer lifecycle 集成冒烟、both 模式的 HTTP + lifecycle 联合冒烟。
- `src/router/router_test.go` 已锁定 `Public / SDK / Portal / Admin / System` 关键路由前缀分离。
- `src/app/http_modules.go` 统一通过各业务模块的 `Module` 暴露 HTTP 能力并聚合进 producer app。
- 启动链补扫后，`src/app` / `src/service/initialization` / `src/interface` / `src/middleware` 生产代码里已无旧 `service/producer` / `handlers/system` / `client/*` 引用，也无残余 `context.Background()` / `GetGateway()` 启动期直拿全局对象。
- 本轮顺手补齐 `src/cmd/aegisctl/output/output.go`，把 CLI 的 JSON / table / info / error 输出能力收回本地包，`cmd/aegisctl` 不再因缺失输出层而阻塞仓库全量构建。
- 本轮继续把启动链残余接口壳压掉：`src/app/producer_init.go` 与 `src/interface/{worker,controller,receiver}/module.go` 已从 `ProducerInitializer` / `LifecycleRunner` 接口切到可直接替换的具体 lifecycle struct，smoke test 也同步改成按具体类型替换，启动编排层又薄了一轮。
- 本轮再顺手清了一批模块仓储纯转发：`src/module/system/repository.go` 的 audit/config/history 查询与写入已直接写实到模块仓储；`src/module/injection/repository.go` 的 groundtruth 更新也不再空转调公共 `repository`；同时 `RegisterProducerInitialization(...)` 已不再额外挂 `CommonResources` 形参。
- 本轮继续把 `src/module/execution/repository.go` 写实：project resolve / execution list/detail/result / execution labels / result save / batch delete / duration update / labels attach 这一整段已直接落回模块仓储，不再散着空转调 `repository/execution.go`、`repository/detector.go`、`repository/granularity.go`、`repository/label.go`。
- 本轮再把 `src/module/container/repository.go` 与 `src/module/dataset/repository.go` 两块成片稳定 CRUD 仓储写实：role resolve、container/dataset CRUD、version CRUD、label relation、helm/env/parameter config、dataset version injection 关系等都已回收到模块仓储；目前这两块只保留 dataset search 对共享 query builder 的调用。
- 本轮继续把剩余一批小模块仓储空转发彻底写回模块：`src/module/{sdk,chaossystem,trace,group,evaluation,task,auth,label}/repository.go` 里的 list/detail/create/update/delete / relation-count / metadata / user-role 查询等都已直接落回模块仓储；`src/module/label/core.go` 也不再直连公共 `repository/label.go`。当前模块侧保留的共享 `repository.ExecuteSearch(...)` 只剩 injection / dataset 两处，作为通用 query builder 基础设施继续复用，不再是无意义兼容层。
- 本轮继续把 search 这条尾巴也收掉：`src/module/dataset/repository.go` 与 `src/module/injection/repository.go` 已不再调用公共 `repository.ExecuteSearch(...)`，而是直接使用 `repository/query_builder.go` 里的通用 builder 组装查询；公共 `ExecuteSearch` 兼容入口已删除，模块侧只保留对底层 query builder 基础设施的显式使用。
- 本轮继续顺手压掉一批 raw client / helper 暴露面：`src/module/auth/token_store.go` 与 `src/module/task/queue_store.go` 已改成依赖 `infra/redis.Gateway`，`src/infra/redis/gateway.go` 补齐 `Set/Subscribe` 语义方法，`src/app/common.go` 的 Fx 公共资源探针也改成依赖 `Redis/Etcd Gateway` 而不是裸 client；同时 `src/module/trace/service.go` / `src/module/trace/stream.go` 把 trace stream processor/read 的包级 helper 收回 service，`src/module/group/service.go` 的 group stream processor 初始化也去掉了无意义的 context 包装。
- 本轮再把剩余 Fx / consumer 暴露面继续压一轮：`src/infra/{redis,etcd}/module.go` 的 `ProvideClient` 已删除，Fx graph 不再向外暴露裸 Redis/Etcd client；`src/service/consumer/{namespace_catalog_store,namespace_lock_store,namespace_status_store,rate_limiter_store}.go` 也都改成直接持有 `infra/redis.Gateway`，`src/service/consumer/{monitor,rate_limiter}.go` 不再在上层显式拿 `gateway.Client()`；另外 `src/app/common.go` 这个仅用于依赖探测的空文件已删除，`src/app/app.go` 不再保留无意义的 `RequireCommonResources` invoke。
- 本轮继续把模块内部 API 面收紧一层：`src/module/{project,team,dataset,rbac,auth,user,execution,injection,container,system}/repository.go` 的 `WithDB` 已统一缩成包内 `withDB`；`src/module/{execution,injection}/repository.go` 的 `EnsureProjectExists`、`src/module/user/repository.go` 的 `EnsureUserExists` 也已缩成包内 helper；`src/module/{execution,injection,container,system,evaluation}` 里原先为 service 暴露的 `DB()` 访问器已删除，service 直接在包内使用 repository 持有的 db。
- 本轮顺手再收一批 `context.Background()` 残点：`src/module/task/log_service.go`、`src/module/injection/handler.go`、`src/module/injection/service.go`、`src/module/system/service.go` 已改成沿调用链传递 request/service context；`src/module/systemmetric/collector.go` 也改成 lifecycle 管理的 collector context，在 `OnStop` 时显式 cancel。
- 本轮继续压掉最后一批显眼的 helper 暴露：`src/module/systemmetric/service.go` 已改成直接使用 `infra/redis.Gateway` 暴露的 `SetMembers / HashGetAll / ZRangeByScore / ZAdd / ZRemRangeByScore` 语义方法，`src/module/system/service.go` 的 Redis 健康检查也切到 `gateway.Ping`；同时 `src/infra/redis/gateway.go` 又补齐这一批语义 API，模块/系统层不再直接拼裸 Redis 命令。当前生产代码里已无 `context.Background()` 残点，剩余 `redisGateway.Client()` 仅收敛在 `service/consumer/*store.go` 这一层 Redis 原语适配代码中。
- 本轮继续把 consumer 最后一层 Redis 原语适配再往 infra 收：`src/service/consumer/{namespace_catalog_store,namespace_status_store,namespace_lock_store,rate_limiter_store}.go` 里残余的 `gateway.Client()` 已全部清掉，分别改成走 `infra/redis/gateway.go` 新增的 `Exists / HashGet / HashSet / SeedNamespaceState / SetRemove / RunScript / Watch` 等语义方法；当前生产代码中 `service/consumer` / `module` / `app` / `interface` 已无直接 `gateway.Client()` 调用，裸 Redis client 已彻底退回 `infra/redis` 内部实现。
- 本轮继续把 infra 边界再收紧一层：`src/infra/{redis,etcd}/gateway.go` 的公开 `Client()` 暴露面已删除，连接初始化/关闭统一收进私有 `clientOrInit()/close()`；`src/infra/k8s/{job,controller}.go` 中原先仅供 gateway 转调的 `CreateJob / GetJobPodLogs / GetVolumeMountConfigMap / NewController` 也已缩成包内私有实现，`Gateway` 成为对外唯一主入口。
- 本轮继续清 initialization 残余全局态：`src/service/initialization/{producer,consumer}.go` 不再持有包级 `producerData / consumerData / resourceIDMap`，初始化配置状态改成局部装配后沿调用链使用；`InitializeSystems(...)` 也已改为显式返回 `error`，producer 启动链不再吞掉系统注册失败。
- 本轮继续压一轮启动壳与模块内部 helper：`src/app/producer_init.go`、`src/interface/{worker,controller,receiver}/module.go`、`src/interface/http/{module,server}.go` 中的 lifecycle/register helper 已全部收成包内私有；`src/module/project/repository.go` 删掉 `loadProjectLabelView(...)`，`src/module/rbac/repository.go` 删掉 `loadWritableRole(...)` 并把 system-role 校验内聚回具体语义方法，`src/module/user/repository.go` 则把 role/container/dataset/project 四组原子 load helper 收成单个 `ensureActiveRecordExists(...)`，`src/module/team/repository.go` 把 role 校验压成 `ensureRoleExists(...)`，`src/module/injection/repository.go` 的 `ensureProjectExists(...)` 也已删掉并改为 service 包内直接使用 repo DB 校验项目存在性。
- 本轮最后再把 execution / user 邻近模块尾巴收掉：`src/module/execution/repository.go` 的 `ensureProjectExists(...)` 已删除，project 存在性校验直接回到 `service.go` 包内用 repo DB 执行；`src/module/user/repository.go` 的 `ensureUserExists(...)` 也已删掉，统一并入已有 `ensureActiveRecordExists(...)`，避免“同一语义多套 helper”继续扩散。
- 最终抛光轮再顺手做了一轮“模块 repo API 面收紧”：`src/module/execution/repository.go` 的 `GetProjectByName / List*View / GetExecution* / ListAvailableExecutionLabels / ListExecutionLabelIDsByKeys`，以及 `src/module/project/repository.go` 的 CRUD/label 管理主方法、`src/module/team/repository.go` 的 team CRUD / list / membership 读取主方法，均已统一缩成包内私有实现，只保留 service 真正需要的模块边界；模块内部语义仍在，但对外可见面进一步变薄。
- 最终抛光轮又继续做了两件事：一是把 `src/module/user/repository.go` 与 `src/module/rbac/repository.go` 里仅供 service 使用的 repo 主方法再统一缩成包内私有，模块命名/API 面进一步一致；二是在 `src/app/startup_smoke_test.go` 补上 producer HTTP 集成冒烟，真实启动 Fx producer app 后校验 `/docs/doc.json` 可访问、`/system/configs/abc` 会经过真实路由与鉴权链返回 `401`，把“能启动”进一步提升到“能接住实际 HTTP 主路径”。
- 本轮已补 `src/module/auth/service_test.go`，覆盖 `register / login / refresh` 成功路径，并顺手修正 `module/auth` / `module/user` 创建用户时密码重复 hash 的问题，避免注册后登录链路天然失效。
- 本轮已扩充 `src/module/project/service_test.go`，覆盖 `create / get detail / list / update / delete` 主路径；`Project CRUD` 现已具备模块级成功路径保护，剩余是更贴近真实依赖的运行态验收。
- 本轮已扩充 `src/module/execution/service_test.go`，覆盖标签列表、列表过滤、detector / granularity 结果上传成功路径；`execution result` 主路径已有模块级保护。
- 本轮新增 `src/module/task/service_test.go`，覆盖 task 列表成功路径与 Loki 历史日志读取；`Task 状态 / 日志` 这条线已具备模块级成功路径保护。
- 本轮继续扩充 `src/module/execution/service_test.go` 与 `src/module/injection/service_test.go`，分别补上 `SubmitAlgorithmExecution` 和 `SubmitDatapackBuilding` 成功路径；`Injection / Execution 提交` 主路径已具备模块级提交保护。
- 为避免真实 Redis 依赖阻塞主线验收，本轮新增 `src/testutil/redisstub.go` 作为极小测试桩，仅覆盖任务提交用到的 Redis 命令，供模块级 submit 测试使用。
- 本轮已补 `src/module/user/service_test.go` 的 create / detail / delete 成功路径，`Admin 用户管理` 主路径现已具备模块级成功路径保护。
- 本轮新增 `src/module/sdk/service_test.go`，覆盖 SDK evaluation / experiment / dataset sample 主路径；`SDK` 主路径已具备模块级成功路径保护。
- 本轮新增 `src/docs/docs_test.go`，校验 `openapi2` / `openapi3` / `converted/sdk.json` 三类文档产物存在且包含核心接口路径；同时 `src/router/router.go` 已显式注册 `aegis/docs/openapi2`，`src/router/router_test.go` 继续锁定 `/docs/doc.json` 可直接返回 Swagger 文档。
- 本轮已完成 `cd src && go test ./...`；当前默认全量测试口径已打通，K8s Job 的真实集群冒烟改为按需用 `RUN_K8S_INTEGRATION=1 go test ./infra/k8s` 单独执行。
- 本轮继续把 `src/infra/k8s/k8s_test.go` 升级成更明确的真实集群验收入口：先做 `Gateway.CheckHealth(...)` 预检，再跑 job create/get/wait/logs/delete 全链路；同时支持 `RUN_K8S_INTEGRATION_NAMESPACE`、`RUN_K8S_INTEGRATION_IMAGE`、`RUN_K8S_INTEGRATION_KEEP_JOB` 三个可选环境变量，便于回填真实环境验收。
- 本轮继续把 `src/infra/k8s/gateway.go` 收成 K8s job 生命周期主入口，补上 `GetJob / WaitForJobCompletion / DeleteJob` 这组 gateway 语义方法；`WaitForJobCompletion(...)` 也改成尊重 context 取消，并在 Job 失败条件出现时尽早返回。
- 本轮继续扩充 `src/app/startup_smoke_test.go`：新增 `TestConsumerOptionsLifecycleIntegrationSmoke`，锁定 worker/controller/receiver 的真实 Fx 启停；新增 `TestBothOptionsHTTPAndLifecycleIntegrationSmoke`，在 both 模式下同时校验 producer 初始化、consumer 生命周期和 `/docs/doc.json` / `/system/configs/:id` 这组真实 HTTP 主路径。

当前说明：

- 第 19 节功能验收已按“模块级成功路径 + 路由/文档产物校验 + app 启动 smoke”口径全部补齐。
- 目前若继续做，已基本进入纯抛光阶段：更激进的命名统一、个别 repo/helper 再折叠、以及更贴近真实外部依赖的集成验收，都不再阻塞 Fx 主线收口。
- 本轮已再次重跑三组主路径验证命令：`go test ./module/auth ./module/project ./module/execution ./module/injection ./module/task ./module/user ./module/sdk ./router ./docs ./app`、`go test ./app ./service/consumer ./service/logreceiver ./interface/controller ./interface/receiver ./interface/worker`、`go test ./app ./router ./interface/http ./middleware`，当前均通过。
- 本轮再补跑 `go test ./...`，当前也已通过。

## 20. 仓库级收尾检查清单

- [x] 默认回归：`cd src && go test ./...`
- [x] Producer Fx 图校验与 HTTP 主路径：`cd src && go test ./app -run 'TestProducerOptionsValidate|TestProducerOptionsStartStopSmoke|TestProducerOptionsHTTPIntegrationSmoke'`
- [x] Consumer / Both 生命周期集成冒烟：`cd src && go test ./app -run 'TestConsumerOptions|TestBothOptions'`
- [x] 路由 / 文档主路径：`cd src && go test ./router ./docs ./interface/http`
- [x] 真实 K8s 集群验收：`cd src && RUN_K8S_INTEGRATION=1 go test ./infra/k8s -run TestK8sGatewayJobLifecycleIntegration`
- [x] 可选真实环境参数已提供：`RUN_K8S_INTEGRATION_NAMESPACE=<ns>`
- [x] 可选真实环境参数已提供：`RUN_K8S_INTEGRATION_IMAGE=<image>`
- [x] 可选真实环境参数已提供：`RUN_K8S_INTEGRATION_KEEP_JOB=1`
- [x] producer / consumer / both 主启动链已无旧 `service/producer` / `handlers/system` / `client/*` 运行态依赖
- [x] K8s / Redis / Etcd / Harbor / Helm / BuildKit 等 infra 主入口已统一收口到 `src/infra/*`
- [x] 仓库级残余兼容面补扫通过
  - 已用 `rg` 对 `service/producer`、`handlers/system`、`GetGateway()`、`CurrentK8s*`、`database.DB`、启动链 `context.Background()` 等模式做补扫；`src/app` / `src/interface` / `src/service` / `src/module` / `src/router` / `src/middleware` 生产代码内未发现这批旧运行态依赖残留。

## 当前建议下一步

从这里开始：

1. 如需复验真实集群，执行 `cd src && RUN_K8S_INTEGRATION=1 go test ./infra/k8s -run TestK8sGatewayJobLifecycleIntegration`
2. 常规回归继续跑 `cd src && go test ./...`
3. 如需继续推进，优先进入第 17 节 SDK 标记治理；其余已基本属于文档/命名/测试抛光

## 21. 微服务拆分主线

目标：在当前 Fx + module + infra 主线已收口的基础上，把运行时进一步演进为“外部 HTTP、内部 gRPC、执行异步队列”的明确微服务架构，而不是继续在单体模式下扩张。

- [x] 微服务设计/治理主文档已收口到 `docs/report-index.md`
- [x] 在 `src/app/` 建立第一批服务边界分组：`gateway / runtime / iam / resource / orchestrator / system`
- [x] 新增第一批可运行服务入口：`src/cmd/api-gateway`、`src/cmd/runtime-worker-service`、`src/cmd/iam-service`
- [x] Runtime Worker Service：补 `runtime.proto` 与 gRPC control-plane（`Ping / GetRuntimeStatus / GetQueueStatus / GetLimiterStatus`）
- [x] IAM Service：补 `iam.proto` 与 token verify / permission check / API key exchange gRPC
- [x] Orchestrator Service：补 `orchestrator.proto`、`src/interface/grpcorchestrator/*` 与 `src/cmd/orchestrator-service`
- [x] Orchestrator Service：首批 submit / cancel RPC 已收口（`Ping / SubmitExecution / SubmitFaultInjection / SubmitDatapackBuilding / CancelTask`）
- [x] Gateway -> Orchestrator：execution / injection submit 主路径已支持通过 `clients.orchestrator.target` 或 `orchestrator.grpc.target` 切到内部 gRPC
- [x] Orchestrator Service：workflow state / task query / dead-letter / retry 首批控制面已收口
  - 当前已新增 `GetTask / ListTasks / GetTrace / ListTraces / ListDeadLetterTasks / RetryTask` 六个内部 RPC，并继续保留 Redis 作为执行异步主通道不变。
- [x] Orchestrator Service：execution owner 的 runtime/evaluation mutation/query facade 已落地
  - 当前已新增 `CreateExecution / CreateInjection / UpdateExecutionState / UpdateInjectionState / UpdateInjectionTimestamps / GetExecution / ListEvaluationExecutionsByDatapack / ListEvaluationExecutionsByDataset` 八个内部 RPC；runtime 状态推进与 evaluation 执行结果查询在配置 `clients.orchestrator.target` 或 `orchestrator.grpc.target` 后已优先走 owner facade。
- [x] Runtime Worker：已切掉对 orchestrator owner 的共享 repository 直读/直写
  - 当前 `src/service/consumer/*` 已不再直接 import `src/repository/*`；执行创建、fault injection 创建、K8s 回调状态推进、结果收集在未配置 orchestrator gRPC 时也只回退到本地 `executionmodule.Service` / `injectionmodule.Service` owner 实现。
- [x] `service/common`：首批 container / label 共享 helper 已回收到 owner 模块
  - 当前 `src/service/common/container.go` 与 `src/service/common/label.go` 已删除；container version/parameter 解析与 label upsert 已分别收回 `src/module/container/*`、`src/module/label/*`，`module/{execution,injection,evaluation,project,container,dataset}` 与 consumer K8s 回调已改走 owner 模块实现。
- [x] `service/common`：dataset version 共享 helper 已回收到 owner 模块
  - 当前 `src/service/common/dataset.go` 已删除；dataset version 解析已收回 `src/module/dataset/resolve.go`，同时 `src/module/dataset/api_types.go` 也已去掉对 `module/injection` 的响应类型依赖，避免再次形成模块循环。
- [x] Resource / System Service：已补独立启动入口 `src/cmd/resource-service`、`src/cmd/system-service`
- [x] Resource Service：首批资源/评估查询 gRPC 已落地（`Ping / ListProjects / GetProject / ListContainers / GetContainer / ListDatasets / GetDataset / ListDatapackEvaluationResults / ListDatasetEvaluationResults / ListEvaluations / GetEvaluation / DeleteEvaluation`）
- [x] System Service：首批系统运维 gRPC 已落地（`Ping / GetHealth / GetMetrics / GetSystemInfo / ListConfigs / GetConfig / ListAuditLogs / GetAuditLog / ListNamespaceLocks / ListQueuedTasks / GetSystemMetrics / GetSystemMetricsHistory`）
- [x] Runtime -> System：namespace locks / queued tasks 首批运行态查询已从 Redis 直读收口到 runtime gRPC
  - 当前 `src/proto/runtime/v1/runtime.proto` 已新增 `GetNamespaceLocks / GetQueuedTasks`，`src/module/system/service.go` 在配置 `clients.runtime.target` 或 `runtime_worker.grpc.target` 后会优先走 `src/internalclient/runtimeclient/*`，未配置时保留本地回退。
- [x] Gateway -> Resource：project / container / dataset / evaluation 主路径已支持通过 `clients.resource.target` 或 `resource.grpc.target` 切到内部 gRPC
- [x] Gateway -> System：`system` / `systemmetric` 首批读路径已支持通过 `clients.system.target` 或 `system.grpc.target` 切到内部 gRPC
- [x] `app.CommonOptions()` 已开始按服务边界拆细
  - 当前已落地 `BaseOptions / ObserveOptions / DataOptions / CoordinationOptions / BuildInfraOptions`，`iam/resource/orchestrator/system` 已切到更窄装配口径。
- [x] 第一轮跨 owner 共享 repository / DB 直查补扫已完成
  - 当前 `src/app` / `src/interface` / `src/internalclient` 侧已无直查 DB；残余主要收敛在 owner 模块内部 repository 和少量本地 fallback 继续压缩。
- [x] Resource / System Service：资源元数据与运维控制面首轮独立服务边界已落地
  - 当前 `resource-service` 已承接 project / container / dataset / evaluation / label / chaos-system 资源元数据主路径，`system-service` 已承接 health / config / audit / monitor / systemmetric 运维控制面主路径；剩余更细粒度拆分进入后续非阻塞治理阶段，不再阻塞当前主线收口。

说明：

- 当前第一轮“文档 + 骨架 + 可运行入口 + 核心 RPC 主路径”已完成；后续如继续演进，重点转向 ownership 深清与发布治理，而不是主线入口缺失。
- 当前 `api-gateway` 语义上对应既有 producer HTTP 栈，`runtime-worker-service` 语义上对应既有 consumer 栈；其余服务的核心内部 RPC 与独立启动入口已落地，后续再按边界细化 owner 职责。
- 本轮已落地 `src/proto/runtime/v1/runtime.proto`、`src/interface/grpcruntime/*` 与 queue/limiter/runtime snapshot 聚合能力，并把 gRPC lifecycle 接入 `ConsumerOptions` / `BothOptions`；默认监听 `:9094`，可通过 `runtime_worker.grpc.addr` 覆盖。
- 本轮继续扩展 `runtime-worker-service` control-plane：当前额外提供 `GetNamespaceLocks / GetQueuedTasks`，用于承接 runtime Redis 运行态对内查询。
- 本轮继续落地 `src/proto/iam/v1/iam.proto`、`src/interface/grpciam/*` 与 `src/cmd/iam-service`，当前 IAM 内部 RPC 已覆盖鉴权、API key、team、user、rbac 五组主路径：除 `VerifyToken / CheckPermission / ExchangeAPIKeyToken` 与 team membership 判定外，也已补齐 `Login / Register / RefreshToken / Logout / ChangePassword / GetProfile / API key CRUD`、`Create/Get/List/Update/Delete user`、user role/permission/resource 绑定、`Create/Get/List/Update/Delete role`、role-permission 绑定以及 permission/resource 查询；默认监听 `:9091`，可通过 `iam.grpc.addr` 覆盖。
- 本轮继续补上 `src/proto/orchestrator/v1/orchestrator.proto`、`src/interface/grpcorchestrator/*` 与 `src/cmd/orchestrator-service`，当前 Orchestrator 内部 RPC 已提供 `Ping / SubmitExecution / SubmitFaultInjection / SubmitDatapackBuilding / CancelTask` 五个入口；默认监听 `:9092`，可通过 `orchestrator.grpc.addr` 覆盖。
- 本轮继续扩展 `orchestrator-service` 控制面：当前额外已提供 `GetTask / ListTasks / GetTrace / ListTraces / ListDeadLetterTasks / RetryTask` 六个入口，用于 workflow state 查询、dead-letter 补偿与手动 retry；同时执行/消费异步仍保持 Redis queue/event 主链不变。
- 本轮继续扩展 `orchestrator-service` owner facade：当前又额外提供 `CreateExecution / CreateInjection / UpdateExecutionState / UpdateInjectionState / UpdateInjectionTimestamps / GetExecution / ListEvaluationExecutionsByDatapack / ListEvaluationExecutionsByDataset` 八个入口，分别承接 runtime 状态回写与 evaluation 执行结果查询。
- Gateway 侧已补 `src/internalclient/orchestratorclient/*`，并通过 `src/app/gateway/options.go` 把 `execution` / `injection` handler 使用的 submit 服务装饰为 remote-aware；配置 `clients.orchestrator.target` 或 `orchestrator.grpc.target` 后，`SubmitAlgorithmExecution / SubmitFaultInjection / SubmitDatapackBuilding` 会优先走内部 gRPC，未配置时继续回退本地实现。
- `src/app/consumer.go` 当前也已补 execution/injection owner 模块，使 `runtime-worker-service` 在未配置 orchestrator gRPC 时改为回退到本地 owner service，而不再直接碰共享 repository；`interface/worker` / `interface/controller` 会把这两个 owner service 显式注入 consumer runtime deps 与 K8s handler。
- 同时已把 `src/cmd/resource-service` 与 `src/cmd/system-service` 补齐，后续可以直接在对应服务边界上继续补 `resource.proto` / `system.proto` 和对内 gRPC。
- 本轮继续补上 `src/proto/resource/v1/resource.proto`、`src/interface/grpcresource/*` 与 `src/app/resource/options.go` 接线，当前 Resource 内部 RPC 已提供 `Ping / ListProjects / GetProject / ListContainers / GetContainer / ListDatasets / GetDataset / ListDatapackEvaluationResults / ListDatasetEvaluationResults / ListEvaluations / GetEvaluation / DeleteEvaluation` 十二个入口；默认监听 `:9093`，可通过 `resource.grpc.addr` 覆盖。
- 本轮继续补上 `src/proto/system/v1/system.proto`、`src/interface/grpcsystem/*` 与 `src/app/system/options.go` 接线，当前 System 内部 RPC 已提供 `Ping / GetHealth / GetMetrics / GetSystemInfo / ListConfigs / GetConfig / ListAuditLogs / GetAuditLog / ListNamespaceLocks / ListQueuedTasks / GetSystemMetrics / GetSystemMetricsHistory` 十二个入口；默认监听 `:9095`，可通过 `system.grpc.addr` 覆盖。
- `src/app/system/options.go` 现已补齐 `k8sinfra.Module` 与 `runtimeclient.Module`，`module/system.Service` 对 `ListNamespaceLocks / ListQueuedTasks` 已优先走 runtime gRPC，把 system/runtime 的首批运行态交互从直接 Redis 读取改成内部 client 边界。
- Gateway 侧已继续补 `src/internalclient/resourceclient/*`，并通过 `src/app/gateway/options.go` 把 `project` / `container` / `dataset` handler 使用的稳定 list/detail 读服务装饰为 remote-aware；配置 `clients.resource.target` 或 `resource.grpc.target` 后，`ListProjects / GetProjectDetail / ListContainers / GetContainer / ListDatasets / GetDataset` 会优先走内部 gRPC，未配置时继续回退本地实现。
- Gateway 侧现已补 `src/internalclient/systemclient/*`，并通过 `src/app/gateway/options.go` 把 `system` / `systemmetric` handler 使用的查询服务装饰为 remote-aware；配置 `clients.system.target` 或 `system.grpc.target` 后，`/system/*` 与 `/api/v2/system/metrics*` 会优先走内部 gRPC，未配置时继续回退本地实现。
- 这一轮继续把 dedicated service 入口往 remote-first 收紧：`src/app/gateway/options.go` 现在会在启动时显式校验 `iam / orchestrator / resource / system` 四类 internal client target，`src/app/runtime/options.go` 会校验 orchestrator target，`src/app/system/options.go` 会校验 runtime target，避免 `api-gateway` / `runtime-worker-service` / `system-service` 这类独立服务入口继续静默回退本地 owner 实现。
- 这一轮又继续把 standalone runtime 边界再压一层：`src/app/runtime/options.go` 已不再直接复用整套 `ConsumerOptions()`，而是去掉 `executionmodule.Module` / `injectionmodule.Module`，`src/interface/worker/module.go` 与 `src/interface/controller/module.go` 中对应 owner service 依赖已改成可选，避免 `runtime-worker-service` 独立入口继续显式装配本地 execution/injection owner。
- 这一轮再把独立服务入口的验收补到位：新增 `src/app/service_entrypoints_test.go`，已覆盖 `api-gateway` 的真实 HTTP 冒烟，以及 `runtime-worker-service / resource-service / system-service` 的真实 gRPC 冒烟；`api-gateway` / `runtime-worker-service` 的独立启动与 runtime control-plane 可用性现在都有自动化保护。
- 这一轮继续把“启动命令 / 配置 / 本地编排”说明收口到 `docs/report-index.md`，并新增 `docker-compose.microservices.yaml` 作为与现有 `docker-compose.yaml` 叠加的多服务本地 compose 骨架；`src/config.dev.toml` 与 `config.dev.toml` 也已补齐 `clients.*.target` 及 `iam/resource/orchestrator/runtime_worker/system` 的 gRPC 默认端口，方便直接按拆分模式起服务。
- 这一轮继续把“镜像入口 / probe 规范 / K8s skeleton”说明也并入 `docs/report-index.md`：`src/main.go` 已增加 `api-gateway / iam-service / resource-service / orchestrator-service / runtime-worker-service / system-service` 六个新子命令，现有镜像可直接用同一二进制起拆分服务；同时 `manifests/microservices/aegislab-microservices.yaml` 已统一 gateway 的 HTTP `/system/health` probe 与五个 gRPC 服务的 health probe，并补上第一版多服务 Deployment/Service 骨架。
- 最终仓库级收尾轮已把补扫与收尾结论并入 `docs/report-index.md`：`src` 生产代码里 `service/producer` / `handlers/system` / `database.DB` / `GetGateway()` / `redisinfra.GetGateway()` 这批旧兼容模式已为零命中，`context.Background()` 残留也只在测试中；当前主线可视为完成，剩余主要转入兼容入口 owner 组合继续压缩与少量跨服务 DB 深清。
- 同一轮里也已把错误码、request-id、观测标签、internal proto、配置命名和 owner 约束统一写实，并收口到 `docs/report-index.md`；治理项已不再是“规范空缺”，当前更多是发布执行层持续收口。
- 继续执行层收口后，HTTP/gRPC 的 request-id 主路径也已正式落地：`src/router/router.go` 现已统一挂 `X-Request-Id` middleware，`src/internalclient/*` 已统一透传 `x-request-id` metadata，治理规范不再只停留在文档。
- 同一批收口里，`src/interface/grpc*` 也已统一在 server 入口提取/补齐 request-id；另外 `module/dataset` / `module/injection` 对旧共享 `repository.NewSearchQueryBuilder` 的依赖已清掉，通用搜索装配收到了独立 `src/searchx`。
- 这一轮又继续把 dedicated `api-gateway` 入口的语义收紧：`src/app/gateway/*` 中经 `iam/resource/orchestrator/system` 的 remote-aware wrapper 已不再静默回退本地 owner service；同时 `src/repository/scope.go` 也已删除，旧共享排序 helper 不再继续扩大。
- 同一轮里，`src/interface/worker` / `src/interface/controller` 也不再直接依赖 `executionmodule.Service` / `injectionmodule.Service`；runtime 执行 owner 已统一改由 `consumer.ExecutionOwner` / `consumer.InjectionOwner` 注入，owner fallback 面进一步收到了 `src/service/consumer/owner_adapter.go` 单点。
- 继续收口后，dedicated `api-gateway` / `runtime-worker-service` / `system-service` 这几条入口已基本形成明确的 remote-required 语义；当前残余 local adapter 主要服务于 `producer` / `consumer` / `both` 兼容入口，而不是新的 dedicated service 主路径。
- 再往下一轮，`resource-service` / `system-service` / `runtime-worker-service` 也已分别通过 `evaluationmodule.RemoteQueryOption()` / `systemmodule.RemoteRuntimeQueryOption()` / `consumer.RemoteOwnerOptions()` 把 dedicated service 路径上的查询/owner 适配器收成 remote-only，进一步减少“同一服务里同时挂本地和远端两套语义”的过渡态。
- 这一轮继续沿跨服务 DB/owner 深清推进 `team -> project` 这条线：`src/module/team/project_reader.go` 新增 remote-aware project reader，team detail 的 project count 与 team project list 在配置 `clients.resource.target` 或 `resource.grpc.target` 后会优先经 `resource-service` 获取，`iam-service` 也已显式要求 `resource-service` target；对应地 `src/module/project.ListProjectReq` / `src/module/project/service.go` / `src/module/project/repository.go` 已补 `team_id` 与 `include_statistics`，使 team 侧远程 count 可直接复用 `ListProjects` 且可跳过 project statistics 聚合，先把 IAM/gateway 对 `projects`、`fault_injections`、`executions` 的这条直查面压掉一层。
- 这一轮继续把 `resource-service` 里的 project statistics 主路径也收进 owner facade：新增 `src/module/project/project_statistics.go`，project service 不再直接在资源侧 repo 中拼 execution/injection 统计，而是统一经 `projectStatisticsSource` 获取；`src/app/resource/options.go` 已用 `projectmodule.RemoteStatisticsOption()` 把 dedicated resource 路径强制到 orchestrator RPC。对应地 `src/proto/orchestrator/v1/orchestrator.proto`、`src/interface/grpcorchestrator/*`、`src/internalclient/orchestratorclient/client.go` 已补 `ListProjectStatistics` 内部 RPC，resource 主路径上的 project detail/list statistics 不再直查 owner 表。
- 这一轮又继续把兼容入口装配层压实到单点：新增 `src/app/compat_options.go`，把 producer 侧 HTTP/K8s/chaos 与 producer init/http server 装配收成 `ProducerCompatibilityOptions / ProducerHTTPEntryOptions`，把 consumer/both 共享的本地 owner runtime 组合继续收成 `CompatibilityRuntimeOptions()`；`src/app/producer.go`、`consumer.go`、`both.go`、`gateway/options.go` 现在不再各自重复拼 `Base/Observe/Data/Coordination/Build + modules + init + http`，同时已删掉 `NormalizeAddr(...)` 与 gateway 专用 `NewProducerInitializerForGateway / RegisterProducerInitializationForGateway` 这类多余壳函数。
- 这一轮继续把 dedicated `api-gateway` 的 metrics 边界收紧：`src/module/metric` 已补 `HandlerService`，`src/app/gateway/metric_services.go` 新增 remote-aware metrics wrapper，gateway 上的 `/api/v2/metrics/injections|executions|algorithms` 不再直接落本地 `fault_injections / executions / containers` 表；其中 injection/execution metrics 已走新增的 orchestrator RPC `GetInjectionMetrics / GetExecutionMetrics`，algorithm metrics 则由 gateway 经 `resource-service` 拉 algorithm 列表后再按算法向 orchestrator 聚合执行指标，先把 dedicated gateway 这块跨 owner 直查面收掉。
- 这一轮继续把 dedicated `api-gateway` 的 team 主路径切到 IAM：`src/module/team` 已补 `HandlerService`，`src/app/gateway/team_services.go` 新增 remote-aware team wrapper，gateway 上的 `/api/v2/teams/*` 现在统一经 `iamclient` 转发 `Create/Get/List/Update/Delete`、member 管理、team project/member 列表，而不再直接吃本地 team owner 实现；对应地 `src/proto/iam/v1/iam.proto`、`src/interface/grpciam/service.go`、`src/internalclient/iamclient/client.go` 已补齐 team RPC 面。同时 `src/module/team/project_reader.go` 又新增 `RemoteProjectReaderOption()`，`src/app/iam/options.go` 已把 dedicated `iam-service` 上的 team->project 视图继续收成 resource RPC-only。
- 这一轮再把 dedicated `api-gateway` 的 IAM 剩余主路径继续收口：`src/module/{auth,user,rbac}` 已补 `HandlerService`，`src/app/gateway/{auth,user,rbac}_services.go` 新增 remote-aware wrapper，gateway 上的 `/api/v2/auth/*`、`/api/v2/api-keys/*`、`/api/v2/users/*`、`/api/v2/roles|permissions|resources/*` 已统一经 `iamclient` 转发，不再在 dedicated `api-gateway` 入口直接吃本地 IAM owner 实现；对应地 `src/proto/iam/v1/iam.proto`、`src/interface/grpciam/service.go`、`src/internalclient/iamclient/client.go` 也已补齐 auth/user/rbac RPC 面。
- `src/app/app.go` 已开始按服务边界拆装配层：当前新增 `BaseOptions / ObserveOptions / DataOptions / CoordinationOptions / BuildInfraOptions`，独立服务启动链不再统一吃满所有 infra。
- `src/app/resource/options.go` 这一轮继续把 standalone 边界推进到 `project / label / container / dataset / evaluation`；`resource-service` 已接入 `orchestratorclient.Module` 承接 evaluation -> orchestrator 的远程查询，gateway 对 `evaluation` handler 也已补上 remote-aware 装饰。
- 这一轮继续把 dedicated `api-gateway` 的 label 主路径切到 Resource：`src/module/label` 已补 `HandlerService`，`src/proto/resource/v1/resource.proto` / `src/interface/grpcresource/service.go` / `src/internalclient/resourceclient/client.go` 已补齐 `Create/Get/List/Update/Delete/BatchDelete label` 对内 RPC；同时 `src/app/gateway/resource_services.go` 与 `src/app/gateway/options.go` 已把 `/api/v2/labels/*` 改成统一经 `resource-service` 转发，dedicated gateway 不再直接承载 label owner 读写。
- 这一轮继续把 dedicated `api-gateway` 的 admin systems 主路径切到 Resource：`src/module/chaossystem` 已补 `HandlerService`，`resource-service` 现已纳入 `chaossystemmodule.Module`，并通过 `src/proto/resource/v1/resource.proto` / `src/interface/grpcresource/service.go` / `src/internalclient/resourceclient/client.go` 承接 `List/Get/Create/Update/Delete chaos system` 与 `metadata upsert/list`；同时 `src/app/gateway/resource_services.go` 与 `src/app/gateway/options.go` 已把 `/api/v2/systems/*` 改成统一经 `resource-service` 转发，继续缩小 dedicated gateway 上的本地 resource owner 面。
- 这一轮继续把 dedicated `api-gateway` 的 task / trace 查询主路径切到 Orchestrator：`src/module/{task,trace}` 已补 `HandlerService`，`src/internalclient/orchestratorclient/client.go` 已补 `GetTask / ListTasks / GetTrace / ListTraces`，`src/app/gateway/orchestrator_services.go` 与 `src/app/gateway/options.go` 已把 `/api/v2/tasks/{id}`、`/api/v2/tasks`、`/api/v2/traces/{id}`、`/api/v2/traces` 改成统一经 `orchestrator-service` 转发；日志 WebSocket 与 trace SSE 仍保留本地实现，留待后续流式通道单独收口。
- 这一轮继续把 dedicated `api-gateway` 的 group stats 查询主路径切到 Orchestrator：`src/module/group` 已补 `HandlerService`，`src/proto/orchestrator/v1/orchestrator.proto` / `src/interface/grpcorchestrator/service.go` / `src/internalclient/orchestratorclient/client.go` 已补 `GetGroupStats` 内部 RPC，`src/app/gateway/orchestrator_services.go` 与 `src/app/gateway/options.go` 已把 `/api/v2/groups/{group_id}/stats` 改成统一经 `orchestrator-service` 转发；group SSE stream 仍保留本地实现，留待后续流式通道单独收口。
- 这一轮继续把 dedicated `api-gateway` 的 SSE 主路径切到 Orchestrator：`src/module/notification` 已补 `HandlerService`，`src/proto/orchestrator/v1/orchestrator.proto` / `src/interface/grpcorchestrator/service.go` / `src/internalclient/orchestratorclient/client.go` 已新增 `GetTraceStreamState / ReadTraceStreamMessages / GetGroupStreamState / ReadGroupStreamMessages / ReadNotificationStreamMessages` 五个内部 RPC，`src/app/gateway/orchestrator_services.go` 与 `src/app/gateway/options.go` 已把 `/api/v2/traces/{trace_id}/stream`、`/api/v2/groups/{group_id}/stream`、`/api/v2/notifications/stream` 改成统一经 `orchestrator-service` 读取流式批次；当前 dedicated gateway 残余主线只剩 task logs WebSocket 尚未收成内部通道。
- 这一轮继续把 dedicated `api-gateway` 的 task logs WebSocket 也切到 Orchestrator：`src/proto/orchestrator/v1/orchestrator.proto` / `src/interface/grpcorchestrator/service.go` / `src/internalclient/orchestratorclient/client.go` 已新增 `PollTaskLogs` 内部 RPC，`src/module/task/service.go` 新增基于 Loki 的 owner-side log poll facade，`src/app/gateway/orchestrator_services.go` 则把 `/api/v2/tasks/{task_id}/logs/ws` 改成由 gateway 继续负责边缘 WebSocket、但日志历史/轮询数据统一经 `orchestrator-service` 获取；dedicated gateway 主线上的 task/trace/group/notification 读写 owner 残余面已基本清空。
- 这一轮再把 `module/evaluation` 的查询源约束收紧一层：`src/module/evaluation/service.go` 里 `Execution` 依赖已改成可选，若既没有 orchestrator client、也没有本地 execution owner，会直接显式报错；同时补了 `src/module/evaluation/service_test.go` 锁住这条行为。
- 这一轮又继续把 gateway -> system 的旧监控接口 fallback 收紧一层：`src/module/system/handler_service.go` / `src/module/system/handler.go` / `src/app/gateway/system_services.go` / `src/interface/grpcsystem/service.go` 里的 `GetMetrics / GetSystemInfo` 已统一改成返回 `(..., error)`；配置 `systemclient` 后不再在 remote 调用失败时静默吞掉错误并回退本地结果。
- 这一轮继续把 runtime 的 owner fallback 收口成单点适配器：新增 `src/service/consumer/owner_adapter.go`，`collect_result / fault_injection / algo_execution / state_store / k8s_handler` 不再各自散落判断 `orchestratorclient` 与本地 owner，而是统一经 `ExecutionOwner / InjectionOwner` 做 remote-first 路由；`src/interface/{worker,controller}/module.go` 也改为只在装配层创建这两个 owner 适配器，把 fallback 面进一步压缩到 consumer 单点。
- 这一轮再把 gateway 请求上下文继续贯穿一层：`src/app/gateway/middleware_service.go` 不再用 `context.Background()` 调 IAM client，`middleware.Service` 的 permission helper 已统一改成显式接收 `context.Context`；同时 `src/module/system/*` / `src/interface/grpcsystem/*` 也把 `GetMetrics / GetSystemInfo / GetAuditLog / ListAuditLogs / GetConfig / ListConfigs` 这批读接口改成透传请求上下文，旧 system remote-aware wrapper 不再自己造背景上下文。
- 这一轮也顺手把 evaluation -> orchestrator 的本地/远程路由收成单点：新增 `src/module/evaluation/execution_query.go`，`module/evaluation.Service` 不再自己持有 `orchestratorclient + execution service` 两套判断，而是统一走 `executionQuerySource` 适配器。
- 这一轮继续把 evaluation 主路径真正并进 `resource-service`：`src/interface/grpcresource/service.go`、`src/internalclient/resourceclient/client.go`、`src/app/gateway/resource_services.go` 已补齐 `ListDatapackEvaluationResults / ListDatasetEvaluationResults / ListEvaluations / GetEvaluation / DeleteEvaluation`，gateway 在配置 `clients.resource.target` 或 `resource.grpc.target` 后会优先走 resource gRPC。
- 这一轮继续把 startup 壳和 system runtime fallback 再压一层：新增 `src/app/runtime_stack.go` 把 runtime worker 的 infra/provider/interface 装配统一抽成共享 stack，`src/app/consumer.go` / `src/app/both.go` / `src/app/runtime/options.go` 不再各自重复拼同一套启动树；同时新增 `src/module/system/runtime_query.go`，`module/system.Service` 对 runtime client / 本地 systemmetric 的切换也已收成单点 `runtimeQuerySource`。
- 本轮继续把 `service/common` 里的 container/label 共享 helper 回收到 owner 模块：`src/module/container/resolve.go` 与 `src/module/label/core.go` 已承接这批逻辑，`module/{execution,injection,evaluation,project,container,dataset}` 主路径不再经由 `service/common` 读 container 参数/版本或创建 labels。
- 本轮继续把 `service/common` 里的 dataset version helper 也收回 owner 模块：`src/service/common/dataset.go` 已删除，`src/module/dataset/resolve.go` 负责 dataset version 解析，`src/module/dataset/api_types.go` 同时去掉了对 `module/injection` 的响应耦合。
- 本轮继续把 `service/common/datapack_resolver.go` 也收回 owner 模块：当前 `src/service/common/datapack_resolver.go` 已删除，datapack 本身与 dataset->datapack 解析已迁入 `src/module/injection/resolve.go`，`module/execution` / `module/injection` 提交主路径不再经由 `service/common`。
- 这一轮再按“模块 repo 写实、少留裸导出 helper”的口径继续内聚了一批仓储逻辑：`src/module/container/{core,resolve}.go`、`src/module/dataset/{core,resolve}.go`、`src/module/injection/resolve.go`、`src/module/label/core.go` 已改成以 `Repository` 方法为主；`service/initialization`、`module/{execution,evaluation,injection,project,container,dataset}`、`service/consumer/k8s_handler.go` 这批调用点已不再直连 `Create*Core` / `MapRefs*WithDB` / `ExtractDatapacksWithDB` / `CreateOrUpdateLabelsFromItems` 之类裸函数，而是显式走各自模块 repo。
- 这一轮继续按同一口径压缩 repo/API 面并深清 interface 残余查询：`src/interface/grpcorchestrator/project_statistics.go` 已不再自己持有 Gorm 聚合 SQL，而是改为复用 `src/module/project.Repository.ListProjectStatistics(...)`；`src/module/team/repository.go` 里重复的 `listTeamProjectStatistics(...)` 也已删除，team 本地 project list statistics 改为复用 project 模块仓储。顺手又把 `src/module/{user,team,rbac}/repository.go` 里一批仅供各自 service 使用的 CRUD / assign / remove / batch helper 收成包内私有方法，继续减少模块 repo 对外暴露面。
- 这一轮再顺着主线补了两处收口：`src/app/compat_options.go` 现已把 consumer 兼容入口里的本地 execution/injection owner 组合显式收成 `CompatibilityOwnerFallbackOptions()` 单点，不再散着写在兼容 runtime 入口里；同时 `src/module/project.Repository.ListProjectStatistics(...)` 已补上对 `fault_injections / executions` 的 `status != deleted` 过滤，和之前 orchestrator interface 的 owner 统计语义重新对齐。`src/module/team` 的 team detail 读取也顺手收成 `loadTeamDetailBase(...)`，避免继续在本地 detail helper 里混入最终由 remote reader 接管的 project count 语义。
- 这一轮又继续把 repo 暴露面按“没用到就删”收了一层：`src/module/{project,team,user,rbac}/repository.go` 的 `Transaction(...)` 已统一收成包内 `transaction(...)`；顺手补扫了当前 `src/module/*/repository.go`，删除了已无任何调用的 `src/module/execution/repository.go:268` `loadExecutionLabelIDsByItems(...)`。同时 `src/app/http_modules.go` / `src/app/compat_options.go` / `src/app/orchestrator/options.go` 现在通过 `ExecutionInjectionOwnerModules()` 复用 execution/injection owner 组合，compat/orchestrator 邻近不再各自散写相同模块列表。
- 这一轮再继续按“整个文件没价值就直接删”的口径清理：由于 `src/repository/*.go` 这批旧共享 repository 文件已无任何业务 import 或有效调用（剩余 `repository.DownloadIndexFile()` 仅为 Helm 官方 `repo` 包别名，不是本项目包），当前已整体删除 `src/repository/{container,dataset,detector,execution,granularity,injection,label}.go`。顺手又补扫了 `src/app` / `src/interface` 里的 `Table/Joins/Raw`，目前已无新的“非 owner 层自己拼 DB 查询”残点，残余数据访问基本都收敛在各自 owner 模块 repo 或 runtime owner 内。
- 这一轮继续顺着你要的两条线往下压：compat 侧新增 `src/app/compat_options.go` 的 `BothCompatibilityOptions(...)`，`src/app/both.go` 不再自己散拼 runtime+HTTP 组合；project/team/user/rbac 这批模块 repo 里又删/折了一批只服务单一路径的内部 helper——例如 `module/project` 把 project label 管理与 label reload、project user count、project list label 装配继续内联回主语义方法，`module/team` 把 team user count、visible team id 查询、team project count / role existence 这批单点 helper 收回主路径或 local reader，`module/user` 把 `ensureUserUnique(...)` 折回 `createUserIfUnique(...)`，`module/rbac` 也继续收掉了旧的 `loadAssignablePermissions(...)` 壳。当前 `project/team/user/rbac` 剩余 repo 方法已基本都对应明确单一 service 语义，不再是“公共但没边界价值”的散 helper。
- 本轮补扫结果表明，当前最高优先级主线残余已进一步收敛到各服务残余 local fallback 的继续压缩。
- 同时已继续把 HTTP 鉴权链往 IAM client 收深一轮：`src/middleware/auth.go` 不再直接依赖 `utils.ValidateToken`，而是通过 `middleware.TokenVerifier` 接口走注入实现；当前默认仍由 `authmodule.Service` 提供，本地功能不变。并且 `src/internalclient/iamclient/*` 已继续补齐 team/project 的 member/admin/public 判定 RPC，`src/app/gateway/options.go` 在配置 `clients.iam.target` 或 `iam.grpc.target` 时，已可优先切到 IAM gRPC 做 token verify、`CheckUserPermission(...)` 与 team/project 权限辅助判断。
- 这一轮继续按“模块内直接写实、删除空包装”的口径再压一层 repo API 面：`src/module/{project,team,user,rbac,execution,injection}/repository.go` 中原先只做 `db.Transaction(...)` / `&Repository{db: tx}` 的 `transaction/Transaction/withDB` 空包装已全部删除，service 组合点统一直接走 `repo.db.Transaction(...)` + `NewRepository(tx)`；同时 `src/module/injection/repository.go` 中只在模块内部使用的一整批方法也已收成包内私有命名，例如 `loadInjection(...)`、`findInjectionByName(...)`、`createInjectionRecord(...)`、`deleteInjectionsCascade(...)`、`listInjectionsView(...)` 等，进一步减少模块仓储对外暴露面并把 compat/local owner 主路径收得更实。
- 紧接着又把同一口径补到 `src/module/execution/repository.go`：`AddExecutionLabels / ClearExecutionLabels / BatchDecreaseLabelUsages / ListExecutionIDsByLabelItems / BatchDeleteExecutions / UpdateExecutionDuration / LoadExecution / CreateExecutionRecord / UpdateExecutionFields / SaveDetectorResults / SaveGranularityResults` 这批仅供模块内 service/runtime owner 使用的方法已全部私有化，`module/execution/service.go` 也同步改成只走模块内语义方法，进一步减少 execution repo 的公开 API 面。
- 这一轮继续把同样的收口扩到 `src/module/{container,dataset,system}`：三处 repo 的 `Transaction(...) / withDB(...)` 空包装都已删除，service 组合点统一改成 `repo.db.Transaction(...) + NewRepository(tx)`；同时 `module/container` / `module/dataset` 中大批仅供模块内部使用的 CRUD、label、version、Helm/Datapack 关系方法已收成包内私有实现，`module/system` 中的 `getAuditLogByID(...)`、`getConfigByID(...)`、`getConfigHistory(...)`、`updateConfig(...)`、`createConfigHistory(...)`、`listConfigHistoriesByConfigID(...)` 也已一并私有化，进一步压缩 repo API 面。对应编译检查已通过：`cd src && GOCACHE=/tmp/aegis-go-cache go test ./module/container ./module/dataset ./module/system ./app ./app/gateway ./app/runtime ./app/orchestrator ./app/iam ./app/resource ./app/system ./service/initialization -run '^$'`。
- 紧接着又继续削了一轮 repo 暴露面和 compat 壳：`src/module/container/repository.go` 的 `ListContainers / ListContainerVersions`、`src/module/dataset/repository.go` 的 `ListDatasets / SearchDatasets / ListDatasetVersions`、`src/module/system/repository.go` 的 `ListAuditLogs / ListConfigs / ListConfigHistories` 已全部收成包内私有实现，当前这三块对外只剩真正有跨模块边界价值的方法（例如 dataset 的 `ListInjectionsByDatasetVersionID(...)`）；同时 `src/app/compat_options.go` 里的 `ConsumerRuntimeOptions()` 空转发已删除，`src/app/consumer.go` 直接走 `CompatibilityRuntimeOptions()`，compat 启动壳再薄一层。对应编译检查已再次通过：`cd src && GOCACHE=/tmp/aegis-go-cache go test ./module/container ./module/dataset ./module/system ./app ./app/gateway ./app/runtime ./app/orchestrator ./app/iam ./app/resource ./app/system ./service/initialization -run '^$'`。
- 这一轮继续把 `app/*/options.go` 这批启动壳里的纯组合 helper 删了一层：`src/app/{iam,resource,system,orchestrator}/options.go` 中仅被各自 `Options(...)` 调用一次的 `Modules()` 已全部内联删除，`src/app/gateway/options.go` 里无调用价值的 `Modules()` 也已直接删除；同时 `src/app/compat_options.go` 内部仅被单点使用的 `ProducerInitializationOptions()`、`HTTPServerOptions()`、`CompatibilityOwnerFallbackOptions()` 也已折回主入口。当前启动链保留的 helper 主要只剩确实复用的 `CommonOptions(...)`、`ProducerHTTPModules()`、`RuntimeWorkerStackOptions()`、`ExecutionInjectionOwnerModules()` 这类有明确边界价值的组合。对应编译检查已通过：`cd src && GOCACHE=/tmp/aegis-go-cache go test ./app ./app/gateway ./app/runtime ./app/orchestrator ./app/iam ./app/resource ./app/system ./module/container ./module/dataset ./module/system ./service/initialization -run '^$'`。
- 这一轮继续把“remote + local fallback” 双态适配器再压一层：`src/service/consumer/owner_adapter.go` 已把 execution/injection 的 local fallback 与 remote-only 两套 adapter 合并成统一结构，通过 `requireRemote` 控制 dedicated runtime-worker 是否允许回落本地 owner；`src/module/team/project_reader.go` 同样把 local / remote fallback / remote-only 三套 reader 合并成单一 `projectReaderAdapter`；并顺手把同类模式的 `src/module/project/project_statistics.go`、`src/module/evaluation/execution_query.go`、`src/module/system/runtime_query.go` 也统一成单 adapter + `requireRemote` 形态，减少重复实现与过渡态暴露面。仓库级补扫结果显示，当前生产代码里这类双 adapter 模式已基本清空；残余 `Transaction/withDB` 包装主要集中在 `module/auth` / `module/label` 这类还未进入本轮主线的模块。对应编译检查已通过：`cd src && GOCACHE=/tmp/aegis-go-cache go test ./service/consumer ./module/team ./module/project ./module/evaluation ./module/system ./app ./app/runtime ./app/iam ./app/resource ./app/system ./app/gateway -run '^$'`。
- 这一轮把前面补扫里最后两块明显残余也收掉了：`src/module/auth/repository.go` 的 `UserRepository/RoleRepository withDB(...) + Transaction(...)` 空包装已删除，`src/module/auth/service.go` 改成直接使用 `userRepo.db.Transaction(...)` + `NewUserRepository(tx)` / `NewRoleRepository(tx)`；`src/module/label/repository.go` 的 `Transaction(...)` 包装也已删除，`src/module/label/service.go` 同步切成 `repo.db.Transaction(...)`。复扫结果显示，当前 `src/app` / `src/module` / `src/service/consumer` 主线里已不再存在这类 repo `withDB(...)` / `Transaction(...)` compat 壳；剩余 `withDB(...)` 命中主要只在 consumer task-state builder 这种内部 fluent helper 上，不再是 repository 兼容层。对应编译检查已通过：`cd src && GOCACHE=/tmp/aegis-go-cache go test ./module/auth ./module/label ./app ./app/gateway ./app/iam ./service/consumer ./module/team ./module/project ./module/evaluation ./module/system -run '^$'`。
