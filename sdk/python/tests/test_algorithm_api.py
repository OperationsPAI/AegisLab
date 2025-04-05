# Run this file:
# uv run pytest -s tests/test_algorithm_api.py
from pprint import pprint
from rcabench import rcabench
import pytest


BASE_URL = "http://localhost:8082"


@pytest.mark.parametrize(
    "algorithms, datasets",
    [
        (
            [["detector", "latest"], ["e-diagnose", "latest"]],
            ["ts-ts-preserve-service-cpu-exhaustion-r4mq88"],
        )
    ],
)
def test_submit_algorithms(algorithms, datasets):
    """测试批量提交算法"""

    sdk = rcabench.RCABenchSDK(BASE_URL)

    data = sdk.algorithm.submit(algorithms, datasets)
    pprint(data)

    traces = data.traces
    if not traces:
        pytest.fail("No traces returned from execution")


# @pytest.mark.asyncio
# async def test_execute_algorithm_and_collection():
#     """测试执行多个算法并验证结果流收集功能

#     验证步骤：
#     1. 初始化 SDK 连接
#     2. 获取可用算法列表
#     3. 为每个算法生成执行参数
#     4. 提交批量执行请求
#     5. 启动流式结果收集
#     6. 验证关键结果字段
#     """
#     sdk = rcabench.RCABenchSDK(BASE_URL)

#     data = sdk.algorithm.submit(algorithms=ALGORITHMS, datasets=DATASETS)
#     pprint(data)

#     traces = data["traces"]
#     if not traces:
#         pytest.fail("No traces returned from execution")

#     task_ids = [trace["head_task_id"] for trace in traces]
#     report = await sdk.start_multiple_stream(
#         task_ids,
#         url="/algorithms/{task_id}/stream",
#         keyword="execution_id",
#         timeout=TIMEOUT,
#     )
#     pprint(report)
