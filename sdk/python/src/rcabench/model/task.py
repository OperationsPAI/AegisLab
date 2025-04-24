from typing import Dict, List, Optional, Union
from .error import ModelHTTPError
from ..const import Task
from ..logger import logger
from pydantic import BaseModel, Field, field_validator
from uuid import UUID


class StreamBatchReq(BaseModel):
    """
    批量流式请求参数

    Attributes:
        task_ids: 需要监控的任务ID列表
        trace_ids: 需要监控的链路ID列表
        client_timeout: 流式连接的最大超时时间（秒）
    """

    task_ids: List[UUID] = Field(
        ...,
        description="List of task IDs to monito",
        json_schema_extra={"example": [UUID("005f94a9-f9a2-4e50-ad89-61e05c1c15a0")]},
    )

    trace_ids: List[UUID] = Field(
        ...,
        description="List of trace IDs to build connection",
        json_schema_extra={"example": [UUID("005f94a9-f9a2-4e50-ad89-61e05c1c15a0")]},
    )

    client_timeout: float = Field(
        ...,
        description="Maximum client timeout in seconds",
        json_schema_extra={"example": 30.0},
        gt=0,
    )


class StreamAllReq(StreamBatchReq):
    """
    全部流式请求参数

    Attributes:
        task_ids: 需要监控的任务ID列表
        trace_ids: 需要监控的链路ID列表
        client_timeout: 流式连接的最大超时时间
        wait_timeout: 等待全部完成的最大超时时间（秒），None表示无超时
    """

    wait_timeout: Optional[float] = Field(
        None,
        description="Maximum wait timeout in seconds (None means no timeout)",
        json_schema_extra={"example": 30.0},
        gt=0,
    )


class StreamSingleReq(BaseModel):
    """
    流式请求参数

    Attributes:
        task_id: 需要监控的任务ID
        trace_id: 需要监控的链路ID
        client_timeout: 流式连接的最大超时时间（秒）
        wait_timeout: 等待全部完成的最大超时时间（秒），None表示无超时
    """

    task_id: UUID = Field(
        ...,
        description="Task ID to monito",
        json_schema_extra={"example": [UUID("005f94a9-f9a2-4e50-ad89-61e05c1c15a0")]},
    )

    trace_id: UUID = Field(
        ...,
        description="Trace ID to build connection",
        json_schema_extra={"example": [UUID("005f94a9-f9a2-4e50-ad89-61e05c1c15a0")]},
    )

    client_timeout: float = Field(
        ...,
        description="Maximum client timeout in seconds must be greater than the interval",
        json_schema_extra={"example": 30.0},
        gt=0,
    )

    wait_timeout: Optional[float] = Field(
        None,
        description="Maximum wait timeout in seconds (None means no timeout)",
        json_schema_extra={"example": 30.0},
        gt=0,
    )


class SSEMessage(BaseModel):
    """
    SSE消息数据模型

    表示服务器发送事件(Server-Sent Events)的消息结构。

    Attributes:
        task_type: 任务类型标识
        dataset: 关联的数据集名称（可选）
        execution_id: 任务执行ID（可选）
        has_detector_result: detector算法执行是否有结果（可选）
    """

    task_type: str = Field(
        ...,
        description="Task type identifier (e.g., FaultInjection/RunAlgorithm)",
        json_schema_extra={"example": "FaultInjection"},
    )

    dataset: Optional[str] = Field(
        None,
        description="Associated dataset name",
        json_schema_extra={"example": "ts-ts-travel2-service-pod-failure-m77s56"},
    )

    execution_id: Optional[int] = Field(
        None,
        description="Task execution ID",
        json_schema_extra={"example": 311},
    )

    has_detector_result: Optional[bool] = Field(
        None,
        description="Whether the result of detector algorithm is empty",
        json_schema_extra={"example": True},
    )


