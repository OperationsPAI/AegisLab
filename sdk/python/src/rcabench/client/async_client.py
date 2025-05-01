from typing import Any, Awaitable, Callable, Dict, Optional, Union
from ..const import EventType, SSEMsgPrefix, Task, TaskStatus
from ..logger import logger
from ..model.error import ModelHTTPError
from ..model.trace import QueueItem, SSEMessage
from uuid import UUID
import aiohttp
import asyncio
import json
import re
import traceback

__all__ = ["AsyncSSEClient", "ClientManager"]


class ClientManager:
    def __init__(self):
        self.client_dict: Dict[UUID, asyncio.Task] = {}
        self.results = {}
        self.errors = {}
        self.close_event = asyncio.Event()
        self.lock = asyncio.Lock()
        self.result_queue = None
        self.client_callbacks: Dict[UUID, Callable[[Any], Awaitable[None]]] = {}

    async def add_client(self, client_id: UUID, task_obj: asyncio.Task) -> None:
        async with self.lock:
            self.client_dict[client_id] = task_obj
            self.close_event.clear()

    async def remove_client(self, client_id: UUID) -> None:
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
        self,
        client_id: UUID,
        result: Optional[Dict[UUID, SSEMessage]] = None,
        error: Optional[Union[Dict[UUID, ModelHTTPError], Exception]] = None,
    ) -> None:
        async with self.lock:
            # 更新错误和结果数据
            if error is not None:
                if isinstance(error, Exception):
                    error = {Task.CLIENT_ERROR_KEY: str(Exception)}
                self.errors[client_id] = error

            if result is not None:
                if client_id not in self.results:
                    self.results[client_id] = result
                else:
                    self.results[client_id].update(result)

            # 生产到队列
            if self.result_queue is not None:
                if client_id in self.errors:
                    # 优先处理错误情况
                    queue_item = {
                        "client_id": client_id,
                        "data": {
                            "error": self.errors[client_id],
                        },
                    }

                    if client_id in self.results:
                        queue_item["data"].update({"result": self.results[client_id]})

                    res = QueueItem.model_validate(queue_item)
                    await self.result_queue.put(res)
                elif client_id in self.results and len(self.results[client_id]) >= 2:
                    # 只有结果且数量足够时才处理
                    res = QueueItem.model_validate(
                        {
                            "client_id": client_id,
                            "data": {
                                "result": self.results[client_id],
                            },
                        }
                    )
                    await self.result_queue.put(res)

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

            # 4. 清空队列
            if self.result_queue is not None:
                while not self.result_queue.empty():
                    try:
                        self.result_queue.get_nowait()
                        self.result_queue.task_done()
                    except asyncio.QueueEmpty:
                        break


