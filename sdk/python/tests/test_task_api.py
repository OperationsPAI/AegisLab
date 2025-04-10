# Run this file:
# uv run pytest -s tests/test_task_api.py
from typing import Any, Dict, List
from conftest import BASE_URL
from pprint import pprint
from rcabench.model.common import SubmitResult
from rcabench.rcabench import RCABenchSDK
from uuid import UUID
import pytest


@pytest.mark.asyncio
@pytest.mark.parametrize(
    "benchmark, interval, pre_duration, specs",
    [
        (
            "clickhouse",
            3,
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
                },
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
                },
            ],
        ),
    ],
)
async def test_injection_and_building_dataset(benchmark, interval, pre_duration, specs):
    sdk = RCABenchSDK(BASE_URL)

    resp = sdk.injection.submit(benchmark, interval, pre_duration, specs)
    pprint(resp)

    if not isinstance(resp, SubmitResult):
        pytest.fail(resp.model_dump_json())

    traces = resp.traces
    if not traces:
        pytest.fail("No traces returned from execution")

    task_ids = [trace.head_task_id for trace in traces]
    await sdk.task.get_stream(task_ids)

    count = 0
    flag = True
    while True:
        extra_time = 0 if flag else interval
        result = await sdk.task.stream.client_manager.get_next_result(
            (1.5 + extra_time) * 60
        )
        print(f"Received result: {result}")

        flag = result.get("type") == "error"
        count += 1
        if count >= 2 * len(traces):
            break


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
        ),
        (
            "clickhouse",
            2,
            0,
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
        ),
        (
            "clickhouse",
            2,
            1,
            [
                {
                    "children": {
                        "1": {
                            "children": {
                                "0": {"value": -1},
                                "1": {"value": 0},
                                "2": {"value": 42},
                            }
                        },
                    },
                    "value": 1,
                }
            ],
        ),
    ],
)
async def test_injection_and_building_dataset_batch(
    benchmark, interval, pre_duration, specs
):
    sdk = RCABenchSDK(BASE_URL)

    resp = sdk.injection.submit(benchmark, interval, pre_duration, specs)
    pprint(resp)

    if not isinstance(resp, SubmitResult):
        pytest.fail(resp.model_dump_json())

    traces = resp.traces
    if not traces:
        pytest.fail("No traces returned from execution")

    task_ids = [trace.head_task_id for trace in traces]
    report = await sdk.task.get_stream_batch(task_ids, timeout=None)
    report = report.model_dump(exclude_unset=True)
    pprint(report)

    return report


@pytest.mark.asyncio
@pytest.mark.parametrize(
    "payload",
    [
        (
            [
                {
                    "algorithm": "e-diagnose",
                    "dataset": "ts-ts-travel2-service-pod-failure-rkxslq",
                }
            ]
        )
    ],
)
async def test_execute_algorithm_and_collection(payload: List[Dict[str, str]]):
    """测试执行多个算法并验证结果流收集功能

    验证步骤：
    1. 初始化 SDK 连接
    2. 获取可用算法列表
    3. 为每个算法生成执行参数
    4. 提交批量执行请求
    5. 启动流式结果收集
    6. 验证关键结果字段
    """
    sdk = RCABenchSDK(BASE_URL)

    resp = sdk.algorithm.submit(payload)
    pprint(resp)

    if not isinstance(resp, SubmitResult):
        pytest.fail(resp.model_dump_json())

    traces = resp.traces
    if not traces:
        pytest.fail("No traces returned from execution")

    task_ids = [trace.head_task_id for trace in traces]
    report = await sdk.task.get_stream_batch(task_ids, timeout=30)
    report = report.model_dump(exclude_unset=True)
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

    payload = []
    algorithms = ["e-diagnose"]
    for algorithm in algorithms:
        for dataset in datasets:
            payload.append({"algorithm": algorithm, "dataset": dataset})

    execution_report = await test_execute_algorithm_and_collection(payload)
    execution_ids = extract_values(execution_report, "execution_id")
    pprint(execution_ids)


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
