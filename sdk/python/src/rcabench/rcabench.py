from typing import Any, AsyncGenerator, Dict, List, Optional, Union
from .api import Algorithm, Dataset
from .client.async_client import AsyncSSEClient, ClientManager
from .client.http_client import HttpClient
from contextlib import asynccontextmanager
import aiohttp
import asyncio
import requests


class RCABenchSDK:
    CLIENT_NAME = "SSE-{client_id}"

    def __init__(self, base_url: str, max_connections: int = 10):
        """
        Initialize the SDK with the base URL of the server.

        :param base_url: Base URL of the RCABench server, e.g., "http://localhost:8080"
        """
        self.base_url = base_url.rstrip("/") + "/api/v1"

        self.client = HttpClient(self.base_url)
        self.algorithm = Algorithm(self.client)
        self.dataset = Dataset(self.client)
        self.evaluation = Evaluation(self)
        self.injection = Injection(self)

        self.client_manager = ClientManager()
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
        self, client_id: str, url: str, keyword: Optional[str]
    ) -> AsyncSSEClient:
        url = f"{self.base_url}{url}"
        return AsyncSSEClient(self.client_manager, client_id, url, keyword)

    async def _stream_client(
        self, client_id: str, url: str, keyword: Optional[str]
    ) -> None:
        retries = 0
        max_retries = 3

        sse_client = await self._create_sse_client(client_id, url, keyword)
        self.active_connections.add(client_id)

        while retries < max_retries:
            try:
                await sse_client.connect()
                break
            except aiohttp.ClientError:
                retries += 1
                await asyncio.sleep(2**retries)

        self.active_connections.discard(client_id)

    async def start_multiple_stream(
        self,
        client_ids: List[str],
        url: str,
        keyword: Optional[str] = None,
        timeout: Optional[float] = None,
    ) -> Dict[str, Any]:
        """批量启动多个SSE流"""
        for client_id in client_ids:
            asyncio.create_task(
                self._stream_client(client_id, url.format(trace_id=client_id), keyword),
                name=self.CLIENT_NAME.format(client_id=client_id),
            )

        report = await self.client_manager.wait_all(timeout)
        await self._cleanup()

        return report

    async def stop_stream(self, client_id: str):
        """停止指定SSE流"""
        for task in asyncio.all_tasks():
            if task.get_name() == self.CLIENT_NAME.format(client_id=client_id):
                task.cancel()
                break

    async def stop_all_streams(self):
        """停止所有SSE流"""
        for client_id in list(self.active_connections):
            await self.stop_stream(client_id)

    async def _cleanup(self):
        """清理所有资源"""
        await self.stop_all_streams()
        while not self.conn_pool.empty():
            session = await self.conn_pool.get()
            await session.close()


class Evaluation:
    URL_PREFIX = "/evaluations"

    URL_ENDPOINTS = {
        "execute": "",
    }

    def __init__(self, sdk: RCABenchSDK):
        self.sdk = sdk

    def execute(self, params: Dict):
        """执行算法评估"""
        url = f"{self.URL_PREFIX}{self.URL_ENDPOINTS['execute']}"
        return self.sdk.client.get(url, params=params)


class Injection:
    URL_PREFIX = "/injections"

    URL_ENDPOINTS = {
        "get_namespace_pod_info": "/namespace_pods",
        "get_parameters": "/parameters",
        "list": "",
        "query": "/{task_id}",
        "submit": "",
    }

    def __init__(self, sdk: RCABenchSDK):
        self.sdk = sdk

    def get_namespace_pod_info(self):
        url = f"{self.URL_PREFIX}{self.URL_ENDPOINTS['get_namespace_pod_info']}"
        return self.sdk.client.get(url)

    def get_parameters(self):
        url = f"{self.URL_PREFIX}{self.URL_ENDPOINTS['/parameters']}"
        return self.sdk.client.get(url)

    def list(self, page_num: int, page_size: int):
        url = f"{self.URL_PREFIX}{self.URL_ENDPOINTS['list']}"
        params = {"page_num": page_num, "page_size": page_size}
        return self.sdk.client.get(url, params=params)

    def query(self, task_id: str):
        endpoint = self.URL_ENDPOINTS["query"].format(task_id=task_id)
        url = f"{self.URL_PREFIX}{endpoint}"
        return self.sdk.client.get(url)

    def submit(self, payloads: Dict[str, Union[bool, int, Dict]]):
        url = f"{self.URL_PREFIX}{self.URL_ENDPOINTS['submit']}"
        return self.sdk.client.post(url, payloads)
