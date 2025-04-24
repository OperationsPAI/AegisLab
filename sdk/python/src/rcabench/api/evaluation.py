from typing import Any, List, Union
from ..client.http_client import HTTPClient
from ..model.error import ModelHTTPError


# TODO 添加模型验证
class Evaluation:
    URL_PREFIX = "/evaluations"

    URL_ENDPOINTS = {
        "execute": "",
    }

    def __init__(self, client: HTTPClient, api_version: str):
        self.client = client
        self.url_prefix = f"{api_version}{self.URL_PREFIX}"

    def execute(
        self,
        execution_ids: List[int],
        algorithms: List[str],
        levels: List[str],
        metrics: List[str],
        rank: int,
    ) -> Union[Any, ModelHTTPError]:
        """
        执行算法评估分析

        Args:
            execution_ids (List[int]): 必需参数，要评估的执行记录ID列表
                - 必须是非空的正整数列表
                - 示例: [101, 102]
            algorithms (Optional[List[str]]): 可选参数，要过滤的算法名称列表
                - 如果提供则必须是非空列表
                - 示例: ["e-diagnose", "nsigma"]
            levels (Optional[List[str]]): 可选参数，要分析的粒度级别列表
                - 如果提供则必须是非空列表
                - 示例: ["service", "pod"]
            metrics (Optional[List[str]]): 可选参数，需要包含的评估指标
                - 如果提供则必须是非空列表
                - 示例: ["accuracy", "f1_score"]
            rank (Optional[int]): 可选参数，结果排名过滤阈值
                - 如果提供则必须是正整数
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
            ...     rank=3
            ... )
        """
        url = f"{self.url_prefix}{self.URL_ENDPOINTS['execute']}"

        params = {
            "execution_ids": execution_ids,
            "algorithms": algorithms,
            "levels": levels,
            "metrics": metrics,
            "rank": rank,
        }
        return self.sdk.client.get(url, params=params)
