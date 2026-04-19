# Backend Refactor TODO

> 更新时间：2026-04-19
> 口径：只保留当前主线状态、验收命令和非阻塞尾项，不再保留逐轮施工日志。

## 1. 当前判断

- [x] Fx + module + infra 主线完成
- [x] `producer / consumer / both` 三种模式可启动
- [x] 六服务入口已落地并可运行
- [x] 旧兼容层已退出主线运行态
- [x] SDK audience / API key 鉴权主线完成
- [x] SDK 路由已统一收口到 `src/router/sdk.go`
- [x] runtime 上传接口并入 `src/router/sdk.go`，并只保留 `RequireServiceTokenAuth()`

结论：当前可按“主线完成”判断，剩余工作主要是非阻塞治理和真实环境补强。

## 2. 主线完成清单

### 2.1 启动与基础设施

- [x] `main.go` 只负责 mode 选择与 Fx 启动
- [x] DB / Redis / Etcd / Tracing / Loki / K8s / Harbor / Helm / BuildKit 已收口到 `src/infra/*`
- [x] HTTP server / worker / controller / receiver 均已纳入 lifecycle
- [x] `src/app` 已按边界拆成基础 options 与服务 options
- [x] `src/interface/grpc/*` 已迁到 `src/interface/grpc/{iam,resource,orchestrator,runtime,system}`

### 2.2 模块边界

- [x] 业务主模块已完成 `module -> service -> repository` 收口
- [x] handler 不再直接依赖全局 DB / 旧 producer / 旧 repository wrapper
- [x] repository 主体回收到各模块 `repository.go`
- [x] 外部系统访问已通过 gateway/store 收口
- [x] middleware 已从旧 producer 依赖切到模块服务 / 独立接口

### 2.3 旧兼容层清理

- [x] `src/service/producer` 已退出生产代码
- [x] `src/handlers/system` 已退出运行态主线
- [x] `database.DB` 已退回 `src/infra/db` 集中管理
- [x] `GetGateway()` / `redisinfra.GetGateway()` / `CurrentK8s*` 这类全局 fallback 已退出主线
- [x] `src/interface/http/router.go`、`src/app/compat_options.go`、`src/router/runtime.go` 这类单层组织文件已继续压缩/删除

### 2.4 路由 / 文档 / SDK / 鉴权

- [x] Public / SDK / Portal / Admin 路由已拆分
- [x] 所有 `@x-api-type {"sdk":"true"}` 运行态入口已统一收口到 `src/router/sdk.go`
- [x] runtime 结果上传接口已作为 `sdk + runtime` 路由并入 `src/router/sdk.go`
- [x] runtime 结果上传接口鉴权已改为仅 `RequireServiceTokenAuth()`
- [x] `portal / admin / sdk / runtime` 四类 audience 路由与 Swagger 标记已完成对齐补扫
- [x] Swagger audience 以 `x-api-type` 为准
- [x] Python SDK 只消费 `sdk.json`
- [x] TypeScript SDK 分别消费 `portal.json` / `admin.json`
- [x] TypeScript OpenAPI Generator 模板目录已压平成 `.openapi-generator/typescript/*`
- [x] `scripts/command` 中 Apifox / SDK / 测试安装 URL 已优先从 `scripts/command/settings.toml` 读取
- [x] API key 主线已统一到 `Key ID / Key Secret` + 签名换 token
- [x] `aegisctl` 与 Python SDK 已切到同一套签名口径

### 2.5 微服务主线

- [x] `api-gateway` 对外 HTTP 入口已形成
- [x] `iam-service` 承接 auth / user / rbac / team / api key
- [x] `resource-service` 承接 project / label / container / dataset / evaluation
- [x] `orchestrator-service` 承接 execution / injection / task / trace / notification / group 控制面
- [x] `runtime-worker-service` 保留 Redis 异步执行链，承接运行态消费、K8s/Helm/BuildKit/Chaos
- [x] `system-service` 承接 config / audit / health / monitor / metrics

## 3. SDK 路由核对结论

已核对 `src/module/*/handler.go` 中所有 `@x-api-type {"sdk":"true"}` 注释，当前运行态路由均由 `src/router/sdk.go` 承接：

