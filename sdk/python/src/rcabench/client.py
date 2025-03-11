from typing import Any, Dict, List, Optional, Union
from .error import HttpClientError
from .logger import CustomLogger
from requests.adapters import HTTPAdapter
from requests.exceptions import HTTPError, RequestException, Timeout
import aiohttp
import asyncio
import json
import logging
import requests
import time


class TaskManager:
    def __init__(self):
        self.task_objects: Dict[str, asyncio.Task] = {}
        self.results = {}
        self.errors = {}
        self.close_event = asyncio.Event()
        self.lock = asyncio.Lock()

    async def add_task(self, task_id: str, task_obj: asyncio.Task) -> None:
        async with self.lock:
            self.task_objects[task_id] = task_obj
            self.close_event.clear()

    async def remove_task(
        self, task_id: str, result: Any = None, error: Exception | None = None
    ) -> None:
        async with self.lock:
            if task_id in self.task_objects:
                task_obj = self.task_objects.pop(task_id, None)
                task_obj.cancel()

            if error:
                self.errors[task_id] = error
            if result:
                self.results[task_id] = result

            if not self.task_objects:
                self.close_event.set()

    async def wait_all(self, timeout: Optional[float] = None) -> Dict[str, Any]:
        try:
            await asyncio.wait_for(self.close_event.wait(), timeout)
        except asyncio.TimeoutError:
            pass

        return {
            "results": self.results,
            "errors": self.errors,
            "pending": list(self.task_objects.keys()),
        }

    def cleanup(self):
        self.results = {}
        self.errors = {}


class AsyncSSEClient:
    def __init__(
        self, task_manager: TaskManager, task_id: str, url: str, keyword: Optional[str]
    ):
        self.task_manager = task_manager
        self.task_id = task_id
        self.url = url
        self.keyword = keyword
        self.logger = CustomLogger().logger
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
                if data.get("status") == "Error":
                    self._close = True
                    error = RuntimeError(data.get("message"))
                    await self.task_manager.remove_task(self.task_id, error=error)
                elif data.get("status") == "Completed":
                    if not self.keyword or (self.keyword and self.keyword in data):
                        self._close = True
                        result = data
                        await self.task_manager.remove_task(self.task_id, result=result)
            except json.JSONDecodeError:
                pass

    async def connect(self):
        async with aiohttp.ClientSession() as session:
            try:
                task = asyncio.current_task()
                await self.task_manager.add_task(self.task_id, task)

                async with session.get(self.url) as resp:
                    async for line in resp.content:
                        if self._close:
                            break
                        await self._process_line(line)
            except asyncio.CancelledError:
                self.logger.info(f"Task {self.task_id} cancelled by manager")
                await self.task_manager.remove_task(
                    self.task_id, error=RuntimeError("Task cancelled by manager")
                )
            except Exception as e:
                await self.task_manager.remove_task(self.task_id, error=e)
                raise
            finally:
                if not self._close:
                    await self.task_manager.remove_task(
                        self.task_id,
                        error=RuntimeError("Connection closed unexpectedly"),
                    )


class HttpClient:
    def __init__(
        self,
        base_url: str,
        timeout: int = 10,
        max_retries: int = 3,
        backoff_factor: float = 0.5,
    ):
        self.base_url = base_url
        self.timeout = timeout
        self.max_retries = max_retries
        self.backoff_factor = backoff_factor
        self.logger = logging.getLogger(self.__class__.__name__)

        # 配置Session对象复用TCP连接
        self.session = requests.Session()
        adapter = HTTPAdapter(max_retries=max_retries)
        self.session.mount("http://", adapter)
        self.session.mount("https://", adapter)

    def _request(
        self,
        method: str,
        url: str,
        params: Optional[Dict] = None,
        json: Optional[Any] = None,
        stream: bool = False,
        retries: int = 3,
    ) -> Any:
        full_url = f"{self.base_url}{url}"

        for attempt in range(retries):
            try:
                self.logger.info(f"Sending {method} request to {full_url}")

                response = self.session.request(
                    method=method,
                    url=full_url,
                    params=params,
                    json=json,
                    timeout=self.timeout,
                    stream=stream,
                )
                response.raise_for_status()

                return response.json() if not stream else response
            except HTTPError as e:
                status_code = e.response.status_code
                if 500 <= status_code < 600 and attempt < self.max_retries:
                    self._handle_retry(attempt, e)
                    continue
                raise HttpClientError(
                    message=f"Server returned {status_code}",
                    status_code=status_code,
                    url=url,
                ) from e
            except (Timeout, RequestException) as e:
                if attempt == self.max_retries:
                    raise HttpClientError(
                        message=f"Request failed after {self.max_retries} retries: {str(e)}",
                        url=url,
                    ) from e
                self._handle_retry(attempt, e)

        return None

    def _handle_retry(self, attempt: int, error: Exception):
        sleep_time = self.backoff_factor * (2**attempt)
        self.logger.warning(
            f"Attempt {attempt + 1} failed: {error}. Retrying in {sleep_time:.1f}s..."
        )
        time.sleep(sleep_time)

    def delete(self, url: str, params: Optional[Dict] = None) -> Any:
        return self._request("DELETE", url, params=params)

    def get(self, url: str, params: Optional[Dict] = None, stream: bool = False) -> Any:
        return self._request("GET", url, params=params, stream=stream)

    def post(self, url: str, payload: Union[Dict, List[Dict]]) -> Any:
        return self._request("POST", url, json=payload)

    def put(self, url: str, payload: Union[Dict, List[Dict]]) -> Any:
        return self._request("PUT", url, json=payload)

    def __enter__(self):
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        self.session.close()
