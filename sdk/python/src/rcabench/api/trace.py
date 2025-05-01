import json
from loguru import logger
import time
from enum import Enum
from typing import Any, Generator, Optional, Dict
import requests
from pydantic import BaseModel, Field


class TaskType(str, Enum):
    DUMMY = ""
    RESTART_SERVICE = "RestartService"
    RUN_ALGORITHM = "RunAlgorithm"
    FAULT_INJECTION = "FaultInjection"
    BUILD_IMAGES = "BuildImages"
    BUILD_DATASET = "BuildDataset"
    COLLECT_RESULT = "CollectResult"


class EventType(str, Enum):
    UPDATE = "update"
    END = "end"


class StreamEvent(BaseModel):
    task_id: str = Field(..., alias="task_id")
    task_type: str = Field(..., alias="task_type")
    file_name: str = Field(..., alias="file_name")
    line: int = Field(..., alias="line")
    event_name: str = Field(..., alias="event_name")
    payload: Any = Field(..., alias="payload")

    model_config = {
        "populate_by_name": True,
    }


class SSEClient:
    def __init__(
        self,
        base_url: str,
        max_retries: int = 3,
        retry_delay: int = 5,
        timeout: int = 30,
        max_backoff: int = 60,
    ):
        self.base_url = base_url
        self.last_id = "0"
        self.max_retries = max_retries
        self.retry_delay = retry_delay
        self.timeout = timeout
        self.max_backoff = max_backoff

    def get_trace_events(self, trace_id: str) -> Generator[StreamEvent, None, None]:
        retries = 0
        backoff = self.retry_delay

        while retries < self.max_retries:
            try:
                url = f"{self.base_url}/api/v1/traces/{trace_id}/stream"
                headers = {
                    "Accept": "text/event-stream",
                    "Cache-Control": "no-cache",
                }

                if self.last_id != "0":
                    headers["Last-Event-ID"] = self.last_id

                logger.info(f"Connecting to {url} with Last-Event-ID: {self.last_id}")
                response = requests.get(
                    url, headers=headers, stream=True, timeout=self.timeout
                )
                response.raise_for_status()

                # Reset backoff on successful connection
                backoff = self.retry_delay

                # Process the SSE stream
                event_data = {}

                for line in response.iter_lines(decode_unicode=True):
                    if not line:
                        # Empty line means end of event
                        if event_data:
                            event_type = event_data.get("event", "message")
                            data = event_data.get("data")

                            if event_type == "update" and data:
                                try:
                                    print(self.last_id)
                                    json_data = json.loads(data)
                                    event = StreamEvent.model_validate(json_data)
                                    yield event
                                except Exception as e:
                                    logger.error(
                                        f"Error parsing event: {e}, data: {data}"
                                    )

                            if event_type == "end":
                                logger.info("Received end event, closing connection")
                                return

                            # Clear event data for the next event
                            event_data = {}
                        continue

                    if ":" not in line:
                        continue

                    field, value = line.split(":", 1)
                    value = value.lstrip()

                    if field == "id":
                        self.last_id = value
                        logger.debug(f"Updated Last-Event-ID: {self.last_id}")
                    elif field == "event":
                        event_data["event"] = value
                    elif field == "data":
                        if "data" not in event_data:
                            event_data["data"] = value
                        else:
                            event_data["data"] += "\n" + value
                    elif field == "retry":
                        try:
                            self.retry_delay = int(value)
                            logger.debug(
                                f"Server requested retry delay: {self.retry_delay}"
                            )
                        except ValueError:
                            pass

                # If we reach here, the connection was closed normally
                logger.info("Connection closed")
                return

            except requests.exceptions.RequestException as e:
                retries += 1
                logger.error(
                    f"Connection error: {e}. Retry {retries}/{self.max_retries}"
                )
                if retries < self.max_retries:
                    # Apply exponential backoff with jitter
                    sleep_time = min(
                        backoff * (0.8 + 0.4 * (time.time() % 1)), self.max_backoff
                    )
                    logger.info(f"Waiting {sleep_time:.1f} seconds before retry")
                    time.sleep(sleep_time)
                    backoff = min(backoff * 2, self.max_backoff)
                else:
                    logger.error("Max retries reached, giving up")
                    raise

    def stream_events(self, trace_id: str) -> Generator[StreamEvent, None, None]:
        """
        Stream events with automatic reconnection.
        This is a convenience wrapper around get_trace_events() that handles reconnection.

        Args:
            trace_id: The trace ID to stream events for

        Yields:
            Validated StreamEvent objects
        """
        backoff = self.retry_delay

        while True:
            try:
                yield from self.get_trace_events(trace_id)
                # If get_trace_events() returns normally, we've received an end event
                return
            except Exception as e:
                logger.error(f"Error in stream_events: {e}")
                sleep_time = min(
                    backoff * (0.8 + 0.4 * (time.time() % 1)), self.max_backoff
                )
                logger.info(f"Waiting {sleep_time:.1f} seconds before reconnection")
                time.sleep(sleep_time)
                backoff = min(backoff * 2, self.max_backoff)

    def filter_events(
        self,
        trace_id: str,
        task_type: Optional[str] = None,
        event_name: Optional[str] = None,
    ) -> Generator[StreamEvent, None, None]:
        """
        Stream events filtered by task_type and/or event_name.

        Args:
            trace_id: The trace ID to stream events for
            task_type: Optional filter for task_type
            event_name: Optional filter for event_name

        Yields:
            Filtered StreamEvent objects
        """
        for event in self.stream_events(trace_id):
            if (task_type is None or event.task_type == task_type) and (
                event_name is None or event.event_name == event_name
            ):
                yield event


# Example usage:
if __name__ == "__main__":
    client = SSEClient("http://10.10.10.46:8082")

    # Basic usage - get all events
    trace_id = "e219195a-7316-4377-a513-41931403e165"
    for event in client.get_trace_events(trace_id):
        print(f"Received event: {event}")
        # Process the event as needed

    # # Advanced usage - filter events
    # for event in client.filter_events(trace_id, task_type="collect_result"):
    #     print(f"Received collect_result event: {event}")
