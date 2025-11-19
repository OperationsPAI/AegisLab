# 测试环境设置说明

本目录包含使用 `python-on-whales` 自动管理 Docker Compose 测试环境的配置。

## 功能特性

- ✅ 自动启动和停止 Docker Compose 服务
- ✅ 等待服务健康检查完成
- ✅ 自动清理测试环境
- ✅ 支持会话级别的 fixture，避免重复启动服务

## 依赖项

测试依赖已添加到 `pyproject.toml` 的 `dev` 依赖组：

```toml
[dependency-groups]
dev = [
    "pytest>=8.3.5",
    "python-dotenv>=1.1.1",
    "python-on-whales>=0.79.0",
]
```

## Docker Compose 配置

测试使用 `/home/nn/workspace/AegisLab/docker-compose.test.yaml` 文件，该文件定义了以下服务：

- **redis** (端口 10000): Redis 缓存服务
- **mysql** (端口 10001): MySQL 数据库
- **jaeger** (端口 10002): 分布式追踪服务
- **buildkitd** (端口 10003): BuildKit 构建服务
- **exp** (端口 10004): RCABench 主应用服务

## 使用方法

### 安装依赖

```bash
cd /home/nn/workspace/AegisLab/sdk/python
uv sync --dev
```

### 运行测试

运行所有测试（会自动启动和停止 Docker Compose）：

```bash
uv run pytest
```

运行特定测试文件：

```bash
uv run pytest tests/test_docker_setup.py -v
```

查看详细输出：

```bash
uv run pytest -v -s
```

## Fixtures 说明

### `docker_compose` (session 级别)

自动管理 Docker Compose 生命周期的 fixture。

**功能：**

- 启动所有 Docker Compose 服务
- 等待服务健康检查通过（最多 60 秒）
- 测试完成后自动停止并清理服务
- 删除卷以确保干净的环境

**用法：**

```python
def test_something(docker_compose):
    # docker_compose 是 DockerClient 实例
    services = docker_compose.compose.ps()
    assert len(services) >= 4
```

### `rcabench_client` (session 级别)

提供配置好的 RCABench 客户端，连接到测试环境。

**功能：**

- 依赖 `docker_compose` fixture
- 自动设置 `RCABENCH_BASE_URL` 为测试服务地址
- 使用上下文管理器确保资源正确清理

**用法：**

```python
def test_api_call(rcabench_client):
    # 使用客户端进行 API 调用
    result = rcabench_client.some_api_method()
    assert result is not None
```

## 测试流程

1. **启动阶段** (首次测试开始前)

   - 读取 `docker-compose.test.yaml`
   - 启动所有服务
   - 等待服务健康检查通过
   - 额外等待 5 秒确保服务完全初始化

2. **测试执行**

   - 所有测试共享同一个 Docker Compose 环境
   - RCABench 客户端连接到 `http://localhost:10004`

3. **清理阶段** (所有测试完成后)
   - 停止所有容器
   - 删除容器和卷
   - 清理网络

## 故障排除

### 服务启动失败

如果服务在 60 秒内未能启动，pytest 会输出：

```
❌ Services failed to start in time
```

**解决方法：**

1. 检查 Docker 是否正在运行
2. 查看 Docker Compose 日志
3. 确保端口 10000-10004 未被占用
4. 检查 `docker-compose.test.yaml` 配置

### 端口冲突

如果端口已被占用，编辑 `docker-compose.test.yaml` 修改端口映射。

### 权限问题

确保当前用户有权限访问 Docker：

```bash
docker ps
```

如果失败，可能需要将用户添加到 `docker` 组：

```bash
sudo usermod -aG docker $USER
newgrp docker
```

## 环境变量

测试会自动设置以下环境变量：

- `RCABENCH_BASE_URL=http://localhost:10004`

可以在 `.env` 文件中配置其他环境变量。

## 最佳实践

1. **会话级别 fixtures**: 使用 `scope="session"` 避免重复启动服务
2. **超时设置**: 给服务足够的启动时间（当前设置为 60 秒）
3. **清理**: 始终使用 `volumes=True` 确保测试环境干净
4. **健康检查**: 依赖 Docker Compose 的健康检查机制
5. **日志**: 使用 `docker_compose.compose.logs()` 调试问题

## 示例测试

```python
def test_service_health(docker_compose):
    """验证所有服务都在运行"""
    services = docker_compose.compose.ps()
    for service in services:
        assert service.state.status == "running"

def test_api_endpoint(rcabench_client):
    """测试 API 端点"""
    # 执行 API 调用
    response = rcabench_client.some_method()
    assert response.status_code == 200
```

## 相关文档

- [python-on-whales 文档](https://gabrieldemarmiesse.github.io/python-on-whales/)
- [pytest fixtures 文档](https://docs.pytest.org/en/stable/fixture.html)
- [Docker Compose 文档](https://docs.docker.com/compose/)
