from typing import Any, AsyncGenerator, Dict, List, Optional, Union
from .client import AsyncSSEClient, HttpClient, TaskManager
from .error import HttpClientError
from .logger import CustomLogger
from contextlib import asynccontextmanager
from functools import wraps
from tqdm import tqdm
from urllib.parse import unquote
import aiohttp
import asyncio
import os
import requests

CLIENT_NAME = "SSE-{task_id}"

logger = CustomLogger().logger


def handle_http_errors(func):
    @wraps(func)
    def wrapper(*args, **kwargs):
        try:
            result = func(*args, **kwargs)
            if result is not None and isinstance(result, dict):
                return result.get("data", None)

        except HttpClientError as e:
            # 统一记录日志或返回错误响应
            logger.error(f"API request failed: {e.url} -> {e.message}")
            return None

        except Exception as e:
            logger.error(f"Unknown error: {str(e)}")
            return None

    return wrapper


class classproperty:
    def __init__(self, func):
        self.func = func

    def __get__(self, instance, owner):
        return self.func(owner)


class BaseRouter:
    URL_PREFIX = ""

    URL_ENDPOINTS = {
        "execute": "",
        "list": "",
    }

    def __init__(self, sdk):
        self.sdk = sdk

    def _build_url(self, endpoint: str) -> str:
        return f"{self.URL_PREFIX}{endpoint}"

    @handle_http_errors
    def execute(self, payloads: Union[Dict, List[Dict]]):
        url = self._build_url(self.URL_ENDPOINTS["execute"])
        return self.sdk.client.post(url, payloads)

    @handle_http_errors
    def list(self) -> Dict[str, Any]:
        url = self._build_url(self.URL_ENDPOINTS["list"])
        return self.sdk.client.get(url)


class Algorithm(BaseRouter):
    URL_PREFIX = "/algorithms"

    _EXTRA_ENDPOINTS = {}

    @classproperty
    def URL_ENDPOINTS(cls):
        return {**super().URL_ENDPOINTS, **cls._EXTRA_ENDPOINTS}


class Dataset(BaseRouter):
    URL_PREFIX = "/datasets"

    _EXTRA_ENDPOINTS = {
        "delete": "",
        "download": "/download",
    }

    @handle_http_errors
    def delete(self, ids: List[int]):
        url = self._build_url(self.URL_ENDPOINTS["delete"])
        return self.sdk.client.delete(url, params={"ids": ids})

    @handle_http_errors
    def download(self, group_ids: List[str], output_path: str):
        url = self._build_url(self.URL_ENDPOINTS["download"])
        response = self.sdk.client.get(
            url, params={"group_ids": group_ids}, stream=True
        )

        if response:
            total_size = int(response.headers.get("content-length", 0))
            progress_bar = tqdm(total=total_size, unit="B", unit_scale=True)

            filename = unquote(url.split("/")[-1])
            if "Content-Disposition" in response.headers:
                content_disposition = response.headers["Content-Disposition"]
                filename = unquote(
                    content_disposition.split("filename=")[-1].strip('"')
                )

            with open(os.path.join(output_path, filename), "wb") as f:
                for chunk in response.iter_content(chunk_size=1024):
                    if chunk:
                        f.write(chunk)
                        progress_bar.update(len(chunk))

            progress_bar.close()
            return

    @handle_http_errors
    def list(self, page_num: int, page_size: int):
        url = self._build_url(self.URL_ENDPOINTS["list"])
        params = {"page_num": page_num, "page_size": page_size}
        return self.sdk.client.get(url, params=params)

    @classproperty
    def URL_ENDPOINTS(cls):
        return {**super().URL_ENDPOINTS, **cls._EXTRA_ENDPOINTS}


class Evaluation(BaseRouter):
    URL_PREFIX = "/evaluations"

    URL_ENDPOINTS = {
        "execute": "",
    }

    @handle_http_errors
    def execute(self, params: Dict):
        url = self._build_url(self.URL_ENDPOINTS["execute"])
        return self.sdk.client.get(url, params=params)


class Injection(BaseRouter):
    URL_PREFIX = "/injections"

    _EXTRA_ENDPOINTS = {
        "execute": "",
        "get_namespace_pod_info": "/namespace_pods",
        "get_parameters": "/parameters",
        "list": "",
    }

    @handle_http_errors
    def get_namespace_pod_info(self):
        url = self._build_url(self.URL_ENDPOINTS["get_namespace_pod_info"])
        return self.sdk.client.get(url)

    @handle_http_errors
    def get_parameters(self):
        url = self._build_url(self.URL_ENDPOINTS["get_parameters"])
        return self.sdk.client.get(url)

    @classproperty
    def URL_ENDPOINTS(cls):
        return {**super().URL_ENDPOINTS, **cls._EXTRA_ENDPOINTS}


class RCABenchSDK:
    def __init__(self, base_url: str, max_connections: int = 10):
        """
        Initialize the SDK with the base URL of the server.

        :param base_url: Base URL of the RCABench server, e.g., "http://localhost:8080"
        """
        self.base_url = base_url.rstrip("/") + "/api/v1"
        self.algorithm = Algorithm(self)
        self.dataset = Dataset(self)
        self.evaluation = Evaluation(self)
        self.injection = Injection(self)

        self.task_manager = TaskManager()
        self.client = HttpClient(self.base_url)
        self.conn_pool = asyncio.Queue(max_connections)
        self.active_connections = set()
        self.loop = asyncio.get_event_loop()

    def _delete(self, url: str, params: Optional[Dict] = None) -> Any:
        url = f"{self.base_url}{url}"
        response = requests.delete(url, params=params)
        response.raise_for_status()
        return response.json()

    @asynccontextmanager
    async def _get_session(self) -> AsyncGenerator[aiohttp.ClientSession, None]:
        session = await self.conn_pool.get()
        try:
            yield session
        finally:
            await self.conn_pool.put(session)

    async def _create_sse_client(
        self, task_id: str, url: str, keyword: Optional[str]
    ) -> AsyncSSEClient:
        url = f"{self.base_url}{url}"
        return AsyncSSEClient(self.task_manager, task_id, url, keyword)

    async def _stream_task(
        self, task_id: str, url: str, keyword: Optional[str]
    ) -> None:
        retries = 0
        max_retries = 3

        sse_client = await self._create_sse_client(task_id, url, keyword)
        self.active_connections.add(task_id)

        while retries < max_retries:
            try:
                await sse_client.connect()
                break
            except aiohttp.ClientError:
                retries += 1
                await asyncio.sleep(2**retries)

        self.active_connections.discard(task_id)

    async def start_multiple_stream(
        self,
        task_ids: List[str],
        url: str,
        keyword: Optional[str] = None,
        timeout: Optional[float] = None,
    ) -> Dict[str, Any]:
        """批量启动多个SSE流"""
        for task_id in task_ids:
            asyncio.create_task(
                self._stream_task(task_id, url.format(task_id=task_id), keyword),
                name=CLIENT_NAME.format(task_id=task_id),
            )

        report = await self.task_manager.wait_all(timeout)
        await self._cleanup()

        return report

    async def stop_stream(self, task_id: str):
        """停止指定SSE流"""
        for task in asyncio.all_tasks():
            if task.get_name() == CLIENT_NAME.format(task_id=task_id):
                task.cancel()
                break

    async def stop_all_streams(self):
        """停止所有SSE流"""
        for task_id in list(self.active_connections):
            await self.stop_stream(task_id)

    async def _cleanup(self):
        """清理所有资源"""
        await self.stop_all_streams()
        while not self.conn_pool.empty():
            session = await self.conn_pool.get()
            await session.close()
