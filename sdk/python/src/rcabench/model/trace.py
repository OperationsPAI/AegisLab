from typing import Any, List, Optional
from ..const import EventType, TaskType
from pydantic import BaseModel, Field
from uuid import UUID


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

    timeout: Optional[float] = Field(
        None,
        description="",
        json_schema_extra={
            "example": [None, 60.0],
        },
    )


class StreamEvent(BaseModel):
    """
    StreamEvent 事件流数据模型

    Attributes:
        task_id (UUID): 任务的唯一标识符，用于关联特定任务实例
            例如: "005f94a9-f9a2-4e50-ad89-61e05c1c15a0"

        task_type (TaskType): 任务类型枚举值，指明事件相关的任务类别
            可选值:
            - BuildDataset: 构建数据集任务
            - CollectResult: 收集结果任务
            - FaultInjection: 故障注入任务
            - RestartService: 服务重启任务
            - RunAlgorithm: 运行算法任务

        event_name (EventType): 事件类型枚举值，表示事件的性质或操作类型

        payload (Any, 可选): 事件相关的附加数据，内容根据事件类型不同而变化
            - 对于错误事件: 包含错误详情和堆栈信息
            - 对于完成事件: 可能包含执行结果数据
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

    payload: Optional[Any] = Field(
        None,
        description="Additional data associated with the event. Content varies based on event_name",
    )


class TraceEvents(BaseModel):
    """
    TraceEvents 跟踪事件集合模型

    Attributes:
        events (List[StreamEvent]): 事件列表，按时间顺序记录链路中的各个状态变更和操作
            - 列表中的每个元素为一个StreamEvent对象
            - 通常按时间先后顺序排列，从链路开始到结束
            - 包含任务生命周期中的所有关键状态变更和操作记录
    """

    events: List[StreamEvent] = Field(
        ...,
        description="Ordered list of events associated with a task trace, capturing the complete execution history from start to finish",
    )
