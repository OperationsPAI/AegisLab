from typing import List, Optional
from .validation import async_validate_request_response
from .stream import Stream
from ..client.http_client import HTTPClient
from ..model.task import (
    QueueItem,
    StreamAllReq,
    StreamBatchReq,
    StreamSingleReq,
    StreamResult,
)
from uuid import UUID
import asyncio

__all__ = ["Task"]


class Task:
    URL_PREFIX = "/tasks"

    URL_ENDPOINTS = {
        "get_detail": "/{task_id}",
        "get_stream": "/{task_id}/stream",
    }

    def __init__(self, client: HTTPClient, api_version: str, max_connections: int = 10):
        self.client = client
        self.stream = Stream(client.base_url, max_connections)
        self.url_prefix = f"{api_version}{self.URL_PREFIX}"

    @async_validate_request_response(StreamSingleReq, QueueItem)
    async def get_stream_single(
        self,
        task_id: UUID,
        trace_id: UUID,
        client_timeout: float,
        wait_timeout: Optional[float] = None,
    ) -> QueueItem:
        url = f"{self.url_prefix}{self.URL_ENDPOINTS['get_stream'].format(task_id=task_id)}"

        try:
            report = await asyncio.wait_for(
                self.stream.start_single_stream(trace_id, url, client_timeout),
                wait_timeout,
            )
            return report
        finally:
            await self.stream.cleanup()

    @async_validate_request_response(StreamBatchReq)
    async def get_stream_batch(
        self, task_ids: List[UUID], trace_ids: List[UUID], client_timeout: float
    ) -> asyncio.Queue[QueueItem]:
        urls = [
            f"{self.url_prefix}{self.URL_ENDPOINTS['get_stream'].format(task_id=task_id)}"
            for task_id in task_ids
        ]

        queue = await self.stream.add_result_queue()
        await self.stream.start_multiple_stream(trace_ids, urls, client_timeout)
        return queue

    @async_validate_request_response(StreamAllReq, StreamResult)
    async def get_stream_all(
        self,
        task_ids: List[UUID],
        trace_ids: List[UUID],
        client_timeout: float,
        wait_timeout: Optional[float] = None,
    ) -> StreamResult:
        urls = [
            f"{self.url_prefix}{self.URL_ENDPOINTS['get_stream'].format(task_id=task_id)}"
            for task_id in task_ids
        ]

        try:
            await self.stream.start_multiple_stream(trace_ids, urls, client_timeout)
            report = await self.stream.client_manager.wait_all(wait_timeout)
            return report
        finally:
            await self.stream.cleanup()
