package router

import (
	"github.com/gin-gonic/gin"
)

/*
===================================================================================
API v2 设计规范 - RESTful API 标准
===================================================================================

v2 API 采用严格的 RESTful 设计规范，与 v1 的杂乱设计形成对比。
v1 API 设计较为随意，方法和路径不规范，v2 将统一按照以下标准执行。

📋 HTTP 方法使用规范：
- GET    : 查询资源（幂等，可缓存）
- POST   : 创建资源 / 复杂查询（非幂等）
- PUT    : 完整更新资源（幂等）
- PATCH  : 部分更新资源（幂等）
- DELETE : 删除资源（幂等）

🎯 URL 设计规范：
1. 资源名称使用复数形式
   ✅ GET /api/v2/users          ❌ GET /api/v2/user
   ✅ GET /api/v2/projects       ❌ GET /api/v2/project

2. 层级关系明确
   ✅ GET /api/v2/users/{id}/projects
   ✅ GET /api/v2/projects/{id}/members

3. 查询参数规范
   ✅ GET /api/v2/users?page=1&size=10&status=active
   ✅ GET /api/v2/tasks?project_id=123&type=injection

📊 标准 CRUD 操作模式：
- GET    /api/v2/{resource}           # 列表查询（支持分页、过滤、排序）
- POST   /api/v2/{resource}           # 创建资源
- GET    /api/v2/{resource}/{id}      # 获取单个资源详情
- PUT    /api/v2/{resource}/{id}      # 完整更新资源
- PATCH  /api/v2/{resource}/{id}      # 部分更新资源
- DELETE /api/v2/{resource}/{id}      # 删除资源

🔍 复杂查询处理：
对于复杂搜索条件，使用专门的搜索端点：
- POST /api/v2/{resource}/search      # 复杂条件搜索
- POST /api/v2/{resource}/query       # 高级查询
- POST /api/v2/{resource}/batch       # 批量操作

🎨 业务操作端点：
语义化的业务操作使用动词形式：
- POST /api/v2/users/{id}/activate    # 激活用户
- POST /api/v2/tasks/{id}/cancel      # 取消任务
- POST /api/v2/injections/{id}/start  # 开始故障注入
- POST /api/v2/containers/{id}/build  # 构建容器

📨 响应格式规范：
1. 成功响应：
   {
     "code": 200,
     "message": "success",
     "data": {...},
     "timestamp": "2024-01-01T12:00:00Z"
   }

2. 列表响应：
   {
     "code": 200,
     "message": "success",
     "data": {
       "items": [...],
       "pagination": {
         "page": 1,
         "size": 10,
         "total": 100,
         "pages": 10
       }
     }
   }

3. 错误响应：
   {
     "code": 400,
     "message": "validation failed",
     "errors": ["field xxx is required"],
     "timestamp": "2024-01-01T12:00:00Z"
   }

🔐 认证授权规范：
- 使用 JWT Bearer Token 认证
- 权限检查基于 RBAC 模型
- 敏感操作需要二次确认

⚡ 性能优化：
- GET 请求支持 ETag 缓存
- 列表查询默认分页（page=1, size=20）
- 支持字段选择 ?fields=id,name,status
- 支持关联查询 ?include=project,labels

注意：v1 API 设计较为混乱，不遵循统一标准，后续逐步迁移到 v2 规范。
===================================================================================
*/

// SetupV2Routes 设置 API v2 路由 - 稳定版本的 API
func SetupV2Routes(router *gin.Engine) {

	v2 := router.Group("/api/v2")

	// 核心业务实体 API 组
	containers := v2.Group("/containers") // 容器管理 - Container 实体
	injections := v2.Group("/injections") // 故障注入管理 - FaultInjectionSchedule 实体
	datasets := v2.Group("/datasets")     // 数据集管理 - Dataset 实体
	executions := v2.Group("/executions") // 执行结果管理 - ExecutionResult 实体
	labels := v2.Group("/labels")         // 标签管理 - Label 实体
	projects := v2.Group("/projects")     // 项目管理 - Project 实体
	tasks := v2.Group("/tasks")           // 任务管理 - Task 实体

	// 权限认证相关 API 组
	roles := v2.Group("/roles")             // 角色管理 - Role 实体
	users := v2.Group("/users")             // 用户管理 - User 实体
	resources := v2.Group("/resources")     // 资源管理 - Resource 实体
	permissions := v2.Group("/permissions") // 权限管理 - Permission 实体
	auth := v2.Group("/auth")               // 认证相关 - 登录/登出/token等

	// 分析检测相关 API 组
	detectors := v2.Group("/detectors")     // 检测器管理 - Detector 实体
	granularity := v2.Group("/granularity") // 粒度结果管理 - GranularityResult 实体
	traces := v2.Group("/traces")           // 追踪管理 - 与 TraceID 相关
	analyzer := v2.Group("/analyzer")       // 分析器相关

	// 系统管理相关 API 组
	monitor := v2.Group("/monitor")       // 监控相关
	health := v2.Group("/health")         // 健康检查
	statistics := v2.Group("/statistics") // 统计信息
	audit := v2.Group("/audit")           // 审计日志

	// 关系管理相关 API 组
	relations := v2.Group("/relations") // 多对多关系管理 (DatasetLabel, UserRole等)

	// 暂时使用空赋值避免编译错误，后续逐步实现具体路由
	_ = containers
	_ = injections
	_ = datasets
	_ = executions
	_ = labels
	_ = projects
	_ = tasks
	_ = roles
	_ = users
	_ = resources
	_ = permissions
	_ = auth
	_ = detectors
	_ = granularity
	_ = traces
	_ = analyzer
	_ = monitor
	_ = health
	_ = statistics
	_ = audit
	_ = relations
}
