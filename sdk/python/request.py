import requests
from dataclasses import dataclass
from typing import List, Dict, Optional


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


class RCABenchSDK:
    def __init__(self, base_url: str):
        """
        Initialize the SDK with the base URL of the server.

        :param base_url: Base URL of the RCABench server, e.g., "http://localhost:8080"
        """
        self.base_url = base_url.rstrip("/")

    def submit_task(self, task_type: str, payload: Dict) -> TaskResponse:
        """
        Submit a task to the server.

        :param task_type: Type of the task (e.g., "FaultInjection", "RunAlgorithm")
        :param payload: Task-specific payload
        :return: TaskResponse object
        """
        url = f"{self.base_url}/tasks?type={task_type}"
        response = requests.post(url, json=payload)
        response.raise_for_status()
        data = response.json()
        return TaskResponse(taskID=data["taskID"], message=data["message"])

    def get_task_status(self, task_id: str) -> TaskStatus:
        """
        Get the status of a task.

        :param task_id: ID of the task
        :return: TaskStatus object
        """
        url = f"{self.base_url}/tasks/{task_id}/status"
        response = requests.get(url)
        response.raise_for_status()
        data = response.json()
        return TaskStatus(
            taskID=data["taskID"], status=data["status"], logs=data["logs"]
        )

    def get_task_details(self, task_id: str) -> TaskDetails:
        """
        Get the details of a specific task.

        :param task_id: ID of the task
        :return: TaskDetails object
        """
        url = f"{self.base_url}/tasks/{task_id}/details"
        response = requests.get(url)
        response.raise_for_status()
        data = response.json()
        return TaskDetails(
            id=data["id"],
            type=data["type"],
            payload=data["payload"],
            status=data["status"],
        )

    def get_algo_bench(self) -> AlgoBenchResponse:
        """
        Retrieve available benchmarks and algorithms.

        :return: AlgoBenchResponse object
        """
        url = f"{self.base_url}/algo_bench"
        response = requests.get(url)
        response.raise_for_status()
        data = response.json()
        return AlgoBenchResponse(
            benchmarks=data["benchmarks"], algorithms=data["algorithms"]
        )

    def get_injection_parameters(self) -> InjectionParameters:
        """
        Retrieve chaos injection parameters.

        :return: InjectionParameters object
        """
        url = f"{self.base_url}/injection_parameters"
        response = requests.get(url)
        response.raise_for_status()
        data = response.json()
        return InjectionParameters(
            specification=data["specification"], keymap=data["keymap"]
        )

    def get_datasets(self) -> DatasetResponse:
        """
        Retrieve available datasets.

        :return: DatasetResponse object
        """
        url = f"{self.base_url}/datasets"
        response = requests.get(url)
        response.raise_for_status()
        data = response.json()
        return DatasetResponse(datasets=data["datasets"])

    def get_namespace_pod(self) -> NamespacePodInfo:
        """
        Retrieve namespace and pod information.

        :return: NamespacePodInfo object
        """
        url = f"{self.base_url}/namespace_pod"
        response = requests.get(url)
        response.raise_for_status()
        data = response.json()
        return NamespacePodInfo(namespace_info=data["namespace_info"])

    def withdraw_task(self, task_id: str) -> WithdrawResponse:
        """
        Withdraw a task by its ID.

        :param task_id: ID of the task to withdraw
        :return: WithdrawResponse object
        """
        url = f"{self.base_url}/tasks/{task_id}/withdraw"
        response = requests.delete(url)
        response.raise_for_status()
        data = response.json()
        return WithdrawResponse(message=data["message"])
