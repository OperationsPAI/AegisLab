from typing import Dict, List, Optional
from datetime import datetime
from pydantic import BaseModel, Field


class DeleteResult(BaseModel):
    """
    数据集批量删除操作结果

    Attributes:
        success_count: 成功删除的数据集数量
        failed_names: 删除失败的数据集名称列表
    """

    success_count: int = Field(
        default=0,
        ge=0,
        description="Number of successfully deleted datasets",
        example=2,
    )

    failed_names: List[str] = Field(
        default_factory=list,
        description="List of dataset names that failed to delete",
        examples=["ts-ts-preserve-service-cpu-exhaustion-znzxcn"],
    )


class InjectionParam(BaseModel):
    duration: int = Field(
        default=1, ge=1, description="Duration of fault injection in minutes"
    )
    fault_type: str = Field(
        ..., description="Type of injected fault", examples="CPUStress"
    )
    namespace: str = Field(
        ...,
        description="Kubernetes namespace where injection occurred",
        examples="default",
    )
    pod: str = Field(
        ...,
        description="Target pod name for fault injection",
        examples="ts-preserve-service",
    )
    spec: Dict[str, int] = Field(
        default_factory=dict, description="Key-value pairs of fault specifications spec"
    )


class DatasetItem(BaseModel):
    name: str = Field(
        ...,
        description="Unique identifier for the dataset",
        examples="ts-ts-preserve-service-cpu-exhaustion-znzxcn",
        max_length=64,
    )
    param: InjectionParam = Field(
        ..., description="Configuration parameters used for fault injection"
    )
    preduration: int = Field(
        default=1,
        ge=1,
        description="Duration of preparing normal time in minutes",
        examples=1,
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


class DetectorRecord(BaseModel):
    span_name: str = Field(...)
    issue: str = Field(...)
    avg_duration: Optional[float] = Field(None)
    succ_rate: Optional[float] = Field(None)
    p90: Optional[float] = Field(None, ge=0, le=1, alias="P90")
    p95: Optional[float] = Field(None, ge=0, le=1, alias="P95")
    p99: Optional[float] = Field(None, ge=0, le=1, alias="P99")


class GranularityRecord(BaseModel):
    level: str = Field(
        ...,
        description="Analysis granularity level (service/pod/span/metric)",
        examples="service",
        max_length=32,
    )
    result: str = Field(
        ...,
        description="Identified root cause description",
        examples="ts-preserve-service",
    )
    rank: int = Field(
        ..., gt=0, description="Severity ranking of the issue", examples=1
    )
    confidence: float = Field(
        ...,
        ge=0,
        le=1,
        description="Confidence score of the analysis result",
        examples=0.8,
    )


class ExecutionRecord(BaseModel):
    algorithm: str = Field(
        ...,
        description="Root cause analysis algorithm name",
        examples="e-diagnose",
    )
    granularity_results: List[GranularityRecord] = Field(
        default_factory=list,
        description="Analysis results across different granularity levels",
    )


class ListResult(BaseModel):
    total: int = Field(
        default=0, ge=0, description="Total number of datasets", examples=20
    )
    datasets: List[DatasetItem] = Field(default_factory=list)


class QueryResult(DatasetItem):
    detector_result: DetectorRecord = Field(
        ..., description="Detailed anomaly detection metrics"
    )
    execution_results: List[ExecutionRecord] = Field(
        default_factory=list,
        description="Collection of root cause analysis results from multiple algorithms",
    )
