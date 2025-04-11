from typing import Dict, List, Optional
from pydantic import BaseModel, Field
from uuid import UUID


class StreamReq(BaseModel):
    """
    流式请求参数

    Attributes:
        task_ids: 需要监控的任务ID列表
        timeout: 流式连接的最大超时时间（秒），None表示无超时
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

    timeout: Optional[float] = Field(
        None,
        description="Maximum connection timeout in seconds (None means no timeout)",
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
