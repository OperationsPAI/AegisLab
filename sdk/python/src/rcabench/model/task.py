from typing import List
from pydantic import BaseModel, Field
from uuid import UUID


class TraceInfo(BaseModel):
    """
    单条任务追踪链元数据
    """

    head_task_id: UUID = Field(
        ...,
        description="Head task UUID in the trace chain",
        example="da1d9598-3a08-4456-bfce-04da8cf850b0",
    )
    trace_id: UUID = Field(
        ...,
        description="Unique identifier for the entire trace",
        example="75430787-c19a-4f90-8c1f-07d215a664b7",
    )


class SubmitResult(BaseModel):
    """
    任务提交结果
    """

    group_id: UUID = Field(
        ...,
        description="Batch task group identifier",
        example="e7cbb5b8-554e-4c82-a018-67f626fc12c6",
    )
    # 至少包含一个追踪链
    traces: List[TraceInfo] = Field(
        ...,
        description="List of trace information objects",
        min_items=1,
    )
