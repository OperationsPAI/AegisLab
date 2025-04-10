from enum import Enum


INJECTION_CONF_MODES = {"display", "engine"}


TIME_FORMAT = "%Y-%m-%dT%H:%M:%S+08:00"
TIME_EXAMPLE = "1970-01-01T00:00:00+08:00"


class InjectionStatusEnum(str, Enum):
    INITIAL = "initial"
    INJECT_SUCCESS = "inject_success"
    INJECT_FAILED = "inject_failed"
    BUILD_SUCCESS = "build_success"
    BUILD_FAILED = "build_failed"
    DELETED = "deleted"


class EventType:
    END = "end"
    UPDATE = "update"


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


class Evaluation:
    ALLOWED_RANKS = {1, 3, 5}
