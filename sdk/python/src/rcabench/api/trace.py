import json
import logging
import time
from enum import Enum
from typing import Any, Generator, Optional, Dict
import requests
from pydantic import BaseModel, Field

# Set up logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


class TaskType(str, Enum):
    COLLECT_RESULT = "collect_result"
    # Add other task types as needed


class EventType(str, Enum):
    UPDATE = "update"
    END = "end"
    # Add other event types as needed


class StreamEvent(BaseModel):
    task_id: str = Field(..., alias="task_id")
    task_type: str = Field(..., alias="task_type")
    file_name: str = Field(..., alias="file_name")
    line: int = Field(..., alias="line")
    event_name: str = Field(..., alias="event_name")
    payload: Any = Field(..., alias="payload")

    class Config:
        allow_population_by_field_name = True


class SSEClient:
    def __init__(self, base_url: str, max_retries: int = 3, retry_delay: int = 5):
        """
        Initialize the SSE client.

        Args:
            base_url: The base URL of the API (e.g., "http://api.example.com")
            max_retries: Maximum number of connection retry attempts
            retry_delay: Delay in seconds between retry attempts
        """
        self.base_url = base_url
        self.last_id = "0"
        self.max_retries = max_retries
        self.retry_delay = retry_delay

    def get_trace_events(self, trace_id: str) -> Generator[StreamEvent, None, None]:
        """
        Connect to the SSE endpoint and yield StreamEvent objects.

        Args:
            trace_id: The trace ID to stream events for

        Yields:
            Validated StreamEvent objects
        """
        retries = 0

        while retries < self.max_retries:
            try:
                url = f"{self.base_url}/api/v1/traces/{trace_id}/stream"
                headers = {
                    "Accept": "text/event-stream",
                    "Cache-Control": "no-cache",
                    "Last-Event-ID": self.last_id,
                }

                logger.info(f"Connecting to {url} with Last-Event-ID: {self.last_id}")
                response = requests.get(url, headers=headers, stream=True)
                response.raise_for_status()

                # Process the SSE stream
                buffer = ""
                for chunk in response.iter_content(chunk_size=1):
                    if not chunk:
                        continue

                    chunk_str = chunk.decode("utf-8")
                    buffer += chunk_str

                    if buffer.endswith("\n\n"):
                        lines = buffer.strip().split("\n")
                        event_type = None
                        data = None

                        for line in lines:
                            if line.startswith("event:"):
                                event_type = line[6:].strip()
                            elif line.startswith("data:"):
                                data = line[5:].strip()
                            elif line.startswith("id:"):
                                self.last_id = line[3:].strip()

                        if event_type == "update" and data:
                            try:
                                event_data = json.loads(data)
                                event = StreamEvent.parse_obj(event_data)
                                yield event
                            except Exception as e:
                                logger.error(f"Error parsing event: {e}, data: {data}")

                        if event_type == "end":
                            logger.info("Received end event, closing connection")
                            return

                        buffer = ""

                # If we reach here, the connection was closed normally
                logger.info("Connection closed")
                return

            except requests.exceptions.RequestException as e:
                retries += 1
                logger.error(
                    f"Connection error: {e}. Retry {retries}/{self.max_retries}"
                )
                if retries < self.max_retries:
                    time.sleep(self.retry_delay)
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
        while True:
            try:
                yield from self.get_trace_events(trace_id)
                # If get_trace_events() returns normally, we've received an end event
                return
            except Exception as e:
                logger.error(f"Error in stream_events: {e}")
                time.sleep(self.retry_delay)

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
