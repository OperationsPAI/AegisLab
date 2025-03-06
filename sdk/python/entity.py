from typing import List, Dict
from dataclasses import dataclass


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
