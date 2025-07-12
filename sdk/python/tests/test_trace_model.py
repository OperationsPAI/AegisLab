# Run this file:
# uv run pytest -s tests/test_trace_api.py
from pprint import pprint
from rcabench.model.trace import ExecutionOptions
import pytest


@pytest.mark.parametrize(
    "json_str",
    [
        (
            '{"algorithm":{"name":"traceback","image":"","tag":""},"dataset":"ts3-ts-order-service-return-qkcshv","execution_id":2314}'
        )
    ],
)
def test_ExecutionOptions(json_str: str) -> None:
    try:
        model = ExecutionOptions.model_validate_json(json_str)
        pprint(model)
    except Exception as e:
        pytest.fail(f"Failed to validate ExecutionOptions: {e}")
