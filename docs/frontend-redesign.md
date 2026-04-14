# AegisLab Frontend Redesign Specification

> Version: 1.1 | Date: 2026-04-13 (updated)
> Status: Implementation Complete — all REQ-812~817 implemented, pending user acceptance

## 1. Design Principles

1. **Flow-first**: UI 围绕核心 pipeline 流程设计，而非 CRUD 表格堆叠
2. **Consistent & restrained**: 统一 Ant Design 用法，不混用风格，克制装饰性元素
3. **Zero mock**: 所有前端组件必须对接真实 API，不允许硬编码数据
4. **Backend-aligned**: 前端数据模型与后端 DTO 一一对应，通过 TypeScript SDK 桥接

## 2. Target Users

**RCA 研究人员** — 核心工作流程：创建实验 → 注入故障 → 生成数据包 → 运行算法 → 查看结果 → 评估对比。

需要关注：
- 实验配置的便捷性（选择系统、配置故障、选算法）
- Pipeline 执行状态的实时可见性
- 结果的快速查阅

## 3. Core Workflows

AegisLab 有两条核心工作流，面向研究人员的不同阶段。

### 3.1 Workflow A — Full Pipeline (故障注入 + 自动运行算法)

**场景**: 研究人员需要生成新的故障数据，并立即用算法分析。

```
┌─────────────┐   ┌──────────────┐   ┌──────────────┐   ┌──────────────┐   ┌──────────────┐
│ Configure   │──>│ Inject Fault │──>│ Build        │──>│ Run          │──>│ Collect      │
│ Experiment  │   │ (CRD)        │   │ Datapack     │   │ Algorithm    │   │ Result       │
└─────────────┘   └──────────────┘   └──────────────┘   └──────────────┘   └──────────────┘
     USER              AUTO               AUTO               AUTO               AUTO
```

`SubmitInjectionReq` 已包含 `algorithms` 字段（`src/dto/injection.go:370`），提交时可同时指定要运行的算法。后端 CRD callback 链自动驱动：

```
FaultInjection CRD → HandleCRDSucceeded → BuildDatapack Job → HandleJobSucceeded → RunAlgorithm Job → HandleJobSucceeded → CollectResult
```

**前端入口**: `/:teamName/:projectName/injections/create`
**前端设计要点**：创建实验时一次性配好 Pedestal + Benchmark + Fault Specs + Algorithms，提交后全自动流转，用户只需在 detail 页观察进度。

### 3.2 Workflow B — Algorithm Benchmark (算法测试)

**场景**: 研究人员开发了一个新的 RCA 算法，需要在已有的 Datapacks 或 Dataset 上批量测试，验证算法效果。

```
┌─────────────┐   ┌───────────────────┐   ┌──────────────┐   ┌──────────────┐
│ Select      │──>│ Select Datapacks  │──>│ Run          │──>│ Collect &    │
│ Algorithm   │   │ or Dataset        │   │ Algorithm    │   │ Compare      │
└─────────────┘   └───────────────────┘   └──────────────┘   └──────────────┘
     USER                USER                   AUTO               AUTO
```

**后端支持**: `SubmitExecutionReq` (`src/dto/execution.go:187`)

```typescript
{
  project_name: string,
  specs: ExecutionSpec[],  // 数组 — 一次提交多个执行
  labels: LabelItem[],
}
```

每个 `ExecutionSpec` 包含：
- `algorithm`: `{ name, version }` — 要测试的算法
- `datapack`: `string` — 单个 datapack 名称 (与 dataset 二选一)
- `dataset`: `{ name, version }` — 整个数据集 (与 datapack 二选一)

**用法**:
- **单算法 × 多 Datapacks**: 同一个 algorithm，生成多个 spec，每个 spec 指向不同 datapack
- **单算法 × 整个 Dataset**: 一个 spec，algorithm + dataset，后端自动展开到该 dataset 下所有 datapacks
- **多算法对比**: 每个算法 × 每个 datapack 生成一个 spec

**前端入口**: `/:teamName/:projectName/executions/new`
**提交后**: 返回 `group_id`，可通过 `GET /groups/:group_id/stream` (SSE) 实时追踪批量执行进度。

