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
    算法服务环境变量配置的数据模型。

    用于环境变量的类型验证和设置管理。

    Attributes:
        algorithm: 算法名称配置
        service: 算法服务配置，用于指定'detector'类算法中的服务项
        workspace: 容器镜像工作目录配置
        input_path: 数据输入路径配置
        output_path: 数据输出路径配置
    """

    model_config = ConfigDict(extra="forbid")

    ALGORITHM: Optional[str] = Field(
        None,
        description="The name of algorithm",
        json_schema_extra={"example": "e-diagnose"},
    )

    SERVICE: Optional[str] = Field(
        None,
        description="The service of the algorithm 'detector'",
        json_schema_extra={"example": "ts-ts-preserve-service"},
    )

    WORKSPACE: Optional[str] = Field(
        None,
        description="The workspace of the image'",
        json_schema_extra={"example": "/app"},
        alias="WORKSPACE",
    )

    INPUT_PATH: Optional[str] = Field(
        None,
        description="The data input_path of the image'",
        json_schema_extra={
            "example": "/data/ts-ts-preserve-service-cpu-exhaustion-znzxcn"
        },
        alias="INPUT_PATH",
    )

    OUTPUT_PATH: Optional[str] = Field(
        None,
        description="The data output_path of the image'",
        json_schema_extra={
            "example": "/data/ts-ts-preserve-service-cpu-exhaustion-znzxcn"
        },
        alias="OUTPUT_PATH",
    )


class SubmitExecutionItem(BaseModel):
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

    payloads: List[SubmitExecutionItem] = Field(
        ...,
        description="Configuration list",
        min_length=1,
    )
