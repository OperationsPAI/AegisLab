# API v2 权限规则设计文档

本文档定义了 API v2 中所有端点的权限要求，包括 **动作（Action）**、**资源（Resource）** 和 **作用域（Scope）**。

## 权限规则格式

```
{resource}:{action}:{scope}
```

- **Resource**: 资源类型（如 `container`, `dataset`, `project` 等）
- **Action**: 动作类型（见下方动作分类）
- **Scope**: `own` (自己的资源), `project` (项目资源), `team` (团队资源), `all` (所有资源)

**示例**:

- `project:read:own` - 只能读取自己的项目
- `project:create:own` - 可以创建自己的项目
- `injection:execute:project` - 可以执行项目内的故障注入
- `dataset:update:team` - 可以修改团队的数据集
- `container:delete:all` - 可以删除所有容器（管理员）
- `dataset_version:download:team` - 可以下载团队的数据集版本
- `role:grant:all` - 可以授予任何角色的权限

---

## 动作原语（Action）分类

### 1. 基础 CRUD 操作

- **`create`** - 创建新资源
- **`read`** - 读取/查看资源
- **`update`** - 更新现有资源
- **`delete`** - 删除资源

### 2. 执行和状态管理

- **`execute`** - 执行任务、注入、构建等操作
- **`stop`** - 停止正在运行的任务或执行
- **`restart`** - 重启服务或任务
- **`activate`** - 激活/启用资源
- **`suspend`** - 暂停/禁用资源

### 3. 文件和数据操作

- **`upload`** - 上传文件、Chart、数据集等
- **`download`** - 下载文件、结果、数据包等
- **`import`** - 导入数据、配置
- **`export`** - 导出数据、报告、备份

### 4. 权限和成员管理

- **`assign`** - 分配用户到资源、角色到用户
- **`grant`** - 授予权限或访问权利
- **`revoke`** - 撤销权限或访问权利

### 5. 配置和管理

- **`configure`** - 配置系统设置、资源参数
- **`manage`** - 完全管理控制（包含所有操作）

### 6. 协作和共享

- **`share`** - 分享资源给他人
- **`clone`** - 克隆/复制资源

### 7. 监控和分析

- **`monitor`** - 监控系统指标、追踪
- **`analyze`** - 分析结果、评估性能
- **`audit`** - 查看审计日志和历史记录

---

## 动作原语使用场景

| 动作        | 典型使用场景       | 示例端点                                                   |
| ----------- | ------------------ | ---------------------------------------------------------- |
| `create`    | 创建新资源         | `POST /projects`, `POST /containers`                       |
| `read`      | 查看资源列表、详情 | `GET /containers`, `GET /datasets/:id`                     |
| `update`    | 更新现有资源       | `PATCH /containers/:id`, `PATCH /projects/:id`             |
| `delete`    | 删除资源           | `DELETE /datasets/:id`, `POST /tasks/batch-delete`         |
| `execute`   | 执行故障注入、算法 | `POST /injections/inject`, `POST /executions/execute`      |
| `stop`      | 停止运行中的任务   | `POST /tasks/:id/stop`, `POST /executions/:id/stop`        |
| `upload`    | 上传文件           | `POST /containers/:id/versions/:vid/helm-chart`            |
| `download`  | 下载数据包、结果   | `GET /datasets/:id/versions/:vid/download`                 |
| `import`    | 导入外部数据       | `POST /datasets/import`, `POST /configurations/import`     |
| `export`    | 导出数据、报告     | `POST /executions/export`, `GET /metrics/export`           |
| `assign`    | 分配角色、成员     | `POST /users/:id/roles/:rid`, `POST /teams/:id/members`    |
| `grant`     | 授予权限           | `POST /roles/:id/permissions/assign`                       |
| `revoke`    | 撤销权限           | `POST /roles/:id/permissions/remove`                       |
| `configure` | 配置系统设置       | `PATCH /system/settings`, `POST /resources/configure`      |
| `manage`    | 全面管理           | `PATCH /teams/:id/members/:uid/role`                       |
| `share`     | 分享资源           | `POST /projects/:id/share`, `POST /datasets/:id/share`     |
| `clone`     | 克隆资源           | `POST /injections/:id/clone`, `POST /containers/:id/clone` |
| `monitor`   | 监控指标           | `GET /system/metrics`, `GET /traces/:id/stream`            |
| `analyze`   | 分析评估           | `POST /evaluations/datasets`, `GET /injections/analysis/*` |
| `audit`     | 查看审计日志       | `GET /audit/logs`, `GET /users/:id/activity`               |

---

## 特殊说明

