from typing import AsyncGenerator, List, Optional
from ..client.async_client import AsyncSSEClient, ClientManager
from ..model.trace import QueueItem
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
        self.index = -1

    @asynccontextmanager
    async def _get_session(self) -> AsyncGenerator[aiohttp.ClientSession, None]:
        session = await self.conn_pool.get()
        try:
            yield session
        finally:
            await self.conn_pool.put(session)

    async def _stream_client(
        self,
        index: int,
        client_id: UUID,
        url: str,
        client_timeout: float,
        to_client_manager: bool,
    ) -> Optional[QueueItem]:
        retries = 0
        max_retries = 3

        sse_client = AsyncSSEClient(
            client_id,
            f"{self.base_url}{url}",
            self.client_manager if to_client_manager else None,
        )
        self.active_connections.add(client_id)

        while retries < max_retries:
            try:
                return await sse_client.connect(client_timeout * (index + 1))
            except aiohttp.ClientError:
                retries += 1
                await asyncio.sleep(2**retries)
            finally:
                await sse_client.close()
                self.active_connections.discard(client_id)

    async def add_result_queue(self) -> asyncio.Queue[QueueItem]:
        queue = asyncio.Queue()
        self.client_manager.result_queue = queue
        return queue

    async def start_single_stream(
        self, client_id: UUID, url: str, client_timeout: float
    ) -> QueueItem:
        self.index += 1
        return await self._stream_client(
            self.index,
            client_id,
            url,
            client_timeout,
            to_client_manager=False,
        )

    async def start_multiple_stream(
        self, client_ids: List[UUID], urls: List[str], client_timeout: float
    ) -> None:
        for idx, (client_id, url) in enumerate(zip(client_ids, urls)):
            self.index += idx + 1
            asyncio.create_task(
                self._stream_client(
                    self.index,
                    client_id,
                    url,
                    client_timeout,
                    to_client_manager=True,
                ),
                name=self.CLIENT_NAME.format(client_id=client_id),
            )

    async def stop_stream(self, client_id: UUID):
        for task in asyncio.all_tasks():
            if task.get_name() == self.CLIENT_NAME.format(client_id=client_id):
                task.cancel()
                break

    async def stop_all_streams(self):
        for client_id in list(self.active_connections):
            await self.stop_stream(client_id)

    async def cleanup(self):
        await self.stop_all_streams()
        while not self.conn_pool.empty():
            session = await self.conn_pool.get()
            await session.close()
        await self.client_manager.cleanup()
