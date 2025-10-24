# Run this file:
# uv run pytest -s tests/test_trace_api.py
from pprint import pprint

import pytest

from rcabench.openapi.api import TracesApi
from rcabench.openapi.api_client import ApiClient
from rcabench.openapi.models import ConstsSSEEventName, DtoStreamEvent


class TestTraces:
    @pytest.fixture(autouse=True)
    def _setup(self, rcabench_client: ApiClient):
        self.traces_api = TracesApi(rcabench_client)

    @pytest.mark.parametrize(
        "id, last_id",
        [("bd1d7e15-f15d-4d77-b4f2-46f82e95ee96", "0")],
    )
    @pytest.mark.v2
    def test_get_trace_stream(self, id: str, last_id: str):
        try:
            sse_client = self.traces_api.api_v2_traces_id_stream_get_sse(id=id, last_id=last_id)
            for e in sse_client.events():
                if e.event == ConstsSSEEventName.EventEnd:
                    pprint("Received 'end' signal. Breaking loop and closing connection.")
                    break

                assert e.event == ConstsSSEEventName.EventUpdate, f"Unexpected event type: {e.event}"

                streamEvent = DtoStreamEvent.from_json(e.data)
                assert isinstance(streamEvent, DtoStreamEvent), "Expected event to be DtoStreamEvent"
                pprint(streamEvent)

        except Exception as e:
            pytest.fail(f"Trace stream failed unexpectedly: {e}")

        finally:
            sse_client.close()

        assert True
