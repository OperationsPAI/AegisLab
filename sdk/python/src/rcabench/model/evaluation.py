from typing import List, Optional
from ..const import Evaluation
from pydantic import BaseModel, Field, field_validator


class EvaluationReq(BaseModel):
    """
    评估请求

    Attributes:
        execution_ids: 需要评估的执行ID列表
        metrics: 可选，要计算的评估指标
        rank: 可选，要计算的排名靠前的结果数量
    """

    execution_ids: List[int] = Field(
        ...,
        description="List of execution IDs to be evaluated",
        json_schema_extra={
            "example": [311],
        },
    )

    metrics: Optional[List[str]] = Field(
        None,
        description="Metrics to be calculated in the evaluation",
        json_schema_extra={
            "example": ["accuracy"],
        },
    )

    rank: Optional[int] = Field(
        None,
        description="Number of top results to be calculated",
        json_schema_extra={
            "example": 5,
        },
    )

    @field_validator("execution_ids")
    @classmethod
    def validate_execution_ids(cls, value: List[int]) -> List[int]:
        for id in value:
            if id <= 0:
                raise ValueError("Execution IDs must be positive integers")

        return value

    @field_validator("rank")
    @classmethod
    def validate_rank(cls, value: int) -> int:
        if value is not None and value not in Evaluation.ALLOWED_RANKS:
            raise ValueError(
                f"Invalid rank value. Must be one of: {', '.join(str(r) for r in Evaluation.ALLOWED_RANKS)}"
            )

        return value
