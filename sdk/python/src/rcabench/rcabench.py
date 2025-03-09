from typing import Any, AsyncGenerator, Dict, List, Optional
from .client import AsyncSSEClient, TaskManager
from contextlib import asynccontextmanager
from .entity import (
    TaskResponse,
    AlgorithmResp,
    EvaluationResp,
    NamespacePodInfo,
    InjectionParameters,
)
import aiohttp
import asyncio
import requests

CLIENT_NAME = "SSE-{task_id}"


class BaseRouter:
    URL_PREFIX = ""

    def __init__(self, sdk):
        self.sdk = sdk

    def _build_url(self, endpoint: str) -> str:
        return f"{self.URL_PREFIX}{endpoint}"


class Algorithm(BaseRouter):
    URL_PREFIX = "/algorithms"

    URL_ENDPOINTS = {
        "execute": "",
        "list": "/list",
    }

    def execute(self, payload: List[Dict]) -> TaskResponse:
        url = self._build_url(self.URL_ENDPOINTS["execute"])
        return self.sdk._post(url, payload)["data"]

    def list(self) -> AlgorithmResp:
        """
        Retrieve available benchmarks and algorithms.
        """
        url = self._build_url(self.URL_ENDPOINTS["list"])
        data = self.sdk._get(url)["data"]
        return AlgorithmResp(algorithms=data["algorithms"])


class Evaluation(BaseRouter):
    URL_PREFIX = "/evaluations"

    URL_ENDPOINTS = {
        "execute": "",
    }

    def execute(self, payload: Dict) -> EvaluationResp:
        url = self._build_url(self.URL_ENDPOINTS["execute"])
        data = self.sdk._get(url, params=payload)["data"]
        return EvaluationResp(results=data["results"])


class Injection(BaseRouter):
    URL_PREFIX = "/injections"

    URL_ENDPOINTS = {
        "execute": "",
        "get_namespace_pod_info": "/namespace_pods",
        "get_parameters": "/parameters",
    }

    def execute(self, payload: List[Dict]) -> TaskResponse:
        url = self._build_url(self.URL_ENDPOINTS["execute"])
        return self.sdk._post(url, payload)["data"]

    def get_namespace_pod_info(self) -> NamespacePodInfo:
        url = self._build_url(self.URL_ENDPOINTS["get_namespace_pod_info"])
        data = self.sdk._get(url)["data"]
        return NamespacePodInfo(namespace_info=data["namespace_info"])

    def get_parameters(self) -> InjectionParameters:
        url = self._build_url(self.URL_ENDPOINTS["get_parameters"])
        data = self.sdk._get(url)["data"]
        return InjectionParameters(
            specification=data["specification"], keymap=data["keymap"]
        )


class RCABenchSDK:
    def __init__(self, base_url: str, max_connections: int = 10):
        """
        Initialize the SDK with the base URL of the server.

        :param base_url: Base URL of the RCABench server, e.g., "http://localhost:8080"
        """
        self.base_url = base_url.rstrip("/") + "/api/v1"
        self.algorithm = Algorithm(self)
        self.evaluation = Evaluation(self)
        self.injection = Injection(self)

        self.task_manager = TaskManager()
        self.conn_pool = asyncio.Queue(max_connections)
        self.active_connections = set()
        self.loop = asyncio.get_event_loop()

    def _get(
        self, url: str, params: Optional[Dict] = None, stream: bool = False
    ) -> Any:
        url = f"{self.base_url}{url}"
        if not stream:
            response = requests.get(url, params=params)
            response.raise_for_status()
            return response.json()
        return requests.get(url, params=params, stream=True)

    def _post(self, url: str, payload: List[Dict]) -> requests.Response:
        url = f"{self.base_url}{url}"
        response = requests.post(url, json=payload)
        response.raise_for_status()
        return response.json()

    @asynccontextmanager
    async def _get_session(self) -> AsyncGenerator[aiohttp.ClientSession, None]:
        session = await self.conn_pool.get()
        try:
            yield session
        finally:
            await self.conn_pool.put(session)

    async def _create_sse_client(self, task_id: str, url: str) -> AsyncSSEClient:
        url = f"{self.base_url}{url}"
        return AsyncSSEClient(self.task_manager, task_id, url)

    async def _stream_task(self, task_id: str, url: str) -> None:
        retries = 0
        max_retries = 3

        sse_client = await self._create_sse_client(task_id, url)
        self.active_connections.add(task_id)

        while retries < max_retries:
            try:
                await sse_client.connect()
                break
            except aiohttp.ClientError:
                retries += 1
                await asyncio.sleep(2**retries)

        self.active_connections.discard(task_id)

    async def start_stream(self, task_id: str, url: str) -> None:
        """启动一个独立的SSE流"""
        asyncio.create_task(
            self._stream_task(task_id, url.format(task_id=task_id)),
            name=CLIENT_NAME.format(task_id=task_id),
        )

    async def start_multiple_stream(
        self, task_ids: List[str], url: str, timeout: Optional[float] = None
    ) -> Dict[str, Any]:
        """批量启动多个SSE流"""
        for task_id in task_ids:
            await self.start_stream(task_id, url.format(task_id=task_id))

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
