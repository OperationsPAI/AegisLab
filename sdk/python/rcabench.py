from typing import Any, List, Dict, Optional
from dataclasses import dataclass
import inspect
import requests


@dataclass
class TaskResponse:
    group_id: str
    task_ids: List[str]


@dataclass
class TaskStatus:
    taskID: str
    status: str
    logs: List[str]


@dataclass
class TaskDetails:
    id: str
    type: str
    payload: str
    status: str


@dataclass
class AlgorithmResp:
    algorithms: List[str]


@dataclass
class EvaluationResp:
    results: List


@dataclass
class InjectionParameters:
    specification: Dict[str, List[Dict]]
    keymap: Dict[str, str]


@dataclass
class NamespacePodInfo:
    namespace_info: Dict[str, List[str]]


@dataclass
class DatasetResponse:
    datasets: List[str]


@dataclass
class WithdrawResponse:
    message: str


@dataclass
class RunAlgorithmPayload:
    algorithm: str
    benchmark: str
    dataset: str


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
        "get_stream": "/{task_id}/stream",
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

    def get_stream(self, task_id: str) -> List[str]:
        endpoint = self.URL_ENDPOINTS["get_stream"].format(task_id=task_id)
        url = self._build_url(endpoint)
        return self.sdk._process_stream(url)


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
        "get_stream": "/{task_id}/stream",
    }

    def execute(self, payload: List[Dict]) -> TaskResponse:
        url = self.URL_ENDPOINTS["execute"]
        return self.sdk._post(url, payload)["data"]

    def get_namespace_pod_info(self) -> NamespacePodInfo:
        url = self.URL_ENDPOINTS["get_namespace_pod_info"]
        data = self.sdk._get(url)["data"]
        return NamespacePodInfo(namespace_info=data["namespace_info"])

    def get_parameters(self) -> InjectionParameters:
        url = self.URL_ENDPOINTS["get_parameters"]
        data = self.sdk._get(url)["data"]
        return InjectionParameters(
            specification=data["specification"], keymap=data["keymap"]
        )

    def get_stream(self, task_id: str) -> List[str]:
        endpoint = self.URL_ENDPOINTS["get_stream"].format(task_id=task_id)
        url = self._build_url(endpoint)
        return self.sdk._process_stream(url, task_id)


class RCABenchSDK:
    def __init__(self, base_url: str):
        """
        Initialize the SDK with the base URL of the server.

        :param base_url: Base URL of the RCABench server, e.g., "http://localhost:8080"
        """
        self.base_url = base_url.rstrip("/") + "/api/v1"
        self.algorithm = Algorithm(self)
        self.evaluation = Evaluation(self)
        self.injection = Injection(self)

    @staticmethod
    def _get_parent_function_name():
        stack = inspect.stack()
        parent_frame = stack[2]
        return parent_frame.function

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

    def _process_stream(self, url: str) -> List[str]:
        response = self._get(url, stream=True)
        lines = []

        try:
            # 持续读取事件流
            for line in response.iter_lines():
                if line:
                    # 解码字节流（过滤心跳包的空行）
                    decoded_line = line.decode("utf-8")
                    print(decoded_line)
                    lines.append(decoded_line)

        except KeyboardInterrupt:
            print("Manual connection termination")
        except Exception as e:
            print(f"Connection error: {str(e)}")
        finally:
            response.close()

        return lines
