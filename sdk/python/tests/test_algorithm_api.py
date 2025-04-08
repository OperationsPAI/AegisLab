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
