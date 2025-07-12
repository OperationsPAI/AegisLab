# Run this file:
# uv run pytest -s tests/test_trace_api.py
from pprint import pprint
from rcabench.model.error import ModelHTTPError
import pytest


@pytest.mark.parametrize(
    "trace_id, last_event_id, timeout",
    [("8c3e4dd4-db86-49d9-b034-06e5a97c632c", "0", 600)],
)
def test_get_trace_events(
    sdk, trace_id: str, last_event_id: str, timeout: float | None
):
    data = sdk.trace.get_trace_events(trace_id, last_event_id, timeout)
    if isinstance(data, ModelHTTPError):
        pytest.fail("Failed to get trace events")

    pprint(data)
