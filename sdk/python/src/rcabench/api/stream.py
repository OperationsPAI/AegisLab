from typing import AsyncGenerator, List
from ..client.async_client import AsyncSSEClient, ClientManager
from contextlib import asynccontextmanager
from uuid import UUID
import aiohttp
import asyncio

__all__ = ["Stream"]


class Stream:
    CLIENT_NAME = "SSE-{client_id}"

    def __init__(self, base_url: str, max_connections: int = 10):
        self.base_url = base_url
        self.client_manager = ClientManager()
        self.conn_pool = asyncio.Queue(max_connections)
        self.active_connections = set()

    @asynccontextmanager
    async def _get_session(self) -> AsyncGenerator[aiohttp.ClientSession, None]:
        session = await self.conn_pool.get()
        try:
            yield session
        finally:
            await self.conn_pool.put(session)

    async def _stream_client(self, client_id: UUID, url: str) -> None:
        retries = 0
        max_retries = 3

        sse_client = AsyncSSEClient(
            self.client_manager, client_id, f"{self.base_url}{url}"
        )
        self.active_connections.add(client_id)

        while retries < max_retries:
            try:
                await sse_client.connect()
                break
            except aiohttp.ClientError:
                retries += 1
                await asyncio.sleep(2**retries)
            finally:
                await sse_client.close()
                self.active_connections.discard(client_id)

    async def add_result_queue(self) -> asyncio.Queue:
        queue = asyncio.Queue()
        self.client_manager.result_queue = queue
        return queue

    async def start_multiple_stream(
        self, client_ids: List[UUID], urls: List[str]
    ) -> None:
        """批量启动多个SSE流"""
        for client_id, url in zip(client_ids, urls):
            asyncio.create_task(
                self._stream_client(client_id, url),
                name=self.CLIENT_NAME.format(client_id=client_id),
            )

    async def stop_stream(self, client_id: UUID):
        """停止指定SSE流"""
        for task in asyncio.all_tasks():
            if task.get_name() == self.CLIENT_NAME.format(client_id=client_id):
                task.cancel()
                break

    async def stop_all_streams(self):
        """停止所有SSE流"""
        for client_id in list(self.active_connections):
            await self.stop_stream(client_id)

    async def cleanup(self):
        """清理所有资源"""
        await self.stop_all_streams()
        while not self.conn_pool.empty():
            session = await self.conn_pool.get()
            await session.close()
