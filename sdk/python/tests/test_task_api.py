# Run this file:
# uv run pytest -s tests/test_task_api.py
from typing import Any, Dict, List, Optional
from conftest import BASE_URL
from pprint import pprint
from rcabench.logger import logger
from rcabench.model.common import SubmitResult
from rcabench.model.task import QueueItem
from rcabench.rcabench import RCABenchSDK
from uuid import UUID
import asyncio
import pytest


@pytest.mark.asyncio
@pytest.mark.parametrize(
    "benchmark, interval, pre_duration, specs, max_items_per_consumer",
    [
        # One spec
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
        ),
        # Many specs
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
                        "4": {
                            "children": {
                                "0": {"value": 1},
                                "1": {"value": 0},
                                "2": {"value": 26},
                                "3": {"value": 10},
                                "4": {"value": 2},
                            }
                        },
                    },
                    "value": 4,
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
            1,
        ),
        # Total timeout
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
        ),
    ],
)
async def test_injection_and_building_dataset_batch(
    benchmark: str,
    interval: int,
    pre_duration: int,
    specs: List[Dict[str, Any]],
    max_items_per_consumer: int,
):
    per_consumer_timeout = (interval + 1) * 60
    total_timeout = len(specs) * (interval + 1) * 60

    results = {}
    try:
        results = await asyncio.wait_for(
            injection_and_building_dataset_batch(
                benchmark,
                interval,
                pre_duration,
                specs,
                max_items_per_consumer,
                per_consumer_timeout,
            ),
            timeout=total_timeout,
        )
    except asyncio.TimeoutError:
        logger.error("Total time out")
    finally:
        logger.info(f"Final results: {results}")


async def injection_and_building_dataset_batch(
    benchmark: str,
    interval: int,
    pre_duration: int,
    specs: List[Dict[str, Any]],
    max_items_per_consumer: int,
    per_consumer_timeout: float,
):
    sdk = RCABenchSDK(BASE_URL)

    resp = sdk.injection.submit(benchmark, interval, pre_duration, specs)
    logger.info(resp)

    traces = resp.traces

    task_ids = [trace.head_task_id for trace in traces]
    trace_ids = [trace.trace_id for trace in traces]
    queue = await sdk.task.get_stream_batch(task_ids, trace_ids, interval * 60)

    num_consumers = len(traces) // max_items_per_consumer
    results = await run_consumers(
        queue,
        num_consumers,
        max_items_per_consumer,
        per_consumer_timeout,
    )

    await sdk.task.stream.cleanup()
    return results


async def run_consumers(
    queue: asyncio.Queue[QueueItem],
    num_consumers: int,
    max_items_per_consumer: int,
    per_consumer_timeout: float,
):
    all_results = {}

    for i in range(num_consumers):
        all_results[f"Batch-{i}"] = await consumer_task(
            i,
            queue,
            max_items_per_consumer,
            per_consumer_timeout,
        )

    return all_results


async def consumer_task(
    consumer_id: int,
    queue: asyncio.Queue[QueueItem],
    max_num: int,
    timeout: float,
) -> Optional[List[Dict[str, Any]]]:
    try:
        results = await consumer(queue, max_num, timeout)
        logger.info(f"Consumer-{consumer_id} completed")
        return results
    except asyncio.TimeoutError:
        logger.error(f"Consumer-{consumer_id} timed out")
        return None
    except Exception as e:
        logger.error(f"Consumer-{consumer_id} failed: {str(e)}")
        return None


async def consumer(
    queue: asyncio.Queue[QueueItem],
    max_num: int,
    timeout: float,
) -> List[Dict[str, Any]]:
    results = []
    count = 0
    while count < max_num:
        try:
            item = await asyncio.wait_for(queue.get(), timeout)
            results.append(item.model_dump(exclude_unset=True))
            count += 1
        except asyncio.TimeoutError:
            logger.warning("Timeout while waiting for queue item")
            raise

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
    "benchmark, interval, pre_duration, specs",
    [
        (
            "clickhouse",
            5,
            1,
            [
                {
                    "children": {
                        "16": {
                            "children": {
                                "0": {"value": 1},
                                "1": {"value": 1},
                                "2": {"value": 10},
                                "3": {"value": -576},
                            }
                        },
                    },
                    "value": 16,
                },
            ],
        ),
    ],
)
async def test_injection_and_building_dataset_single(
    benchmark, interval, pre_duration, specs
):
    sdk = RCABenchSDK(BASE_URL)

    if len(specs) != 1:
        pytest.fail("The length of specs must be 1")

    resp = sdk.injection.submit(benchmark, interval, pre_duration, specs)
    pprint(resp)

    if not isinstance(resp, SubmitResult):
        pytest.fail(resp.model_dump_json())

    traces = resp.traces
    if not traces:
        pytest.fail("No traces returned from execution")

    task_id = [trace.head_task_id for trace in traces][0]
    trace_id = [trace.trace_id for trace in traces][0]

    timeout = (interval + 1) * 60
    report = await sdk.task.get_stream_single(task_id, trace_id, timeout)
    report = report.model_dump(exclude_unset=True)
    pprint(report)

    await sdk.task.stream.cleanup()


@pytest.mark.asyncio
@pytest.mark.parametrize(
    "payload",
    [
        (
            [
                {
                    "image": "e-diagnose",
                    "dataset": "ts-ts-rebook-service-pod-failure-xdqs9v",
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
    report = await sdk.task.get_stream_all(task_ids, trace_ids, client_timeout=60)
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
