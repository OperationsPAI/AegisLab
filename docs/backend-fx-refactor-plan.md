# AegisLab Backend Fx Refactor Plan

> 创建日期：2026-04-15  
> 状态：Draft  
> 范围：后端模块边界、Fx 依赖装配、生命周期管理、HTTP / worker / controller / receiver 多入口治理

## TL;DR

AegisLab 后端不只是一个 HTTP API 服务。它同时包含：

- HTTP producer server
- background consumer
- scheduler
- K8s controller
- OTLP log receiver
- DB / Redis / Etcd / K8s / Loki / tracing 等基础设施资源

因此当前问题不是单纯缺少依赖注入，而是：

- 模块边界不清
- 全局初始化散落
- 资源生命周期没有统一管理
- HTTP / worker / controller 等入口互相交叉
- handler / service / repository 依赖方向不够硬

结论：**优先采用 Fx，而不是继续沿旧 DI 骨架扩张。**

Fx 在这里的价值不是“自动 new 对象”，而是：

1. 把 app 启动和模块装配收回到 app 层。
2. 用 module 明确业务域和基础设施边界。
3. 用 lifecycle 管理 DB、Redis、HTTP server、consumer、scheduler、receiver、controller 的启动和关闭。
4. 让 producer / consumer / both 三种模式共享基础模块，但启用不同入口。

## 1. Current Problems

### 1.1 全局初始化散落

当前启动流程里存在多个全局初始化点：

- `database.InitDB()`
- `client.InitTraceProvider()`
- `initChaosExperiment()`
- `k8s.GetK8sController()`
- `client.GetRedisClient()`
- `consumer.StartScheduler(ctx)`
- `consumer.ConsumeTasks(ctx)`
- `logreceiver.NewOTLPLogReceiver(...).Start(ctx)`

这些初始化分散在 `main.go`、`client`、`service`、`repository` 等多个包里。结果是：

- 启动顺序靠人工记忆。
- 新人很难判断资源从哪里来。
- 关闭逻辑不统一。
- 测试很难替换基础设施。
- producer / consumer / both 三种模式重复装配逻辑。

### 1.2 分层边界不够硬

期望依赖方向：

```text
cmd
  -> app
    -> interface
      -> module
        -> domain
        -> infra interface
          -> infra implementation
```

当前实际情况：

- handler 直接调用 `service/producer` 包级函数。
- 少数 handler 直接 import `database` / `repository`。
- service 大量直接使用全局 `database.DB`、Redis、K8s、Loki 等 client。
- repository 中混入 Redis queue / token blacklist 等非 DB 能力。
- middleware 直接依赖具体 producer service。

### 1.3 多入口没有统一 app 模型

当前有三种运行模式：

- `producer`: HTTP API server
- `consumer`: background worker / scheduler / K8s controller / receiver
- `both`: 同时启动 producer 和 consumer 能力

这些模式本质上应该是三套 Fx option：

```text
CommonOptions + ProducerOptions
CommonOptions + ConsumerOptions
CommonOptions + ProducerOptions + ConsumerOptions
```

而不是在 `main.go` 中手写多份初始化流程。

## 2. Target Module Boundary

先确定模块边界，再谈 Fx 注入。

### 2.1 App Layer

职责：

- 程序启动入口
- Fx app 创建
- producer / consumer / both option 选择
- 生命周期统一管理
- graceful shutdown

建议目录：

```text
src/app/
  app.go
  options.go
  producer.go
  consumer.go
  both.go
```

### 2.2 Interface Layer

职责：

- HTTP / Gin router
- middleware
- handler
- worker entry
- scheduler entry
- K8s controller entry
- OTLP receiver entry

建议目录：

```text
src/interface/
  http/
    module.go
    router.go
    routes_public.go
    routes_sdk.go
    routes_portal.go
    routes_admin.go
    routes_system.go
  worker/
    module.go
    consumer.go
    scheduler.go
  controller/
    module.go
    k8s.go
  receiver/
    module.go
    otlp.go
```

过渡期可以先不移动现有 `handlers/`、`router/`、`service/consumer/` 文件，只在 Fx module 中包装它们。

### 2.3 Business Module Layer

按业务域拆模块，每个模块只暴露 `Module`、`NewService`、`NewHandler`、`NewRepository`、必要接口。

建议业务模块：

```text
src/module/
  auth/
  user/
  rbac/
  team/
  project/
  container/
  dataset/
  injection/
  execution/
  task/
  evaluation/
  trace/
  metrics/
  notification/
  audit/
  system/
  dynamicconfig/
```

每个模块的目标形态：

```go
var Module = fx.Module("project",
    fx.Provide(
        NewRepository,
        NewService,
        NewHandler,
    ),
)
```

### 2.4 Domain Layer

职责：

- 核心业务规则
- domain entity / value object
- 纯逻辑校验
- 不依赖 Gin、Gorm、Redis、K8s

建议目录：

```text
src/domain/
  project/
  task/
  injection/
  execution/
  permission/
```

