# Model / DTO Refactor TODO

> 创建日期：2026-04-17  
> 目标：把原 `database` 语义收缩为持久化模型层 `model`，并把当前全局 `dto` 逐步拆回各模块，避免存储模型和接口契约继续混在一起。

## 设计原则

- `src/model` 只放持久化模型、GORM hook、scanner / valuer、只读 view model。
- `src/infra/db` 负责连接、迁移、生命周期、view 创建。
- 不把 `dto` 直接并入 `model`。
- `dto` 优先按模块下沉到 `src/module/*`，只保留极少数真正跨模块共享类型。
- 先改命名和目录边界，再做更细的 DTO 下沉，避免一轮里同时改太多语义。

## 阶段 1：`database` -> `model`

- [x] 新建 `src/model`
- [x] 将 `src/database/*` 迁到 `src/model/*`
- [x] 将包名从 `database` 改为 `model`
- [x] 批量更新仓库内 `aegis/database` import
- [x] 批量更新 `database.*` 类型引用
- [x] 跑主链测试确认编译通过
  - 已执行：`cd src && go test ./app -count=1`
  - 已执行：`cd src && go test ./module/... ./router/... ./repository ...`
  - 备注：沙箱内执行 `cd src && go test ./...` 时，`app` 中两条 loopback smoke test 在整仓并行场景下触发 `listen tcp 127.0.0.1:0: socket: operation not permitted`，单包执行通过，属于环境限制而非本轮重命名回归。

## 阶段 2：继续压缩 `src/model`

- [x] 复查 `src/model` 是否只剩实体 / view model / scanner / valuer
- [x] 将模块专用读模型从 `src/model` 下沉回对应模块
- [x] 优先处理 SDK 只读模型
  - `src/model/sdk_entities.go` 已删除
  - SDK 只读模型已迁到 `src/module/sdk/models.go`

## 阶段 3：拆全局 `dto`

- [x] 明确 `dto` 中每个文件对应的模块归属
  - 当前剩余 `src/dto/*` 已收敛为共享分页/搜索/响应壳与跨模块运行时载荷：`common.go`、`response.go`、`search.go`、`permission.go`、`project.go`、`dynamic_config.go`、`container.go`、`dataset.go`、`injection.go`、`task.go`、`trace.go`、`log.go`、`label.go`
- [x] 优先试点 `auth` / `sdk` / `system`
  - `src/module/auth/api_types.go` 已落地，`src/dto/auth.go` 已删除
  - `src/module/sdk/api_types.go` / `src/module/sdk/models.go` 已落地，`src/dto/sdk_evaluation.go` 已删除
  - `src/module/system/api_types.go` 已落地，`src/dto/audit.go` 已删除，并缩减 `src/dto/system.go` / `src/dto/dynamic_config.go`
