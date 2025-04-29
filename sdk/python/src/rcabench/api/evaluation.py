from typing import Any, List, Optional, Union
from .validation import validate_request_response
from ..client.http_client import HTTPClient
from ..model.error import ModelHTTPError
from ..model.evaluation import EvaluationReq


# TODO 添加模型验证
class Evaluation:
    URL_PREFIX = "/evaluations"

    URL_ENDPOINTS = {
        "execute": "",
    }

    def __init__(self, client: HTTPClient, api_version: str):
        self.client = client
        self.url_prefix = f"{api_version}{self.URL_PREFIX}"

    @validate_request_response(EvaluationReq)
    def execute(
        self,
        execution_ids: List[int],
        metrics: Optional[List[str]],
        rank: Optional[int],
    ) -> Union[Any, ModelHTTPError]:
        """
        执行算法评估分析

        Args:
            execution_ids (List[int]): 必需参数，要评估的执行记录ID列表
                - 必须是非空的正整数列表
                - 示例: [101, 102]
            metrics (Optional[List[str]]): 可选参数，需要包含的评估指标
                - 如果为 None 或者是空列表则表示全部
                - 示例: ["accuracy", "f1_score"]
            rank (Optional[int]): 可选参数，结果排名过滤阈值
                - 如果提供则必须是正整数，且在1, 3, 5之间
                - 示例: 5

        Returns:
            dict: 包含评估结果的字典，结构示例:
                {
                    "summary": {...},
                    "details": [...]
                }

        Raises:
            TypeError: 当参数类型不符合要求时抛出
            ValueError: 当参数值不符合要求时抛出

        Example:
            >>> result = evaluation.execute(
            ...     execution_ids=[101, 102],
            ...     algorithms=["e-diagnose"],
            ...     rank=5
            ... )
        """
        url = f"{self.url_prefix}{self.URL_ENDPOINTS['execute']}"

        params = {
            "execution_ids": execution_ids,
            "metrics": metrics,
            "rank": rank,
        }
        return self.client.get(url, params=params)