- 🔓 **公开访问**: 无需权限或仅需 JWT 认证
- 👤 **所有者检查**: 需要额外检查是否为资源所有者
- 👥 **团队成员检查**: 需要检查是否为团队成员
- ⚡ **复合权限**: 需要满足多个权限条件之一

---

## 1. 认证相关 API (`/api/v2/auth`)

| HTTP方法 | 端点                    | 权限规则             | 说明               |
| -------- | ----------------------- | -------------------- | ------------------ |
| POST     | `/auth/login`           | 🔓 公开              | 用户登录           |
| POST     | `/auth/register`        | 🔓 公开              | 用户注册           |
| POST     | `/auth/refresh`         | 🔓 公开              | 刷新令牌           |
| POST     | `/auth/logout`          | JWT 认证             | 用户登出           |
| POST     | `/auth/change-password` | JWT 认证 + 👤 所有者 | 修改密码（仅自己） |
| GET      | `/auth/profile`         | JWT 认证             | 获取当前用户信息   |

---

## 2. 容器管理 API (`/api/v2/containers`)

### 2.1 容器主资源

| HTTP方法 | 端点                               | 权限规则                                                                                       | 说明                              |
| -------- | ---------------------------------- | ---------------------------------------------------------------------------------------------- | --------------------------------- |
| GET      | `/containers`                      | `container:read:own` ⚡ `container:read:team` ⚡ `container:read:all`                          | 列出容器（根据scope返回不同范围） |
| GET      | `/containers/:container_id`        | `container:read:own` ⚡ `container:read:team` ⚡ `container:read:all` + 👤 所有者检查          | 获取容器详情                      |
| POST     | `/containers`                      | `container:create:own`                                                                         | 创建容器（创建者自动成为owner）   |
| POST     | `/containers/build`                | `container:execute:own` ⚡ `container:execute:team` ⚡ `container:execute:all` + 👤 所有者检查 | 构建容器                          |
| PATCH    | `/containers/:container_id`        | `container:update:own` ⚡ `container:update:team` ⚡ `container:update:all` + 👤 所有者检查    | 更新容器                          |
| PATCH    | `/containers/:container_id/labels` | `container:update:own` ⚡ `container:update:team` ⚡ `container:update:all` + 👤 所有者检查    | 管理容器标签                      |
| DELETE   | `/containers/:container_id`        | `container:delete:own` ⚡ `container:delete:all` + 👤 所有者检查                               | 删除容器                          |

### 2.2 容器版本子资源

| HTTP方法 | 端点                                                         | 权限规则                                                                                                       | 说明             |
| -------- | ------------------------------------------------------------ | -------------------------------------------------------------------------------------------------------------- | ---------------- |
| GET      | `/containers/:container_id/versions`                         | `read:container_version:own` ⚡ `read:container_version:team` ⚡ `read:container_version:all` + 父容器权限检查 | 列出容器版本     |
| GET      | `/containers/:container_id/versions/:version_id`             | `read:container_version:own` ⚡ `read:container_version:team` ⚡ `read:container_version:all` + 父容器权限检查 | 获取版本详情     |
| POST     | `/containers/:container_id/versions`                         | `container_version:create:own` ⚡ `container_version:create:team` + 父容器权限检查                             | 创建容器版本     |
| POST     | `/containers/:container_id/versions/:version_id/helm-chart`  | `container_version:update:own` ⚡ `container_version:update:team` + 父容器权限检查                             | 上传 Helm Chart  |
| POST     | `/containers/:container_id/versions/:version_id/helm-values` | `container_version:update:own` ⚡ `container_version:update:team` + 父容器权限检查                             | 上传 Helm Values |
| PATCH    | `/containers/:container_id/versions/:version_id`             | `container_version:update:own` ⚡ `container_version:update:team` + 父容器权限检查                             | 更新容器版本     |
| DELETE   | `/containers/:container_id/versions/:version_id`             | `delete:container_version:own` ⚡ `delete:container_version:all` + 父容器权限检查                              | 删除容器版本     |

---

## 3. 数据集管理 API (`/api/v2/datasets`)

### 3.1 数据集主资源

