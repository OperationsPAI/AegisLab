from enum import Enum


TIME_FORMAT = "%Y-%m-%dT%H:%M:%S+08:00"
TIME_EXAMPLE = "1970-01-01T00:00:00+08:00"


class SSEMsgPrefix:
    DATA = "data"
    EVENT = "event"


class TaskStatus(str, Enum):
    Canceled = "Canceled"
    COMPLETED = "Completed"
    ERROR = "Error"
    Pending = "Pending"
    Running = "Running"
    Rescheduled = "Rescheduled"
    Scheduled = "Scheduled"


class Task:
    CLIENT_ERROR_KEY = "Client Error"
    HTTP_ERROR_STATUS_CODE = 500


class EventType(str, Enum):
    EventAlgoRunSucceed = "algorithm.run.succeed"
    EventAlgoRunFailed = "algorithm.run.failed"

    EventAlgoResultCollection = "algorithm.collect_result"
    EventDatasetResultCollection = "dataset.result.collection"
    EventDatasetNoAnomaly = "dataset.no_anomaly"
    EventDatasetNoConclusionFile = "dataset.no_conclusion_file"
    EventDatasetBuildSucceed = "dataset.build.succeed"
    EventDatasetBuildFailed = "dataset.build.failed"

    EventTaskStatusUpdate = "task.status.update"
    EventTaskRetryStatus = "task.retry.status"
    EventTaskStarted = "task.started"

    EventImageBuildSucceed = "image.build.succeed"

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
    BUILD_IMAGE = "BuildImage"
    COLLECT_RESULT = "CollectResult"
    FAULT_INJECTION = "FaultInjection"
    RESTART_SERVICE = "RestartService"
    RUN_ALGORITHM = "RunAlgorithm"
