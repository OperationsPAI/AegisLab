from typing import Any, Dict, List, Optional
from ..const import INJECTION_CONF_MODES, InjectionStatus, TIME_EXAMPLE
from datetime import datetime
from pydantic import BaseModel, Field, field_validator, model_validator
from uuid import UUID


class GetConfReq(BaseModel):
    """
    配置获取请求

    Attributes:
        mode: 配置模式标识符 (display, engine)
    """

    namespace: str = Field(
        ...,
        description="Target Kubernetes namespace",
        json_schema_extra={"example": "ts"},
    )

    mode: str = Field(
        ...,
        description="Choose the config mode (display, engine)",
        json_schema_extra={"example": "display"},
    )

    @field_validator("mode")
    @classmethod
    def validate_mode(cls, value: str) -> str:
        if value not in INJECTION_CONF_MODES:
            raise ValueError(
                f"Injection conf mode must be one of {INJECTION_CONF_MODES}"
            )

        return value


class InjectionItem(BaseModel):
    """
    注入任务元数据信息

    Attributes:
        id: 注入任务的唯一标识符
        task_id: 所属父任务的ID
        fault_type: 注入的故障类型
        spec: 故障注入的具体参数配置
        status: 当前任务状态
        start_time: 注入窗口开始时间
        end_time: 注入窗口结束时间
    """

    id: int = Field(
        default=1,
        description="Unique identifier for the injection",
        json_schema_extra={"example": 1},
        gt=0,
    )

    task_id: UUID = Field(
        ...,
        description="Unique identifier for the task which injection belongs to",
        json_schema_extra={"example": "005f94a9-f9a2-4e50-ad89-61e05c1c15a0"},
    )

    fault_type: str = Field(
        ...,
        description="Type of injected fault",
        json_schema_extra={"example": "CPUStress"},
    )

    spec: Dict[str, Any] = Field(
        ...,
        description="Specification parameters for the fault injection",
    )

    status: InjectionStatus = Field(
        ...,
        description="Status value:initial, inject_success, inject_failed, build_success, build_failed, deleted",
        json_schema_extra={"example": ["initial"]},
    )

    start_time: datetime = Field(
        ...,
        description="Start timestamp of injection window",
        json_schema_extra={"example": TIME_EXAMPLE},
    )

    end_time: datetime = Field(
        ...,
        description="End timestamp of injection window",
        json_schema_extra={"example": TIME_EXAMPLE},
    )


class ListResult(BaseModel):
    """
    分页查询结果

    Attributes:
        total: 注入任务总数
        total_pages: 注入记录总页数
        items: 注入任务列表
    """

    total: int = Field(
        default=0,
        ge=0,
        description="Total number of injections",
        json_schema_extra={"example": 20},
    )

    total_pages: int = Field(
        default=0,
        ge=0,
        description="Total number of injections pages",
        json_schema_extra={"example": 20},
    )

    items: List[InjectionItem] = Field(
        default_factory=list,
        description="List of injections",
    )


class SpecNode(BaseModel):
    """
    分层配置节点结构

    Attributes:
        name: 配置节点名称标识
        range: 允许的数值范围[min, max]
        description: 配置项功能描述
        children: 子节点层级结构
        value: 当前节点的配置数值
    """

    name: Optional[str] = Field(
        None,
        min_length=1,
        description="Unique identifier for the configuration node",
        json_schema_extra={"example": "CPUStress"},
    )

    range: Optional[List[int]] = Field(
        None,
        description="Allowed value range [min, max] for validation",
        json_schema_extra={"example": [1, 60]},
    )

    description: Optional[str] = Field(
        None,
        min_length=3,
        description="Human-readable explanation of the node's purpose",
    )

    children: Optional[Dict[str, "SpecNode"]] = Field(
        None,
        description="Child nodes forming a hierarchical structure",
        json_schema_extra={"example": {"1": {"name": "sub_node", "value": 42}}},
    )

    value: Optional[int] = Field(
        None,
        description="Numerical value for this configuration node",
        json_schema_extra={"example": 1},
    )

    @model_validator(mode="after")
    def check_required_fields(self):
        # 验证值是否符合范围
        if self.range and self.value is not None:
            if len(self.range) != 2:
                raise ValueError("Range must contain exactly 2 elements")

            if self.value != 0 and not (self.range[0] <= self.value <= self.range[1]):
                raise ValueError(f"Value {self.value} out of range {self.range}")

        return self


class SubmitReq(BaseModel):
    """
    故障注入请求参数

    Attributes:
        benchmark: 基准测试名称，如果为空则不执行数据采集
        interval: 故障注入间隔时间（分钟）
        pre_duration: 故障注入前的正常观测时间（分钟）
        specs: 分层配置参数树
    """

    interval: int = Field(
        ...,
        description="Fault injection interval (minute)",
        json_schema_extra={"example": 2},
        gt=0,
    )

    pre_duration: int = Field(
        ...,
        description="Normal time before fault injection (minute)",
        json_schema_extra={"example": 1},
        gt=0,
    )

    specs: List[SpecNode] = Field(
        ...,
        description="Hierarchical configuration parameter tree, each element represents a parameter branch",
        min_length=1,
    )

    benchmark: Optional[str] = Field(
        None,
        description="Benchmark name",
        json_schema_extra={"example": "clichhouse"},
    )