| HTTP方法 | 端点                           | 权限规则                                                                              | 说明           |
| -------- | ------------------------------ | ------------------------------------------------------------------------------------- | -------------- |
| GET      | `/datasets`                    | `dataset:read:own` ⚡ `dataset:read:team` ⚡ `dataset:read:all`                       | 列出数据集     |
| GET      | `/datasets/:dataset_id`        | `dataset:read:own` ⚡ `dataset:read:team` ⚡ `dataset:read:all` + 👤 所有者检查       | 获取数据集详情 |
| POST     | `/datasets`                    | `dataset:create:own`                                                                  | 创建数据集     |
| PATCH    | `/datasets/:dataset_id`        | `dataset:update:own` ⚡ `dataset:update:team` ⚡ `dataset:update:all` + 👤 所有者检查 | 更新数据集     |
| PATCH    | `/datasets/:dataset_id/labels` | `dataset:update:own` ⚡ `dataset:update:team` ⚡ `dataset:update:all` + 👤 所有者检查 | 管理数据集标签 |
| DELETE   | `/datasets/:dataset_id`        | `dataset:delete:own` ⚡ `dataset:delete:all` + 👤 所有者检查                          | 删除数据集     |

### 3.2 数据集版本子资源

| HTTP方法 | 端点                                                    | 权限规则                                                                                                   | 说明           |
| -------- | ------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------- | -------------- |
| GET      | `/datasets/:dataset_id/versions`                        | `read:dataset_version:own` ⚡ `read:dataset_version:team` ⚡ `read:dataset_version:all` + 父数据集权限检查 | 列出数据集版本 |
| GET      | `/datasets/:dataset_id/versions/:version_id`            | `read:dataset_version:own` ⚡ `read:dataset_version:team` ⚡ `read:dataset_version:all` + 父数据集权限检查 | 获取版本详情   |
| GET      | `/datasets/:dataset_id/versions/:version_id/download`   | `read:dataset_version:own` ⚡ `read:dataset_version:team` ⚡ `read:dataset_version:all` + 父数据集权限检查 | 下载数据集版本 |
| POST     | `/datasets/:dataset_id/versions`                        | `dataset_version:create:own` ⚡ `dataset_version:create:team` + 父数据集权限检查                           | 创建数据集版本 |
| PATCH    | `/datasets/:dataset_id/versions/:version_id`            | `dataset_version:update:own` ⚡ `dataset_version:update:team` + 父数据集权限检查                           | 更新数据集版本 |
| PATCH    | `/datasets/:dataset_id/versions/:version_id/injections` | `dataset_version:update:own` ⚡ `dataset_version:update:team` + 父数据集权限检查                           | 管理版本注入   |
| DELETE   | `/datasets/:dataset_id/versions/:version_id`            | `delete:dataset_version:own` ⚡ `delete:dataset_version:all` + 父数据集权限检查                            | 删除数据集版本 |

---

## 4. 项目管理 API (`/api/v2/projects`)

| HTTP方法 | 端点                               | 权限规则                                                                               | 说明         |
| -------- | ---------------------------------- | -------------------------------------------------------------------------------------- | ------------ |
| GET      | `/projects`                        | `project:read:own` ⚡ `project:read:team` ⚡ `project:read:all`                        | 列出项目     |
| GET      | `/projects/:project_id`            | `project:read:own` ⚡ `project:read:team` ⚡ `project:read:all` + 项目成员检查         | 获取项目详情 |
| GET      | `/projects/:project_id/injections` | `project:read:own` ⚡ `project:read:team` ⚡ `project:read:all` + 项目成员检查         | 列出项目注入 |
| GET      | `/projects/:project_id/executions` | `project:read:own` ⚡ `project:read:team` ⚡ `project:read:all` + 项目成员检查         | 列出项目执行 |
| POST     | `/projects`                        | `project:create:own`                                                                   | 创建项目     |
| PATCH    | `/projects/:project_id`            | `project:update:own` ⚡ `project:update:team` ⚡ `project:update:all` + 项目管理员检查 | 更新项目     |
| PATCH    | `/projects/:project_id/labels`     | `project:update:own` ⚡ `project:update:team` ⚡ `project:update:all` + 项目管理员检查 | 管理项目标签 |
| DELETE   | `/projects/:project_id`            | `project:delete:own` ⚡ `project:delete:all` + 项目管理员检查                          | 删除项目     |

---

## 5. 团队管理 API (`/api/v2/teams`)

| HTTP方法 | 端点                                    | 权限规则                                                | 说明                                 |
| -------- | --------------------------------------- | ------------------------------------------------------- | ------------------------------------ |
| GET      | `/teams`                                | JWT 认证                                                | 列出公开团队和用户所属团队           |
| GET      | `/teams/:team_id`                       | `team:read:own` + 👥 团队成员检查 OR 公开团队           | 获取团队详情                         |
| GET      | `/teams/:team_id/members`               | `team:read:own` + 👥 团队成员检查 OR 公开团队           | 列出团队成员                         |
| GET      | `/teams/:team_id/projects`              | `team:read:own` + 👥 团队成员检查 OR 公开团队           | 列出团队项目                         |
| POST     | `/teams`                                | JWT 认证                                                | 创建团队（创建者自动成为团队管理员） |
| PATCH    | `/teams/:team_id`                       | `team:update:own` + 团队管理员检查 OR `team:manage:all` | 更新团队                             |
| DELETE   | `/teams/:team_id`                       | `team:delete:own` + 团队管理员检查 OR `team:delete:all` | 删除团队                             |
| POST     | `/teams/:team_id/members`               | `team:manage:own` + 团队管理员检查 OR `team:manage:all` | 添加团队成员                         |
| DELETE   | `/teams/:team_id/members/:user_id`      | `team:manage:own` + 团队管理员检查 OR `team:manage:all` | 移除团队成员                         |
| PATCH    | `/teams/:team_id/members/:user_id/role` | `team:manage:own` + 团队管理员检查 OR `team:manage:all` | 更新成员角色                         |