class AsyncSSEClient:
    def __init__(
        self,
        client_id: UUID,
        url: str,
        client_manager: Optional[ClientManager] = None,
    ):
        self.client_id = client_id
        self.url = url
        self._close = False
        self._session = None

        self.to_client_manager = client_manager is not None
        if self.to_client_manager:
            self.client_manager = client_manager
        else:
            self.result = None
            self.error = None
            self.queue_item = None

    @staticmethod
    def _pattern_msg(prefix: str, text: str):
        pattern = re.compile(rf"{re.escape(prefix)}:\s*(.*)", re.DOTALL)

        match = pattern.search(text)
        if not match:
            return None

        return match.group(1).strip()

    async def _set_client_item(
        self,
        result: Optional[Dict[UUID, SSEMessage]] = None,
        error: Optional[Union[Dict[UUID, ModelHTTPError], Exception]] = None,
    ) -> None:
        if error is not None:
            if isinstance(error, Exception):
                self.error = {Task.CLIENT_ERROR_KEY: str(error)}
            else:
                self.error = error

        if result is not None:
            if self.result is None:
                self.result = result
            else:
                self.result.update(result)

        if self.error is not None:
            queue_item = {
                "client_id": self.client_id,
                "data": {
                    "error": self.error,
                },
            }

            if self.result is not None:
                queue_item["data"]["result"] = self.result

            self.queue_item = QueueItem.model_validate(queue_item)
        elif self.result is not None and len(self.result) >= 2:
            self.queue_item = QueueItem.model_validate(
                {
                    "client_id": self.client_id,
                    "data": {
                        "result": self.result,
                    },
                }
            )

    async def _process_line(self, line_bytes: bytes):
        decoded_line = line_bytes.decode()
        if decoded_line.startswith(SSEMsgPrefix.EVENT):
            event_type = self._pattern_msg(SSEMsgPrefix.EVENT, decoded_line)
            if event_type and event_type == EventType.END:
                self._close = True
                if self.to_client_manager:
                    await self.client_manager.remove_client(self.client_id)

        if decoded_line.startswith(SSEMsgPrefix.DATA):
            lines = decoded_line.strip().split("\n")

            data_parts = []
            for line in lines:
                data_part = self._pattern_msg(SSEMsgPrefix.DATA, line)
                data_parts.append(data_part)

            combined_data = "".join(data_parts)
            data = json.loads(combined_data)

            message = SSEMessage.model_validate(data)
            if message.status == TaskStatus.COMPLETED:
                result = {message.task_id: message}
                if self.to_client_manager:
                    await self.client_manager.set_client_item(
                        self.client_id,
                        result=result,
                    )
                else:
                    await self._set_client_item(result=result)

            if message.status == TaskStatus.ERROR:
                error = {
                    message.task_id: ModelHTTPError(
                        status_code=Task.HTTP_ERROR_STATUS_CODE,
                        detail=message.payload.error,
                    )
                }
                if self.to_client_manager:
                    await self.client_manager.set_client_item(
                        self.client_id,
                        error=error,
                    )
                else:
                    await self._set_client_item(error=error)

    async def _connect_with_return(self, client_timeout: float) -> None:
        try:
            timeout = aiohttp.ClientTimeout(total=client_timeout)
            self._session = aiohttp.ClientSession(
                timeout=timeout, headers={"Accept": "text/event-stream"}
            )
            async with self._session.get(self.url) as resp:
                try:
                    async for line in resp.content:
                        if self._close:
                            break
                        await self._process_line(line)
                except asyncio.TimeoutError:
                    logger.warning(f"Client {self.client_id} stream timeout")
                    raise

        except asyncio.CancelledError:
            logger.warning(f"Client {self.client_id} cancelled by manager")
            self._close = True
            await self._set_client_item(
                error=RuntimeError("Client cancelled by manager")
            )

        except Exception as e:
            logger.error(
                f"Client {self.client_id} exception occured: {traceback.format_exc()}"
            )
            self._close = True
            await self._set_client_item(error=e)

        finally:
            if self._session is not None:
                await self._session.close()
                self._session = None

            if not self._close:
                await self._set_client_item(
                    error=RuntimeError("Connection closed unexpectedly"),
                )

    async def _connect_to_client_manager(self, client_timeout: float) -> None:
        try:
            await self.client_manager.add_client(self.client_id, asyncio.current_task())

            timeout = aiohttp.ClientTimeout(total=client_timeout)
            self._session = aiohttp.ClientSession(
                timeout=timeout, headers={"Accept": "text/event-stream"}
            )
            async with self._session.get(self.url) as resp:
                try:
                    async for line in resp.content:
                        if self._close or self.client_manager.close_event.is_set():
                            break
                        await self._process_line(line)
                except asyncio.TimeoutError:
                    logger.warning(f"Client {self.client_id} stream timeout")
                    raise

        except asyncio.CancelledError:
            logger.warning(f"Client {self.client_id} cancelled by manager")
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
            if self._session is not None:
                await self._session.close()
                self._session = None

            if not self._close and not self.client_manager.close_event.is_set():
                await self.client_manager.set_client_item(
                    self.client_id,
                    error=RuntimeError("Connection closed unexpectedly"),
                )
                await self.client_manager.remove_client(self.client_id)

    async def connect(self, client_timeout: float) -> Optional[QueueItem]:
        if self.to_client_manager:
            await self._connect_to_client_manager(client_timeout)
            return None
        else:
            await self._connect_with_return(client_timeout)
            return self.queue_item

    async def close(self):
        """关闭SSE连接并清理资源"""
        if self._close:
            return

        self._close = True
        if self._session is not None:
            await self._session.close()
            self._session = None

        if self.to_client_manager:
            await self.client_manager.remove_client(self.client_id)
