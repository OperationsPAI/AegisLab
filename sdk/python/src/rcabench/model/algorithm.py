from typing import List, Optional
from pydantic import BaseModel, ConfigDict, Field


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


class EnvVar(BaseModel):
    """
    算法镜像环境变量模型

    Attributes:
        ALGORITHM: 算法名称配置
        SERVICE: 算法服务配置，用于指定 detector 类算法中的服务项
        VENV: 算法启动的 python 虚拟环境
    """

    model_config = ConfigDict(extra="forbid")

    ALGORITHM: Optional[str] = Field(
        None,
        description="The name of algorithm",
        json_schema_extra={"example": "e-diagnose"},
    )

    SERVICE: Optional[str] = Field(
        None,
        description="The service of the algorithm detector",
        json_schema_extra={"example": "ts-ts-preserve-service"},
    )

    VENV: Optional[str] = Field(
        None,
        description="The python environment of the algorithm",
        json_schema_extra={"example": "default"},
    )


class ExecutionPayload(BaseModel):
    """
    算法执行任务配置

    Attributes:
        image: 算法镜像名称
        tag: 镜像 tag（如果为空的话，服务器会选择 harbor 中最新的）
        dataset: 数据集名称
        env_vars: 环境变量
    """

    image: str = Field(
        ...,
        description="The name of algorithm image",
        json_schema_extra={"example": "e-diagnose"},
    )

    tag: Optional[str] = Field(
        None,
        description="The tag of algorithm image in harbor. If tag is none, the server will get the latest one.",
        json_schema_extra={"example": "latest"},
    )

    dataset: str = Field(
        ...,
        description="The name of dataset",
        json_schema_extra={"example": "ts-ts-preserve-service-cpu-exhaustion-znzxcn"},
    )

    env_vars: Optional[EnvVar] = Field(
        None,
        description="The enviroment vars of the image",
    )


class SubmitReq(BaseModel):
    """
    算法执行请求参数
    """

    payloads: List[ExecutionPayload] = Field(
        ...,
        description="List of payloads to execute algorithm",
        min_length=1,
    )