> **Implementation Status**: `ExecutionCreatePage.tsx` 已实现 Datapack/Dataset 双模式 (REQ-814)。提交后 `group_id` 通过 URL query param 传递到 Execution List 页 (REQ-815)。

## 4. Information Architecture

### 4.1 Page Structure (simplified)

```
/login                                  → 登录

/home                                   → Dashboard (简化)
/projects                               → 项目列表
/projects/new                           → 创建项目

/:teamName/:projectName                 → 项目概览 (Pipeline 状态一览)
/:teamName/:projectName/injections      → Datapack 列表 (注入产生的数据包)
/:teamName/:projectName/injections/create → 创建实验 (配置注入 + 算法)
/:teamName/:projectName/injections/:id  → Datapack 详情 (Pipeline 视图)
/:teamName/:projectName/executions      → 算法执行列表
/:teamName/:projectName/executions/:id  → 执行详情 (结果查看)
/:teamName/:projectName/executions/new  → 手动创建执行
/:teamName/:projectName/evaluations     → 评估列表
/:teamName/:projectName/evaluations/:id → 评估详情
/:teamName/:projectName/settings        → 项目设置

/:teamName                              → 团队概览
/:teamName/projects                     → 团队项目
/:teamName/users                        → 团队成员
/:teamName/settings                     → 团队设置

/profile                                → 个人信息 (精简)
/settings                               → 用户设置
/tasks                                  → 全局任务队列 (高级/调试)

/admin/users                            → 用户管理
/admin/containers                       → 容器管理 (算法、Benchmark、Pedestal)
/admin/datasets                         → 数据集管理
/admin/system                           → 系统设置
```

### 4.2 Removed Pages/Features

| Feature | Reason |
|---------|--------|
| `/notifications` + `NotificationsPage` | 非核心，通知功能删除 |
| Star/收藏功能 | 后端无 API，前端 stub 删除 |
| `ProfilePage` 的 `StarsTab` | 依赖不存在的 Star API |
| `ProfilePage` 的 `ActivityGraph` | 无真实数据源 |
| `SystemMetrics.tsx` (硬编码组件) | 100% mock 数据，HomePage 已用真实 API 替代 |
| `QuickActions.tsx`, `ActivityFeed.tsx` | 未实际接入 API，作为 dead code 清理 |

### 4.3 Container Registration (研究人员可用)

研究人员需要注册自己的 RCA 算法容器。Container 管理目前在 `/admin/containers`，但需要分两级暴露：

- **Admin**: 管理所有类型 Container (algorithm, benchmark, pedestal) — 保持 `/admin/containers`
- **Researcher**: 仅 `type=algorithm` 的 Container 和 ContainerVersion 的注册/管理 — 在 Workspace 内新增入口

具体做法：在 WorkspaceSidebar 中增加 "Algorithms" 入口，对应路由 `/:teamName/:projectName/algorithms`，复用 Container CRUD API 但 filter `type=algorithm`。

## 5. Page-by-Page Design

### 5.1 Login (`/login`)

**现状**: 基本可用。
**改动**: 无大改。确保样式与全局主题统一。

### 5.2 Home Dashboard (`/home`)

**现状**: 有 Welcome + Quick Actions + Metrics Cards + Recent Projects + Getting Started guide。
**问题**:
- Getting Started 静态文本对老用户无用
- Metrics 卡片信息密度低

**改动**:
- 保留 Welcome + Recent Projects + Metrics 概览
- 删除 Getting Started 卡片（或仅在用户无项目时显示）
- Metrics 卡片合并精简：只展示 Total Injections、Total Executions、System Status
- 删除 Execution Metrics Summary 的 Running/Completed/Failed 行 (信息冗余)
- 接入真实 system API（当前 HomePage 已部分接入，验证是否完整）

### 5.3 Project List (`/projects`)

**现状**: Table/Card 切换，有搜索。
**问题**: Star toggle 是 no-op。
**改动**:
- 删除 Star toggle
- 保持 table 列表视图
- 确保 `totalExperiments` 统计接入真实 API（当前 hardcoded to 0）

### 5.4 Project Overview (`/:teamName/:projectName`)

