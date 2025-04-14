# Run this file:
# uv run pytest -s tests/test_algorithm_api.py
from rcabench.model.common import SubmitResult
from pprint import pprint
import pytest


@pytest.mark.parametrize(
    "payloads",
    [
        (
            [
                {
                    "image": "e-diagnose",
                    "dataset": "ts-ts-ui-dashboard-pod-failure-c27jzh",
                    "env_vars": {"SERVICE": "ts-ui-dashboard"},
                }
            ]
        ),
        (
            [
                {
                    "image": "rcabench-rcaeval-baro",
                    "dataset": "ts-ts-ui-dashboard-pod-failure-c27jzh",
                }
            ]
        ),
    ],
)
def test_submit_algorithms(sdk, payloads):
    """测试批量提交算法"""
    resp = sdk.algorithm.submit(payloads)
    pprint(resp)

    if not isinstance(resp, SubmitResult):
        pytest.fail(resp.model_dump_json())

    traces = resp.traces
    if not traces:
        pytest.fail("No traces returned from execution")