过渡期可以先继续使用 `database` entity 和 `dto`，等模块稳定后再抽 domain。

### 2.5 Infra Layer

职责：

- 配置
- 日志
- DB
- Redis
- Etcd
- K8s
- Loki
- Jaeger / tracing
- Harbor
- Helm
- BuildKit
- Chaos client

建议目录：

```text
src/infra/
  config/
  logger/
  db/
  redis/
  etcd/
  k8s/
  loki/
  tracing/
  harbor/
  helm/
  buildkit/
  chaos/
```

每个 infra module 要明确：

- 创建什么资源
- 返回什么接口或 client
- 是否需要 `fx.Lifecycle`
- `OnStart` 做什么
- `OnStop` 做什么

## 3. Dependency Rules

### 3.1 允许依赖

```text
cmd -> app
app -> interface / module / infra
interface -> module service interface
module -> domain / infra interface
infra implementation -> external libraries
```

### 3.2 禁止依赖

```text
domain -> gin / gorm / redis / k8s
repository -> handler
repository -> service
handler -> database.DB
handler -> repository implementation
middleware -> concrete producer service
business module -> another module's implementation
```

跨业务模块调用优先依赖接口。例如 project 需要 RBAC 能力：

```go
type PermissionChecker interface {
    CheckUserPermission(ctx context.Context, params *dto.CheckPermissionParams) (bool, error)
}
```

由 rbac module 提供实现。

## 4. Fx App Design

### 4.1 Common Options

所有模式共享：

```go
func CommonOptions() fx.Option {
    return fx.Options(
        config.Module,
        logger.Module,
        db.Module,
        redis.Module,
        tracing.Module,
        etcd.Module,
        BusinessModules(),
    )
}
```

### 4.2 Producer Options

HTTP server 模式：

```go
func ProducerOptions() fx.Option {
    return fx.Options(
        CommonOptions(),
        http.Module,
    )
}
```

### 4.3 Consumer Options

后台任务模式：

```go
func ConsumerOptions() fx.Option {
    return fx.Options(
        CommonOptions(),
        k8s.Module,
        chaos.Module,
        worker.Module,
        controller.Module,
        receiver.Module,
    )
}
```

### 4.4 Both Options

本地或一体化部署模式：

```go
func BothOptions() fx.Option {
    return fx.Options(
        CommonOptions(),
        k8s.Module,
        chaos.Module,
        http.Module,
        worker.Module,
        controller.Module,
        receiver.Module,
    )
}
```

### 4.5 main.go 目标形态

```go
func main() {
    mode := parseMode()

    var opts fx.Option
    switch mode {
    case "producer":
        opts = app.ProducerOptions()
    case "consumer":
        opts = app.ConsumerOptions()
    case "both":
        opts = app.BothOptions()
    }

    fx.New(opts).Run()
}
```

`main.go` 不再直接初始化 DB、Redis、K8s controller、HTTP server、scheduler。

## 5. Lifecycle Plan

Fx lifecycle 应统一管理这些资源：

### 5.1 DB

- `fx.Provide(NewGormDB)`
- `OnStop`: close underlying sql DB

### 5.2 Redis

- `fx.Provide(NewRedisClient)`
- `OnStop`: `Close()`

### 5.3 HTTP Server

- `fx.Provide(NewGinEngine, NewHTTPServer)`
- `OnStart`: `server.ListenAndServe()` in goroutine
- `OnStop`: `server.Shutdown(ctx)`

### 5.4 K8s Controller

- `fx.Provide(NewK8sController)`
- `OnStart`: start controller in goroutine
- `OnStop`: cancel controller context

### 5.5 Worker / Scheduler

- `fx.Provide(NewTaskConsumer, NewScheduler)`
- `OnStart`: start goroutines
- `OnStop`: cancel context and wait if needed

### 5.6 OTLP Receiver

- `fx.Provide(NewOTLPReceiver)`
- `OnStart`: start receiver
- `OnStop`: shutdown receiver

### 5.7 Tracing

- `fx.Provide(NewTraceProvider)`
- `OnStop`: flush / shutdown provider if supported

## 6. HTTP Boundary Plan

HTTP routes should be split by audience, not by current file size.

### 6.1 Public

- login
- register
- refresh
- health
- docs

### 6.2 SDK

Stable programmatic API:

- project list / get / create
- container / dataset / version query
- submit injection / build / execution
- task status / logs
- injection / execution / evaluation query
- metrics query
- datapack download / query

### 6.3 Portal

普通登录用户前端页面 API：

- profile
- teams / projects
- labels
- notifications
- user-scoped container / dataset / injection / execution
- upload / download / query

### 6.4 Admin

系统管理 API：

- users
- roles
- permissions
- resources
- audit
- system configs
- global injections / executions
- chaos systems
- batch delete

第一阶段只拆注册函数，不改 URL。

## 7. Migration Strategy

### Phase 0: Stop Legacy DI Expansion