**现状**: 项目概览，有 summary statistics。
**问题**: 缺乏 pipeline 进度视图。

**改动 — 核心页面重设计**:

这是项目的 landing page，应该让研究员一眼看到：
1. **项目统计摘要**: Injections count, Executions count, Datapacks count
2. **最近 Pipeline 活动**: 最近 N 条注入及其 pipeline 状态（参见 5.6 Pipeline 状态组件）
3. **Quick Action**: "New Experiment" 按钮跳转到 injection create

### 5.5 Experiment Creation (`/:teamName/:projectName/injections/create`)

**现状**: 多步表单 (InjectionCreate) with VisualCanvas, FaultConfigPanel, FaultTypePanel, AlgorithmSelector, TagManager。
**问题**: 各组件松散，需确认是否完整使用 `SubmitInjectionReq` 所有字段。

**改动 — 一键流程**:

表单步骤：
1. **选择目标系统**: Pedestal (microservice system under test) — 选容器 + 版本
2. **选择 Benchmark**: Benchmark 容器 (用于检测) — 选容器 + 版本
3. **配置故障注入**: Interval, PreDuration, Fault Specs (chaos nodes)
4. **选择 RCA 算法** (可选): 从 `type=algorithm` 的 containers 中选，支持多选
5. **标签 & 确认**: Labels, 名称，Review 配置摘要

对应 `SubmitInjectionReq`:
```typescript
{
  project_name: string,        // from URL context
  pedestal: ContainerSpec,     // Step 1
  benchmark: ContainerSpec,    // Step 2
  interval: number,            // Step 3
  pre_duration: number,        // Step 3
  specs: Node[][],             // Step 3 — fault configs
  algorithms: ContainerSpec[], // Step 4 — optional
  labels: LabelItem[],         // Step 5
}
```

**API**: `POST /api/v2/projects/:project_id/injections/inject`

提交后跳转到 Injection Detail 页，实时观察 pipeline 进度。

### 5.6 Injection/Datapack Detail (`/:teamName/:projectName/injections/:id`)

**现状**: Tabbed detail view (Overview, Files, Logs)。
**已实现 (REQ-812)**: Pipeline Progress Bar 已集成到 Overview Tab 顶部。

**Pipeline Progress 组件** (`PipelineProgress.tsx`):

在 Overview Tab 顶部显示 **Pipeline Progress Bar**：

```
┌─────────────────────────────────────────────────────────────────────┐
│  ● Inject Fault  ──>  ● Build Datapack  ──>  ● Run Algorithm  ──>  ● Collect Result  │
│    [completed]          [running]               [pending]              [pending]       │
└─────────────────────────────────────────────────────────────────────┘
```

**数据来源**:
- `FaultInjection.state` (DatapackState) 表示当前 datapack 状态
- 关联的 Trace (通过 Task → Trace) 包含整个 pipeline 的 task DAG
- 使用 `GET /api/v2/traces/:trace_id/stream` (SSE) 实时更新状态

**Pipeline 状态映射**:
- DatapackState = `Initial` / `Pending` → 注入阶段
- DatapackState = `Building` → 数据包构建阶段
- DatapackState = `Built` / `Ready` → 可用 / 算法运行中
- 关联 Execution 的 state → 算法运行和结果收集阶段

**Tab 结构 (当前实现)**:
- **Overview**: Pipeline 进度 (PipelineProgress) + 基本信息 (名称、故障类型、时间窗口、Groundtruth、Config)
- **Files**: Datapack parquet 文件列表 + ArrowViewer
- **Logs**: Pipeline 日志流 (LogsTab)

### 5.7 Execution List (`/:teamName/:projectName/executions`)

**现状**: WorkspaceTable 列表 + BatchProgressBanner。
**已实现 (REQ-815)**:
- WorkspaceTable 列表，列: Name, Notes, Algorithm, Status, Injection, Datapack, Runtime, Created, Labels
- **Batch Progress Banner** (`BatchProgressBanner.tsx`): 当 URL 含 `?group_id=xxx` 时显示
  - SSE 连接 `GET /api/v2/groups/:group_id/stream` 实时更新
  - 初始状态从 `GET /api/v2/groups/:group_id/stats` 获取
  - 显示: completed/total count, failure count, progress bar
  - 可折叠/关闭 (dismiss 后清除 URL param)
  - Alert 类型动态变化: info (进行中), success (全部完成), warning (有失败)
