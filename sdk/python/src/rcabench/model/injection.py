from typing import Any, Dict, List, Optional
from ..const import InjectionStatusEnum
from datetime import datetime
from pydantic import BaseModel, Field, model_validator


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
        default=1, gt=1, description="Unique identifier for the injection", examples="1"
    )

    task_id: str = Field(
        ...,
        description="Unique identifier for the task which injection belongs to",
        examples="005f94a9-f9a2-4e50-ad89-61e05c1c15a0",
        max_length=64,
    )

    fault_type: str = Field(
        ...,
        description="Type of injected fault",
        examples="CPUStress",
    )

    spec: Dict[str, Any] = Field(
        ..., description="Specification parameters for the fault injection"
    )

    status: InjectionStatusEnum = Field(
        ...,
        description="Status value:initial, inject_success, inject_failed, build_success, build_failed, deleted",
        examples=["initial", "deleted"],
    )

    start_time: datetime = Field(
        ...,
        description="Start timestamp of injection window",
        examples="2025-03-23T12:05:42+08:00",
    )

    end_time: datetime = Field(
        ...,
        description="End timestamp of injection window",
        examples="2025-03-23T12:06:42+08:00",
    )


class ListResult(BaseModel):
    """
    分页查询结果

    Attributes:
        total: 注入任务总数
        injections: 注入任务列表
    """

    total: int = Field(
        default=0,
        ge=0,
        description="Total number of injections",
        examples=20,
    )

    injections: List[InjectionItem] = Field(
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
        examples=["CPUStress", "PodChaos"],
    )

    range: Optional[List[int]] = Field(
        None,
        description="Allowed value range [min, max] for validation",
        examples=[[1, 60], [1, 100]],
    )

    description: Optional[str] = Field(
        None,
        min_length=3,
        description="Human-readable explanation of the node's purpose",
    )

    children: Optional[Dict[str, "SpecNode"]] = Field(
        None,
        description="Child nodes forming a hierarchical structure",
        examples=[{"1": {"name": "sub_node", "value": 42}}],
    )

    value: Optional[int] = Field(
        None,
        description="Numerical value for this configuration node",
        examples=[1, 23, 60],
    )

    @model_validator(mode="after")
    def check_required_fields(self):
        # 验证至少存在 value 或 children
        if self.value is None and self.children is None:
            raise ValueError("A node must contain at least one of value or children")

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
        benchmark: 基准测试名称
        interval: 故障注入间隔时间（分钟）
        pre_duration: 故障注入前的正常观测时间（分钟）
        specs: 分层配置参数树
    """

    benchmark: str = Field(
        ...,
        description="Benchmark name",
        examples=["clickhouse"],
    )

    interval: int = Field(
        ...,
        gt=0,
        description="Fault injection interval (minute)",
        examples=[2],
    )

    pre_duration: int = Field(
        ...,
        gt=0,
        description="Normal time before fault injection (minute)",
        examples=[1],
    )

    specs: List[SpecNode] = Field(
        ...,
        min_length=1,
        description="Hierarchical configuration parameter tree, each element represents a parameter branch",
    )