- [x] 再推进下一批明显模块内聚 DTO
  - `src/module/chaossystem/api_types.go` 已落地，`src/dto/chaos_system.go` 已删除
  - `src/module/team/api_types.go` 已落地，`src/dto/team.go` 已删除
  - `src/module/label/api_types.go` 已落地，`src/dto/label.go` 已裁剪为仅保留共享 `LabelItem`
  - `src/module/rbac/api_types.go` 已落地，`src/dto/resource.go` 已删除，`src/dto/role.go` 已裁剪掉 role mutation/list 请求类型
  - `src/module/user/api_types.go` 已落地，`src/dto/user.go` 已裁剪掉 user CRUD/detail 请求响应类型
  - `src/module/project/api_types.go` 已落地，`src/dto/project.go` 已裁剪为仅保留共享 search/statistics 结构
  - `src/module/dataset/api_types.go` 已落地，`src/dto/dataset.go` 已裁剪掉 dataset CRUD/detail/label 管理请求响应类型
  - `src/module/container/api_types.go` 已落地，`src/dto/container.go` 已裁剪掉 container CRUD/detail/label 管理请求响应类型
  - `src/module/evaluation/api_types.go` 已落地，`src/dto/evaluation.go` 已删除，并把批量评估逻辑收回 `src/module/evaluation/service.go`
  - `src/module/metric/api_types.go` 已落地，`src/dto/metrics.go` 已删除
  - `src/module/task/api_types.go` 已落地，`src/dto/task.go` 已裁剪掉 task list/batch-delete/detail/queue 这批模块内 API 类型；trace 仍复用共享 `TaskResp`
  - `src/module/notification/api_types.go` 已落地，`src/dto/notification.go` 已删除
  - `src/module/group/api_types.go` 已落地，`src/dto/group.go` 已删除，并把 group stats/stream 相关类型从 `src/dto/trace.go` 收回模块
  - `src/module/trace/api_types.go` 已落地，`src/dto/trace.go` 已裁剪掉 trace list/detail/stream 请求响应类型；`src/repository/trace.go` 也已去掉对 `dto.ListTraceFilters` 的依赖
  - `src/module/systemmetric/api_types.go` 已落地，`src/dto/system.go` 已删除；system 通过模块别名复用监控响应类型
  - `src/module/execution/api_types.go` 已落地，`src/dto/execution.go` 已删除；evaluation 改为直接复用 execution 模块公开执行引用类型
  - `src/module/execution/result_types.go` 已落地，执行结果上传请求/响应与 detector / granularity 结果项已迁回模块，`src/dto/algorithm_result.go` 已删除
  - `src/module/rbac/api_types.go` 已继续扩充 role / permission 响应与 permission list 查询契约，`src/dto/role.go` 已删除，`src/dto/permission.go` 已裁剪为仅保留 middleware / repository 共享的 `CheckPermissionParams`
  - `src/module/injection/api_types.go` 已落地，`src/dto/injection.go` 已裁剪为仅保留 consumer / task 共享的 `InjectionItem`
  - `src/module/injection/time_range.go` 已落地，注入分析查询时间窗契约已迁回模块，`src/dto/request.go` 已删除
  - `src/module/dataset/api_types.go` 已继续接管 search / dataset version / datapack relation 契约，`src/dto/dataset.go` 已裁剪为仅保留共享 `DatasetRef`
  - `src/module/auth/api_types.go` 已接管 profile 响应契约，`src/dto/user.go` 已删除未使用的 `UserSearchReq` 并移出 `UserProfileResp`
  - `src/module/user/api_types.go` 已继续接管 permission assignment / resource-role 视图契约，`src/dto/user.go` 已删除
  - `src/dto/project.go` 已删除未使用的 `SearchProjectReq`，当前仅保留 project/team 共用的 `ProjectStatistics`
  - `src/dto/dynamic_config.go` 已删除未使用的 `ConfigStatsResp`，当前仅保留跨 `service/common` / `module/system` 共用的 `ConfigUpdateResponse`
  - `src/module/task/log_types.go` 已落地，任务日志 WebSocket 消息已迁回模块，`src/dto/log.go` 已裁剪为仅保留共享 `LogEntry`
  - container 构建请求已直接复用共享 `dto.BuildOptions`，模块内重复定义已删除
  - 未被引用的遗留全局 DTO 已继续清理：`src/dto/analyzer.go`、`src/dto/debug.go`、`src/dto/redis.go` 已删除；`src/dto/trace.go` 中未使用的 `TraceQuery` 已移除
- [x] 将模块专用 request / response 移到 `src/module/*`
  - Auth 请求/响应类型已迁到 `src/module/auth/api_types.go`
  - SDK 请求/响应类型已迁到 `src/module/sdk/api_types.go`
  - System 请求/响应类型已迁到 `src/module/system/api_types.go`
  - ChaosSystem 请求/响应类型已迁到 `src/module/chaossystem/api_types.go`
  - Team 请求/响应类型已迁到 `src/module/team/api_types.go`
  - Label 请求/响应类型已迁到 `src/module/label/api_types.go`
  - RBAC 的 role/resource 请求类型与 resource 响应类型已迁到 `src/module/rbac/api_types.go`
  - User 的 CRUD/detail 请求响应类型已迁到 `src/module/user/api_types.go`
  - Project 的 CRUD/detail/label 管理请求响应类型已迁到 `src/module/project/api_types.go`
  - Dataset 的 CRUD/detail/label 管理请求响应类型已迁到 `src/module/dataset/api_types.go`
  - Container 的 CRUD/detail/label 管理请求响应类型已迁到 `src/module/container/api_types.go`
  - Evaluation 的 list/detail/batch evaluate 请求响应类型已迁到 `src/module/evaluation/api_types.go`
  - Metric 的 query/response 类型已迁到 `src/module/metric/api_types.go`
  - Task 的 list/batch-delete/detail/queue 请求响应类型已迁到 `src/module/task/api_types.go`
  - Notification 的 stream 请求/事件类型已迁到 `src/module/notification/api_types.go`
  - Group 的 stats/stream 请求响应类型已迁到 `src/module/group/api_types.go`
  - Trace 的 list/detail/stream 请求响应类型已迁到 `src/module/trace/api_types.go`
  - SystemMetric 的 metrics/namespace-lock 请求响应类型已迁到 `src/module/systemmetric/api_types.go`
  - Execution 的 list/detail/submit/batch-delete 请求响应类型已迁到 `src/module/execution/api_types.go`
  - Execution 的 detector/granularity 结果上传请求响应类型已迁到 `src/module/execution/result_types.go`
  - RBAC 的 role / permission 响应类型与 permission list 请求类型已迁到 `src/module/rbac/api_types.go`
  - Injection 的 list/search/submit/build/label/file/upload 请求响应类型已迁到 `src/module/injection/api_types.go`
  - Injection 的时间窗查询类型已迁到 `src/module/injection/time_range.go`
  - Dataset 的 search / version CRUD / datapack relation 请求响应类型已迁到 `src/module/dataset/api_types.go`
  - Auth 的 profile 响应类型已迁到 `src/module/auth/api_types.go`
  - User 的 permission assignment / resource relation 响应类型已迁到 `src/module/user/api_types.go`