- "New Execution" 按钮 → 跳转 `executions/new`
- 支持搜索、排序、分组、列管理、批量操作

### 5.8 Execution Detail (`/:teamName/:projectName/executions/:id`)

**已实现 (REQ-817)**:
- 顶部显示关键信息：Algorithm (name + version), Datapack ID, State, Duration
- **Artifacts Tab** (`ArtifactsTab.tsx`) 展示 `DetectorResults` 和 `GranularityResults`
  - DetectorResults: 表格展示每个 span 的 normal vs abnormal 指标 (avg_duration, succ_rate)
  - 可展开的 Percentile 列: p90, p95, p99 (通过 "Show Percentiles" 按钮切换)
  - GranularityResults: 按 rank 排序展示定位结果 (level, result, confidence progress bar)
  - 类型对齐 SDK: 使用 `DetectorResultItem` 和 `GranularityResultItem` from `@rcabench/client`
- 其他 Tab: Overview, Logs, Files

### 5.9 Algorithm Benchmark (`/:teamName/:projectName/executions/new`)

**已实现 (REQ-814)**: `ExecutionCreatePage.tsx`

**场景**: 研究员开发了新 RCA 算法，需要在已有 Datapacks 或 Dataset 上批量测试效果。这是 Workflow B 的核心页面。

**表单设计 — 两种模式 (Tab 切换)**:

#### Mode A: Datapack 模式 (逐个选择)

```
┌────────────────────────────────────────────────────┐
│ Step 1: Select Algorithm                           │
│ ┌────────────────────┐ ┌──────────────┐            │
│ │ Algorithm ▼        │ │ Version ▼    │            │
│ └────────────────────┘ └──────────────┘            │
│                                                    │
│ Step 2: Select Datapacks                           │
│ ┌──────────────────────────────────────────────┐   │
│ │ ☑ datapack-train-ticket-network-001          │   │
│ │ ☑ datapack-train-ticket-pod-002              │   │
│ │ ☐ datapack-sock-shop-stress-001              │   │
│ │ ...                     [Select All] [Clear] │   │
│ └──────────────────────────────────────────────┘   │
│ Showing 24 datapacks · 2 selected                  │
│                                                    │
│ Step 3: Labels (optional)                          │
│ ┌────────────────────────────────────────────────┐ │
│ │ + Add Label                                    │ │
│ └────────────────────────────────────────────────┘ │
│                                                    │
│              [Cancel]  [Run 2 Executions]          │
└────────────────────────────────────────────────────┘
```

生成的 `SubmitExecutionReq`:
```typescript
{
  project_name: "my-project",
  specs: [
    { algorithm: { name: "MicroRCA", version: "1.0.0" }, datapack: "datapack-train-ticket-network-001" },
    { algorithm: { name: "MicroRCA", version: "1.0.0" }, datapack: "datapack-train-ticket-pod-002" },
  ],
  labels: [...]
}
```

#### Mode B: Dataset 模式 (整个数据集)

```
┌────────────────────────────────────────────────────┐
│ Step 1: Select Algorithm                           │
│ ┌────────────────────┐ ┌──────────────┐            │
│ │ Algorithm ▼        │ │ Version ▼    │            │
│ └────────────────────┘ └──────────────┘            │
│                                                    │
│ Step 2: Select Dataset                             │
│ ┌────────────────────┐ ┌──────────────┐            │
│ │ Dataset ▼          │ │ Version ▼    │            │
│ └────────────────────┘ └──────────────┘            │
│ Contains 48 datapacks                              │
│                                                    │
│              [Cancel]  [Run on Dataset]             │
└────────────────────────────────────────────────────┘
```

生成的 `SubmitExecutionReq`:
```typescript
{
  project_name: "my-project",
  specs: [
    { algorithm: { name: "MicroRCA", version: "1.0.0" }, dataset: { name: "train-ticket-v1", version: "1.0.0" } }
  ],
  labels: [...]
}
```

