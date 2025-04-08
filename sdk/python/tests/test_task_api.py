# Run this file:
# uv run pytest -s tests/test_task_api.py
from typing import Any, Dict, List
from pprint import pprint
from rcabench import rcabench
from uuid import UUID
import pytest


BASE_URL = "http://localhost:8082"


@pytest.mark.asyncio
@pytest.mark.parametrize(
    "benchmark, interval, pre_duration, specs",
    [
        (
            "clickhouse",
            2,
            1,
            [
                {
                    "children": {
                        "1": {
                            "children": {
                                "0": {"value": 1},
                                "1": {"value": 0},
                                "2": {"value": 42},
                            }
                        },
                    },
                    "value": 1,
                }
            ],
        )
    ],
)
async def test_injection_and_building_dataset(benchmark, interval, pre_duration, specs):
    sdk = rcabench.RCABenchSDK(BASE_URL)

    data = sdk.injection.submit(benchmark, interval, pre_duration, specs)
    pprint(data)

    traces = data.traces
    if not traces:
        pytest.fail("No traces returned from execution")

    task_ids = [trace.head_task_id for trace in traces]
    report = await sdk.task.get_stream(task_ids, timeout=1.5 * 60)
    pprint(report)

    return report


def extract_values(data: Dict[UUID, Any], key: str) -> List[str]:
    """递归提取嵌套结构中的所有value值

    Args:
        data: 输入的嵌套字典结构，键可能为UUID

    Returns:
        所有找到的value值列表
    """
    values = []

    def _recursive_search(node):
        if isinstance(node, dict):
            # 检查当前层级是否有dataset字段
            if key in node:
                values.append(node[key])
            # 递归处理所有子节点
            for value in node.values():
                _recursive_search(value)
        elif isinstance(node, (list, tuple)):
            # 处理可迭代对象
            for item in node:
                _recursive_search(item)

    _recursive_search(data)
    return values


@pytest.mark.asyncio
@pytest.mark.parametrize(
    "algorithms, datasets",
    [([["e-diagnose"]], ["ts-ts-ui-dashboard-pod-failure-ncbfpc"])],
)
async def test_execute_algorithm_and_collection(
    algorithms: List[List[str]], datasets: List[str]
):
    """测试执行多个算法并验证结果流收集功能

    验证步骤：
    1. 初始化 SDK 连接
    2. 获取可用算法列表
    3. 为每个算法生成执行参数
    4. 提交批量执行请求
    5. 启动流式结果收集
    6. 验证关键结果字段
    """
    sdk = rcabench.RCABenchSDK(BASE_URL)

    data = sdk.algorithm.submit(algorithms, datasets)
    pprint(data)

    traces = data.traces
    if not traces:
        pytest.fail("No traces returned from execution")

    task_ids = [trace.head_task_id for trace in traces]
    report = await sdk.task.get_stream(task_ids, timeout=30)
    pprint(report)

    return report


@pytest.mark.asyncio
async def test_workflow():
    injection_payload = {
        "benchmark": "clickhouse",
        "interval": 2,
        "pre_duration": 1,
        "specs": [
            {
                "children": {
                    "1": {
                        "children": {
                            "0": {"value": 1},
                            "1": {"value": 0},
                            "2": {"value": 42},
                        }
                    },
                },
                "value": 1,
            }
        ],
    }

    injection_report = await test_injection_and_building_dataset(**injection_payload)
    datasets = extract_values(injection_report, "dataset")
    pprint(datasets)

    algorithms = [["e-diagnose"]]
    execution_report = await test_execute_algorithm_and_collection(algorithms, datasets)
    execution_ids = extract_values(execution_report, "execution_id")
    pprint(execution_ids)
