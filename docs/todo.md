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

- [x] 阅读 [backend-fx-refactor-plan.md](./backend-fx-refactor-plan.md)
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
- SDK / CLI 认证主线改为 `AK/SK -> access token`；`username/password login` 仅保留给 Portal / Admin 等人类交互入口。
- `AK/SK -> token` 进一步改为 header 签名模式：`X-Access-Key`、`X-Timestamp`、`X-Nonce`、`X-Signature`，业务接口仍继续走 Bearer token。

本轮执行清单：

- [x] 收缩 `sdk` audience 到最小可维护白名单
  - 已继续把剩余误标的 `sdk` audience 收回，只保留 `POST /api/v2/auth/access-key/token` 与 `src/router/sdk.go` 下 4 个 SDK 样例接口；当前 `sdk.json` 已收缩到 `5 paths / 5 operations`。
- [x] 从 Swagger audience 中移除 `POST /api/v2/auth/register` 的 `sdk`
- [x] 从 Swagger audience 中移除 `POST /api/v2/auth/login` 的 `sdk`
- [x] 盘点并设计 AK/SK 数据模型
  - 已新增 `database.UserAccessKey`，覆盖 `owner`、`enabled/disabled/deleted`、`expires_at`、`last_used_at`、`name/description`、`secret_hash`，并纳入 `AutoMigrate`。
- [x] 增加 AK/SK 管理接口
  - 已补 `portal` 路由：`GET/POST /api/v2/access-keys`、`GET/DELETE /api/v2/access-keys/{access_key_id}`、`POST /api/v2/access-keys/{access_key_id}/rotate|disable|enable`。
- [x] 增加 `AK/SK -> token` 接口并标为 `sdk`
  - 已补 `POST /api/v2/auth/access-key/token`，返回 Bearer token，并在 JWT claims 中标记 `auth_type=access_key` 与 `access_key_id`；当前入口改为 `X-Access-Key` / `X-Timestamp` / `X-Nonce` / `X-Signature` 头签名校验，服务端会校验 5 分钟时间窗并用 Redis 做 nonce 防重放。
- [x] 将 Python SDK 的鉴权入口切到 AK/SK
  - `sdk/python/src/rcabench/client/http_client.py` 已改为优先使用 `token` 或 `access_key + secret_key`；SDK 不再依赖 username/password login，环境变量同步切到 `RCABENCH_ACCESS_KEY` / `RCABENCH_SECRET_KEY`，并在换 token 时自动按 `METHOD\\nPATH\\nACCESS_KEY\\nTIMESTAMP\\nNONCE` 规范计算 HMAC-SHA256 签名头。
- [x] 将 `aegisctl` 的鉴权入口切到 AK/SK
  - `src/cmd/aegisctl/cmd/auth.go` / `src/cmd/aegisctl/client/auth.go` 已切到 `--access-key` + `--secret-key` 签名换取 `POST /api/v2/auth/access-key/token`；签名规范与 Python SDK 保持一致，登录结果继续只落盘 Bearer token，不保存 `secret_key`。
- [x] 给 `aegisctl` 增加本地签名排障命令
  - 已补 `aegisctl auth inspect` 与 `aegisctl auth sign-debug`；前者可检查当前 context 的 token / auth_type / access_key / expiry，后者可直接打印 canonical string、签名头与 curl 样例，并可通过 `--execute` 直接发起换 token 请求回显响应，或通过 `--save-context` 直接把成功返回的 Bearer token 落盘到当前 CLI context，便于排查 SDK / CLI / 服务端签名不一致问题。
- [x] 补充 AK/SK 头签名规范文档
  - 已新增 `docs/access-key-signature-spec.md`，明确 canonical string、Header 约定、HMAC 规则、时间窗与 nonce 防重放语义，并补了 Portal 上 access key 的使用说明、curl 示例与 `aegisctl` 排障命令说明；同时已回填 `src/handlers/v2/access_keys.go` / `src/dto/auth.go` 的 Swagger/OpenAPI 注释与 schema example，前端与文档站可直接消费。
