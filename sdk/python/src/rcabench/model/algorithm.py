from typing import List, Optional
from pydantic import BaseModel, Field


class ListResult(BaseModel):
    """
    查询算法结果

    Attributes:
        algorithms: 算法条目列表
    """

    algorithms: List[str] = Field(
        ...,
        description="List of algorithms",
        json_schema_extra={"example": ["e-diagnose"]},
    )


class SubmitExecutionItem(BaseModel):
    """
    算法执行任务配置

    Attributes:
        algorithm: 算法名称
        dataset: 数据集名称
        tag: 镜像 tag（如果为空的话，服务器会选择 harbor 中最新的）
    """

    algorithm: str = Field(
        ...,
        description="The name of algorithm",
        json_schema_extra={"example": "e-diagnose"},
    )

    dataset: str = Field(
        ...,
        description="The name of dataset",
        json_schema_extra={"example": "ts-ts-preserve-service-cpu-exhaustion-znzxcn"},
    )

    tag: Optional[str] = Field(
        None,
        description="The tag of algorithm image in harbor. If tag is none, the server will get the latest one.",
        json_schema_extra={"example": "latest"},
    )


class SubmitReq(BaseModel):
    """
    算法执行请求参数
    """

    payload: List[SubmitExecutionItem] = Field(
        ...,
        description="Configuration list",
    )
