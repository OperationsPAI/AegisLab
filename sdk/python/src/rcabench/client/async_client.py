from typing import Any, Dict, Optional
from ..const import EventType, SSEMsgPrefix, TaskStatus
import aiohttp
import asyncio
import json
import re

__all__ = ["AsyncSSEClient", "ClientManager"]


class ClientManager:
    def __init__(self):
        self.client_dict: Dict[str, asyncio.Task] = {}
        self.results = {}
        self.errors = {}
        self.close_event = asyncio.Event()
        self.lock = asyncio.Lock()

    async def add_client(self, client_id: str, task_obj: asyncio.Task) -> None:
        async with self.lock:
            self.client_dict[client_id] = task_obj
            self.close_event.clear()

    async def remove_client(self, client_id: str) -> None:
        async with self.lock:
            if client_id in self.client_dict:
                task_obj = self.client_dict.pop(client_id, None)
                task_obj.cancel()

            if not self.client_dict:
                self.close_event.set()

    async def set_client_item(
        self, client_id: str, result: Any = None, error: Exception | None = None
    ) -> None:
        async with self.lock:
            if error:
                self.errors[client_id] = error
            if result:
                self.results[client_id] = result

    async def wait_all(self, timeout: Optional[float] = None) -> Dict[str, Any]:
        try:
            await asyncio.wait_for(self.close_event.wait(), timeout)
        except asyncio.TimeoutError:
            pass

        return {
            "results": self.results,
            "errors": self.errors,
            "pending": list(self.client_dict.keys()),
        }

    def cleanup(self):
        self.results = {}
        self.errors = {}


class AsyncSSEClient:
    def __init__(
        self,
        client_manager: ClientManager,
        client_id: str,
        url: str,
        keyword: Optional[str],
    ):
        self.client_manager = client_manager
        self.client_id = client_id
        self.url = url
        self.keyword = keyword
        self._close = False

    @staticmethod
    def _pattern_msg(prefix: str, text: str):
        pattern = re.compile(rf"{re.escape(prefix)}:\s*(.*)", re.DOTALL)

        match = pattern.search(text)
        if not match:
            return None

        return match.group(1).strip()

    async def _process_line(self, line_bytes: bytes):
        decoded_line = line_bytes.decode()
        if decoded_line.startswith(SSEMsgPrefix.EVENT):
            event_type = self._pattern_msg(SSEMsgPrefix.EVENT, decoded_line)
            if event_type and event_type == EventType.END:
                self._close = True
                await self.client_manager.remove_client(self.client_id)

        if decoded_line.startswith(SSEMsgPrefix.DATA):
            lines = decoded_line.strip().split("\n")

            data_parts = []
            for line in lines:
                data_part = self._pattern_msg(SSEMsgPrefix.DATA, line)
                data_parts.append(data_part)

            combined_data = "".join(data_parts)

            try:
                data = json.loads(combined_data)
                if data.get("status") == TaskStatus.COMPLETED:
                    await self.client_manager.set_client_item(
                        self.client_id, result=data
                    )

                if data.get("status") == TaskStatus.ERROR:
                    error = RuntimeError(data.get("message"))
                    await self.client_manager.set_client_item(
                        self.client_id, error=error
                    )

            except json.JSONDecodeError:
                pass

    async def connect(self):
        async with aiohttp.ClientSession() as session:
            try:
                await self.client_manager.add_client(
                    self.client_id, asyncio.current_task()
                )

                async with session.get(self.url) as resp:
                    async for line in resp.content:
                        if self._close:
                            break
                        await self._process_line(line)

            except asyncio.CancelledError:
                self.logger.error(f"Client {self.client_id} cancelled by manager")
                await self.client_manager.set_client_item(
                    self.client_id, error=RuntimeError("Client cancelled by manager")
                )
                await self.client_manager.remove_client(self.client_id)

            except Exception as e:
                await self.client_manager.set_client_item(self.client_id, error=e)
                await self.client_manager.remove_client(self.client_id)
                raise

            finally:
                if not self._close and not self.client_manager.close_event.is_set():
                    await self.client_manager.set_client_item(
                        self.client_id,
                        error=RuntimeError("Connection closed unexpectedly"),
                    )
                    await self.client_manager.remove_client(self.client_id)