**提交后行为** (已实现):
- API 返回 `SubmitExecutionResp` 包含 `group_id` 和每个 execution 的 `trace_id`/`task_id`
- 跳转到 Execution List 页 (`?group_id=xxx`)，顶部 BatchProgressBanner 自动显示
- SSE 通过 `GET /api/v2/groups/:group_id/stream` 实时追踪批量进度

**Datapack 选择器数据源**:
- `GET /api/v2/projects/:project_id/injections` — 列出项目内所有 datapacks
- 支持按 fault_type、category、labels 筛选
- 支持搜索

**Algorithm 选择器数据源**:
- `GET /api/v2/containers?type=algorithm` — 列出所有算法容器
- 选中容器后, `GET /api/v2/containers/:id/versions` — 列出版本

**API**: `POST /api/v2/projects/:project_id/executions/execute`

### 5.10 Evaluation List & Detail

**现状**: 列表 + 详情页。
**改动**: 保持现有结构。评估功能在核心流程跑通后再深化。确保：
- 评估列表从真实 API 加载
- 评估创建使用 `POST /api/v2/evaluations/datapacks` 或 `POST /api/v2/evaluations/datasets`

### 5.11 Algorithms (新增入口)

**已实现 (REQ-813)**: `AlgorithmListPage.tsx`

**路由**: `/:teamName/:projectName/algorithms`
**功能**: 展示和管理项目可用的 RCA 算法容器

- 列表: 调用 `GET /api/v2/containers?type=algorithm`，Table 展示 Name, Status, Public, Created
- 注册新算法: Modal 调用 `POST /api/v2/containers` (type=0, is_public=true)
- 添加版本: Modal 调用 `POST /api/v2/containers/:id/versions` (name, image_ref, command)
- 展开行显示版本列表: `AlgorithmVersions` 子组件
- 删除: Popconfirm + `DELETE /api/v2/containers/:id`
- "Run Benchmark" 按钮 → 跳转 `executions/new`

### 5.12 Team Pages

**现状**: TeamDetailPage with Overview/Projects/Users/Settings tabs。
**问题**: 
- `teamApi.addMember` 发送 `{email}` 但后端期望 `{username}` (BUG)
- `Team.member_count` vs 后端 `TeamDetailResp.user_count` (类型不匹配)

**改动**:
- Fix addMember request body: `{username, role_id}`
- 对齐字段名：使用后端 `user_count`
- 删除 ProjectsTab 中的 Star toggle

### 5.13 Profile (`/profile`)

**现状**: 3 个 tab (Profile, Projects, Stars)，大量 mock。
**改动**: 大幅精简
- 只保留 **Profile Tab**: 显示用户基本信息（从 `GET /api/v2/auth/profile` 获取）
- 删除 ProjectsTab（用户的项目在 `/projects` 已有）
- 删除 StarsTab
- 删除 ProfileSidebar 中硬编码的假数据
- 删除 ActivityGraph, RecentRuns 组件

### 5.14 Admin Pages

保持现有结构：
- `/admin/users` — 用户 CRUD
- `/admin/containers` — 所有类型容器管理 (algorithm + benchmark + pedestal)
- `/admin/datasets` — 数据集管理
- `/admin/system` — 系统配置

### 5.15 Tasks (`/tasks`)

保留为全局调试/高级页面。从侧栏降低优先级（放到底部或折叠在"高级"里）。

### 5.16 Traces (降级)

`/:teamName/:projectName/traces` 保留但降级为高级功能。Trace 信息主要通过 Injection Detail 的 Pipeline 视图间接展示。

## 6. Layout & Navigation

### 6.1 MainLayout (Global)

```
┌──────────────────────────────────────────────────┐
│ [Logo] AegisLab        [Theme] [User Dropdown]   │  ← AppHeader
├──────────┬───────────────────────────────────────┤
│ Home     │                                       │
│          │                                       │
│ Projects │         Main Content                  │
│  ├ Proj1 │                                       │
│  ├ Proj2 │                                       │
│          │                                       │
│ Teams    │                                       │
│  ├ Team1 │                                       │
│          │                                       │
│ ──────── │                                       │
│ Admin    │                                       │  ← Only if admin
│  ├ Users │                                       │
│  ├ ...   │                                       │
│ ──────── │                                       │
│ Tasks    │                                       │  ← 底部/低优先级
└──────────┴───────────────────────────────────────┘
```

