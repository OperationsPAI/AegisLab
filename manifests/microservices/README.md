# Microservice Kubernetes Skeleton

这目录承接当前六服务拆分后的第一版 Kubernetes skeleton。

当前文件：

- `aegislab-microservices.yaml`

用途：

- 给 `api-gateway / iam-service / resource-service / orchestrator-service / runtime-worker-service / system-service` 提供第一版 Deployment/Service 骨架
- 明确端口、启动命令、probe 约定、配置挂载方式

注意：

- 这是一份 skeleton，不是最终生产部署方案
- 当前仍假设：
  - MySQL / Redis / Etcd / Jaeger / BuildKit 等基础依赖已由其他清单或平台层提供
  - `ConfigMap/aegislab-config` 已准备好并包含 `config.toml`
- 这份清单优先编码“边界和启动方式”，而不是覆盖全部生产级资源策略