- [x] 保留一个极薄的跨模块共享 DTO 层，避免继续养大而全 `dto`
  - 当前共享 DTO 只保留分页/搜索/统一响应、权限检查参数，以及 consumer / runtime / trace / log 等跨模块载荷

## 当前进展

- [x] `auth` 模块已完成本地 API 类型收口并通过校验
  - 已执行：`cd src && go test ./module/auth ./router ./docs`
- [x] `sdk` / `system` 模块已完成前序下沉并继续保持通过
  - 已执行：`cd src && go test ./module/system ./module/sdk ./module/auth ./router ./docs`
- [x] `chaossystem` 模块已完成本地 API 类型下沉并通过校验
  - 已执行：`cd src && go test ./module/chaossystem ./router ./docs`
- [x] `team` 模块已完成本地 API 类型下沉并通过校验
  - 已执行：`cd src && go test ./module/team ./router ./docs`
- [x] `label` 模块已完成本地 API 类型下沉并通过校验
  - 已执行：`cd src && go test ./module/label ./router ./docs`
- [x] `rbac` 模块已完成一轮本地 API 类型下沉并通过校验
  - 已执行：`cd src && go test ./module/rbac ./router ./docs`
- [x] `user` 模块已完成一轮本地 API 类型下沉并通过校验
  - 已执行：`cd src && go test ./module/user ./module/rbac ./router ./docs`
- [x] `project` 模块已完成一轮本地 API 类型下沉并通过校验
  - 已执行：`cd src && go test ./module/project ./module/team ./router ./docs`
- [x] `dataset` 模块已完成一轮本地 API 类型下沉并通过校验
  - 已执行：`cd src && go test ./module/dataset ./module/project ./router ./docs`
- [x] `container` 模块已完成一轮本地 API 类型下沉并通过校验
  - 已执行：`cd src && go test ./module/container ./module/project ./router ./docs`
- [x] `evaluation` 模块已完成一轮本地 API 类型下沉并通过校验
  - 已执行：`cd src && go test ./module/evaluation ./router ./docs`
- [x] `metric` / `task` 模块已完成一轮本地 API 类型下沉并通过校验
  - 已执行：`cd src && go test ./module/metric ./module/task ./module/systemmetric ./module/system ./router ./docs`
- [x] `notification` / `group` 模块已完成一轮本地 API 类型下沉并通过校验
  - 已执行：`cd src && go test ./module/notification ./module/group ./service/consumer ./module/docs ./router ./docs`
- [x] `trace` 模块已完成一轮本地 API 类型下沉并通过校验
  - 已执行：`cd src && go test ./module/trace ./router ./docs`
- [x] `systemmetric` / `system` 模块已完成一轮本地 API 类型下沉并通过校验
  - 已执行：`cd src && go test ./module/systemmetric ./module/system ./router ./docs`
- [x] `execution` / `rbac` / `user` 模块已继续完成一轮本地 API 类型收缩并通过校验
  - 已执行：`cd src && go test ./module/execution ./module/rbac ./module/user ./module/evaluation ./router ./docs`
- [x] `injection` / `dataset` / `project` / `auth` 模块已继续完成一轮本地 API 类型收缩并通过校验
  - 已执行：`cd src && go test ./module/auth ./module/injection ./module/dataset ./module/project ./router ./docs`
