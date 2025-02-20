import inspect
import requests
from dataclasses import dataclass
from typing import Any, List, Dict, Optional


@dataclass
class TaskResponse:
    taskID: str
    message: str


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
class AlgoBenchResponse:
    benchmarks: List[str]
    algorithms: List[str]


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


class RCABenchSDK:
    URL_DICT = {
        "get_algorithms": "{}/algorithms",
        "inject": "{}/injections",
        "get_injection_parameters": "{}/injections/parameters",
        "get_injection_namespace_pod_info": "{}/injections/namespace_pods",
    }

    def __init__(self, base_url: str):
        """
        Initialize the SDK with the base URL of the server.

        :param base_url: Base URL of the RCABench server, e.g., "http://localhost:8080"
        """
        self.base_url = base_url.rstrip("/") + "/api/v1"

    @staticmethod
    def _get_parent_function_name():
        stack = inspect.stack()
        parent_frame = stack[2]
        return parent_frame.function

    def _get(self, task_id: Optional[str] = None) -> Any:
        url = self.URL_DICT[self._get_parent_function_name()].format(self.base_url)
        if task_id:
            url = url.format(task_id)
        response = requests.get(url)
        response.raise_for_status()
        return response.json()

    def _post(self, payload: List[Dict]) -> requests.Response:
        url = self.URL_DICT[self._get_parent_function_name()].format(self.base_url)
        json_payload = [item for item in payload]
        return requests.post(url, json=json_payload)

    def get_algorithms(self) -> AlgoBenchResponse:
        """
        Retrieve available benchmarks and algorithms.

        :return: AlgoBenchResponse object
        """
        data = self._get()["data"]
        return AlgoBenchResponse(algorithms=data["algorithms"])

    def inject(self, payload: List[Dict]) -> TaskResponse:
        return self._post(payload).json()

    def get_injection_parameters(self) -> InjectionParameters:
        data = self._get()["data"]
        return InjectionParameters(
            specification=data["specification"], keymap=data["keymap"]
        )

    def get_injection_namespace_pod_info(self) -> NamespacePodInfo:
        data = self._get()["data"]
        return NamespacePodInfo(namespace_info=data["namespace_info"])