---

## 6. 标签管理 API (`/api/v2/labels`)

| HTTP方法 | 端点                   | 权限规则                                                 | 说明         |
| -------- | ---------------------- | -------------------------------------------------------- | ------------ |
| GET      | `/labels`              | `label:read:own` ⚡ `label:read:all`                     | 列出标签     |
| GET      | `/labels/:label_id`    | `label:read:own` ⚡ `label:read:all` + 👤 所有者检查     | 获取标签详情 |
| POST     | `/labels`              | `label:create:own`                                       | 创建标签     |
| PATCH    | `/labels/:label_id`    | `label:update:own` ⚡ `label:update:all` + 👤 所有者检查 | 更新标签     |
| DELETE   | `/labels/:label_id`    | `label:delete:own` ⚡ `label:delete:all` + 👤 所有者检查 | 删除标签     |
| POST     | `/labels/batch-delete` | `label:delete:own` ⚡ `label:delete:all`                 | 批量删除标签 |

---

## 7. 用户管理 API (`/api/v2/users`)

| HTTP方法 | 端点                                                      | 权限规则                                                     | 说明                 |
| -------- | --------------------------------------------------------- | ------------------------------------------------------------ | -------------------- |
| GET      | `/users`                                                  | `user:read:all`                                              | 列出用户（仅管理员） |
| GET      | `/users/:user_id/detail`                                  | `user:read:all` OR 👤 自己                                   | 获取用户详情         |
| POST     | `/users`                                                  | `user:create:all`                                            | 创建用户（仅管理员） |
| PATCH    | `/users/:user_id`                                         | `user:update:all`                                            | 更新用户（仅管理员） |
| DELETE   | `/users/:user_id`                                         | `user:delete:all`                                            | 删除用户（仅管理员） |
| POST     | `/users/:user_id/roles/:role_id`                          | `user:assign:all` + `role:manage:all`                        | 分配全局角色         |
| DELETE   | `/users/:user_id/roles/:role_id`                          | `user:assign:all` + `role:manage:all`                        | 移除全局角色         |
| POST     | `/users/:user_id/projects/:project_id/roles/:role_id`     | `user:assign:all` OR `project:manage:own` + 项目管理员检查   | 分配项目角色         |
| DELETE   | `/users/:user_id/projects/:project_id`                    | `user:assign:all` OR `project:manage:own` + 项目管理员检查   | 移除项目成员         |
| POST     | `/users/:user_id/permissions/assign`                      | `user:assign:all` + `permission:manage:all`                  | 分配用户权限         |
| POST     | `/users/:user_id/permissions/remove`                      | `user:assign:all` + `permission:manage:all`                  | 移除用户权限         |
| POST     | `/users/:user_id/containers/:container_id/roles/:role_id` | `user:assign:all` OR `container:manage:own` + 容器管理员检查 | 分配容器角色         |
| DELETE   | `/users/:user_id/containers/:container_id`                | `user:assign:all` OR `container:manage:own` + 容器管理员检查 | 移除容器成员         |
| POST     | `/users/:user_id/datasets/:dataset_id/roles/:role_id`     | `user:assign:all` OR `dataset:manage:own` + 数据集管理员检查 | 分配数据集角色       |
| DELETE   | `/users/:user_id/datasets/:dataset_id`                    | `user:assign:all` OR `dataset:manage:own` + 数据集管理员检查 | 移除数据集成员       |

---

## 8. 角色管理 API (`/api/v2/roles`)

