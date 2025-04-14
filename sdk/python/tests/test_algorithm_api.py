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
                }
            ]
        ),
        (
            [
                {
                    "image": "rcabench-rcaeval-generic",
                    "dataset": "ts-ts-travel-service-delay-ftz7lb",
                    "env_vars": {"ALGORITHM": "baro", "VENV": "default"},
                },
                {
                    "image": "rcabench-rcaeval-generic",
                    "dataset": "ts-ts-travel-service-delay-ftz7lb",
                    "env_vars": {"ALGORITHM": "nsigma", "VENV": "default"},
                },
                {
                    "image": "rcabench-rcaeval-generic",
                    "dataset": "ts-ts-travel-service-delay-ftz7lb",
                    "env_vars": {"ALGORITHM": "circa", "VENV": "default"},
                },
                {
                    "image": "rcabench-rcaeval-generic",
                    "dataset": "ts-ts-travel-service-delay-ftz7lb",
                    "env_vars": {"ALGORITHM": "rcd", "VENV": "rcd"},
                },
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
