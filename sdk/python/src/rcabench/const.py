from enum import Enum


INJECTION_CONF_MODES = {"display", "engine"}


TIME_FORMAT = "%Y-%m-%dT%H:%M:%S+08:00"
TIME_EXAMPLE = "1970-01-01T00:00:00+08:00"


class InjectionStatus(str, Enum):
    INITIAL = "initial"
    INJECT_SUCCESS = "inject_success"
    INJECT_FAILED = "inject_failed"
    BUILD_SUCCESS = "build_success"
    BUILD_FAILED = "build_failed"
    DELETED = "deleted"


class SSEMsgPrefix:
    DATA = "data"
    EVENT = "event"


class TaskStatus:
    COMPLETED = "Completed"
    ERROR = "Error"


class Pagination:
    DEFAULT_PAGE_NUM = 1
    ALLOWED_PAGE_SIZES = {10, 20, 50}
    DEFAULT_PAGE_SIZE = 10


class Dataset:
    DEFAULT_SORT = "desc"
    ALLOWED_SORTS = {"asc", "desc"}


class Evaluation:
    ALLOWED_RANKS = {1, 3, 5}


class Task:
    CLIENT_ERROR_KEY = "Client Error"
    HTTP_ERROR_STATUS_CODE = 500


class EventStatus(str, Enum):
    END = "end"
    UPDATE = "update"


class EventType(str, Enum):
    ALGO_RESULT_COLLECTION = "algorithm.collect_result"
    DATASET_RESULT_COLLECTION = "algorithm.dataset_result"
    TASK_STATUS_UPDATE = "task.status.update"
    TASK_STARTED = "task.started"
    NO_NAMESPACE_AVAILABLE = "no.namespace.available"
    RESTART_SERVICE_STARTED = "restart.service.started"
    RESTART_SERVICE_COMPLETED = "restart.service.completed"
    RESTART_SERVICE_FAILED = "restart.service.failed"


class TaskType(str, Enum):
    DUMMY = ""
    BUILD_DATASET = "BuildDataset"
    COLLECT_RESULT = "CollectResult"
    FAULT_INJECTION = "FaultInjection"
    RESTART_SERVICE = "RestartService"
    RUN_ALGORITHM = "RunAlgorithm"
