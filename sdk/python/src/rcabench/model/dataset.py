from typing import Any, Dict, List, Optional
from ..const import TIME_EXAMPLE, Dataset
from pydantic import BaseModel, ConfigDict, Field, field_validator, model_validator
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
        min_length=1,
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
        names: 需要下载的数据集列表（与 group_ids 互斥）
        output_path: 文件保存的目标路径（需可写权限）
    """

    group_ids: Optional[List[UUID]] = Field(
        None,
        description="List of task groups",
        json_schema_extra={"example": [UUID("550e8400-e29b-41d4-a716-446655440000")]},
    )

    names: Optional[List[str]] = Field(
        None,
        description="List of datasets preparing for being downloaded",
        json_schema_extra={"example": ["ts-ts-preserve-service-cpu-exhaustion-znzxcn"]},
    )

    output_path: str = Field(
        ...,
        description="The path to save package.zip",
        json_schema_extra={"example": os.getcwd()},
    )

    @field_validator("group_ids")
    @classmethod
    def validate_group_ids(cls, value: Optional[List[UUID]]) -> Optional[List[UUID]]:
        if value is not None:
            if not value:
                raise ValueError("GroupIDs cannot be empty if provided")

        return value

    @field_validator("names")
    @classmethod
    def validate_names(cls, value: Optional[List[str]]) -> Optional[List[str]]:
        if value is not None:
            if not value:
                raise ValueError("Names cannot be empty if provided")

            for name in value:
                if not name:
                    raise ValueError("Dataset name cannot be empty string")

                if len(name) < 1 or len(name) > 64:
                    raise ValueError(
                        f"The length of dataset must be in range [1-64]: {name}"
                    )

        return value

    @field_validator("output_path")
    @classmethod
    def validate_output_path(cls, value: str) -> str:
        if not os.path.isdir(value):
            raise FileNotFoundError(f"Output path does not exist: {value}")

        if not os.access(value, os.W_OK):
            raise PermissionError(f"No write permission on path: {value}")

        return value

    @model_validator(mode="after")
    def validate_exclusive_fields(self) -> "DownloadReq":
        group_ids_set = self.group_ids is not None and len(self.group_ids) > 0
        names_set = self.names is not None and len(self.names) > 0

        if group_ids_set and names_set:
            raise ValueError("Cannot specify both 'group_ids' and 'names'")

        if not group_ids_set and not names_set:
            raise ValueError("Must specify either 'group_ids' or 'names'")

        return self


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
        total_pages: 数据集记录总页数
        items: 数据集条目列表
    """

    total: int = Field(
        default=0,
        description="Total number of datasets",
        json_schema_extra={"example": 20},
        ge=0,
    )

    total_pages: int = Field(
        default=0,
        ge=0,
        description="Total number of injections pages",
        json_schema_extra={"example": 20},
    )

    items: List[DatasetItem] = Field(
        default_factory=list,
        description="List of datasets",
    )


class DetectorRecord(BaseModel):
    """
    detector 算法指标记录

    Attributes:
        span_name: 存在问题的Span名称
        issue: 检测到的异常描述
        abnormal_avg_duration: 异常时段的平均持续时间指标（秒）
        normal_avg_duration: 正常时段的平均持续时间指标（秒）
        abnormal_succ_rate: 异常时段的成功率百分比（0-1范围）
        normal_succ_rate: 正常时段的成功率百分比（0-1范围）
        abnormal_p90: 异常时段的90分位延迟测量值
        normal_p90: 正常时段的90分位延迟测量值
        abnormal_p95: 异常时段的95分位延迟测量值
        normal_p95: 正常时段的95分位延迟测量值
        abnormal_p99: 异常时段的99分位延迟测量值
        normal_p99: 正常时段的99分位延迟测量值
    """

    span_name: str = Field(
        ...,
        description="Identified span name with issues",
    )

    issue: str = Field(
        ...,
        description="Description of detected anomaly",
    )

    abnormal_avg_duration: Optional[float] = Field(
        None,
        description="Average duration metric for abnormal period (seconds)",
    )

    normal_avg_duration: Optional[float] = Field(
        None,
        description="Average duration metric for normal period (seconds)",
    )

    abnormal_succ_rate: Optional[float] = Field(
        None,
        description="Success rate percentage for abnormal period (0-1 scale)",
        ge=0,
        le=1,
    )

    normal_succ_rate: Optional[float] = Field(
        None,
        description="Success rate percentage for normal period (0-1 scale)",
        ge=0,
        le=1,
    )

    abnormal_p90: Optional[float] = Field(
        None,
        alias="abnormal_P90",
        description="90th percentile latency measurement for abnormal period",
        ge=0,
    )

    normal_p90: Optional[float] = Field(
        None,
        alias="normal_P90",
        description="90th percentile latency measurement for normal period",
        ge=0,
    )

    abnormal_p95: Optional[float] = Field(
        None,
        alias="abnormal_P95",
        description="95th percentile latency measurement for abnormal period",
        ge=0,
    )

    normal_p95: Optional[float] = Field(
        None,
        alias="normal_P95",
        description="95th percentile latency measurement for normal period",
        ge=0,
    )

    abnormal_p99: Optional[float] = Field(
        None,
        alias="abnormal_P99",
        description="99th percentile latency measurement for abnormal period",
        ge=0,
    )

    normal_p99: Optional[float] = Field(
        None,
        alias="normal_P99",
        description="99th percentile latency measurement for normal period",
        ge=0,
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


class QueryReq(BaseModel):
    name: str = Field(
        ...,
        description="Unique identifier for the dataset",
        json_schema_extra={"example": "ts-ts-preserve-service-cpu-exhaustion-znzxcn"},
        max_length=64,
    )

    sort: str = Field(
        ...,
        description="Dataset sorting method",
        json_schema_extra={"example": "desc"},
    )

    @field_validator("sort")
    @classmethod
    def validate_sort(cls, value: str) -> str:
        if value not in Dataset.ALLOWED_SORTS:
            raise ValueError(
                f"Invalid sort value. Must be one of: {', '.join(s for s in Dataset.ALLOWED_SORTS)}"
            )

        return value


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


class EnvVar(BaseModel):
    """
    采集数据镜像环境变量模型

    Attributes:
        NAMESPACE: 待注入命名空间
        SERVICE: 筛选服务名称
    """

    model_config = ConfigDict(extra="forbid")

    NAMESPACE: Optional[str] = Field(
        None,
        description="Target Kubernetes namespace",
        json_schema_extra={"example": "ts"},
    )


class BuildPayload(BaseModel):
    benchmark: str = Field(
        ...,
        description="Detailed anomaly detection metrics",
        json_schema_extra={"example": "clickhouse"},
    )

    name: str = Field(
        ...,
        description="Unique identifier for the dataset",
        json_schema_extra={"example": "ts-ts-preserve-service-cpu-exhaustion-znzxcn"},
        max_length=64,
    )

    pre_duration: int = Field(
        ...,
        description="Normal time before fault injection (minute)",
        json_schema_extra={"example": 1},
        gt=0,
    )

    env_vars: Optional[EnvVar] = Field(
        None,
        description="The enviroment variables of the image",
    )


class SubmitReq(BaseModel):
    """
    数据集构建请求
    """

    payloads: List[BuildPayload] = Field(
        ...,
        description="List of payloads to build dataset",
        min_length=1,
    )
