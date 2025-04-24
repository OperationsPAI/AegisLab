from typing import List
from ..const import Pagination
from pydantic import BaseModel, ConfigDict, Field, field_validator
from uuid import UUID


class PaginationReq(BaseModel):
    """
    分页请求参数模型

    Attributes:
        page_num (int): 当前页码，必须大于0
        page_size (int): 每页数据量，必须在允许的范围内
    """

    page_num: int = Field(
        ...,
        description="",
        json_schema_extra={"example": 1},
        gt=0,
    )

    page_size: int = Field(
        ...,
        description="",
        json_schema_extra={"example": 1},
    )

    @field_validator("page_size")
    @classmethod
    def validate_page_size(cls, value: int) -> int:
        if value not in Pagination.ALLOWED_PAGE_SIZES:
            raise ValueError(
                f"Page size must be one of {Pagination.ALLOWED_PAGE_SIZES}"
            )

        return value


class TraceInfo(BaseModel):
    """
    单条任务追踪链元数据
    """

    model_config = ConfigDict(arbitrary_types_allowed=True)

    head_task_id: UUID = Field(
        ...,
        description="Head task UUID in the trace chain",
        json_schema_extra={"example": UUID("da1d9598-3a08-4456-bfce-04da8cf850b0")},
    )

    trace_id: UUID = Field(
        ...,
        description="Unique identifier for the entire trace",
        json_schema_extra={"example": UUID("75430787-c19a-4f90-8c1f-07d215a664b7")},
    )


class SubmitResult(BaseModel):
    """
    任务提交结果
    """

    model_config = ConfigDict(arbitrary_types_allowed=True)

    group_id: UUID = Field(
        ...,
        description="Batch task group identifier",
        json_schema_extra={"example": UUID("e7cbb5b8-554e-4c82-a018-67f626fc12c6")},
    )

    traces: List[TraceInfo] = Field(
        ...,
        description="List of trace information objects",
    )
