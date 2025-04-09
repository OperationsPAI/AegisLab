from typing import Any, Dict, Optional
from ..const import EventType, SSEMsgPrefix, TaskStatus
from ..logger import logger
from uuid import UUID
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
            if client_id not in self.client_dict:
                return

            task_obj = self.client_dict.pop(client_id)
            if task_obj and not task_obj.done():
                task_obj.cancel()
                try:
                    await task_obj
                except (asyncio.CancelledError, Exception):
                    pass

            # 检查是否所有客户端都已移除
            if not self.client_dict:
                self.close_event.set()

    async def set_client_item(
        self, key: str, result: Any = None, error: Exception | None = None
    ) -> None:
        async with self.lock:
            if error:
                self.errors[key] = error
            if result:
                if key not in self.results:
                    self.results[key] = result
                self.results[key].update(result)

    async def wait_all(self, timeout: Optional[float] = None) -> Dict[str, Any]:
        start_time = asyncio.get_event_loop().time()

        # 1. 先等待所有客户端被移除（close_event被设置）
        if not self.close_event.is_set():
            try:
                if timeout is not None:
                    await asyncio.wait_for(self.close_event.wait(), timeout)
                else:
                    await self.close_event.wait()
            except asyncio.TimeoutError:
                pass

        # 2. 如果有任务仍在运行，等待它们完成
        if self.client_dict:
            tasks = list(self.client_dict.values())
            try:
                if timeout is not None:
                    elapsed = asyncio.get_event_loop().time() - start_time
                    remaining_timeout = max(0, timeout - elapsed)
                    if remaining_timeout > 0:
                        await asyncio.wait_for(
                            asyncio.gather(*tasks, return_exceptions=True),
                            remaining_timeout,
                        )
                else:
                    await asyncio.gather(*tasks, return_exceptions=True)
            except (asyncio.TimeoutError, asyncio.CancelledError):
                pass

        # 3. 清理已完成的任务
        async with self.lock:
            pending_tasks = {
                client_id: task
                for client_id, task in self.client_dict.items()
                if not task.done()
            }
            self.client_dict = pending_tasks

        return {
            "results": self.results,
            "errors": self.errors,
            "pending": list(self.client_dict.keys()),
        }

    async def cleanup(self):
        """清理所有任务和资源"""
        async with self.lock:
            # 1. 取消所有任务
            for task in self.client_dict.values():
                if not task.done():
                    task.cancel()

            # 2. 等待所有任务完成（即使是被取消的任务）
            if self.client_dict:
                await asyncio.gather(
                    *self.client_dict.values(),
                    return_exceptions=True,
                )

            # 3. 清空字典和结果
            self.client_dict.clear()
            self.results = {}
            self.errors = {}

            print(self.results)


class AsyncSSEClient:
    def __init__(self, client_manager: ClientManager, client_id: str, url: str):
        self.client_manager = client_manager
        self.client_id = client_id
        self.url = url
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
                task_id = UUID(data.pop("task_id"))
                status = data.pop("status")
                if status == TaskStatus.COMPLETED:
                    await self.client_manager.set_client_item(
                        self.client_id, result={task_id: data}
                    )

                if status == TaskStatus.ERROR:
                    error = RuntimeError(data.get("message"))
                    await self.client_manager.set_client_item(
                        self.client_id, error={task_id: error}
                    )

            except json.JSONDecodeError:
                pass

    async def connect(self):
        session = None
        try:
            await self.client_manager.add_client(self.client_id, asyncio.current_task())

            session = aiohttp.ClientSession()
            async with session.get(self.url) as resp:
                async for line in resp.content:
                    if self._close or self.client_manager.close_event.is_set():
                        break
                    await self._process_line(line)

        except asyncio.CancelledError:
            logger.error(f"Client {self.client_id} cancelled by manager")
            self._close = True
            await self.client_manager.set_client_item(
                self.client_id, error=RuntimeError("Client cancelled by manager")
            )
            await self.client_manager.remove_client(self.client_id)

        except Exception as e:
            logger.error(f"Client {self.client_id} exception occured: {str(e)}")
            self._close = True
            await self.client_manager.set_client_item(self.client_id, error=e)
            await self.client_manager.remove_client(self.client_id)

        finally:
            if session is not None:
                await session.close()

            if not self._close and not self.client_manager.close_event.is_set():
                await self.client_manager.set_client_item(
                    self.client_id,
                    error=RuntimeError("Connection closed unexpectedly"),
                )
                await self.client_manager.remove_client(self.client_id)
