import json
from typing import Any
from uuid import UUID

from pydantic import BaseModel, ConfigDict, Field, field_validator

from ..const import EventType, TaskStatus, TaskType


class AlgorithmItem(BaseModel):
    """Algorithm item model"""

    model_config = ConfigDict(extra="ignore", validate_by_name=True)  # type: ignore

    name: str = Field(..., description="Algorithm name")
    image: str = Field(..., description="Algorithm image")
    tag: str = Field(..., description="Algorithm image tag")


class DatasetOptions(BaseModel):
    """Dataset options model"""

    model_config = ConfigDict(extra="ignore", validate_by_name=True)  # type: ignore

    dataset: str = Field(..., description="Dataset name")


class DetectorRecord(BaseModel):
    """Detector record model"""

    model_config = ConfigDict(extra="ignore", validate_by_name=True)  # type: ignore

    span_name: str = Field(..., alias="SpanName", description="Span name")
    issues: dict[str, Any] = Field(..., description="Issues detected")
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

    model_config = ConfigDict(extra="ignore", validate_by_name=True)  # type: ignore

    algorithm: AlgorithmItem = Field(..., description="Algorithm item")
    dataset: str = Field(..., description="dataset")
    execution_id: int = Field(..., description="Execution ID")


class InfoPayload(BaseModel):
    """Info payload model"""

    model_config = ConfigDict(extra="ignore", validate_by_name=True)  # type: ignore

    status: TaskStatus = Field(..., description="Status of the task")
    message: str = Field(..., alias="msg", description="Message associated with the task status")


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