| HTTP方法 | 端点                                 | 权限规则                                    | 说明                     |
| -------- | ------------------------------------ | ------------------------------------------- | ------------------------ |
| GET      | `/roles`                             | `role:read:all`                             | 列出角色                 |
| GET      | `/roles/:role_id`                    | `role:read:all`                             | 获取角色详情             |
| GET      | `/roles/:role_id/users`              | `role:read:all`                             | 列出角色下的用户         |
| POST     | `/roles`                             | `role:create:all`                           | 创建角色（仅系统管理员） |
| PATCH    | `/roles/:role_id`                    | `role:update:all`                           | 更新角色（仅系统管理员） |
| DELETE   | `/roles/:role_id`                    | `role:delete:all`                           | 删除角色（仅系统管理员） |
| POST     | `/roles/:role_id/permissions/assign` | `role:grant:all` + `permission:manage:all`  | 分配权限给角色           |
| POST     | `/roles/:role_id/permissions/remove` | `role:revoke:all` + `permission:manage:all` | 移除角色权限             |

---

## 9. 权限管理 API (`/api/v2/permissions`)

| HTTP方法 | 端点                                | 权限规则                | 说明                     |
| -------- | ----------------------------------- | ----------------------- | ------------------------ |
| GET      | `/permissions`                      | `permission:read:all`   | 列出权限                 |
| GET      | `/permissions/:permission_id`       | `permission:read:all`   | 获取权限详情             |
| GET      | `/permissions/:permission_id/roles` | `permission:read:all`   | 列出权限关联的角色       |
| POST     | `/permissions`                      | `permission:create:all` | 创建权限（仅系统管理员） |
| PUT      | `/permissions/:permission_id`       | `permission:update:all` | 更新权限（仅系统管理员） |
| DELETE   | `/permissions/:permission_id`       | `permission:delete:all` | 删除权限（仅系统管理员） |

---

## 10. 资源管理 API (`/api/v2/resources`)

| HTTP方法 | 端点                                  | 权限规则            | 说明           |
| -------- | ------------------------------------- | ------------------- | -------------- |
| GET      | `/resources`                          | `resource:read:all` | 列出资源       |
| GET      | `/resources/:resource_id`             | `resource:read:all` | 获取资源详情   |
| GET      | `/resources/:resource_id/permissions` | `resource:read:all` | 列出资源的权限 |

---

## 11. 任务管理 API (`/api/v2/tasks`)

| HTTP方法 | 端点                  | 权限规则                                                                                               | 说明         |
| -------- | --------------------- | ------------------------------------------------------------------------------------------------------ | ------------ |
| GET      | `/tasks`              | `task:read:own` ⚡ `task:read:project` ⚡ `task:read:team` ⚡ `task:read:all`                          | 列出任务     |
| GET      | `/tasks/:task_id`     | `task:read:own` ⚡ `task:read:project` ⚡ `task:read:team` ⚡ `task:read:all` + 👤 所有者/项目成员检查 | 获取任务详情 |
| POST     | `/tasks/batch-delete` | `task:delete:own` ⚡ `task:delete:all` + 👤 所有者检查                                                 | 批量删除任务 |

---

## 12. 故障注入管理 API (`/api/v2/injections`)

| HTTP方法 | 端点                               | 权限规则                                                                                                                   | 说明                                 |
| -------- | ---------------------------------- | -------------------------------------------------------------------------------------------------------------------------- | ------------------------------------ |
| GET      | `/injections`                      | `injection:read:own` ⚡ `injection:read:project` ⚡ `injection:read:team` ⚡ `injection:read:all`                          | 列出注入                             |
| GET      | `/injections/:id`                  | `injection:read:own` ⚡ `injection:read:project` ⚡ `injection:read:team` ⚡ `injection:read:all` + 👤 所有者/项目成员检查 | 获取注入详情                         |
| GET      | `/injections/:id/download`         | `injection:read:own` ⚡ `injection:read:project` ⚡ `injection:read:team` ⚡ `injection:read:all` + 👤 所有者/项目成员检查 | 下载数据包                           |
| GET      | `/injections/metadata`             | `injection:read:own` ⚡ `injection:read:project` ⚡ `injection:read:team` ⚡ `injection:read:all`                          | 获取注入元数据                       |
| GET      | `/injections/analysis/no-issues`   | `injection:read:all`                                                                                                       | 列出无问题的注入                     |
| GET      | `/injections/analysis/with-issues` | `injection:read:all`                                                                                                       | 列出有问题的注入                     |
| POST     | `/injections/search`               | `injection:read:own` ⚡ `injection:read:project` ⚡ `injection:read:team` ⚡ `injection:read:all`                          | 高级搜索注入                         |
| POST     | `/injections/inject`               | `injection:execute:own` ⚡ `injection:execute:project` + 项目成员检查                                                      | 提交故障注入                         |
| POST     | `/injections/build`                | `injection:execute:own` ⚡ `injection:execute:project` + 项目成员检查                                                      | 提交数据包构建                       |
| POST     | `/injections/:id/clone`            | `injection:clone:own`                                                                                                      | 克隆注入（克隆后成为新注入的所有者） |
| PATCH    | `/injections/:id/labels`           | `injection:update:own` ⚡ `injection:update:project` + 👤 所有者/项目管理员检查                                            | 管理注入标签                         |
| PATCH    | `/injections/labels/batch`         | `injection:update:own` ⚡ `injection:update:project` ⚡ `injection:update:all`                                             | 批量管理注入标签                     |
| POST     | `/injections/batch-delete`         | `injection:delete:own` ⚡ `injection:delete:all` + 👤 所有者检查                                                           | 批量删除注入                         |