当前已有的旧 DI 骨架可以视为短期试验。后续不要继续沿旧 provider 骨架深挖。

处理方式：

- 暂时保留也可以，避免立即制造回滚噪音。
- 开始引入 Fx 后，用 Fx app 替换 `app.InitializeProducerApp()`。
- 最终删除旧 DI 骨架相关文件和依赖。

### Phase 1: Add Fx Skeleton

目标：引入 Fx，但不重写业务逻辑。

任务：

- 增加 `go.uber.org/fx`
- 新建 `app` Fx options
- 新建 `infra/config`、`infra/logger`、`infra/db` 的 module 草案
- 先包装现有 `config.Init`、`database.InitDB`
- 保持现有 router / handler / service 行为

验收：

- producer 可以通过 Fx 启动
- consumer 旧逻辑暂不迁移或只包装
- 现有 API 路由不变

### Phase 2: Move Lifecycle Into Fx

目标：把启动和停止资源收回 app。

迁移顺序：

1. DB lifecycle
2. Redis lifecycle
3. tracing lifecycle
4. HTTP server lifecycle
5. OTLP receiver lifecycle
6. scheduler lifecycle
7. consumer lifecycle
8. K8s controller lifecycle

验收：

- `main.go` 不再手写资源启动顺序
- producer / consumer / both 使用不同 Fx options
- 资源关闭有 `OnStop`

### Phase 3: Module Boundary Wrapper

目标：先建立业务 module 壳，不急着重写内部逻辑。

优先模块：

1. project
2. auth
3. task
4. injection
5. execution

每个模块先暴露：

```go
var Module = fx.Module("project",
    fx.Provide(NewHandler),
)
```

如果 service / repository 尚未 struct 化，可以先由 handler wrapper 调旧函数。

验收：

- router 依赖 module handler
- 新模块入口清晰
- 旧包级函数逐步减少

### Phase 4: Structify Service / Repository

目标：逐个业务模块把包级函数改成 struct。

每个模块执行：

- `Repository` struct 化
- `Service` struct 化
- `Handler` struct 化
- service 注入 repository / store / gateway
- handler 注入 service
- 移除 handler direct repository / database import
- 移除 service direct global DB usage

验收：

- 新迁移模块可单测
- 依赖从 Fx 图中可见
- 无新增全局 client 访问

### Phase 5: Split Store / Gateway

目标：把基础设施访问从 repository / service 中抽出。

- Redis token blacklist -> `infra/redis` 或 `module/auth.TokenStore`
- Redis task queue -> `module/task.QueueStore`
- Loki -> `infra/loki.Gateway`
- K8s -> `infra/k8s.Gateway`
- Etcd -> `infra/etcd.Client`
- Harbor / Helm / BuildKit -> gateway

验收：

- repository 只处理 DB
- 外部系统访问都有接口边界

### Phase 6: SDK / Portal / Admin Governance

目标：让 API 受众边界和 SDK 生成一致。

- 拆 route registration
- 审核 OpenAPI3 `x-api-type` audience 扩展
- 修正 `sdk / portal / admin` 归属
- 更新 SDK 生成脚本

## 8. First PR Scope

第一批建议只做：

1. 新增 Fx 依赖。
2. 新增 `app` Fx options。
3. 新增 `infra/config`、`infra/db`、`interface/http` module 壳。
4. producer 模式通过 Fx 启动 HTTP server。
5. 保持业务 handler/service/repository 不动。
6. 标记当前旧 DI 骨架文件为待删除，或直接在本 PR 中移除旧骨架。

第一批不建议做：

- 迁移所有 service
- 搬目录
- 改 URL
- 清理所有 SDK 标记
- 重写 consumer
- 重写 repository

## 9. Completion Criteria

最终完成后应满足：

1. `main.go` 只负责解析 mode 和启动 Fx app。
2. producer / consumer / both 由 Fx options 组合。
3. DB / Redis / HTTP / worker / receiver / controller 都有 lifecycle。
4. HTTP routes 按 Public / SDK / Portal / Admin / System 拆分。
5. handler 不直接 import `database` / repository implementation。
6. service 不直接使用全局 `database.DB`。
7. repository 不操作 Redis / K8s / Loki / Etcd。
8. business module 之间依赖接口，不依赖实现。
9. 新模块只需要暴露 `Module` 和构造函数。
10. SDK 只包含稳定外部 API。

## 10. Open Questions

1. 是否要物理移动目录到 `module/`、`infra/`、`interface/`，还是先保持旧目录、只用 Fx module 约束？
2. `service/producer` / `service/consumer` 是否改名？
3. Redis task queue 放在 `infra/redis` 还是 `module/task`？
4. Permission checker 接口归属 `module/rbac` 还是 `interface/http/middleware`？
5. 是否在本轮移除已有旧 DI 骨架，还是等 Fx producer 跑通后再删？

建议：先不做大规模目录迁移。第一阶段用 Fx module 包装旧代码，等启动生命周期稳定后，再逐个业务模块搬迁。
