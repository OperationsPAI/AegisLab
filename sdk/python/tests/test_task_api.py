# Run this file:
# uv run pytest -s tests/test_task_api.py
from typing import Any, Dict, List, Optional
from conftest import BASE_URL
from pprint import pprint
from rcabench.logger import logger
from rcabench.model.common import SubmitResult
from rcabench.rcabench import RCABenchSDK
from uuid import UUID
import asyncio
import pytest


@pytest.mark.asyncio
@pytest.mark.parametrize(
    "benchmark, interval, pre_duration, specs, num_consumers, max_items_per_consumer, per_consumer_timeout",
    [
        # PodKill
        (
            "clickhouse",
            3,
            1,
            [
                {
                    "children": {
                        "0": {
                            "children": {
                                "0": {"value": 1},
                                "1": {"value": 0},
                                "2": {"value": 42},
                            }
                        },
                    },
                    "value": 0,
                },
            ],
            1,
            1,
            2 * 60,
        ),
        # PodFailure
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
            ],
            1,
            1,
            3 * 60,
        ),
        # HttpRequestAbort
        (
            "clickhouse",
            3,
            1,
            [
                {
                    "children": {
                        "5": {
                            "children": {
                                "0": {"value": 1},
                                "1": {"value": 0},
                                "2": {"value": 42},
                            }
                        },
                    },
                    "value": 5,
                },
            ],
            1,
            1,
            2 * 60,
        ),
    ],
)
async def test_injection_and_building_dataset_batch(
    benchmark: str,
    interval: int,
    pre_duration: int,
    specs: List[Dict[str, Any]],
    num_consumers: int,
    max_items_per_consumer: int,
    per_consumer_timeout: float,
):
    assert num_consumers <= len(specs)
    sdk = RCABenchSDK(BASE_URL)

    resp = sdk.injection.submit(benchmark, interval, pre_duration, specs)
    pprint(resp)

    if not isinstance(resp, SubmitResult):
        pytest.fail(resp.model_dump_json())

    traces = resp.traces
    if not traces:
        pytest.fail("No traces returned from execution")

    task_ids = [trace.head_task_id for trace in traces]
    trace_ids = [trace.trace_id for trace in traces]
    queue = await sdk.task.get_stream_batch(task_ids, trace_ids)

    try:
        results = await run_consumers(
            queue,
            num_consumers,
            max_items_per_consumer,
            per_consumer_timeout,
            per_consumer_timeout * (num_consumers + 1),
        )
        logger.info(f"Final results: {results}")
    finally:
        await sdk.task.stream.cleanup()


async def run_consumers(
    queue: asyncio.Queue,
    num_consumers: int,
    max_items_per_consumer: int,
    per_consumer_timeout: Optional[float] = None,
    total_timeout: Optional[float] = None,
):
    all_results = {}

    start_time = asyncio.get_event_loop().time()
    for i in range(num_consumers):
        if total_timeout:
            elapsed = asyncio.get_event_loop().time() - start_time
            if elapsed >= total_timeout:
                logger.warning(
                    f"Total execution time exceeded {total_timeout} seconds, stopping subsequent tasks"
                )
                break

            # Calculate remaining timeout for current task
            remaining = total_timeout - elapsed
            current_timeout = min(per_consumer_timeout, remaining)
        else:
            current_timeout = per_consumer_timeout

        logger.info(f"Starting consumer {i}, timeout: {current_timeout} seconds")

        try:
            results = await consumer(queue, max_items_per_consumer, current_timeout)
            all_results[f"Batch-{i}"] = results
            logger.info(f"Consumer-{i} completed")
        except asyncio.TimeoutError:
            logger.error(f"Consumer-{i} execution timeout")
        except Exception as e:
            logger.error(f"Consumer-{i} execution error: {e}")

    return all_results


async def consumer(queue: asyncio.Queue, max_num: int, timeout: Optional[float] = None):
    results = []
    count = 0
    while count < max_num:
        got_item = False
        try:
            item = await asyncio.wait_for(queue.get(), timeout)
            got_item = True

            results.append(item)

            count += 1
        finally:
            if got_item:
                queue.task_done()

    return results


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
async def test_injection_and_building_dataset_all(
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
    trace_ids = [trace.trace_id for trace in traces]
    report = await sdk.task.get_stream_all(task_ids, trace_ids, timeout=None)
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
    trace_ids = [trace.trace_id for trace in traces]
    report = await sdk.task.get_stream_all(task_ids, trace_ids, timeout=30)
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

    injection_report = await test_injection_and_building_dataset_all(
        **injection_payload
    )
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
            # 检查当前层级是否有 key 字段
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