- [x] `POST /api/v2/auth/api-key/token`
- [x] `GET /api/v2/sdk/evaluations`
- [x] `GET /api/v2/sdk/evaluations/experiments`
- [x] `GET /api/v2/sdk/evaluations/{id}`
- [x] `GET /api/v2/sdk/datasets`
- [x] `GET /api/v2/datasets/{dataset_id}/versions/{version_id}/download`
- [x] `PATCH /api/v2/datasets/{dataset_id}/version/{version_id}/injections`
- [x] `GET /api/v2/projects/{project_id}/injections`
- [x] `GET /api/v2/projects/{project_id}/injections/analysis/no-issues`
- [x] `GET /api/v2/projects/{project_id}/injections/analysis/with-issues`
- [x] `POST /api/v2/projects/{project_id}/injections/inject`
- [x] `POST /api/v2/projects/{project_id}/injections/build`
- [x] `GET /api/v2/projects/{project_id}/executions`
- [x] `POST /api/v2/projects/{project_id}/executions/execute`
- [x] `POST /api/v2/evaluations/datapacks`
- [x] `POST /api/v2/evaluations/datasets`
- [x] `GET /api/v2/evaluations`
- [x] `GET /api/v2/evaluations/{id}`
- [x] `GET /api/v2/executions/{id}`
- [x] `PATCH /api/v2/executions/{id}/labels`
- [x] `GET /api/v2/injections/metadata`
- [x] `GET /api/v2/injections/{id}`
- [x] `POST /api/v2/injections/{id}/clone`
- [x] `GET /api/v2/injections/{id}/download`
- [x] `GET /api/v2/injections/{id}/files`
- [x] `GET /api/v2/injections/{id}/files/download`
- [x] `GET /api/v2/injections/{id}/files/query`
- [x] `PATCH /api/v2/injections/{id}/labels`
- [x] `GET /api/v2/metrics/algorithms`
- [x] `GET /api/v2/metrics/executions`
- [x] `GET /api/v2/metrics/injections`
- [x] `POST /api/v2/executions/{execution_id}/detector_results`
- [x] `POST /api/v2/executions/{execution_id}/granularity_results`

补充说明：

- `src/module/docs/swagger_models.go` 里的 `sdk` 标记只用于 Swagger model 聚合，不对应独立运行态路由。
- `GET /api/v2/executions/labels` 当前不是 `sdk:true`，所以仍保留在 Portal 侧，不在本次 SDK 收口范围内。

## 4. 验收命令

- [x] 默认回归
  - `cd src && go test ./...`
- [x] Producer Fx 图与 HTTP 冒烟
  - `cd src && go test ./app -run 'TestProducerOptionsValidate|TestProducerOptionsStartStopSmoke|TestProducerOptionsHTTPIntegrationSmoke'`
- [x] Consumer / Both 生命周期冒烟
  - `cd src && go test ./app -run 'TestConsumerOptions|TestBothOptions'`
- [x] 路由 / 文档主路径
  - `cd src && go test ./router ./docs ./interface/http`
- [x] 真实 K8s 集群验收
  - `cd src && RUN_K8S_INTEGRATION=1 go test ./infra/k8s -run TestK8sGatewayJobLifecycleIntegration`

## 5. 主线完成 / 非阻塞剩余项

### 5.1 主线完成

- [x] 单体 Fx 启动主线完成
- [x] 六服务边界主线完成
- [x] 旧兼容层主线完成清扫
- [x] SDK / audience / API key 主线完成
- [x] SDK 路由统一收口完成
- [x] 真实 K8s 集群验收入口完成

### 5.2 非阻塞剩余项

- [ ] 人工确认 Fx 启动日志是否需要进一步裁剪
- [ ] 少量 dedicated service 的 local fallback 还可以继续压窄
- [ ] 少量跨 owner 直查仍可继续按 owner 深清
- [ ] 发布层可继续补 values/HPA/Ingress/镜像策略等环境治理
- [ ] 更贴近真实外部依赖的集成回归仍可继续补强

## 6. 参考文档

- `docs/report-index.md`
- `docs/package-rename-todo.md`
- `docs/api-key-auth-execution-todo.md`
- `docs/python-runtime-wrapper-design.md`
- `docs/python-runtime-wrapper-todo.md`
- `docs/swagger-audience-unmarked-report.md`