class QueueDataItem(BaseModel):
    """
    队列数据项模型

    表示异步任务队列中携带的任务处理结果数据。

    Attributes:
        error: 任务错误信息（可选）
        result: 任务成功结果（可选）
    """

    error: Optional[Union[Dict[str, str], Dict[UUID, ModelHTTPError]]] = Field(
        None,
        description="A dictionary capturing errors that occurred during task processing",
        json_schema_extra={
            "example": {
                UUID("792aa5aa-2dc3-4284-a852-b48fda567dff"): {
                    Task.CLIENT_ERROR_KEY,
                    "",
                },
                UUID("7e16011f-adbd-4361-82b0-7570701153ee"): ModelHTTPError(
                    status_code=Task.HTTP_ERROR_STATUS_CODE,
                    detail="",
                ),
            },
        },
    )

    result: Optional[Dict[UUID, SSEMessage]] = Field(
        None,
        description="A dictionary of successfully processed task results",
        json_schema_extra={
            "example": {
                UUID("792aa5aa-2dc3-4284-a852-b48fda567dff"): SSEMessage(
                    task_type="fault_injection"
                )
            }
        },
    )

    @field_validator("error")
    @classmethod
    def validate_error(
        cls, value: Optional[Union[Dict[str, str], Dict[UUID, ModelHTTPError]]]
    ) -> Optional[Union[Dict[str, str], Dict[UUID, ModelHTTPError]]]:
        if value is not None:
            if len(value) != 1:
                raise ValueError("The length of error must be 1")

            key, error_data = list(value.items())[0]
            if isinstance(error_data, str):
                if key != Task.CLIENT_ERROR_KEY:
                    raise ValueError(
                        f"The client error must contain '{Task.CLIENT_ERROR_KEY}' key, "
                        f"but got: {list(error_data.keys()[0])}"
                    )

            if isinstance(error_data, ModelHTTPError):
                if error_data.status_code != Task.HTTP_ERROR_STATUS_CODE:
                    raise ValueError(
                        f"HTTP error for task {key} must have status_code {Task.HTTP_ERROR_STATUS_CODE}, "
                        f"but got: {error_data.status_code}"
                    )

        return value


class QueueItem(BaseModel):
    """
    异步队列消息项模型

    表示异步任务队列中的标准消息项结构。

    Attributes:
        client_id: 客户端唯一标识
        data: 任务处理结果数据
    """

    client_id: UUID = Field(
        ...,
        description="The unique identifier of the async client",
        json_schema_extra={"example": "FaultInjection"},
    )

    data: QueueDataItem = Field(
        ...,
        description="The processed data",
    )


class StreamResult(BaseModel):
    """
    流式处理结果

    Attributes:
        results: 已完成任务的结果字典，格式为 {链路ID: {任务ID: 消息详情}}
        errors: 失败任务的错误信息字典，格式为 {链路ID: 错误描述}
        pending: 待处理的链路ID列表
    """

    results: Dict[UUID, Dict[UUID, SSEMessage]] = Field(
        default_factory=dict,
        description="Dictionary of completed task results (nested structure)",
        json_schema_extra={
            "example": {
                UUID("12da92c5-4075-4634-8a50-61920f94ca1e"): {
                    UUID("12da92c5-4075-4634-8a50-61920f94ca1e"): {
                        "execution_id": 311,
                        "status": "Completed",
                        "task_type": "RunAlgorithm",
                    },
                }
            }
        },
    )

    errors: Dict[UUID, str] = Field(
        default_factory=dict,
        description="Dictionary of failed task errors",
        json_schema_extra={
            "example": {
                UUID("12da92c5-4075-4634-8a50-61920f94ca1e"): "Task execution timeout"
            }
        },
    )

    pending: List[UUID] = Field(
        default_factory=list,
        description="List of pending task IDs",
        json_schema_extra={"example": [UUID("12da92c5-4075-4634-8a50-61920f94ca1e")]},
    )
