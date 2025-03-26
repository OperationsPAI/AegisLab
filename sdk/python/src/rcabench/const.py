class EventType:
    END = "end"
    ERROR = "error"
    UPDATE = "update"


class SSEMsgPrefix:
    DATA = "data"
    EVENT = "event"


class TaskStatus:
    COMPLETED = "completed"
    ERROR = "error"


class Pagination:
    DEFAULT_PAGE_NUM = 1
    ALLOWED_PAGE_SIZES = {10, 20, 50}
    DEFAULT_PAGE_SIZE = 10
