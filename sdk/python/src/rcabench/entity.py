from typing import List, Dict
from dataclasses import dataclass


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
class Dataset:
    id: int
    name: str


@dataclass
class DatasetResp:
    total: int
    datasets: List[Dataset]


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
class WithdrawResponse:
    message: str


@dataclass
class RunAlgorithmPayload:
    algorithm: str
    benchmark: str
    dataset: str


@dataclass
class SubmitResp:
    group_id: str
    task_ids: List[str]
