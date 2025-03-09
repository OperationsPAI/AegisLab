from typing import Any, Dict, Optional
import aiohttp
import asyncio
import json


class TaskManager:
    def __init__(self):
        self.active_tasks = set()
        self.results = {}
        self.errors = {}
        self.close_event = asyncio.Event()
        self.lock = asyncio.Lock()

    async def add_task(self, task_id: str) -> None:
        async with self.lock:
            self.active_tasks.add(task_id)
            self.close_event.clear()

    async def remove_task(
        self, task_id: str, result: Any = None, error: Exception | None = None
    ) -> None:
        async with self.lock:
            self.active_tasks.discard(task_id)
            if error:
                self.errors[task_id] = error
            if result:
                self.results[task_id] = result

            if not self.active_tasks:
                self.close_event.set()

    async def wait_all(self, timeout: Optional[float] = None) -> Dict[str, Any]:
        try:
            await asyncio.wait_for(self.close_event.wait(), timeout)
        except asyncio.TimeoutError:
            pass

        return {
            "results": self.results,
            "errors": self.errors,
            "pending": list(self.active_tasks),
        }


class AsyncSSEClient:
    def __init__(self, task_manager: TaskManager, task_id: str, url: str):
        self.task_manager = task_manager
        self.task_id = task_id
        self.url = url
        self._close = False

    async def _process_line(self, line_bytes: bytes):
        line = line_bytes.decode()
        if line.startswith("data"):
            lines = line.strip().split("\n")

            data_parts = []
            for line in lines:
                data_part = line[len("data:") :].strip()
                data_parts.append(data_part)

            combined_data = "".join(data_parts)

            try:
                data = json.loads(combined_data)
                if data.get("status") in ["Completed", "Error"]:
                    self._close = True
                    result = data if data["status"] == "Completed" else None
                    error = (
                        RuntimeError(data.get("message"))
                        if data["status"] == "Error"
                        else None
                    )
                    await self.task_manager.remove_task(
                        self.task_id, result=result, error=error
                    )
            except json.JSONDecodeError:
                pass

    async def connect(self):
        async with aiohttp.ClientSession() as session:
            try:
                await self.task_manager.add_task(self.url)
                async with session.get(self.url) as resp:
                    async for line in resp.content:
                        await self._process_line(line)
            except Exception as e:
                await self.task_manager.remove_task(self.task_id, error=e)
                raise
            finally:
                if not self._close:
                    await self.task_manager.remove_task(
                        self.task_id,
                        error=RuntimeError("Connection closed unexpectedly"),
                    )
