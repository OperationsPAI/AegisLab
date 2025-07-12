from typing import Any
from ..const import EventType, TaskStatus, TaskType
from ..openapi import DtoAlgorithmItem
from pydantic import BaseModel, ConfigDict, Field, field_validator
from uuid import UUID
import json


class GetTraceEventsReq(BaseModel):
    trace_id: UUID = Field(
        ...,
        description="Unique identifier for the entire trace",
        json_schema_extra={"example": UUID("75430787-c19a-4f90-8c1f-07d215a664b7")},
    )

    last_event_id: str = Field(
        default="0",
        description="",
        json_schema_extra={"example": "0"},
    )

    timeout: float | None = Field(
        None,
        description="",
        json_schema_extra={
            "example": [None, 60.0],
        },
    )


class AlgorithmItem(BaseModel):
    """Algorithm item model"""

    name: str = Field(..., description="Algorithm name")
    image: str = Field(..., description="Algorithm image")
    tag: str = Field(..., description="Algorithm image tag")


class DatasetOptions(BaseModel):
    """Dataset options model"""

    dataset: str = Field(..., description="Dataset name")


class DetectorRecord(BaseModel):
    """Detector record model"""

    model_config = ConfigDict(extra="ignore", validate_by_name=True)

    span_name: str = Field(..., alias="SpanName", description="Span name")
    issues: dict[str, Any] = Field(
        ..., default_factory=dict, description="Issues detected"
    )
    abnormal_avg_duration: float = Field(..., alias="AbnormalAvgDuration")
    normal_avg_duration: float = Field(..., alias="NormalAvgDuration")
    abnormal_succ_rate: float = Field(..., alias="AbnormalSuccRate")
    normal_succ_rate: float = Field(..., alias="NormalSuccRate")
    abnormal_p99: float = Field(..., alias="AbnormalP99")
    normal_p99: float = Field(..., alias="NormalP99")

    @field_validator("issues", mode="before")
    def parse_issues(cls, v):
        if isinstance(v, str):
            try:
                return json.loads(v)
            except json.JSONDecodeError:
                return {}

        return v if isinstance(v, dict) else {}


class ExecutionOptions(BaseModel):
    """Execution options model"""

    algorithm: DtoAlgorithmItem = Field(..., description="Algorithm name")
    dataset: str = Field(..., description="Dataset name")
    execution_id: int = Field(..., description="Execution ID", ge=1)


class InfoPayload(BaseModel):
    """Info payload model"""

    model_config = ConfigDict(extra="ignore", validate_by_name=True)

    status: TaskStatus = Field(..., description="Status of the task")
    message: str = Field(
        ..., alias="msg", description="Message associated with the task status"
    )


class StreamEvent(BaseModel):
    """
    StreamEvent data model for the event stream.

    Attributes:
        task_id (UUID): A unique identifier for the task, used to associate with a specific task instance.
            Example: "005f94a9-f9a2-4e50-ad89-61e05c1c15a0"

        task_type (TaskType): An enum value indicating the category of the task associated with the event.
            Possible values:
            - BuildDataset: Task for building a dataset.
            - BuildImage: Task for building a Docker image.
            - CollectResult: Task for collecting results.
            - FaultInjection: Task for fault injection.
            - RestartService: Task for restarting a service.
            - RunAlgorithm: Task for running an algorithm.

        event_name (EventType): An enum value indicating the nature or type of the operation or status change.

        payload (Any, optional): Additional data associated with the event. The content varies depending on the event type.
            - For error events: Contains error details and stack trace information.
            - For completion events: May contain execution result data.
    """

    task_id: UUID = Field(
        ...,
        description="Unique identifier for the task which injection belongs to",
        json_schema_extra={"example": "005f94a9-f9a2-4e50-ad89-61e05c1c15a0"},
    )

    task_type: TaskType = Field(
        ...,
        description="TaskType value:BuildDatset, CollectResult, FaultInjection, RestartService, RunAlgorithm",
        json_schema_extra={"example": ["BuildDataset"]},
    )

    event_name: EventType = Field(
        ...,
        description="Type of event being reported in the stream. Indicates the nature of the operation or status change.",
        json_schema_extra={"example": ["task.start"]},
    )

    payload: Any | None = Field(
        None,
        description="Additional data associated with the event. Content varies based on event_name",
    )


class TraceEvents(BaseModel):
    """
    TraceEvents 跟踪事件集合模型

    Attributes:
        last_event_id (str): 返回的事件集合中最后一个事件的ID
            - 格式通常为 "{timestamp}-{sequence}"，来自Redis Stream的消息ID体系
            - 用于后续请求的分页标记，实现增量获取事件
            - 客户端应在下次请求中将此值作为起始ID传递，以获取更新的事件

        events (List[StreamEvent]): 事件列表，按时间顺序记录链路中的各个状态变更和操作
            - 列表中的每个元素为一个StreamEvent对象
            - 通常按时间先后顺序排列，从链路开始到结束
            - 包含任务生命周期中的所有关键状态变更和操作记录
    """

    last_event_id: str = Field(
        ...,
        description="ID of the last event in the returned collection, used for pagination and incremental event retrieval. ",
    )

    events: list[StreamEvent] = Field(
        ...,
        description="Ordered list of events associated with a task trace, capturing the complete execution history from start to finish",
    )
