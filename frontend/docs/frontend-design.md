# RCABench 前端设计文档

## 目录
1. [系统概览](#系统概览)
2. [用户角色与权限](#用户角色与权限)
3. [页面架构](#页面架构)
4. [页面详细设计](#页面详细设计)
5. [通用组件](#通用组件)
6. [API 映射](#api-映射)
7. [技术栈建议](#技术栈建议)
8. [实时功能](#实时功能)

---

## 系统概览

RCABench (AegisLab) 是一个微服务根因分析(RCA)基准测试平台,提供故障注入、算法执行和评估能力。系统的核心业务流程包括:

1. **资源管理**: 项目、容器(工作负载/检测器/算法)、数据集
2. **实验编排**: 数据包构建 → 故障注入 → 结果收集 → RCA 算法执行
3. **结果评估**: 算法预测与真实故障对比分析

### 核心概念

- **Pedestal**: 被测系统(SUT)容器
- **Benchmark**: 检测器容器,用于异常检测
- **Algorithm**: RCA 算法容器
- **Datapack**: 构建的数据包,包含 traces/logs
- **Fault Injection**: 故障注入配置与执行
- **Execution**: 算法在数据包上的运行结果

---

## 用户角色与权限

### 角色类型
1. **系统管理员**: 管理用户、角色、权限
2. **项目管理员**: 管理项目资源
3. **研究员**: 创建实验、执行算法
4. **访客**: 只读访问公开资源

### 权限维度
- **全局权限**: 系统级操作
- **项目级权限**: 特定项目的操作
- **容器级权限**: 特定容器的操作
- **数据集级权限**: 特定数据集的操作

---

## 页面架构

### 主导航结构

```
┌─ Dashboard (仪表盘)
├─ Projects (项目管理)
│  ├─ 项目列表
│  ├─ 项目详情
│  │  ├─ 关联容器
│  │  ├─ 关联数据集
│  │  ├─ 故障注入记录
│  │  └─ 算法执行记录
│  └─ 创建/编辑项目
├─ Containers (容器管理)
│  ├─ 容器列表 (按类型筛选: Pedestal/Benchmark/Algorithm)
│  ├─ 容器详情
│  │  ├─ 版本列表
│  │  ├─ 构建历史
│  │  └─ 使用统计
│  ├─ 创建容器
│  └─ 容器版本管理
├─ Datasets (数据集管理)
│  ├─ 数据集列表
│  ├─ 数据集详情
│  │  ├─ 版本列表
│  │  └─ 关联故障注入
│  ├─ 创建数据集
│  └─ 上传数据集版本
├─ Injections (故障注入)
│  ├─ 注入列表
│  ├─ 注入详情
│  │  ├─ 故障配置
│  │  ├─ 执行状态
│  │  ├─ 检测结果
│  │  └─ 相关算法执行
│  ├─ 创建注入实验
│  │  ├─ 配置向导 (步骤式)
│  │  ├─ 故障编排器 (可视化)
│  │  └─ 批量配置
│  └─ 分析视图
│     ├─ 无异常注入
│     └─ 有异常注入
├─ Executions (算法执行)
│  ├─ 执行列表
│  ├─ 执行详情
│  │  ├─ 算法信息
│  │  ├─ 输入数据
│  │  ├─ 检测结果
│  │  ├─ 粒度结果 (可视化)
│  │  └─ 性能指标
│  └─ 创建执行任务
├─ Evaluations (评估)
│  ├─ Datapack 评估
│  ├─ Dataset 评估
│  ├─ 算法对比
│  └─ 可视化报表
│     ├─ 准确率分析
│     ├─ Top-K 定位
│     └─ 混淆矩阵
├─ Tasks (任务监控)
│  ├─ 任务列表
│  ├─ 任务详情
│  │  ├─ 实时日志流 (SSE)
│  │  ├─ 任务依赖树
│  │  └─ 性能指标
│  └─ 任务统计
├─ System (系统管理)
│  ├─ 用户管理
│  ├─ 角色管理
│  ├─ 权限管理
│  ├─ 标签管理
│  └─ 资源管理
└─ Settings (个人设置)
   ├─ 个人资料
   └─ 修改密码
```

---

## 页面详细设计

### 1. Dashboard (仪表盘)

#### 功能描述
展示系统整体状态、关键指标和最近活动。

#### 页面组件

**1.1 关键指标卡片**
- 项目总数
- 活跃实验数
- 待处理任务数
- 今日执行数

**1.2 任务状态分布 (饼图)**
- Pending / Running / Completed / Error

**1.3 最近活动流**
- 最近 10 条操作记录
- 类型: 创建项目、提交注入、执行完成等

**1.4 快捷入口**
- 创建项目
- 提交故障注入
- 执行算法
- 查看报告

**1.5 系统健康状态**
- Redis 连接状态
- MySQL 连接状态
- Kubernetes 集群状态

#### 关联 API
```
GET /api/v2/projects?page=1&size=5 (获取最近项目)
GET /api/v2/tasks?page=1&size=10 (获取最近任务)
GET /api/v2/injections?page=1&size=5 (获取最近注入)
GET /api/v2/executions?page=1&size=5 (获取最近执行)
```

---

### 2. Projects (项目管理)

#### 2.1 项目列表页

**功能描述**: 展示所有项目,支持筛选、搜索和排序。

**页面组件**:
- **搜索栏**: 项目名称关键词搜索
- **筛选器**:
  - 可见性: 公开/私有
  - 状态: 活跃/已归档
  - 标签: 多选标签筛选
- **项目卡片/表格**:
  - 项目名称
  - 描述
  - 关联容器数/数据集数
  - 最近活动时间
  - 创建者
  - 操作按钮: 查看详情、编辑、删除
- **分页器**

**关联 API**:
```
GET /api/v2/projects?page={page}&size={size}&is_public={bool}&status={int}&label={key:value}
DELETE /api/v2/projects/{project_id}
```

#### 2.2 项目详情页

**功能描述**: 展示项目详细信息及关联资源。

**页面组件**:

**顶部信息卡**:
- 项目名称、描述、创建者、创建时间
- 标签列表 (可添加/删除)
- 编辑按钮

**Tab 切换**:
1. **关联容器** (表格)
   - 容器名称、类型、版本数、最新版本
   - 操作: 查看详情、解除关联
2. **关联数据集** (表格)
   - 数据集名称、类型、版本数、最新版本
   - 操作: 查看详情、解除关联
3. **故障注入记录** (表格)
   - 注入名称、故障类型、状态、开始时间
   - 操作: 查看详情、重新执行
4. **算法执行记录** (表格)
   - 算法名称、执行时间、状态、准确率
   - 操作: 查看结果

**关联 API**:
```
GET /api/v2/projects/{project_id}
PATCH /api/v2/projects/{project_id}/labels (管理标签)
GET /api/v2/containers?project_id={project_id} (关联容器)
GET /api/v2/datasets?project_id={project_id} (关联数据集)
GET /api/v2/injections?project_id={project_id} (注入记录)
GET /api/v2/executions?project_id={project_id} (执行记录)
```

#### 2.3 创建/编辑项目页

**功能描述**: 创建新项目或编辑现有项目。

**表单字段**:
- 项目名称 (必填)
- 项目描述 (富文本编辑器)
- 可见性: 公开/私有 (单选)
- 标签: 标签选择器 (支持创建新标签)
- 关联容器: 多选下拉框
- 关联数据集: 多选下拉框

**关联 API**:
```
POST /api/v2/projects (创建)
PATCH /api/v2/projects/{project_id} (更新)
GET /api/v2/labels?category=project (获取可用标签)
GET /api/v2/containers?is_public=true (获取可关联容器)
GET /api/v2/datasets?is_public=true (获取可关联数据集)
```

---

### 3. Containers (容器管理)

#### 3.1 容器列表页

**功能描述**: 展示所有容器,支持按类型筛选。

**页面组件**:

**筛选器**:
- 类型: Pedestal / Benchmark / Algorithm
- 可见性: 公开/私有
- 状态: 活跃/已归档
- 标签筛选

**容器卡片**:
- 容器名称
- 容器类型 (徽章)
- README 摘要
- 版本数量
- 使用次数
- 最新版本信息
- 操作: 查看详情、编辑、删除

**关联 API**:
```
GET /api/v2/containers?type={type}&is_public={bool}&status={int}&label={key:value}
DELETE /api/v2/containers/{container_id}
```

#### 3.2 容器详情页

**功能描述**: 展示容器详细信息、版本列表和构建历史。

**页面组件**:

**顶部信息卡**:
- 容器名称、类型、README (Markdown 渲染)
- 标签列表
- 创建时间、更新时间

**Tab 切换**:

1. **版本列表** (表格)
   - 列: 版本号、Registry、Repository、Tag、Command、创建时间、使用次数
   - 操作: 查看详情、编辑、删除、设置为默认
   - 新增版本按钮

2. **构建历史** (时间线)
   - 构建任务列表
   - 状态: Pending / Running / Completed / Error
   - 构建日志链接

3. **使用统计** (图表)
   - 版本使用趋势 (折线图)
   - 最近使用记录 (表格)

4. **配置详情** (JSON 查看器)
   - Helm 配置
   - 参数配置
   - 环境变量

**关联 API**:
```
GET /api/v2/containers/{container_id}
GET /api/v2/containers/{container_id}/versions
POST /api/v2/containers/{container_id}/versions (创建版本)
PATCH /api/v2/containers/{container_id}/versions/{version_id} (更新版本)
DELETE /api/v2/containers/{container_id}/versions/{version_id}
GET /api/v2/tasks?resource_type=container&resource_id={container_id} (构建任务)
```

#### 3.3 创建容器页

**功能描述**: 创建新容器及其初始版本。

**表单字段**:

**基础信息**:
- 容器名称 (必填)
- 容器类型 (单选: Pedestal/Benchmark/Algorithm)
- README (Markdown 编辑器)
- 可见性: 公开/私有
- 标签选择

**版本信息**:
- 版本号 (语义化版本)
- Registry (例: docker.io)
- Repository (例: nginx)
- Tag (例: latest)
- 完整镜像引用 (自动拼接显示)
- 启动命令 (可选)
- 环境变量 (键值对列表)

**Helm 配置** (可选):
- Repo URL
- Chart Name
- Values 文件上传

**关联 API**:
```
POST /api/v2/containers
POST /api/v2/containers/{container_id}/versions
POST /api/v2/containers/{container_id}/versions/{version_id}/helm-values
```

#### 3.4 容器构建页

**功能描述**: 提交容器构建任务。

**表单字段**:
- 容器选择 (下拉框)
- 构建配置 (JSON 编辑器)
- Dockerfile 内容 (代码编辑器)

**关联 API**:
```
POST /api/v2/containers/build
```

---

### 4. Datasets (数据集管理)

#### 4.1 数据集列表页

**功能描述**: 展示所有数据集。

**页面组件**:

**筛选器**:
- 类型: Trace / Log / Metric
- 可见性: 公开/私有
- 标签筛选

**数据集卡片**:
- 数据集名称
- 数据集类型 (徽章)
- 描述摘要
- 版本数量
- 最新版本
- 关联注入数
- 操作: 查看详情、编辑、删除

**关联 API**:
```
GET /api/v2/datasets?type={type}&is_public={bool}&label={key:value}
POST /api/v2/datasets/search (高级搜索)
DELETE /api/v2/datasets/{dataset_id}
```

#### 4.2 数据集详情页

**功能描述**: 展示数据集详细信息和版本列表。

**页面组件**:

**顶部信息卡**:
- 数据集名称、类型、描述
- 标签列表
- 创建时间、更新时间

**Tab 切换**:

1. **版本列表** (表格)
   - 列: 版本号、大小、Checksum、创建时间、状态
   - 操作: 下载、查看详情、删除
   - 新增版本按钮

2. **关联注入** (表格)
   - 注入名称、故障类型、状态、创建时间
   - 操作: 查看注入详情

3. **使用统计** (图表)
   - 版本下载趋势
   - 关联注入数量

**关联 API**:
```
GET /api/v2/datasets/{dataset_id}
GET /api/v2/datasets/{dataset_id}/versions
GET /api/v2/datasets/{dataset_id}/versions/{version_id}
GET /api/v2/datasets/{dataset_id}/versions/{version_id}/download
POST /api/v2/datasets/{dataset_id}/versions (创建版本)
PATCH /api/v2/datasets/{dataset_id}/version/{version_id}/injections (管理注入关联)
```

#### 4.3 创建/编辑数据集页

**表单字段**:
- 数据集名称 (必填)
- 数据集类型 (单选: Trace/Log/Metric)
- 描述 (富文本)
- 可见性: 公开/私有
- 标签选择

**关联 API**:
```
POST /api/v2/datasets
PATCH /api/v2/datasets/{dataset_id}
```

#### 4.4 上传数据集版本页

**表单字段**:
- 版本号 (语义化版本)
- 文件上传 (支持拖拽上传, ZIP 格式)
- 描述 (可选)
- 自动计算 Checksum

**关联 API**:
```
POST /api/v2/datasets/{dataset_id}/versions
```

---

### 5. Injections (故障注入)

#### 5.1 注入列表页

**功能描述**: 展示所有故障注入记录。

**页面组件**:

**筛选器**:
- 时间范围: 1小时/24小时/7天/自定义
- 故障类型: Network/CPU/Memory/IO/Pod/Container
- 状态: Pending/Running/Completed/Error
- 标签筛选

**注入表格**:
- 列:
  - 注入名称
  - 故障类型 (徽章)
  - Pedestal 容器
  - Benchmark 容器
  - 状态 (进度条)
  - 开始时间
  - 持续时间
  - 标签
- 操作: 查看详情、查看日志、删除

**批量操作**:
- 批量删除
- 批量添加标签

**关联 API**:
```
GET /api/v2/injections?lookback={duration}&fault_type={type}&state={int}&label={key:value}
POST /api/v2/injections/search (高级搜索)
POST /api/v2/injections/batch-delete
PATCH /api/v2/injections/labels/batch
GET /api/v2/injections/analysis/no-issues (无异常注入)
GET /api/v2/injections/analysis/with-issues (有异常注入)
```

#### 5.2 注入详情页

**功能描述**: 展示单个故障注入的详细信息。

**页面组件**:

**顶部信息卡**:
- 注入名称
- 状态徽章 (实时更新)
- 故障类型
- 开始时间、结束时间、持续时间
- Pedestal 和 Benchmark 信息
- 标签列表

**Tab 切换**:

1. **故障配置** (可视化展示)
   - Display Config (用户友好的配置展示)
   - Engine Config (JSON 查看器, 折叠/展开)
   - 故障时间线 (甘特图)
   - 并行/串行批次展示

2. **执行状态** (实时更新)
   - 当前状态: BUILDING_DATAPACK / DATAPACK_READY / INJECTING / INJECTION_COMPLETE / COLLECTING_RESULT / COMPLETED
   - 进度条
   - 关联任务列表 (任务树形结构)
   - 任务状态实时流 (SSE)

3. **检测结果** (DetectorResults)
   - 表格展示:
     - Span 名称
     - 异常类型
     - 平均延迟 (正常 vs 异常)
     - 成功率 (正常 vs 异常)
     - P90/P95/P99 对比
   - 图表: 延迟对比柱状图
   - 筛选: 仅显示异常 Span

4. **真实故障信息** (Groundtruth)
   - 表格:
     - 故障级别 (Service/Pod/Span/Metric)
     - 目标组件
     - 预期影响时间
   - 故障影响范围可视化 (服务拓扑图)

5. **算法执行** (关联的 Execution 列表)
   - 表格:
     - 算法名称/版本
     - 执行时间
     - 状态
     - 准确率指标
   - 操作: 查看执行详情

6. **任务日志** (SSE 实时流)
   - 实时日志输出
   - 日志级别筛选 (INFO/WARN/ERROR)
   - 自动滚动/暂停
   - 下载日志

**关联 API**:
```
GET /api/v2/injections/{id}
GET /api/v2/traces/{trace_id}/stream (SSE 实时日志)
GET /api/v2/tasks?injection_id={id} (关联任务)
GET /api/v2/executions?datapack_id={id} (关联执行)
PATCH /api/v2/injections/{id}/labels
```

#### 5.3 创建注入实验页 (步骤向导)

**功能描述**: 引导用户逐步配置故障注入实验。

**步骤 1: 基础配置**
- 项目选择 (下拉框)
- 实验名称前缀 (可选)
- 标签选择

**步骤 2: 选择容器**
- Pedestal 容器选择:
  - 容器名称下拉
  - 版本选择
  - Namespace 输入
- Benchmark 容器选择:
  - 容器名称下拉
  - 版本选择
  - Namespace 输入
- 可选: 算法容器选择 (多选)
  - 算法容器列表 (多选复选框)
  - 版本选择

**步骤 3: 时间配置**
- Interval (实验总时长, 分钟)
- PreDuration (正常数据收集时长, 分钟, < Interval)
- 时间线预览 (可视化)

**步骤 4: 故障编排** (核心功能)

**界面布局**:
- 左侧: 故障类型面板 (可折叠分组)
  - Network Faults
  - CPU Faults
  - Memory Faults
  - IO Faults
  - Pod Faults
  - Container Faults
- 中间: 故障编排画布
  - 批次 (Batch) 列表
  - 每个批次包含并行执行的故障节点
  - 拖拽添加故障节点
  - 可视化批次依赖关系 (箭头连接)
- 右侧: 故障配置面板
  - 选中故障节点的详细配置表单
  - 目标选择器 (Service/Pod/Container)
  - 故障参数配置 (根据故障类型动态表单)

**故障配置元数据获取**:
```
GET /api/v2/injections/metadata?system={SystemType}
返回:
- FaultTypeConfigs (各故障类型的可配置参数)
- Resources (系统中可用的目标资源列表)
```

**批次管理**:
- 添加批次按钮
- 批次重排序 (拖拽)
- 删除批次
- 批次内故障节点重排序

**故障节点配置示例** (Network Delay):
- 目标选择: Service/Pod/Container (下拉)
- 延迟时间 (ms)
- 抖动 (jitter, ms)
- 持续时间 (秒)
- 概率 (百分比)

**步骤 5: 预览与提交**
- 实验配置摘要
- 批次数量、故障节点总数
- 预计执行时间
- JSON 配置预览 (只读)
- 提交按钮

**关联 API**:
```
GET /api/v2/injections/metadata?system={SystemType} (获取故障配置元数据)
POST /api/v2/injections/inject (提交注入任务)
```

**SubmitInjectionReq 结构**:
```json
{
  "project_name": "string",
  "pedestal": {
    "name": "string",
    "version": "string",
    "namespace": "string"
  },
  "benchmark": {
    "name": "string",
    "version": "string",
    "namespace": "string"
  },
  "interval": 30,
  "pre_duration": 5,
  "specs": [
    [ /* Batch 1 - 并行执行 */
      { /* Fault Node 1 */ },
      { /* Fault Node 2 */ }
    ],
    [ /* Batch 2 - 串行执行 */ ]
  ],
  "algorithms": [
    { "name": "algorithm1", "version": "v1.0.0" }
  ],
  "labels": [
    { "key": "env", "value": "test" }
  ]
}
```

#### 5.4 Datapack 构建页

**功能描述**: 提交独立的 Datapack 构建任务 (不执行故障注入)。

**表单字段**:
- Benchmark 容器选择
- 数据源选择 (二选一):
  - 现有 Datapack (下拉)
  - 数据集 + 版本
- PreDuration (可选覆盖)

**关联 API**:
```
POST /api/v2/injections/build
```

---

### 6. Executions (算法执行)

#### 6.1 执行列表页

**功能描述**: 展示所有算法执行记录。

**页面组件**:

**筛选器**:
- 状态: Pending/Running/Completed/Error
- 算法名称 (多选下拉)
- Datapack 名称
- 标签筛选

**执行表格**:
- 列:
  - Execution ID
  - 算法名称/版本
  - Datapack 名称
  - 执行时长 (秒)
  - 状态 (徽章)
  - 创建时间
  - 标签
- 操作: 查看详情、查看结果、删除

**批量操作**:
- 批量删除

**关联 API**:
```
GET /api/v2/executions?state={int}&status={int}&label={key:value}
POST /api/v2/executions/batch-delete
```

#### 6.2 执行详情页

**功能描述**: 展示单个算法执行的详细信息和结果。

**页面组件**:

**顶部信息卡**:
- Execution ID
- 算法名称/版本
- Datapack 名称
- 执行状态徽章
- 执行时长
- 创建时间
- 标签列表

**Tab 切换**:

1. **输入数据**
   - Datapack 信息:
     - 关联故障注入链接
     - 数据集信息
     - Benchmark 信息
   - 检测结果 (DetectorResults) 表格
   - Ground Truth 表格

2. **粒度结果** (GranularityResults)

**可视化展示** (核心功能):

**服务拓扑图** (交互式图表):
- 节点: 微服务
- 边: 服务调用关系
- 颜色编码: 根据算法预测的故障概率
- 悬浮提示: 显示详细指标
- 点击节点: 展开 Pod 级详情

**分层结果表格**:
- Service Level Results:
  - 列: Rank, Service Name, Confidence, Ground Truth Match (✓/✗)
- Pod Level Results:
  - 列: Rank, Pod Name, Service, Confidence, Ground Truth Match
- Span Level Results:
  - 列: Rank, Span Name, Service, Confidence, Ground Truth Match
- Metric Level Results:
  - 列: Rank, Metric Name, Confidence, Ground Truth Match

**准确率指标卡片**:
- Top-1 Accuracy (Service/Pod/Span/Metric)
- Top-3 Accuracy
- Top-5 Accuracy
- Mean Reciprocal Rank (MRR)

**混淆矩阵** (Service Level):
- 行: 真实故障服务
- 列: 预测故障服务
- 热力图展示

3. **性能指标**
   - 执行时间分解 (饼图):
     - 数据加载
     - 图构建
     - 算法计算
     - 结果输出
   - 资源使用:
     - CPU 使用率
     - 内存使用量
     - 网络 I/O

4. **任务日志** (SSE 实时流)
   - 算法执行日志
   - 错误日志

**关联 API**:
```
GET /api/v2/executions/{id}
GET /api/v2/traces/{trace_id}/stream (SSE 日志流)
```

#### 6.3 创建执行任务页

**功能描述**: 手动提交算法执行任务 (通常由系统自动触发)。

**表单字段**:
- 算法选择 (容器 + 版本)
- Datapack 选择 (下拉)
- 标签选择

**关联 API**:
```
POST /api/v2/executions/execute
```

---

### 7. Evaluations (评估)

#### 7.1 Datapack 评估页

**功能描述**: 对比多个算法在同一 Datapack 上的表现。

**页面组件**:

**配置表单**:
- 添加评估规格 (可添加多条):
  - 算法选择 (容器 + 版本)
  - Datapack 选择
  - 标签筛选 (过滤执行记录)
- 提交评估按钮

**结果展示**:

**对比表格**:
- 行: 算法名称/版本
- 列: Datapack 名称, Top-1/3/5 Accuracy, MRR, 执行时长
- 支持排序

**可视化对比**:
- 雷达图: 多维度对比 (准确率、速度、资源消耗)
- 柱状图: 准确率对比
- 折线图: 执行时长对比

**详细结果列表** (可展开):
- 每个算法的执行列表
- 链接到执行详情页

**关联 API**:
```
POST /api/v2/evaluations/datapacks
请求体:
{
  "specs": [
    {
      "algorithm": { "name": "algo1", "version": "v1.0.0" },
      "datapack": "datapack-name",
      "filter_labels": [{"key": "env", "value": "test"}]
    }
  ]
}

响应:
[
  {
    "algorithm": "algo1",
    "algorithm_version": "v1.0.0",
    "datapack": "datapack-name",
    "groundtruths": [...],
    "execution_refs": [
      {
        "execution_id": 123,
        "execution_duration": 45.2,
        "detector_results": [...],
        "predictions": [...],
        "executed_at": "2025-01-01T00:00:00Z"
      }
    ]
  }
]
```

#### 7.2 Dataset 评估页

**功能描述**: 对比多个算法在整个数据集上的表现。

**配置表单**:
- 添加评估规格:
  - 算法选择
  - 数据集 + 版本选择
  - 标签筛选

**结果展示**:
- 类似 Datapack 评估
- 聚合统计: 平均准确率、中位数、标准差

**关联 API**:
```
POST /api/v2/evaluations/datasets
```

#### 7.3 算法对比页

**功能描述**: 全局对比多个算法的综合表现。

**页面组件**:
- 算法选择器 (多选)
- 时间范围选择
- 对比维度选择:
  - 准确率
  - 执行速度
  - 资源消耗
  - 鲁棒性 (不同故障类型的表现)

**可视化图表**:
- 散点图: 准确率 vs 速度
- 热力图: 算法 × 故障类型准确率矩阵
- 小提琴图: 准确率分布

**关联 API**:
```
GET /api/v2/executions?algorithm_id={id}&lookback={duration}
(客户端聚合分析)
```

---

### 8. Tasks (任务监控)

#### 8.1 任务列表页

**功能描述**: 展示所有后台任务。

**页面组件**:

**筛选器**:
- 任务类型: SubmitInjection / BuildDatapack / FaultInjection / CollectResult / AlgorithmExecution
- 状态: Pending/Running/Completed/Error/Cancelled
- 是否立即执行 (Immediate)
- Trace ID / Group ID
- 项目 ID

**任务表格**:
- 列:
  - Task ID (UUID)
  - 任务类型 (徽章)
  - 状态 (进度条)
  - 关联 Trace
  - 关联 Group
  - 创建时间
  - 开始时间
  - 结束时间
  - 重试次数
- 操作: 查看详情、查看日志、取消任务

**批量操作**:
- 批量删除

**关联 API**:
```
GET /api/v2/tasks?task_type={type}&state={int}&status={int}&trace_id={id}&group_id={id}&project_id={id}
POST /api/v2/tasks/batch-delete
```

#### 8.2 任务详情页

**功能描述**: 展示单个任务的详细信息和实时日志。

**页面组件**:

**顶部信息卡**:
- Task ID
- 任务类型
- 状态徽章 (实时更新)
- 父任务 ID (链接)
- 关联 Trace ID (链接)
- 关联 Group ID
- 创建/开始/结束时间
- 重试次数/最大重试次数

**任务依赖树** (可视化):
- 树形图展示任务层级关系
- Grandfather (Group) → Father (Trace) → Task
- 子任务列表 (如果有)

**任务 Payload** (JSON 查看器):
- 任务输入参数
- 折叠/展开

**任务日志流** (SSE 实时):
- 实时日志输出
- 日志级别筛选
- 时间戳
- 自动滚动/暂停
- 下载完整日志

**关联 API**:
```
GET /api/v2/tasks/{task_id}
GET /api/v2/traces/{trace_id}/stream (SSE 日志流)
```

#### 8.3 任务统计页

**功能描述**: 展示任务执行的全局统计。

**页面组件**:

**统计卡片**:
- 今日任务总数
- 当前运行任务数
- 待处理任务数
- 失败任务数

**任务类型分布** (饼图):
- 按任务类型统计数量

**任务状态趋势** (折线图):
- 时间 × 任务状态
- 最近 7 天趋势

**任务耗时分析** (柱状图):
- 各任务类型的平均耗时

**失败任务分析** (表格):
- 最近失败任务列表
- 失败原因分类

**关联 API**:
```
GET /api/v2/tasks?page=1&size=1000 (客户端聚合统计)
GET /api/v2/traces/group/stats?group_id={id} (组统计)
```

---

### 9. System (系统管理)

#### 9.1 用户管理页

**功能描述**: 管理系统用户。

**页面组件**:

**用户表格**:
- 列:
  - 用户名
  - 邮箱
  - 状态 (活跃/禁用)
  - 全局角色
  - 创建时间
- 操作: 查看详情、编辑、删除、分配角色

**创建/编辑用户表单**:
- 用户名 (必填)
- 密码 (创建时必填)
- 邮箱
- 状态: 活跃/禁用
- 全局角色选择 (多选)

**用户详情页**:
- Tab 切换:
  1. **基本信息**
  2. **全局角色** (表格)
     - 角色名称、分配时间
     - 操作: 移除角色
  3. **全局权限** (表格)
     - 权限名称、操作、资源类型
     - 操作: 添加/移除权限
  4. **项目分配** (表格)
     - 项目名称、角色、分配时间
     - 操作: 移除分配
  5. **容器分配** (表格)
  6. **数据集分配** (表格)

**关联 API**:
```
GET /api/v2/users?page={page}&size={size}
POST /api/v2/users (创建用户)
PATCH /api/v2/users/{id} (更新用户)
DELETE /api/v2/users/{id}
GET /api/v2/users/{id}/detail

POST /api/v2/users/{user_id}/role/{role_id} (分配角色)
DELETE /api/v2/users/{user_id}/roles/{role_id} (移除角色)

POST /api/v2/users/{user_id}/permissions/assign (分配权限)
POST /api/v2/users/{user_id}/permissions/remove (移除权限)

POST /api/v2/users/{user_id}/projects/{project_id}/roles/{role_id} (项目分配)
DELETE /api/v2/users/{user_id}/projects/{project_id}

POST /api/v2/users/{user_id}/containers/{container_id}/roles/{role_id} (容器分配)
DELETE /api/v2/users/{user_id}/containers/{container_id}

POST /api/v2/users/{user_id}/datasets/{dataset_id}/roles/{role_id} (数据集分配)
DELETE /api/v2/users/{user_id}/datasets/{dataset_id}
```

#### 9.2 角色管理页

**功能描述**: 管理系统角色。

**页面组件**:

**角色表格**:
- 列:
  - 角色名称
  - 描述
  - 是否系统角色 (系统角色不可删除)
  - 权限数量
  - 用户数量
  - 创建时间
- 操作: 查看详情、编辑、删除、分配权限

**创建/编辑角色表单**:
- 角色名称 (必填)
- 描述
- 权限选择 (多选, 分组显示)

**角色详情页**:
- Tab 切换:
  1. **基本信息**
  2. **权限列表** (表格)
     - 权限名称、操作、资源类型
     - 操作: 移除权限、添加权限
  3. **用户列表** (表格)
     - 用户名、邮箱、分配时间

**关联 API**:
```
GET /api/v2/roles?page={page}&size={size}
POST /api/v2/roles (创建角色)
GET /api/v2/roles/{id}
PATCH /api/v2/roles/{id} (更新角色)
DELETE /api/v2/roles/{id}

POST /api/v2/roles/{role_id}/permissions/assign (分配权限)
POST /api/v2/roles/{role_id}/permissions/remove (移除权限)

GET /api/v2/roles/{role_id}/users (获取用户列表)
```

#### 9.3 权限管理页

**功能描述**: 管理系统权限。

**页面组件**:

**权限表格**:
- 列:
  - 权限名称
  - 操作 (read/write/delete/execute)
  - 资源类型 (project/container/dataset/injection/execution)
  - 描述
  - 是否系统权限 (系统权限不可删除)
  - 状态
- 筛选: 操作类型、资源类型、是否系统权限
- 操作: 查看详情、编辑、删除

**创建/编辑权限表单**:
- 权限名称 (必填, 格式: action:resource)
- 操作类型 (下拉: read/write/delete/execute)
- 资源类型 (下拉)
- 描述

**权限详情页**:
- Tab 切换:
  1. **基本信息**
  2. **角色列表** (拥有此权限的角色)

**关联 API**:
```
GET /api/v2/permissions?action={action}&is_system={bool}&status={int}
POST /api/v2/permissions
GET /api/v2/permissions/{id}
PUT /api/v2/permissions/{id}
DELETE /api/v2/permissions/{id}
GET /api/v2/permissions/{permission_id}/roles
```

#### 9.4 标签管理页

**功能描述**: 管理自定义标签。

**页面组件**:

**标签表格**:
- 列:
  - Key
  - Value
  - 分类 (project/container/dataset/injection/execution)
  - 是否系统标签
  - 使用次数
  - 状态
- 筛选: Key, Value, 分类, 是否系统标签
- 操作: 编辑、删除

**批量操作**:
- 批量删除

**创建/编辑标签表单**:
- Key (必填)
- Value (必填)
- 分类 (下拉)
- 描述

**关联 API**:
```
GET /api/v2/labels?key={key}&value={value}&category={category}&is_system={bool}
POST /api/v2/labels
GET /api/v2/labels/{label_id}
PATCH /api/v2/labels/{label_id}
DELETE /api/v2/labels/{label_id}
POST /api/v2/labels/batch-delete
```

#### 9.5 资源管理页

**功能描述**: 查看系统资源 (只读)。

**页面组件**:

**资源表格**:
- 列:
  - 资源名称
  - 资源类型 (project/container/dataset/etc.)
  - 分类
  - 创建时间
- 操作: 查看详情、查看权限

**资源详情页**:
- 资源基本信息
- 关联权限列表

**关联 API**:
```
GET /api/v2/resources?type={type}&category={category}
GET /api/v2/resources/{id}
GET /api/v2/resources/{id}/permissions
```

---

### 10. Settings (个人设置)

#### 10.1 个人资料页

**功能描述**: 查看和编辑个人信息。

**页面组件**:

**个人信息卡片**:
- 用户名 (只读)
- 邮箱 (可编辑)
- 头像 (上传)
- 创建时间

**编辑表单**:
- 邮箱
- 提交按钮

**关联 API**:
```
GET /api/v2/auth/profile
PATCH /api/v2/users/{id} (更新邮箱)
```

#### 10.2 修改密码页

**功能描述**: 修改登录密码。

**表单字段**:
- 当前密码 (密码输入框)
- 新密码 (密码输入框)
- 确认新密码 (密码输入框)
- 提交按钮

**关联 API**:
```
POST /api/v2/auth/change-password
请求体:
{
  "old_password": "string",
  "new_password": "string"
}
```

---

## 通用组件

### 1. 标签选择器 (LabelSelector)

**功能**: 多选标签, 支持创建新标签。

**Props**:
- `value`: 已选标签数组
- `category`: 标签分类 (过滤可选标签)
- `onChange`: 选择变化回调

**API**:
```
GET /api/v2/labels?category={category}
POST /api/v2/labels (创建新标签)
```

---

### 2. 容器选择器 (ContainerSelector)

**功能**: 选择容器及其版本。

**Props**:
- `containerType`: 容器类型 (Pedestal/Benchmark/Algorithm)
- `value`: { name, version }
- `onChange`: 选择变化回调

**API**:
```
GET /api/v2/containers?type={type}
GET /api/v2/containers/{container_id}/versions
```

---

### 3. 实时日志流组件 (LogStream)

**功能**: SSE 实时日志显示。

**Props**:
- `traceId`: Trace ID
- `autoScroll`: 是否自动滚动

**API**:
```
GET /api/v2/traces/{trace_id}/stream (SSE)
```

**实现要点**:
- 使用 EventSource 建立 SSE 连接
- 自动重连机制
- 日志级别颜色编码
- 虚拟滚动 (处理大量日志)

---

### 4. 任务状态徽章 (TaskStatusBadge)

**功能**: 显示任务状态的彩色徽章。

**Props**:
- `state`: Pending/Running/Completed/Error/Cancelled
- `status`: 数值状态码

**样式映射**:
- Pending: 灰色
- Running: 蓝色 (带动画)
- Completed: 绿色
- Error: 红色
- Cancelled: 橙色

---

### 5. 故障类型图标 (FaultTypeIcon)

**功能**: 根据故障类型显示对应图标。

**Props**:
- `faultType`: Network/CPU/Memory/IO/Pod/Container

---

### 6. 时间选择器 (TimeRangePicker)

**功能**: 选择时间范围或使用快捷选项。

**Props**:
- `value`: { start, end }
- `shortcuts`: ['1h', '24h', '7d']
- `onChange`: 回调

---

### 7. 分页器 (Pagination)

**功能**: 标准分页组件。

**Props**:
- `page`: 当前页
- `size`: 每页大小
- `total`: 总数
- `onChange`: 回调

---

### 8. 状态进度条 (StateProgressBar)

**功能**: 展示多阶段任务的进度。

**Props**:
- `stages`: 阶段数组 ['BUILDING', 'READY', 'INJECTING', 'COMPLETE']
- `currentStage`: 当前阶段

---

### 9. JSON 查看器 (JsonViewer)

**功能**: 格式化显示 JSON, 支持折叠/展开。

**Props**:
- `data`: JSON 对象
- `collapsible`: 是否可折叠

---

### 10. Markdown 渲染器 (MarkdownRenderer)

**功能**: 渲染 Markdown 内容。

**Props**:
- `content`: Markdown 字符串

**库推荐**: react-markdown

---

## API 映射

### 认证相关

| API 端点 | 方法 | 页面/组件 | 功能 |
|---------|------|----------|------|
| `/api/v2/auth/register` | POST | 注册页 | 用户注册 |
| `/api/v2/auth/login` | POST | 登录页 | 用户登录 |
| `/api/v2/auth/refresh` | POST | 全局 | 刷新 Token |
| `/api/v2/auth/logout` | POST | 全局 | 登出 |
| `/api/v2/auth/profile` | GET | 个人资料页 | 获取个人信息 |
| `/api/v2/auth/change-password` | POST | 修改密码页 | 修改密码 |

---

### 项目管理

| API 端点 | 方法 | 页面/组件 | 功能 |
|---------|------|----------|------|
| `/api/v2/projects` | GET | 项目列表页 | 获取项目列表 |
| `/api/v2/projects` | POST | 创建项目页 | 创建项目 |
| `/api/v2/projects/{id}` | GET | 项目详情页 | 获取项目详情 |
| `/api/v2/projects/{id}` | PATCH | 编辑项目页 | 更新项目 |
| `/api/v2/projects/{id}` | DELETE | 项目列表页 | 删除项目 |
| `/api/v2/projects/{id}/labels` | PATCH | 项目详情页 | 管理标签 |

---

### 容器管理

| API 端点 | 方法 | 页面/组件 | 功能 |
|---------|------|----------|------|
| `/api/v2/containers` | GET | 容器列表页 | 获取容器列表 |
| `/api/v2/containers` | POST | 创建容器页 | 创建容器 |
| `/api/v2/containers/{id}` | GET | 容器详情页 | 获取容器详情 |
| `/api/v2/containers/{id}` | PATCH | 编辑容器页 | 更新容器 |
| `/api/v2/containers/{id}` | DELETE | 容器列表页 | 删除容器 |
| `/api/v2/containers/{id}/versions` | GET | 容器详情页 | 获取版本列表 |
| `/api/v2/containers/{id}/versions` | POST | 容器详情页 | 创建版本 |
| `/api/v2/containers/{id}/versions/{vid}` | GET | 容器详情页 | 获取版本详情 |
| `/api/v2/containers/{id}/versions/{vid}` | PATCH | 编辑版本页 | 更新版本 |
| `/api/v2/containers/{id}/versions/{vid}` | DELETE | 容器详情页 | 删除版本 |
| `/api/v2/containers/{id}/versions/{vid}/helm-values` | POST | 容器详情页 | 上传 Helm Values |
| `/api/v2/containers/build` | POST | 容器构建页 | 提交构建任务 |

---

### 数据集管理

| API 端点 | 方法 | 页面/组件 | 功能 |
|---------|------|----------|------|
| `/api/v2/datasets` | GET | 数据集列表页 | 获取数据集列表 |
| `/api/v2/datasets` | POST | 创建数据集页 | 创建数据集 |
| `/api/v2/datasets/search` | POST | 数据集列表页 | 高级搜索 |
| `/api/v2/datasets/{id}` | GET | 数据集详情页 | 获取数据集详情 |
| `/api/v2/datasets/{id}` | PATCH | 编辑数据集页 | 更新数据集 |
| `/api/v2/datasets/{id}` | DELETE | 数据集列表页 | 删除数据集 |
| `/api/v2/datasets/{id}/versions` | GET | 数据集详情页 | 获取版本列表 |
| `/api/v2/datasets/{id}/versions` | POST | 上传版本页 | 创建版本 |
| `/api/v2/datasets/{id}/versions/{vid}` | GET | 数据集详情页 | 获取版本详情 |
| `/api/v2/datasets/{id}/versions/{vid}` | PATCH | 编辑版本页 | 更新版本 |
| `/api/v2/datasets/{id}/versions/{vid}` | DELETE | 数据集详情页 | 删除版本 |
| `/api/v2/datasets/{id}/versions/{vid}/download` | GET | 数据集详情页 | 下载版本 |
| `/api/v2/datasets/{id}/version/{vid}/injections` | PATCH | 数据集详情页 | 管理注入关联 |

---

### 故障注入

| API 端点 | 方法 | 页面/组件 | 功能 |
|---------|------|----------|------|
| `/api/v2/injections` | GET | 注入列表页 | 获取注入列表 |
| `/api/v2/injections/search` | POST | 注入列表页 | 高级搜索 |
| `/api/v2/injections/{id}` | GET | 注入详情页 | 获取注入详情 |
| `/api/v2/injections/batch-delete` | POST | 注入列表页 | 批量删除 |
| `/api/v2/injections/metadata` | GET | 创建注入页 | 获取故障配置元数据 |
| `/api/v2/injections/{id}/labels` | PATCH | 注入详情页 | 管理标签 |
| `/api/v2/injections/labels/batch` | PATCH | 注入列表页 | 批量管理标签 |
| `/api/v2/injections/analysis/no-issues` | GET | 注入列表页 | 无异常注入 |
| `/api/v2/injections/analysis/with-issues` | GET | 注入列表页 | 有异常注入 |
| `/api/v2/injections/inject` | POST | 创建注入页 | 提交注入任务 |
| `/api/v2/injections/build` | POST | Datapack 构建页 | 提交构建任务 |

---

### 算法执行

| API 端点 | 方法 | 页面/组件 | 功能 |
|---------|------|----------|------|
| `/api/v2/executions` | GET | 执行列表页 | 获取执行列表 |
| `/api/v2/executions/{id}` | GET | 执行详情页 | 获取执行详情 |
| `/api/v2/executions/batch-delete` | POST | 执行列表页 | 批量删除 |
| `/api/v2/executions/labels` | GET | 执行列表页 | 获取可用标签 |
| `/api/v2/executions/{id}/labels` | PATCH | 执行详情页 | 管理标签 |
| `/api/v2/executions/execute` | POST | 创建执行页 | 提交执行任务 |
| `/api/v2/executions/{id}/detector_results` | POST | 执行详情页 | 上传检测结果 |
| `/api/v2/executions/{id}/granularity_results` | POST | 执行详情页 | 上传粒度结果 |

---

### 评估

| API 端点 | 方法 | 页面/组件 | 功能 |
|---------|------|----------|------|
| `/api/v2/evaluations/datapacks` | POST | Datapack 评估页 | 评估 Datapack |
| `/api/v2/evaluations/datasets` | POST | Dataset 评估页 | 评估 Dataset |

---

### 任务管理

| API 端点 | 方法 | 页面/组件 | 功能 |
|---------|------|----------|------|
| `/api/v2/tasks` | GET | 任务列表页 | 获取任务列表 |
| `/api/v2/tasks/{task_id}` | GET | 任务详情页 | 获取任务详情 |
| `/api/v2/tasks/batch-delete` | POST | 任务列表页 | 批量删除 |
| `/api/v2/traces/{trace_id}/stream` | GET (SSE) | 任务详情页/注入详情页 | 实时日志流 |
| `/api/v2/traces/group/stats` | GET | 任务统计页 | 获取组统计 |

---

### 用户管理

| API 端点 | 方法 | 页面/组件 | 功能 |
|---------|------|----------|------|
| `/api/v2/users` | GET | 用户列表页 | 获取用户列表 |
| `/api/v2/users` | POST | 创建用户页 | 创建用户 |
| `/api/v2/users/{id}/detail` | GET | 用户详情页 | 获取用户详情 |
| `/api/v2/users/{id}` | PATCH | 编辑用户页 | 更新用户 |
| `/api/v2/users/{id}` | DELETE | 用户列表页 | 删除用户 |
| `/api/v2/users/{uid}/role/{rid}` | POST | 用户详情页 | 分配角色 |
| `/api/v2/users/{uid}/roles/{rid}` | DELETE | 用户详情页 | 移除角色 |
| `/api/v2/users/{uid}/permissions/assign` | POST | 用户详情页 | 分配权限 |
| `/api/v2/users/{uid}/permissions/remove` | POST | 用户详情页 | 移除权限 |
| `/api/v2/users/{uid}/projects/{pid}/roles/{rid}` | POST | 用户详情页 | 项目分配 |
| `/api/v2/users/{uid}/projects/{pid}` | DELETE | 用户详情页 | 移除项目 |
| `/api/v2/users/{uid}/containers/{cid}/roles/{rid}` | POST | 用户详情页 | 容器分配 |
| `/api/v2/users/{uid}/containers/{cid}` | DELETE | 用户详情页 | 移除容器 |
| `/api/v2/users/{uid}/datasets/{did}/roles/{rid}` | POST | 用户详情页 | 数据集分配 |
| `/api/v2/users/{uid}/datasets/{did}` | DELETE | 用户详情页 | 移除数据集 |

---

### 角色管理

| API 端点 | 方法 | 页面/组件 | 功能 |
|---------|------|----------|------|
| `/api/v2/roles` | GET | 角色列表页 | 获取角色列表 |
| `/api/v2/roles` | POST | 创建角色页 | 创建角色 |
| `/api/v2/roles/{id}` | GET | 角色详情页 | 获取角色详情 |
| `/api/v2/roles/{id}` | PATCH | 编辑角色页 | 更新角色 |
| `/api/v2/roles/{id}` | DELETE | 角色列表页 | 删除角色 |
| `/api/v2/roles/{rid}/permissions/assign` | POST | 角色详情页 | 分配权限 |
| `/api/v2/roles/{rid}/permissions/remove` | POST | 角色详情页 | 移除权限 |
| `/api/v2/roles/{rid}/users` | GET | 角色详情页 | 获取用户列表 |

---

### 权限管理

| API 端点 | 方法 | 页面/组件 | 功能 |
|---------|------|----------|------|
| `/api/v2/permissions` | GET | 权限列表页 | 获取权限列表 |
| `/api/v2/permissions` | POST | 创建权限页 | 创建权限 |
| `/api/v2/permissions/{id}` | GET | 权限详情页 | 获取权限详情 |
| `/api/v2/permissions/{id}` | PUT | 编辑权限页 | 更新权限 |
| `/api/v2/permissions/{id}` | DELETE | 权限列表页 | 删除权限 |
| `/api/v2/permissions/{id}/roles` | GET | 权限详情页 | 获取角色列表 |

---

### 标签管理

| API 端点 | 方法 | 页面/组件 | 功能 |
|---------|------|----------|------|
| `/api/v2/labels` | GET | 标签列表页/标签选择器 | 获取标签列表 |
| `/api/v2/labels` | POST | 标签列表页/标签选择器 | 创建标签 |
| `/api/v2/labels/{id}` | GET | 标签详情页 | 获取标签详情 |
| `/api/v2/labels/{id}` | PATCH | 编辑标签页 | 更新标签 |
| `/api/v2/labels/{id}` | DELETE | 标签列表页 | 删除标签 |
| `/api/v2/labels/batch-delete` | POST | 标签列表页 | 批量删除 |

---

### 资源管理

| API 端点 | 方法 | 页面/组件 | 功能 |
|---------|------|----------|------|
| `/api/v2/resources` | GET | 资源列表页 | 获取资源列表 |
| `/api/v2/resources/{id}` | GET | 资源详情页 | 获取资源详情 |
| `/api/v2/resources/{id}/permissions` | GET | 资源详情页 | 获取资源权限 |

---

## 技术栈建议

### 前端框架
- **React 18+** 或 **Vue 3+**
- **TypeScript** (强类型安全)

### UI 组件库
- **Ant Design** (React) 或 **Element Plus** (Vue)
- **Tailwind CSS** (样式定制)

### 状态管理
- **Zustand** (React, 轻量级) 或 **Pinia** (Vue)
- **React Query** / **TanStack Query** (服务端状态管理, API 缓存)

### 路由
- **React Router v6** (React) 或 **Vue Router** (Vue)

### 数据可视化
- **ECharts** (图表库, 支持复杂图表)
- **D3.js** (自定义可视化, 服务拓扑图)
- **Cytoscape.js** (图数据库可视化, 任务依赖树)

### 实时通信
- **EventSource** (SSE, 浏览器原生 API)
- **reconnecting-eventsource** (自动重连封装)

### 代码编辑器
- **Monaco Editor** (VS Code 同款, JSON/YAML 编辑)
- **CodeMirror** (轻量级代码编辑器)

### Markdown 渲染
- **react-markdown** (React) 或 **vue-markdown** (Vue)

### 文件上传
- **react-dropzone** (React) 或 **vue-upload-component** (Vue)

### 表单验证
- **React Hook Form** (React) 或 **VeeValidate** (Vue)
- **Zod** (TypeScript 类型安全的 schema 验证)

### API 客户端
- **Axios** (HTTP 客户端)
- **OpenAPI Generator** (根据 Swagger 生成 TypeScript SDK)

### 构建工具
- **Vite** (快速构建工具)

### 测试
- **Vitest** (单元测试)
- **Playwright** (E2E 测试)

---

## 实时功能

### SSE (Server-Sent Events) 集成

RCABench 使用 SSE 提供实时日志流和任务状态更新。

#### 实现要点

**1. 建立 SSE 连接**:
```typescript
const eventSource = new EventSource(`/api/v2/traces/${traceId}/stream`, {
  headers: {
    'Authorization': `Bearer ${token}`
  }
});

eventSource.onmessage = (event) => {
  const logEntry = JSON.parse(event.data);
  appendLog(logEntry);
};

eventSource.onerror = (error) => {
  console.error('SSE connection error:', error);
  // 实现重连逻辑
};
```

**2. 日志事件格式**:
```json
{
  "timestamp": "2025-01-01T00:00:00Z",
  "level": "INFO",
  "message": "Task started",
  "task_id": "uuid",
  "trace_id": "trace-id"
}
```

**3. 自动重连**:
- 使用 `reconnecting-eventsource` 库
- 指数退避策略
- 断线后自动恢复

**4. 连接管理**:
- 组件卸载时关闭连接
- 避免重复连接
- 处理网络切换

#### 实时更新页面

1. **任务详情页**: 实时日志流
2. **注入详情页**: 任务状态实时更新
3. **执行详情页**: 任务进度实时更新
4. **Dashboard**: 最近活动流

---

## 附录

### A. 任务状态映射

| State | 名称 | 含义 | 前端显示 |
|-------|------|------|---------|
| 0 | PENDING | 待处理 | 灰色徽章 |
| 1 | RUNNING | 运行中 | 蓝色徽章 (动画) |
| 2 | COMPLETED | 已完成 | 绿色徽章 |
| 3 | ERROR | 错误 | 红色徽章 |
| 4 | CANCELLED | 已取消 | 橙色徽章 |

### B. 故障注入状态映射

| State | 名称 | 含义 |
|-------|------|------|
| 0 | BUILDING_DATAPACK | 构建数据包中 |
| 1 | DATAPACK_READY | 数据包就绪 |
| 2 | INJECTING | 注入故障中 |
| 3 | INJECTION_COMPLETE | 注入完成 |
| 4 | COLLECTING_RESULT | 收集结果中 |
| 5 | COMPLETED | 已完成 |
| -1 | ERROR | 错误 |

### C. 执行状态映射

| State | 名称 | 含义 |
|-------|------|------|
| 0 | PENDING | 待执行 |
| 1 | RUNNING | 执行中 |
| 2 | COMPLETED | 已完成 |
| -1 | ERROR | 错误 |

### D. 故障类型列表

- **Network**: NetworkDelay, NetworkLoss, NetworkCorrupt, NetworkDuplicate, NetworkPartition
- **CPU**: CPUStress
- **Memory**: MemoryStress
- **IO**: IOStress, IOFault
- **Pod**: PodKill, PodFailure
- **Container**: ContainerKill

### E. 粒度级别

- **Service**: 服务级定位
- **Pod**: Pod 级定位
- **Span**: Span 级定位
- **Metric**: 指标级定位

---

## 总结

本设计文档涵盖了 RCABench 前端应用的完整页面结构、功能描述、API 映射和技术栈建议。

**核心页面**:
1. Dashboard (仪表盘)
2. Projects (项目管理)
3. Containers (容器管理)
4. Datasets (数据集管理)
5. Injections (故障注入, 包含可视化故障编排器)
6. Executions (算法执行, 包含结果可视化)
7. Evaluations (评估与对比)
8. Tasks (任务监控)
9. System (系统管理: 用户/角色/权限/标签/资源)
10. Settings (个人设置)

**关键功能**:
- 故障注入可视化编排 (批次 + 并行故障节点)
- 实时日志流 (SSE)
- 算法结果可视化 (服务拓扑图、分层结果、准确率分析)
- 多维度评估与对比
- RBAC 权限管理
- 标签系统

**技术要点**:
- TypeScript 类型安全
- React Query 服务端状态管理
- ECharts / D3.js 可视化
- SSE 实时通信
- OpenAPI Generator 生成 SDK

该文档可作为前端团队的开发指南和交付依据。
