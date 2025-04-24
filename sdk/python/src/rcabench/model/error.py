from typing import List, Optional
from ..const import TIME_EXAMPLE, TIME_FORMAT
from datetime import datetime, timedelta, timezone
from pydantic import BaseModel, Field


class ModelHTTPError(BaseModel):
    """
    HTTP标准错误响应模型

    Attributes:
        status_code: HTTP状态码
        detail: 错误详情信息
        timestamp: 错误发生时间（带时区偏移）
        path: 请求路径（可选）
        method: HTTP请求方法（可选）
    """

    status_code: int = Field(
        ...,
        description="HTTP status code indicating the error type",
        json_schema_extra={"example": 404},
    )

    detail: str = Field(
        ...,
        description="Human-readable error message explaining what went wrong",
        json_schema_extra={"example": "Item not found"},
    )

    timestamp: str = Field(
        default_factory=lambda: datetime.now(timezone(timedelta(hours=8))).strftime(
            TIME_FORMAT
        ),
        description="ISO 8601 format time (with time zone offset)",
        json_schema_extra={"example": TIME_EXAMPLE},
    )

    path: Optional[str] = Field(
        None,
        description="API endpoint path that triggered the error",
        json_schema_extra={"example": "/api/items/123"},
    )

    method: Optional[str] = Field(
        None,
        description="HTTP method used for the request",
        json_schema_extra={"example": "GET"},
    )


class FieldValidationIssue(BaseModel):
    """
    单个字段验证错误详情

    Attributes:
        field: 发生错误的字段路径
        message: 具体错误描述
        error_type: 错误类型标识
    """

    field: str = Field(
        ...,
        description="Dot-path to the invalid field",
        json_schema_extra={"example": "user_id"},
    )

    message: str = Field(
        ...,
        description="Human-readable error message",
        json_schema_extra={"example": "Value is not a valid UUID"},
    )

    error_type: str = Field(
        ...,
        description="Error type identifier",
        json_schema_extra={"example": "value_error.uuid"},
    )


class ModelValidationError(BaseModel):
    """
    模型验证错误响应模型

    Attributes:
        detail: 错误信息
        issuse: 详细错误列表
        timestamp 错误发生时间（带时区偏移）
    """

    detail: str = Field(
        ...,
        description="Error Summary",
    )

    issues: List[FieldValidationIssue] = Field(
        ...,
        description="Detailed list of validation errors",
        min_length=1,
    )

    timestamp: str = Field(
        default_factory=lambda: datetime.now(timezone(timedelta(hours=8))).strftime(
            TIME_FORMAT
        ),
        description="ISO 8601 format time (with time zone offset)",
        json_schema_extra={"example": TIME_EXAMPLE},
    )
