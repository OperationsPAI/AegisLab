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
    EventAlgoRunSucceed = "algorithm.run.succeed"

    EventAlgoResultCollection = "algorithm.collect_result"
    EventDatasetResultCollection = "dataset.result.collection"
    EventDatasetNoAnomaly = "dataset.no_anomaly"
    EventDatasetNoConclusionFile = "dataset.no_conclusion_file"
    EventDatasetBuildSucceed = "dataset.build.succeed"

    EventTaskStatusUpdate = "task.status.update"
    EventTaskRetryStatus = "task.retry.status"
    EventTaskStarted = "task.started"

    EventNoNamespaceAvailable = "no.namespace.available"
    EventRestartServiceStarted = "restart.service.started"
    EventRestartServiceCompleted = "restart.service.completed"
    EventRestartServiceFailed = "restart.service.failed"

    EventFaultInjectionStarted = "fault.injection.started"
    EventFaultInjectionCompleted = "fault.injection.completed"
    EventFaultInjectionFailed = "fault.injection.failed"

    EventAcquireLock = "acquire.lock"
    EventReleaseLock = "release.lock"


class TaskType(str, Enum):
    DUMMY = ""
    BUILD_DATASET = "BuildDataset"
    COLLECT_RESULT = "CollectResult"
    FAULT_INJECTION = "FaultInjection"
    RESTART_SERVICE = "RestartService"
    RUN_ALGORITHM = "RunAlgorithm"
