from typing import List
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
        example=["ts-ts-preserve-service-cpu-exhaustion-znzxcn"],
    )
