from typing import Any, Dict, List, Optional
from ..const import TIME_EXAMPLE
from pydantic import BaseModel, Field, field_validator
from uuid import UUID
import os
import zipfile


class DeleteReq(BaseModel):
    """
    数据集删除请求

    Attributes:
        names: 待删除的数据集名称列表
    """

    names: List[str] = Field(
        ...,
        description="List of datasets preparing for being deleted",
        json_schema_extra={"example": ["ts-ts-preserve-service-cpu-exhaustion-znzxcn"]},
    )

    @field_validator("names")
    @classmethod
    def validate_names(cls, value: List[str]) -> List[str]:
        for name in value:
            if not name:
                raise ValueError("Dataset name cannot be empty string")
            if len(name) < 1 or len(name) > 64:
                raise ValueError(
                    f"The length of dataset must be in range [1-64]: {name}"
                )

        return value


class DeleteResult(BaseModel):
    """
    数据集批量删除操作结果

    Attributes:
        success_count: 成功删除的数据集数量
        failed_names: 删除失败的数据集名称列表
    """

    success_count: int = Field(
        default=0,
        description="Number of successfully deleted datasets",
        json_schema_extra={"exampmle": 2},
        ge=0,
    )

    failed_names: List[str] = Field(
        default_factory=list,
        description="List of dataset names that failed to delete",
        json_schema_extra={"example": ["ts-ts-preserve-service-cpu-exhaustion-znzxcn"]},
    )


class DownloadReq(BaseModel):
    """
    文件下载请求

    用于定义批量下载任务组所需参数，包含任务组ID列表和输出路径校验。

    Attributes:
        group_ids: 需要下载的任务组ID列表
        output_path: 文件保存的目标路径（需可写权限）
    """

    group_ids: List[UUID] = Field(
        ...,
        description="List of task groups",
        json_schema_extra={"example": [UUID("550e8400-e29b-41d4-a716-446655440000")]},
    )

    output_path: str = Field(
        ...,
        description="The path to save package.zip",
        json_schema_extra={"example": os.getcwd()},
    )

    @field_validator("output_path")
    @classmethod
    def validate_output_path(cls, value: str) -> str:
        if not os.path.isdir(value):
            raise FileNotFoundError(f"Output path does not exist: {value}")

        if not os.access(value, os.W_OK):
            raise PermissionError(f"No write permission on path: {value}")

        return value


class DownloadResult(BaseModel):
    """
    文件下载结果

    Attributes:
        file_path: 文件保存路径
    """

    file_path: str = Field(
        ...,
        description="The saving path",
        json_schema_extra={"example": os.path.join(os.getcwd(), "package.zip")},
    )

    @field_validator("file_path")
    @classmethod
    def validate(cls, value: str) -> str:
        if not os.path.exists(value):
            raise FileNotFoundError(f"File does not exist: {value}")
        if not os.path.isfile(value):
            raise ValueError(f"Path is not a regular file: {value}")
        if os.path.getsize(value) == 0:
            raise ValueError(f"File is empty: {value}")
        if not value.lower().endswith(".zip"):
            raise ValueError(f"File is not a ZIP archive: {value}")

        try:
            with zipfile.ZipFile(value, "r") as zip_ref:
                # 查看 ZIP 是否包含文件
                if not zip_ref.namelist():
                    raise ValueError(f"ZIP archive contains no files: {value}")

                # ZIP 内所有文件是否可读
                corrupt_file = zip_ref.testzip()
                if corrupt_file is not None:
                    raise ValueError(
                        f"ZIP archive contains corrupt file: {corrupt_file}"
                    )

        except zipfile.BadZipFile:
            raise ValueError(f"Invalid ZIP file format: {value}")

        return value


