# Run this file:
# uv run pytest -s tests/test_algorithm_api.py
from rcabench.model.common import SubmitResult
from pprint import pprint
import pytest


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
        ),
        (
            [
                {
                    "algorithm": "rcabench-rcaeval-baro",
                    "dataset": "ts-ts-preserve-service-cpu-exhaustion-j4pjlb",
                }
            ]
        ),
    ],
)
def test_submit_algorithms(sdk, payload):
    """测试批量提交算法"""
    resp = sdk.algorithm.submit(payload)
    pprint(resp)

    if not isinstance(resp, SubmitResult):
        pytest.fail(resp.model_dump_json())

    traces = resp.traces
    if not traces:
        pytest.fail("No traces returned from execution")