**与现有的区别**:
- 删除 Notifications 入口
- Tasks 移到底部
- 保持 Projects + Teams 作为主要导航

### 6.2 WorkspaceLayout (Project-scoped)

```
┌──────────────────────────────────────────────────┐
│ [≡] AegisLab > Team > Project          [...]     │  ← Breadcrumb
├──────────┬───────────────────────────────────────┤
│ Overview │                                       │
│ Datapacks│         Workspace Content             │
│ Executions│                                      │
│ Evaluations│                                     │
│ Algorithms│                                      │  ← 新增
│ ──────── │                                       │
│ Settings │                                       │
└──────────┴───────────────────────────────────────┘
```

**与现有的区别**:
- 新增 "Algorithms" 入口
- "Injections" 重命名为 "Datapacks" (与后端概念一致: FaultInjection = Datapack)
- Traces 从侧栏移除（内容通过 Datapack Detail 的 Pipeline 视图展示）

## 7. Frontend-Backend Alignment (不一致修复清单)

### 7.1 Critical Bugs

| # | Issue | Fix |
|---|-------|-----|
| 1 | `teamApi.addMember` sends `{email}` but backend expects `{username}` | 改前端请求体 |
| 2 | Frontend `Team.member_count` vs backend `TeamDetailResp.user_count` | 统一用 `user_count` |
| 3 | `AddLabelDropdown` uses `getMockLabelSuggestions()` | 改用 `GET /api/v2/labels` |
| 4 | Admin check `(user as Record<string, unknown>)?.is_superuser` | 后端 `UserProfileResp` 需新增 `is_admin` 字段，或前端读 role |

### 7.2 Mock Data Removal

| Component | Issue | Fix |
|-----------|-------|-----|
| `SystemMetrics.tsx` | 100% hardcoded fake data | 删除组件（HomePage 已接入真实 API） |
| `ProfileSidebar.tsx` | 硬编码假团队 'Personal Capital' | 从 API 获取用户团队 |
| `ProfilePage/ProjectsTab.tsx` | `projectsData = {items:[], total:0}` | 使用 `useProjects()` hook |
| `ProfilePage/StarsTab.tsx` | 全部 stub | 删除 |
| `ActivityGraph.tsx` | 无真实数据源 | 删除 |
| `QuickActions.tsx` | 可能未使用 | 确认后删除 |
| `ActivityFeed.tsx` | 可能未使用 | 确认后删除 |

### 7.3 Backend TODOs (影响前端)

| Backend Issue | Impact | Action |
|---|---|---|
| `GetInjectionLogs` 返回空 `[]string{}` | Logs Tab 无数据 | 后端修复，或前端暂显示"日志功能开发中" |
| `GetTaskDetail` 返回空 logs | Task Detail 无日志 | 后端修复 |
| Notification read state 无后端持久化 | 通知已删除，不影响 | N/A |

### 7.4 Type Migration

`types/api.ts` 中手写的 `Team`, `TeamMember` 等类型需迁移到 SDK 生成类型 (`@rcabench/client`)。

步骤：
1. 确保所有相关 API 在后端标注了 `@x-api-type {"sdk":"true"}`
2. `just swag-init && just generate-typescript-client`
3. 前端 `import type { ... } from '@rcabench/client'` 替换手写类型

## 8. UI/UX Guidelines

UI/UX 规范独立维护，参见 **[frontend-ui-guidelines.md](./frontend-ui-guidelines.md)**。

该文档覆盖：颜色系统、Layout 模式、组件模式 (List/Detail/Form)、Loading/Empty/Error 状态、Typography、Spacing、导航、表格、表单、动画、可访问性、Anti-Patterns 禁止清单。

与本文档的关系：本文档定义"做什么"（业务流程和页面设计），UI Guidelines 定义"怎么做"（视觉和交互规范）。两者正交。

## 9. Component Cleanup

### 9.1 To Delete

