#!/usr/bin/env python3
"""
RCABench 客户端
支持 with 语句的面向对象客户端
"""

import sys
import os

# 添加src目录到Python路径
sys.path.insert(0, os.path.join(os.path.dirname(__file__), "src"))

import rcabench.openapi
from rcabench.openapi.models.dto_login_request import DtoLoginRequest


class RCABenchClient:
    """RCABench API客户端，支持自动认证"""

    def __init__(
        self, base_url="http://10.10.10.46:8082", username="admin", password="admin123"
    ):
        self.base_url = base_url
        self.username = username
        self.password = password
        self.access_token = None
        self._api_client = None

    def __enter__(self):
        """进入上下文时自动登录并获取认证客户端"""
        self._login()
        return self._get_authenticated_client()

    def __exit__(self, exc_type, exc_val, exc_tb):
        """退出上下文时清理资源"""
        if self._api_client:
            # 如果需要的话，可以在这里添加清理逻辑
            pass

    def _login(self):
        """登录获取访问令牌"""
        if self.access_token:
            return

        config = rcabench.openapi.Configuration(host=self.base_url)
        with rcabench.openapi.ApiClient(config) as api_client:
            auth_api = rcabench.openapi.AuthenticationApi(api_client)
            login_request = DtoLoginRequest(
                username=self.username, password=self.password
            )
            response = auth_api.api_v2_auth_login_post(login_request)
            self.access_token = response.data.token

    def _get_authenticated_client(self):
        """获取已认证的API客户端"""
        if not self.access_token:
            self._login()

        # 创建带认证的配置
        auth_config = rcabench.openapi.Configuration(
            host=self.base_url,
            api_key={"BearerAuth": self.access_token},
            api_key_prefix={"BearerAuth": "Bearer"},
        )

        self._api_client = rcabench.openapi.ApiClient(auth_config)
        return self._api_client

    def get_client(self):
        """获取认证客户端（不自动登录）"""
        if not self._api_client:
            self._api_client = self._get_authenticated_client()
        return self._api_client


# 使用示例
if __name__ == "__main__":
    # 方式1: 使用 with 语句（推荐）
    with RCABenchClient() as api_client:
        print("✅ 获取到认证客户端")
        print(f"客户端类型: {type(api_client)}")

        # 使用客户端调用API
        task_api = rcabench.openapi.TaskApi(api_client)
        tasks = task_api.api_v1_tasks_get()
        print(f"任务数量: {len(tasks.data)}")

    # 方式2: 直接获取客户端
    client = RCABenchClient()
    api_client = client.get_client()
    print("✅ 直接获取认证客户端成功")
