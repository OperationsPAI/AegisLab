from typing import Dict, List, Union
from .validation import validate_request_response
from ..client.http_client import HTTPClient
from ..model.algorithm import ListResult, SubmitReq
from ..model.common import SubmitResult
from ..model.error import ModelHTTPError

__all__ = ["Algorithm"]


class Algorithm:
    URL_PREFIX = "/algorithms"

    URL_ENDPOINTS = {
        "list": "",
        "submit": "",
    }

    def __init__(self, client: HTTPClient, api_version: str):
        self.client = client
        self.url_prefix = f"{api_version}{self.URL_PREFIX}"

    @validate_request_response(response_model=ListResult)
    def list(self) -> Union[ListResult, ModelHTTPError]:
        """
        获取可用的算法列表

        Returns:
            ListResult: 包含算法列表的结构化响应对象

        Raises:
            ModelValidationError: 当输入参数不符合Pydantic模型验证规则时抛出
            ModelHTTPError: 当API请求失败（4xx/5xx状态码）时抛出
        """
        url = f"{self.url_prefix}{self.URL_ENDPOINTS['list']}"

        return self.client.get(url)

    @validate_request_response(SubmitReq, SubmitResult)
    def submit(
        self, payloads: List[Dict[str, str]]
    ) -> Union[SubmitResult, ModelHTTPError]:
        """
        批量提交算法任务

        Args:
            payloads: 预定义任务字典列表

        Returns:
            SubmitResult: 包含任务组ID和追踪链信息的结构化响应对象，

        Raises:
            ModelValidationError: 当输入参数不符合Pydantic模型验证规则时抛出
            ModelHTTPError: 当API请求失败（4xx/5xx状态码）时抛出

        Examples:
            >>> submit(
                    payloads=[
                        {"algorithm": ["a1"], "dataset": "d1"},
                        {"algorithm": ["a2"], "dataset": "d3", "tag":"latest"},
                    ]
                )
        """
        url = f"{self.url_prefix}{self.URL_ENDPOINTS['submit']}"

        return self.client.post(url, json=payloads)