```
src/pages/notifications/NotificationsPage.tsx    — feature removed
src/api/notifications.ts                          — feature removed
src/hooks/useNotifications.ts                     — feature removed
src/components/dashboard/ActivityFeed.tsx          — unused / mock
src/components/dashboard/QuickActions.tsx          — unused / mock
src/components/dashboard/SystemMetrics.tsx         — 100% mock, replaced
src/components/profile/ActivityGraph.tsx           — no data source
src/components/profile/RecentRuns.tsx              — no data source (verify)
src/pages/profile/tabs/StarsTab.tsx               — no backend API
```

### 9.2 To Simplify

```
src/pages/profile/ProfilePage.tsx     — remove Stars/Activity tabs, keep Profile only
src/components/profile/ProfileSidebar.tsx — remove hardcoded data, fetch from API
src/pages/profile/tabs/ProjectsTab.tsx — 删除或改用 useProjects() 获取真实数据
```

### 9.3 Added (implemented)

```
src/pages/projects/algorithms/AlgorithmListPage.tsx   — 研究员算法管理 (REQ-813) ✓
src/components/workspace/PipelineProgress.tsx         — Pipeline 进度组件 (REQ-812) ✓
src/components/workspace/BatchProgressBanner.tsx      — 批量执行进度 Banner (REQ-815) ✓
src/pages/executions/ExecutionCreatePage.tsx           — 算法测试页 Workflow B (REQ-814) ✓
```

## 10. API Coverage Matrix

前端必须对接的后端 API 完整清单：

### Auth
| Endpoint | Frontend | Status |
|----------|----------|--------|
| `POST /auth/login` | Login.tsx | OK |
| `POST /auth/register` | (admin) | OK |
| `GET /auth/profile` | auth store | OK |
| `POST /auth/change-password` | Settings | OK |

### Projects
| Endpoint | Frontend | Status |
|----------|----------|--------|
| `GET /projects` | ProjectList | OK |
| `POST /projects` | ProjectEdit | OK |
| `GET /projects/:id` | ProjectOverview | OK |
| `PATCH /projects/:id` | ProjectSettings | OK |
| `DELETE /projects/:id` | ProjectSettings | OK |

### Injections (Project-scoped)
| Endpoint | Frontend | Status |
|----------|----------|--------|
| `GET /projects/:id/injections` | ProjectInjectionList | OK |
| `POST /projects/:id/injections/inject` | InjectionCreate | **Verify algorithms field** |
| `POST /projects/:id/injections/search` | (advanced search) | Deferred |
| `GET /injections/:id` | ProjectInjectionDetail | OK |
| `GET /injections/:id/files` | FilesTab | OK |
| `GET /injections/:id/files/query` | ArrowViewer | OK |
| `GET /injections/:id/download` | (download action) | OK |
| `PUT /injections/:id/groundtruth` | GroundTruthTable | **Verify** |
| `POST /injections/upload` | ProjectInjectionList (Upload Modal) | OK (REQ-816) |
| `POST /injections/batch-delete` | BulkActionBar | OK |

### Executions (Project-scoped)
| Endpoint | Frontend | Status |
|----------|----------|--------|
| `GET /projects/:id/executions` | ProjectExecutionList | OK |
| `POST /projects/:id/executions/execute` | ExecutionCreatePage | OK (REQ-814) |
| `GET /executions/:id` | ProjectExecutionDetail | OK |

### Evaluations
| Endpoint | Frontend | Status |
|----------|----------|--------|
| `GET /evaluations` | EvaluationList | OK |
| `GET /evaluations/:id` | EvaluationDetail | OK |
| `POST /evaluations/datapacks` | EvaluationForm | **Verify** |
| `POST /evaluations/datasets` | EvaluationForm | **Verify** |

### Containers
| Endpoint | Frontend | Status |
|----------|----------|--------|
| `GET /containers` | ContainerList + AlgorithmList | OK |
| `POST /containers` | ContainerForm + AlgorithmForm | OK |
| `GET /containers/:id` | ContainerDetail | OK |
| `PATCH /containers/:id` | ContainerForm | OK |
| `GET /containers/:id/versions` | ContainerVersions | OK |
| `POST /containers/:id/versions` | ContainerVersions + AlgorithmForm | OK |

