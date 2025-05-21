from typing import Optional, Union, Generator
from .validation import validate_request_response
from ..client.http_client import HTTPClient
from ..model.error import ModelHTTPError
from ..model.trace import GetTraceEventsReq, StreamEvent, TraceEvents
from ..logger import logger


class Trace:
    URL_PREFIX = "/traces"

    URL_ENDPOINTS = {
        "get_trace_events": "/{trace_id}/stream",
    }

    def __init__(
        self,
        client: HTTPClient,
        api_version: str,
    ):
        self.client = client
        self.url_prefix = f"{api_version}{self.URL_PREFIX}"

    @validate_request_response(GetTraceEventsReq, TraceEvents)
    def get_trace_events(
        self, trace_id: str, last_event_id: str = "0", timeout: Optional[float] = None
    ) -> Union[TraceEvents, ModelHTTPError]:
        url = f"{self.url_prefix}{self.URL_ENDPOINTS['get_trace_events']}".format(
            trace_id=trace_id
        )
        headers = {
            "Accept": "text/event-stream",
            "Cache-Control": "no-cache",
        }

        logger.info(f"Connecting to {url} with Last-Event-ID: {last_event_id}")
        response = self.client.get(
            url,
            headers=headers,
            params={"last_event_id": last_event_id},
            stream=True,
            timeout=timeout,
        )
        if isinstance(response, ModelHTTPError):
            return response

        events = []
        event_data = {}
        for line in response.iter_lines(decode_unicode=True):
            if not line:
                if event_data:
                    event_type = event_data.get("event", "message")
                    data = event_data.get("data")

                    if event_type == "update" and data:
                        try:
                            events.append(StreamEvent.model_validate_json(data))
                        except Exception as e:
                            logger.error(f"Error parsing event: {e}, data: {data}")

                    if event_type == "end":
                        logger.info("Received end event, closing connection")
                        return {"last_event_id": last_event_id, "events": events}

                    event_data = {}
                    continue

            if ":" not in line:
                continue

            field, value = line.split(":", 1)
            value = value.lstrip()

            if field == "id":
                last_event_id = value
            elif field == "event":
                event_data["event"] = value
            elif field == "data":
                if "data" not in event_data:
                    event_data["data"] = value
                else:
                    event_data["data"] += "\n" + value

        logger.info("Connection closed")
        return {"last_event_id": last_event_id, "events": events}

    def stream_trace_events(
        self, trace_id: str, last_event_id: str = "0", timeout: Optional[float] = None
    ) -> Generator[Union[StreamEvent, ModelHTTPError], None, None]:
        url = f"{self.url_prefix}{self.URL_ENDPOINTS['get_trace_events']}".format(
            trace_id=trace_id
        )
        headers = {
            "Accept": "text/event-stream",
            "Cache-Control": "no-cache",
        }

        logger.info(f"Connecting to {url} with Last-Event-ID: {last_event_id}")
        response = self.client.get(
            url,
            headers=headers,
            params={"last_event_id": last_event_id},
            stream=True,
            timeout=timeout,
        )

        if isinstance(response, ModelHTTPError):
            logger.error(f"Error connecting to stream: {response}")
            yield response
            return

        event_data = {}
        for line in response.iter_lines(decode_unicode=True):
            if not line:
                if event_data:
                    event_type = event_data.get("event", "message")
                    data = event_data.get("data")

                    if event_type == "update" and data:
                        try:
                            yield StreamEvent.model_validate_json(data)
                        except Exception as e:
                            logger.error(f"Error parsing event: {e}, data: {data}")

                    if event_type == "end":
                        logger.info("Received end event, closing connection")
                        break

                    event_data = {}
                    continue

            if ":" not in line:
                continue

            field, value = line.split(":", 1)
            value = value.lstrip()

            if field == "id":
                last_event_id = value
            elif field == "event":
                event_data["event"] = value
            elif field == "data":
                if "data" not in event_data:
                    event_data["data"] = value
                else:
                    event_data["data"] += "\n" + value

        logger.info("Connection closed")
