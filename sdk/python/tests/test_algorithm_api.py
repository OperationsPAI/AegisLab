# Run this file:
# uv run pytest -s tests/test_algorithm_api.py
from pprint import pprint
import pytest


@pytest.mark.parametrize(
    "algorithms, datasets",
    [
        (
            [["detector", "latest"], ["e-diagnose", "latest"]],
            ["ts-ts-preserve-service-cpu-exhaustion-r4mq88"],
        )
    ],
)
def test_submit_algorithms(sdk, algorithms, datasets):
    """测试批量提交算法"""
    data = sdk.algorithm.submit(algorithms, datasets)
    pprint(data)

    traces = data.traces
    if not traces:
        pytest.fail("No traces returned from execution")