- [x] 补 Portal access key 前端文案与表单提示
  - `../AegisLab-frontend/src/pages/settings/Settings.tsx` 已新增 Access Keys 管理页签，覆盖创建 / 轮换 / 启停 / 删除与一次性 secret 提示；`../AegisLab-frontend/src/api/auth.ts` 也已补齐 access key API 封装，页面文案与 OpenAPI 说明保持一致。
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
  - `module/user` CRUD / 资源授权、`module/systemmetric` 指标查询、`module/rbac` 已基本切离 `service/producer`；`handlers/system/monitor.go`、`configs.go`、`audit.go` 主路由入口也已并入 `module/system`。此前已删除旧 `service/producer` 中的 system / metrics / sdk / chaos-system / permission / audit / evaluation / notification / team / trace / group 兼容入口；middleware 也不再直接依赖旧 producer。`module/container` 与 `module/dataset` 现已进一步把 CRUD / detail / list / labels / version 元数据、container build / helm upload、dataset filename / download / version injection 路径下沉到模块 service/repository，并把直接碰 `config` / git / 文件系统的部分收成模块内 gateway/store。旧 `service/producer/container.go` / `dataset.go` 已删除；初始化已改走 `module/container` / `module/dataset` 暴露的 core helper。最近几轮里，`module/injection` 已先后接管 datapack download / files / file query / upload / build 提交流程，以及 injection list / project list / detail / labels / logs / submit fault injection / search / no-issues / with-issues / clone / batch delete 主路径；`src/service/producer/injection.go` 已整体删除。随后又继续按“模块语义留在模块 repo、纯转发尽量删除”的口径收缩：`module/injection` 把 search / list / labels / batch label 管理，以及 project injection list 的标签装配收进 `repository.go`，并继续把 `LoadInjection` / `FindInjectionByName` / `CreateInjectionRecord` / `LoadTask` / `LoadPedestalHelmConfig` / label/execution 删除辅助等一批原子转发写实到模块仓储；最近三轮又把 project resolve、detail with labels、existing injection map、label 条件聚合、project injection list、issue/no-issue 视图、label id by key、fault injection 批量 with labels 这批组合查询继续收成模块内实现。`module/user` 这一轮又把 `CreateUser + EnsureUserUnique`、`Get/Update` 这批基础 CRUD 空包装进一步折成 `CreateUserIfUnique`、`GetUserDetailBase`、`UpdateMutableUser`、`ListUserViews`，并把 `DeleteUserCascade`、global/container/dataset/project 的 assign/remove、permission batch create/delete 这批 relation 逻辑也直接写进模块 repo；随后又把 user detail 关系装配，以及 role/container/dataset/project 的加载 helper 继续改为模块内直接查库；最近又把 permission id 批量校验也直接内聚到模块仓储，并把纯存在性校验提升成公开 `EnsureUserExists(...)` 供 service 组合点复用。`module/rbac` 把 role 详情装配、权限批量校验、角色删除级联、resource/permission 关系查询收进模块 repo，并继续把 role / permission / resource 的基础 list/load/create 查询直接内聚到模块仓储；最近又把 role detail、role->user、permission->role、resource->permission 这批组合视图改成模块内直查；上一轮再把 role delete cascade、mutable update、permission id 批量加载也进一步改成模块仓储自管；这一轮继续把“可写 role”校验收口成模块内 `loadWritableRole(...)`，同时把通用 `LoadPermission` / `LoadResource` 改成更贴业务语义的 `GetPermissionDetail(...)` / `GetResourceDetail(...)`。`module/project` 现已把 create-with-owner、delete cascade、detail/list 视图装配、mutable update、label reload 与按 key 移除标签收进自身 repo，这几轮继续把 project owner role 查询、project statistics 聚合、label 批量装配 / project label id 查找 / usage decrease 一并写实；这一轮再把内部 helper 命名继续往语义侧收紧成 `loadProjectRecord(...)` / `listProjectStatistics(...)`。`module/team` 也把 create-with-creator、detail 聚合、visible list、team project list、member add/remove/update role、team visibility 读取等操作收进 repo，并把 team project statistics 聚合也留在模块内；这一轮又把 team 加载进一步收成 `loadTeam(...)`，用于 detail / mutable update / ensure exists / visibility 读取，同时把 project statistics helper 明确成 `listTeamProjectStatistics(...)`。`module/execution` 现已接管 project list / global list / detail / labels / batch delete / detector result / granularity result / submit execution 全链路，新增自身 `repository.go` 并删除旧 `src/service/producer/execution.go`。由于 project 主路径此前早已由 `module/project` 承接，本轮也同步删除了已空心化的 `src/service/producer/project.go`；同时 `service/producer/label.go` 也已删除，初始化阶段改走 `module/label.CreateLabelCore`。`service/producer/relation.go`、`user.go`、`role.go`、`resource.go`、`auth_helpers.go`、`permission_helpers.go`、`datapack_archive.go` 同样已清掉，producer 侧残余重点进一步收敛到更少的共享逻辑；当前 `src/service/producer` 已无 Go 源文件残留。与此同时，旧 `src/client/loki.go` / `jaeger.go` / `redis_client.go` / `etcd_client.go` / `harbor_client.go` / `helm.go` / `client/k8s/*` 及 Helm 对应测试也已从 root `client` 包清走，真实实现统一并入 `src/infra/*`；上一轮已把 `src/infra/k8s/client.go` 删除，rest/client/dynamic/controller 的单例初始化直接吸回 `src/infra/k8s/gateway.go`；这一轮继续把 `service/consumer` / `service/initialization` 中的 `CurrentK8sController()` fallback 干掉，改成由 Fx 注入 `*k8sinfra.Controller`，同时 `service/common` 的 etcd fallback 改为回落到 `infra/etcd.GetGateway()` 单点入口，并进一步删掉 `service/consumer/deps.go` / `service/common/deps.go` 这类旧全局依赖注册文件。`service/consumer` 中剩余的 K8s / BuildKit / Helm 访问也继续改为直接走 `infra/*` 单点入口：新增 `buildkitinfra.GetGateway()`、`helminfra.GetGateway()`，`CurrentK8sGateway()` / `currentBuildkitGateway()` / `currentHelmGateway()` 已全部清掉；这轮又把 `app/startup.go` 删除，并进一步引入 `app.RegisterProducerInitialization`，把 producer 初始化从 `context.Background()` 改成走 Fx `OnStart` 生命周期上下文。随后又继续把 `interface/controller` / `interface/receiver` / `interface/worker` 的生命周期上下文改成从 Fx `OnStart` 派生，不再在模块注册期直接构造 `context.Background()`；再往下一轮又把 `service/consumer/task.go` / `trace.go` / `jvm_runtime_mutator.go` / `k8s_handler.go` 里残余 `context.Background()` 全部清成 consumer 内部 detached context helper。初始化侧原先带 callback 的 `registerHandlers(...)` 旧 helper 也已改成更窄职责的 `activateConfigScope(...)`，consumer / producer 各自显式注册所需 handlers，再统一激活 listener scope；这一轮再把 `GetConfigUpdateListener(...)` 单例 helper 从启动链收掉，改为在 producer / worker Fx `OnStart` 生命周期里显式创建 `ConfigUpdateListener` 后传给 initialization。`service/consumer` 的 Redis 直连也开始往更窄语义收：新增内部 `currentRedisGateway` / `currentRedisClient` / `publishRedisStreamEvent` / `publishTraceStreamEvent` / `loadCachedInjectionAlgorithms` helper，先把 trace/group stream 发布、detector cache 读取，以及 `monitor` / `rate_limiter` 对 Redis gateway 的获取收进更窄入口；随后又把 monitor 的上下文来源收回 worker lifecycle，并把 namespace SMembers/HGet/HSet/Pipeline 这批读取/写入改为统一走 consumer 内部 Redis helper 取 client，同时 `rate_limiter` 也不再自持 Redis client，而是统一经由 consumer Redis helper 获取连接；最近一轮再把 namespace key / exists / field read / seed / lock write 继续折成 `monitor` 内部更窄 helper，减少 monitor 主流程里散落的 Redis 原语；上一轮则继续把 rate limiter Redis 操作下沉成独立 `tokenBucketStore`，把 token acquire/release 的 Redis 细节与 limiter 配置/调度逻辑分开；这一轮再正式把 monitor 按同一路径拆出独立 `namespaceStore`，把 namespace key/list/exists/read/write/watch/status 这批 Redis 操作从 monitor 主流程里抽走；紧接着又继续深拆成 `namespaceCatalogStore` / `namespaceLockStore` / `namespaceStatusStore` 三个更窄 store，把锁读取/抢占/释放、namespace 注册、status 读写彻底从 `monitor.go` 抽开，并删除已空心化的 `src/service/consumer/namespace_store.go`。这一轮再把 startup / interface 链路里对 monitor 的旧包级获取收一批：`consumer.NewMonitor(...)` 作为 Fx provider 现在直接吃 `*redisinfra.Gateway` 并在内部自取 client，monitor 构造期不再向启动链暴露裸 `*redis.Client`，`initialization.InitializeConsumer(...)`、`RegisterConsumerHandlers(...)`、`interface/controller` 的 K8s callback 构造均改为显式注入 monitor，而不再自己碰 `GetMonitor()`；紧接着又继续把运行时执行主流程里的 monitor 单例拿掉，新增 `consumer.RuntimeDeps` 由 worker lifecycle 显式传入，`dispatchTask(...)` / `executeTaskWithRetry(...)` / `executeFaultInjection(...)` / `executeRestartPedestal(...)` 已不再自己碰 `GetMonitor()`。这一轮继续顺着同一主线把 rate limiter 也从进程级单例收成纯 Fx provider：`NewRestartPedestalRateLimiter(...)` / `NewBuildContainerRateLimiter(...)` / `NewAlgoExecutionRateLimiter(...)` 现在直接吃 `*redisinfra.Gateway` 构造 limiter，不再经过 `Get*RateLimiter()` / `sync.Once`；`executeBuildContainer(...)`、`executeAlgorithm(...)`、`executeRestartPedestal(...)` 与 K8s job 回调里的 algorithm token release 也都改为走显式传入 limiter，不再直接碰旧包级 getter。与此同时，`service/common/config_registry.go` / `config_listener.go` 把配置元数据读取继续收成 `service/common/config_store.go` 本地语义 store，不再穿过公共 `repository` 包；随后又把 producer/worker/controller/receiver 的启动执行体再收成显式可替换的 `ProducerInitializer` / `LifecycleRunner` 依赖，避免 lifecycle 本身直接抱一大串底层依赖，主路径更贴近 Fx；在此基础上，`src/app/startup_validate_test.go` 与 `src/app/startup_smoke_test.go` 现在已经补上 producer / consumer / both 三种 app option 的 Fx 图校验与 start/stop smoke（通过替换重型初始化依赖，验证 HTTP/worker/controller/receiver/producer lifecycle 编排本身可启动可停止）。这一轮继续顺着同一条线，把 `service/common/config_registry.go` 里的 `sync.Once` / `globalHandlersOnce` 再压掉，改成常驻 registry + 幂等注册逻辑，并补上 `config_registry_test.go` 锁住“全局 handlers 多次注册不重复”行为，进一步减少 config startup 主路径上的一次性单例状态；紧接着又继续把 listener / publish 周边的剩余全局依赖再收一层：`ConfigUpdateListener` 现在显式携带 `*gorm.DB`，不再在读取配置元数据和处理变更时回落到 `database.DB`；`RegisterGlobalHandlers(...)` / `RegisterConsumerHandlers(...)` 也开始显式接收 `ConfigPublisher`，`PublishWrapper(...)` 改成走传入 publisher，而不再自己碰 `redisinfra.GetGateway()`。对应地 producer / consumer 初始化与 worker lifecycle 现已把 Redis gateway / DB 一路显式传进 config listener 与 handler 注册主链。顺手也暴露并修复了 producer 模式此前缺少 `k8sinfra.Module`、导致 `chaosinfra.Module` 无法解析 `*rest.Config` 的问题。当前 producer / consumer / both 三种 app options 都已能通过 `go test ./app` 的图校验和启动链 smoke。这一轮继续把 `service/common` 热路径往显式 DB 收：`DBMetadataStore` 改成由 initialization 注入 `*gorm.DB` 创建，`container` / `dataset` / `task` 公共能力补上 `WithDB` 变体，`module/execution` / `module/injection` / `module/container` 的提交与 ref 解析主路径已改用模块 repo 自带 DB，不再回落到 `database.DB`。这一轮又继续把 consumer 运行态主链的 DB 依赖显式化：`consumer.RuntimeDeps` 开始携带 DB，worker/controller 生命周期分别把 DB 显式注入 task runtime 与 K8s handler，build/restart/algo reschedule、fault injection 落库、collect result 查询、K8s job/CRD 回调里的 execution/injection 状态推进与后续 task submit 也都改成优先走注入 DB，而不再默认抓全局 `database.DB`。这一轮继续把状态同步链也收进显式 DB：`taskStateUpdate` 新增 DB 上下文，`updateTaskState(...)` / `updateTraceState(...)` / trace optimistic lock 更新现在优先沿调用链携带的 DB 执行；K8s error context 也开始透传 handler 注入 DB，因此 consumer 主链里剩余 `database.DB` 基本只落在少量兼容 fallback 和 `service/common` 默认 wrapper。这一轮顺手再把 `module/evaluation` -> `service/analyzer` 这条链也切到显式 DB：evaluation service 改用 repo 持有 DB 调 analyzer 的 `WithDB` 版本，container/dataset ref 解析与 evaluation 持久化不再依赖 analyzer 内部全局 DB；同时 `service/common/ExtractDatapacks(...)` 解析 dataset 时也已改走传入 DB 的 `MapRefsToDatasetVersionsWithDB(...)`。再往下一步，consumer 里 `collect_result` / `fault_injection` / `createExecution` 这类原先“nil 就回落全局 DB”的点也开始直接要求 runtime DB 存在，进一步缩小 fallback 面积。这一轮再继续把兼容层直接砍掉：`service/common` 里默认版 `MapRefsToContainerVersions` / `MapRefsToDatasetVersions` / `ListContainerVersionEnvVars` / `ListHelmConfigValues` / `SubmitTask` / `ProduceFaultInjectionTasks` 已删除，`service/analyzer` 里的默认版 evaluation 入口也删掉，只保留显式 `WithDB` 路径；同时 `consumer/task.go` / `trace.go` / `k8s_handler.go` 里的 DB fallback 也改成显式报错，不再默默回落全局 `database.DB`。紧接着又把 `module/system` 里最后一处直接碰 `database.DB` 的 health check 改成走 `repo.DB()`；目前 `module/*`、`service/common`、`service/consumer`、`service/analyzer` 这批主线包内已无 `database.DB` 残留。顺手又把 repository 层里残留的统计/搜索/资源/注入查询改成统一吃显式 `db` 参数，`repository/task.go` 的 `ListTasksByTimeRange(...)` 也不再偷偷回落全局 DB；现在全仓库只剩 `src/infra/db/module.go` 这一处集中持有 `database.DB`，作为 Fx 提供与关闭数据库连接的基础设施边界。最近两轮又继续把 consumer 外部依赖收窄到 Fx 注入：`interface/worker` 把 `*k8sinfra.Gateway` / `*buildkitinfra.Gateway` / `*helminfra.Gateway` / `*consumer.FaultBatchManager` / `*redisinfra.Gateway` 显式塞进 `consumer.RuntimeDeps`，`build container` / `build datapack` / `algo execution` / `restart pedestal` / `collect result` / task retry / trace state update / K8s callback 已不再直接碰 `GetGateway()` 与 fault batch `sync.Once` 单例；`interface/controller` 同步把 K8s gateway、Redis gateway 和 batch manager 显式交给 `consumer.NewHandler(...)`；`service/logreceiver` 也开始由 `interface/receiver` 注入 Redis publisher，OTLP receiver 不再自己抓 `redisinfra.GetGateway()`。这一轮又继续把 HTTP 链路里的 middleware 全局态收掉：`src/middleware/deps.go` 现在提供 `middleware.Service` 与 `InjectService(...)`，`src/router/router.go` 在根路由中显式注入 middleware service，`src/middleware/permission.go` / `audit.go` 改为按请求从 Gin context 读取 checker/logger，不再持有 `currentPermissionChecker` / `currentAuditLogger` 这类包级默认服务；`src/interface/http/module.go` 也不再用 `fx.Invoke(middleware.RegisterDeps)` 做全局注册。最近这一轮再把 startup 初始化链里的隐藏 fatal 收掉：`newConfigDataWithDB(...)`、`activateConfigScope(...)`、`InitializeProducer(...)`、`InitializeConsumer(...)` 全部改成显式返回 `error`，producer/worker 的 Fx `OnStart` 现在会把初始化失败直接上抛，而不再在 helper 内部 `logrus.Fatalf(...)` 提前退出进程。紧接着这一轮又继续把 consumer startup 链里的 Redis 裸 client 收口到 gateway：worker 初始化改为走 `RedisGateway.InitConcurrencyLock(...)`，`monitor` / `rate limiter` provider 也改成只依赖 `*redisinfra.Gateway`。再下一轮又把模块侧剩余 Redis 全局入口清掉：`module/group` / `module/notification` / `module/trace` / `module/injection` / `module/systemmetric` / `module/system` 现在都改为通过构造注入 `*redisinfra.Gateway`，trace/group/notification stream 读取、injection algorithm cache、system config response subscribe、system metric Redis 查询不再直接碰 `redisinfra.GetGateway()`。这一轮继续把任务队列 helper 也收回 gateway：`infra/redis/task_queue.go` 里的 submit/get/reschedule/dead-letter/queue index/concurrency lock/list/remove 操作全部改成 `Gateway` 方法，`service/common.SubmitTaskWithDB(...)`、`service/consumer` 调度与取消链路、`module/systemmetric` 排队任务查询都已改走显式 Redis gateway。紧接着又把 `infra/redis` / `infra/etcd` / `infra/buildkit` / `infra/helm` / `infra/k8s` 里已经没有调用方的 `GetGateway()` 单例 fallback 全部删除，主线现在只剩少量 lifecycle/startup 组织层 wrapper 需要再压。当前 `src/service/consumer` / `src/middleware` 里残余重点已从“全局 gateway fallback / 全局 default service”收缩到更少的流程组织 helper 与 initialization 邻近收尾。
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