---

## 13. 执行结果管理 API (`/api/v2/executions`)

| HTTP方法 | 端点                                            | 权限规则                                                                                                                   | 说明               |
| -------- | ----------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------- | ------------------ |
| GET      | `/executions`                                   | `execution:read:own` ⚡ `execution:read:project` ⚡ `execution:read:team` ⚡ `execution:read:all`                          | 列出执行结果       |
| GET      | `/executions/:execution_id`                     | `execution:read:own` ⚡ `execution:read:project` ⚡ `execution:read:team` ⚡ `execution:read:all` + 👤 所有者/项目成员检查 | 获取执行详情       |
| GET      | `/executions/labels`                            | `execution:read:own` ⚡ `execution:read:project` ⚡ `execution:read:team` ⚡ `execution:read:all`                          | 列出可用的执行标签 |
| POST     | `/executions/execute`                           | `execution:execute:own` ⚡ `execution:execute:project` + 项目成员检查                                                      | 提交算法执行       |
| POST     | `/executions/:execution_id/detector_results`    | `execution:update:own` ⚡ `execution:update:project` + 👤 所有者检查                                                       | 上传检测器结果     |
| POST     | `/executions/:execution_id/granularity_results` | `execution:update:own` ⚡ `execution:update:project` + 👤 所有者检查                                                       | 上传粒度结果       |
| PATCH    | `/executions/:execution_id/labels`              | `execution:update:own` ⚡ `execution:update:project` + 👤 所有者检查                                                       | 管理执行标签       |
| POST     | `/executions/batch-delete`                      | `execution:delete:own` ⚡ `execution:delete:all` + 👤 所有者检查                                                           | 批量删除执行       |

---

## 14. 追踪管理 API (`/api/v2/traces`)

| HTTP方法 | 端点                       | 权限规则                                                                     | 说明              |
| -------- | -------------------------- | ---------------------------------------------------------------------------- | ----------------- |
| GET      | `/traces/group/stats`      | `trace:read:all`                                                             | 获取追踪组统计    |
| GET      | `/traces/:trace_id/stream` | `trace:read:own` ⚡ `trace:read:project` ⚡ `trace:read:all` + 👤 所有者检查 | 获取追踪流（SSE） |

---

## 15. 通知 API (`/api/v2/notifications`)

| HTTP方法 | 端点                    | 权限规则 | 说明              |
| -------- | ----------------------- | -------- | ----------------- |
| GET      | `/notifications/stream` | JWT 认证 | 全局通知流（SSE） |

---

## 16. 评估 API (`/api/v2/evaluations`)

| HTTP方法 | 端点                     | 权限规则                                                        | 说明               |
| -------- | ------------------------ | --------------------------------------------------------------- | ------------------ |
| POST     | `/evaluations/datasets`  | `dataset:read:own` ⚡ `dataset:read:team` ⚡ `dataset:read:all` | 获取数据集评估结果 |
| POST     | `/evaluations/datapacks` | `dataset:read:own` ⚡ `dataset:read:team` ⚡ `dataset:read:all` | 获取数据包评估结果 |

---

## 17. 指标 API (`/api/v2/metrics`)

| HTTP方法 | 端点                  | 权限规则                                                                                          | 说明             |
| -------- | --------------------- | ------------------------------------------------------------------------------------------------- | ---------------- |
| GET      | `/metrics/injections` | `injection:read:own` ⚡ `injection:read:project` ⚡ `injection:read:team` ⚡ `injection:read:all` | 获取注入指标     |
| GET      | `/metrics/executions` | `execution:read:own` ⚡ `execution:read:project` ⚡ `execution:read:team` ⚡ `execution:read:all` | 获取执行指标     |
| GET      | `/metrics/algorithms` | `execution:read:own` ⚡ `execution:read:project` ⚡ `execution:read:team` ⚡ `execution:read:all` | 获取算法对比指标 |

---

## 18. 系统指标 API (`/api/v2/system`)