class DatasetItem(BaseModel):
    """
    数据集元数据信息

    Attributes:
        name: 数据集的唯一标识符
        param: 数据集的配置参数
        start_time: 注入窗口的开始时间戳
        end_time: 注入窗口的结束时间戳
    """

    name: str = Field(
        ...,
        description="Unique identifier for the dataset",
        json_schema_extra={"example": "ts-ts-preserve-service-cpu-exhaustion-znzxcn"},
        max_length=64,
    )

    param: Dict[str, Any] = Field(
        ...,
        description="Configuration parameters for dataset",
    )

    start_time: str = Field(
        ...,
        description="Start timestamp of injection window",
        json_schema_extra={"example": TIME_EXAMPLE},
    )

    end_time: str = Field(
        ...,
        description="End timestamp of injection window",
        json_schema_extra={"example": TIME_EXAMPLE},
    )


class ListResult(BaseModel):
    """
    分页数据集查询结果

    Attributes:
        total: 数据集总数
        datasets: 数据集条目列表
    """

    total: int = Field(
        default=0,
        description="Total number of datasets",
        json_schema_extra={"example": 20},
        ge=0,
    )

    datasets: List[DatasetItem] = Field(
        default_factory=list,
        description="List of datasets",
    )


class DetectorRecord(BaseModel):
    """
    detector 算法指标记录

    Attributes:
        span_name: 存在问题的Span名称
        issue: 检测到的异常描述
        avg_duration: 平均持续时间指标（分钟）
        succ_rate: 成功率百分比（0-1范围）
        p90: 90分位延迟测量值
        p95: 95分位延迟测量值
        p99: 99分位延迟测量值
    """

    span_name: str = Field(
        ...,
        description="Identified span name with issues",
    )

    issue: str = Field(
        ...,
        description="Description of detected anomaly",
    )

    avg_duration: Optional[float] = Field(
        None,
        description="Average duration metric (seconds)",
    )

    succ_rate: Optional[float] = Field(
        None,
        description="Success rate percentage (0-1 scale)",
        ge=0,
        le=1,
    )

    p90: Optional[float] = Field(
        None,
        alias="P90",
        description="90th percentile latency measurement",
        ge=0,
        le=1,
    )

    p95: Optional[float] = Field(
        None,
        alias="P95",
        description="95th percentile latency measurement",
        ge=0,
        le=1,
    )

    p99: Optional[float] = Field(
        None,
        alias="P99",
        description="99th percentile latency measurement",
        ge=0,
        le=1,
    )


class GranularityRecord(BaseModel):
    """
    粒度分析结果记录

    Attributes:
        level: 分析粒度级别（service/pod/span/metric）
        result: 识别到的根因描述
        rank: 问题严重性排名
        confidence: 分析结果的置信度评分
    """

    level: str = Field(
        ...,
        description="Analysis granularity level (service/pod/span/metric)",
        json_schema_extra={"example": "service"},
        max_length=32,
    )

    result: str = Field(
        ...,
        description="Identified root cause description",
        json_schema_extra={"example": "ts-preserve-service"},
    )

    rank: int = Field(
        ...,
        description="Severity ranking of the issue",
        json_schema_extra={"example": 1},
        gt=0,
    )

    confidence: float = Field(
        ...,
        description="Confidence score of the analysis result",
        json_schema_extra={"example": 0.8},
        ge=0,
        le=1,
    )


class ExecutionRecord(BaseModel):
    """
    根因分析执行记录

    Attributes:
        algorithm: 根因分析算法名称
        granularity_results: 不同粒度层级的分析结果
    """

    algorithm: str = Field(
        ...,
        description="Root cause analysis algorithm name",
        json_schema_extra={"example": "e-dianose"},
    )

    granularity_records: List[GranularityRecord] = Field(
        default_factory=list,
        description="Analysis results across different granularity levels",
    )


class QueryResult(DatasetItem):
    """
    包含诊断信息的扩展数据集查询结果

    Attributes:
        detector_result: 详细的异常检测指标
        execution_results: 多算法根因分析结果集合
    """

    detector_result: DetectorRecord = Field(
        ...,
        description="Detailed anomaly detection metrics",
    )

    execution_results: List[ExecutionRecord] = Field(
        default_factory=list,
        description="Collection of root cause analysis results from multiple algorithms",
    )
