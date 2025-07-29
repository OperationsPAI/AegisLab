#!/usr/bin/env python3
"""
RCABench 客户端使用示例
"""

import sys
import os

sys.path.insert(0, os.path.join(os.path.dirname(__file__), "src"))

from rcabench_client import RCABenchClient
import rcabench.openapi


def example_with_statement():
    """使用 with 语句的示例"""
    print("=== 使用 with 语句 ===")

    with RCABenchClient() as api_client:
        # 获取任务列表
        task_api = rcabench.openapi.TaskApi(api_client)
        tasks = task_api.api_v1_tasks_get()
        print(f"任务数量: {len(tasks.data)}")

        # 获取容器列表
        container_api = rcabench.openapi.ContainersApi(api_client)
        containers = container_api.api_v2_containers_get()
        print(f"容器数量: {len(containers.data.items)}")


def example_direct_usage():
    """直接使用的示例"""
    print("\n=== 直接使用 ===")

    client = RCABenchClient()
    api_client = client.get_client()

    # 使用API客户端
    task_api = rcabench.openapi.TaskApi(api_client)
    tasks = task_api.api_v1_tasks_get()
    print(f"任务数量: {len(tasks.data)}")


def example_custom_config():
    """自定义配置的示例"""
    print("\n=== 自定义配置 ===")

    # 自定义服务器地址和凭据
    with RCABenchClient(
        base_url="http://10.10.10.46:8082", username="admin", password="admin123"
    ) as api_client:
        task_api = rcabench.openapi.TaskApi(api_client)
        tasks = task_api.api_v1_tasks_get()
        print(f"任务数量: {len(tasks.data)}")


if __name__ == "__main__":
    example_with_statement()
    example_direct_usage()
    example_custom_config()

    print("\n✅ 所有示例运行成功！")