- [x] `execution` / `injection` / `label` / `task` / `container` 已继续完成最后一轮共享 DTO 收缩并通过校验
  - 已执行：`cd src && go test ./module/execution ./module/evaluation ./module/injection ./module/label ./module/container ./module/task ./router ./docs`
- [x] 继续按模块清点 `src/dto/*` 中剩余仅被单模块消费的类型
- [x] 再清一轮已空心化 `repository` / helper 边界壳
  - `src/repository/project.go`、`src/repository/user.go` 已删除；相关 project/user 访问已完全由模块仓储接管
  - `src/module/injection/repository.go` 已删除仅做 label item 转条件的空包装，service 直接传递 label condition
  - 已执行：`cd src && go test ./module/project ./module/user ./module/injection ./module/team ./repository ./router ./docs`
- [x] 继续删旧仓储中已无人引用的模块专用壳文件
  - `src/repository/system.go`、`src/repository/evaluation.go`、`src/repository/role.go`、`src/repository/resource.go`、`src/repository/permission.go`、`src/repository/team.go` 已删除
  - 这些能力已分别由 `src/module/chaossystem`、`src/module/evaluation`、`src/module/rbac`、`src/module/team` 或 middleware / initialization 内聚实现接管
  - 已执行：`cd src && go test ./module/chaossystem ./module/evaluation ./module/rbac ./middleware ./service/initialization ./repository ./router ./docs`
  - 已执行：`cd src && go test ./module/team ./middleware ./service/initialization ./repository ./router ./docs`
- [x] 再收一轮 `service/common` / `service/consumer` 直连旧仓储 helper
  - `src/repository/dynamic_config.go`、`src/repository/task.go`、`src/repository/trace.go`、`src/repository/system_metadata.go` 已删除
  - 配置创建、task/trace upsert、trace 查询、system metadata 查询已分别内聚回 `src/service/common` / `src/service/consumer`
  - 当前 `src/repository/*` 仅剩 container/dataset/execution/injection/label/search builder 等跨模块共享查询能力
  - 已执行：`cd src && go test ./service/common ./service/consumer ./module/system ./module/group ./module/trace ./repository ./router ./docs`
- [x] 最后一轮共享层命名 / 文件抛光
  - `src/repository/common.go` 已删除，剩余共享仓储不再保留无语义公共常量文件
  - container/dataset/injection 共享仓储里的 `active_name` omit 常量已改成各文件自解释命名
  - 修正残余命名/注释噪音：如 `contaierType`、`BatchDelteInjections`
  - 已执行：`cd src && go test ./repository ./router ./docs ./service/common ./service/consumer`

## 边界口径

- [x] `src/model` 继续只承载持久化实体 / view model / scanner / valuer
- [x] 跨模块共享的 API 请求/响应暂不并入 `src/model`
  - 原因：共享 DTO 仍属于接口契约层，不是持久化模型；直接并入 `model` 会重新把存储边界和 HTTP/API 边界混在一起
  - 后续方向：继续缩小 `src/dto`，必要时再拆成更明确的共享契约包，而不是回灌到 `model`
  - 当前保留例子：`src/dto/trace.go` 仍保留 trace 自身 stream/list/detail 契约；group 侧统计/stream DTO 已拆回 `src/module/group`
  - 更新：`src/dto/trace.go` 现在只保留 trace stream 事件负载等共享结构，trace handler/service 自身契约已迁回模块，未使用 `TraceQuery` 已删除
  - 更新：`src/dto/task.go` 现在只保留 `UnifiedTask` 这类调度/运行时共享结构；原先重复保留的 `TaskResp` 已删除，trace 直接复用 `src/module/task/api_types.go`
  - 更新：`src/dto/log.go` 现在只保留 Loki / OTLP / task log 共用的 `LogEntry`；WebSocket 消息壳已迁回 `src/module/task/log_types.go`

## 当前决定

- [x] DB 初始化、迁移、生命周期已转入 `src/infra/db`
- [x] `scope` 查询辅助已从原 `database` 迁到 `src/repository`
- [x] 明确不采用“把 `dto` 并入 `model`”方案
- [x] 完成第一阶段目录重命名
- [x] DTO / model 主线重构已完成
  - 当前保留的 `src/repository/*` 主要是 consumer / service/common / metadata / search builder 等跨模块共享查询能力，不再属于本轮“模块专用旧壳”
  - 后续若继续做，只剩增量优化，不再是本轮主线阻塞项