| HTTP方法 | 端点                      | 权限规则          | 说明                         |
| -------- | ------------------------- | ----------------- | ---------------------------- |
| GET      | `/system/metrics`         | `system:read:all` | 获取当前系统指标（仅管理员） |
| GET      | `/system/metrics/history` | `system:read:all` | 获取历史系统指标（仅管理员） |

---

## 权限检查逻辑说明

### 1. 作用域优先级

权限检查按以下优先级进行：

```
own (自己的资源) < project (项目资源) < team (团队资源) < all (所有资源)
```

- 拥有 `all` 作用域的权限可以访问所有资源
- 拥有 `team` 作用域的权限可以访问团队资源、项目资源和自己的资源
- 拥有 `project` 作用域的权限可以访问项目内的资源和自己的资源
- 拥有 `own` 作用域的权限只能访问自己创建的资源

### 权限继承规则

权限系统支持三种继承机制：**项目级继承**、**团队级继承**和**全局作用域**。

#### 1. 项目级权限继承 (Project-Level Inheritance)

对于项目内的资源（**injection、execution、task、trace**），会自动继承项目的权限：

- 如果用户拥有 `project:read:project` 权限，则自动拥有该项目内所有 injection、execution、task、trace 的读取权限
- 如果用户拥有 `project:execute:project` 权限，则自动拥有该项目内 injection、execution 的执行权限
- 如果用户拥有 `project:manage:project` 权限，则自动拥有该项目内所有资源的完全管理权限

**项目级继承示例**：

```
用户 A 是项目 P1 的成员，拥有 project:read:project 权限
  ↓ 自动继承
- injection:read:project (可以读取 P1 内的所有故障注入)
- execution:read:project (可以读取 P1 内的所有执行结果)
- task:read:project (可以读取 P1 内的所有任务)
- trace:read:project (可以读取 P1 内的所有追踪)
```

#### 2. 团队级权限继承 (Team-Level Inheritance)

对于团队内的资源（**container、dataset、project、label** 及其子资源），会自动继承团队的权限：

- 如果用户拥有 `team:read:team` 权限，则自动拥有该团队内所有 container、dataset、project、label 的读取权限
- 如果用户拥有 `team:create:team` 权限，则可以在该团队内创建新资源
- 如果用户拥有 `team:manage:team` 权限，则自动拥有该团队内所有资源的完全管理权限

**团队级继承示例**：

```
用户 B 是团队 T1 的成员，拥有 team:read:team 权限
  ↓ 自动继承
- container:read:team (可以读取 T1 内的所有容器)
- container_version:read:team (可以读取 T1 容器的所有版本)
- dataset:read:team (可以读取 T1 内的所有数据集)
- dataset_version:read:team (可以读取 T1 数据集的所有版本)
- project:read:team (可以读取 T1 内的所有项目)
- label:read:team (可以读取 T1 内的所有标签)
```

#### 3. 全局作用域 (All Scope)

**`all` 作用域不需要继承机制**，因为它本身就是顶级权限：

- 拥有 `:all` 作用域的权限直接绕过所有权限检查和继承逻辑
- 这是系统管理员级别的权限，可以访问系统中的所有资源
- 不受项目成员、团队成员、资源所有者等限制

**全局作用域示例**：

```
用户 C 拥有 container:read:all 权限
  ↓ 直接访问（无需继承）
- 可以读取系统中的所有容器，无论是谁创建的、属于哪个团队或项目
- 不需要检查团队成员关系或项目成员关系
- 权限检查直接通过
```

#### 继承优先级

当用户同时拥有多个作用域的权限时，按以下优先级处理：

```
all (全局) > team (团队) > project (项目) > own (自己)
```

- 如果有 `all` 权限，直接通过，不检查其他
- 如果有 `team` 权限且是团队成员，允许访问团队资源（包括团队内的项目及其资源）
- 如果有 `project` 权限且是项目成员，允许访问项目资源
- 如果有 `own` 权限，只能访问自己创建的资源

### 2. 所有者检查

对于需要 👤 **所有者检查** 的端点：

1. 如果用户是资源的所有者（`creator_id == user_id`），允许访问
2. 如果用户拥有 `{resource}:{action}:all` 权限，允许访问
3. 如果资源属于某个团队，且用户拥有 `{resource}:{action}:team` 权限并是团队成员，允许访问
4. 否则拒绝访问

### 3. 团队成员检查

对于需要 👥 **团队成员检查** 的端点：

1. 检查用户是否为团队成员（通过 `team_members` 表）
2. 如果是公开团队（`is_public = true`），允许访问
3. 如果用户拥有 `team:manage:all` 权限（系统管理员），允许访问
4. 否则拒绝访问

### 4. 项目成员检查

对于项目相关的端点：