### Teams
| Endpoint | Frontend | Status |
|----------|----------|--------|
| `GET /teams` | MainSidebarContent | OK |
| `POST /teams` | Sidebar modal | OK |
| `GET /teams/:id` | TeamDetailPage | OK |
| `GET /teams/:id/members` | UsersTab | OK |
| `POST /teams/:id/members` | UsersTab | **BUG: email→username** |
| `GET /teams/:id/projects` | ProjectsTab | OK |

### Labels
| Endpoint | Frontend | Status |
|----------|----------|--------|
| `GET /labels` | AddLabelDropdown | **Replace mock** |
| `POST /labels` | AddLabelDropdown | OK |

### Tasks
| Endpoint | Frontend | Status |
|----------|----------|--------|
| `GET /tasks` | TaskList | OK |
| `GET /tasks/:id` | TaskDetail | OK |
| `GET /tasks/:id/logs/ws` | useTaskLogs | OK |

### Traces
| Endpoint | Frontend | Status |
|----------|----------|--------|
| `GET /traces` | TracesPage | OK |
| `GET /traces/:id/stream` | useTraceSSE | OK — **用于 Pipeline 视图** |

### System
| Endpoint | Frontend | Status |
|----------|----------|--------|
| `GET /system/metrics` | HomePage | OK |
| `GET /metrics/injections` | HomePage | OK |
| `GET /metrics/executions` | HomePage | OK |
| `GET /metrics/algorithms` | HomePage | OK |

### Chaos Systems
| Endpoint | Frontend | Status |
|----------|----------|--------|
| `GET /systems` | InjectionCreate | **Verify** — 用于选择目标系统 |

## 11. Implementation Priority & Status

### Phase 1: Pipeline 跑通 (P0) — COMPLETE

1. ~~Fix `InjectionCreate` — 确保 `algorithms` 字段正确传入 `SubmitInjectionReq`~~ (verified)
2. ~~实现 `PipelineProgress` 组件~~ → `PipelineProgress.tsx` (REQ-812)
3. ~~Fix Team bugs (addMember, type mismatch)~~ (fixed)
4. ~~Fix Label mock → real API~~ (fixed)
5. ~~删除 mock 组件 (SystemMetrics, ActivityGraph, etc.)~~ (18 files deleted)

### Phase 2: UI 统一 (P1) — COMPLETE

6. ~~统一 Loading/Empty/Error 状态处理~~
7. ~~精简 Profile 页~~
8. ~~删除 Notifications 相关代码~~
9. ~~删除 Stars 相关代码~~
10. ~~新增 AlgorithmListPage~~ (REQ-813)
11. ~~WorkspaceSidebar 新增 Algorithms 入口~~
12. ~~Algorithm Benchmark Page~~ → `ExecutionCreatePage.tsx` (REQ-814)
13. ~~Batch Execution Progress Banner~~ → `BatchProgressBanner.tsx` (REQ-815)
14. ~~Execution Detail Results Display~~ → `ArtifactsTab.tsx` aligned with SDK (REQ-817)

### Phase 3: 补全 (P2) — COMPLETE

15. ~~Manual Datapack Upload UI~~ → Upload modal in `ProjectInjectionList.tsx` (REQ-816)
16. ~~Type migration (hand-written → SDK types)~~ (ArtifactsTab migrated to SDK types)
17. ~~清理 unused CSS/components~~ (ExecutionForm.tsx deleted)

### Remaining Work

- Type migration for remaining hand-written types (teams, etc.) → SDK types
- User acceptance testing for all REQ-812~817
- Backend fixes: `GetInjectionLogs` returns empty, `GetTaskDetail` logs empty

## 12. Backend Changes Required

| Change | Description | Priority |
|--------|-------------|----------|
| `UserProfileResp` add `is_admin` | 前端需要可靠的管理员判断 | P0 |
| Fix `GetInjectionLogs` | 当前返回空数组 | P1 |
| Fix `GetTaskDetail` logs | 当前返回空数组 | P1 |
| Verify `SubmitInjectionReq.Algorithms` end-to-end | 确保 CRD callback chain 带着 algorithms 一直跑到 RunAlgorithm | P0 |