1. 检查用户是否为项目成员（通过 `project_users` 表）
2. 检查用户在项目中的角色权限
3. 如果用户拥有 `project:manage:all` 权限（系统管理员），允许访问
4. 否则拒绝访问

### 5. 子资源权限继承

子资源（如容器版本、数据集版本）的权限检查需要：

1. 首先检查父资源的访问权限
2. 然后检查子资源的操作权限
3. 两者都满足才允许访问

例如：要访问 `/containers/:container_id/versions/:version_id`，需要：

- 拥有容器的读权限：`container:read:own/team/all`
- 拥有版本的读权限：`container_version:read:own/team/all`

---

## 默认角色权限映射

根据 `system.go` 中的定义，各角色的默认权限：

### 系统角色

- **super_admin**: 拥有所有权限（不受限制）
- **admin**: 拥有所有资源的 `read/write/delete/execute/manage` 权限（`all` 作用域）
- **user**: 基础角色，无默认权限，需要额外分配

### 容器角色

- **container_admin**: `container:read/create/update/delete/manage:all` + `container_version:read/create/update/delete/manage:all`
- **container_developer**: `container:read/create/update:team` + `container_version:read/create/update:team`
- **container_viewer**: `container:read:all` + `container_version:read:all`

### 数据集角色

- **dataset_admin**: `dataset:read/create/update/delete/manage:all` + `dataset_version:read/create/update/delete/manage:all`
- **dataset_developer**: `dataset:read/create/update:team` + `dataset_version:read/create/update:team`
- **dataset_viewer**: `dataset:read:all` + `dataset_version:all`

### 项目角色

- **project_admin**: 项目的所有操作权限 + 项目内资源的管理权限
- **project_developer**: 项目的读权限 + 项目内资源的创建/更新/执行权限
- **project_viewer**: 项目的读权限 + 项目内资源的读权限

### 团队角色

- **team_admin**: `team:read/create/update/delete/manage:own`（仅限所在团队）
- **team_member**: `team:read/update:own`（仅限所在团队）
- **team_viewer**: `team:read:own`（仅限所在团队）

---

## 实现建议

### 1. 中间件层面

建议实现以下中间件函数：

```go
// 基础权限检查中间件
func RequirePermission(action ActionName, resource ResourceName, scopes ...ResourceScope) gin.HandlerFunc

// 资源所有者检查中间件
func RequireResourceOwnership(resourceType string, idParam string) gin.HandlerFunc

// 团队成员检查中间件
func RequireTeamMembership(teamIDParam string) gin.HandlerFunc

// 项目成员检查中间件
func RequireProjectMembership(projectIDParam string) gin.HandlerFunc
```

### 2. 数据库层面

建议在数据库中添加以下索引以优化权限查询：

```sql
-- 用户-角色关系
CREATE INDEX idx_user_roles_user_id ON user_roles(user_id);

-- 角色-权限关系
CREATE INDEX idx_role_permissions_role_id ON role_permissions(role_id);

-- 资源所有权
CREATE INDEX idx_resources_creator_id ON {resource_table}(creator_id);

-- 团队成员关系
CREATE INDEX idx_team_members_user_id ON team_members(user_id);
CREATE INDEX idx_team_members_team_id ON team_members(team_id);

-- 项目成员关系
CREATE INDEX idx_project_users_user_id ON project_users(user_id);
CREATE INDEX idx_project_users_project_id ON project_users(project_id);
```

### 3. 权限缓存

建议使用 Redis 缓存用户权限信息：

```
Key: "user:permissions:{user_id}"
Value: JSON array of PermissionRule objects
TTL: 15 minutes
```

在用户角色或权限变更时，清除对应缓存。

---

## 审计日志

建议对以下操作进行审计日志记录：

1. **所有写操作**（POST, PATCH, PUT, DELETE）
2. **权限变更**（角色分配、权限分配）
3. **敏感资源访问**（数据集下载、容器构建）
4. **故障注入执行**
5. **算法执行提交**

审计日志应包含：

- 用户ID
- 操作时间
- 操作类型（HTTP方法 + 路径）
- 资源类型和ID
- 操作结果（成功/失败）
- 失败原因（如权限不足）

---

## 总结

本权限设计遵循以下原则：

1. **最小权限原则**: 用户默认只能访问自己的资源
2. **层级权限**: 通过 `own -> team -> all` 的作用域实现权限层级
3. **继承性**: 子资源权限继承父资源权限
4. **灵活性**: 支持角色权限和直接权限分配
5. **可审计性**: 所有权限检查都有明确的日志记录

请审查以上设计，如有问题请提出修改意见。确认无误后，我将开始实现相关的中间件和权限检查逻辑。
