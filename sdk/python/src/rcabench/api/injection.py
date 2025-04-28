from typing import Any, Dict, List, Union
from .validation import validate_request_response
from ..client.http_client import HTTPClient
from ..const import Pagination
from ..model.common import PaginationReq, SubmitResult
from ..model.error import ModelHTTPError
from ..model.injection import GetConfReq, ListResult, SpecNode, SubmitReq

__all__ = ["Injection"]


class Injection:
    URL_PREFIX = "/injections"

    URL_ENDPOINTS = {
        "get_conf": "/conf",
        "list": "",
        "submit": "",
    }

    def __init__(self, client: HTTPClient, api_version: str):
        self.client = client
        self.url_prefix = f"{api_version}{self.URL_PREFIX}"

    @validate_request_response(request_model=GetConfReq)
    def get_conf(self, namespace: str, mode: str) -> Union[Dict[str, Any], SpecNode]:
        """
        获取指定模式的注入配置信息

        Args:
            namespace (str): k8s命名空间
            mode (str): 配置模式，必须存在于 INJECTION_CONF_MODES 中，可选值：["display", "engine"]

        Returns:
            SpecNode|dict: 配置数据对象
            - 模式为 'display' 时返回原始响应字典
            - 模式为 'engine' 时返回 SpecNode 模型实例，

        Raises:
            ValueError: 参数值不在允许范围内时抛出
            ModelHTTPError: 当API请求失败（4xx/5xx状态码）时抛出

        Examples:
            >>> client.get_conf(mode="engine")
        """
        url = f"{self.url_prefix}{self.URL_ENDPOINTS['get_conf']}"

        result = self.client.get(url, params={"namespace": namespace, "mode": mode})
        if mode == "engine":
            return SpecNode.model_validate(result)
        else:
            return result

    @validate_request_response(PaginationReq, ListResult)
    def list(
        self,
        page_num: int = Pagination.DEFAULT_PAGE_NUM,
        page_size: int = Pagination.DEFAULT_PAGE_SIZE,
    ) -> Union[ListResult, ModelHTTPError]:
        """
        分页查询注入记录

        Args:
            page_num (int):  页码（从1开始的正整数），默认为1
            page_size (int): 每页数据量，仅允许 10/20/50 三种取值，默认为10

        Returns:
            ListResult: 包含注入任务基本信息和分页结果的数据模型实例

        Raises:
            TypeError: 参数类型错误时抛出
            ValueError: 参数值不符合要求时抛出
            ModelHTTPError: 当API请求失败（4xx/5xx状态码）时抛出

        Examples:
            >>> client.list(page_num=1, page_size=10)
        """
        url = f"{self.url_prefix}{self.URL_ENDPOINTS['list']}"

        params = {"page_num": page_num, "page_size": page_size}
        return self.client.get(url, params=params)

    @validate_request_response(SubmitReq, SubmitResult)
    def submit(
        self,
        benchmark: str,
        interval: int,
        pre_duration: int,
        specs: List[Dict[str, Any]],
    ) -> Union[SubmitResult, ModelHTTPError]:
        """
        提交批量故障注入任务

        Args:
            benchmark (str): 基准测试数据库
            interval (int): 故障注入间隔时间（分钟），必须 ≥1
            pre_duration (int): 注入前的正常运行时长（分钟），必须 ≥1
            specs (List[Dict]): 分层参数配置列表，每个元素应符合 SpecNode 结构定义

        Returns:
            SubmitResult: 包含任务提交结果的数据模型实例

        Raises:
            ModelValidationError: 当输入参数不符合Pydantic模型验证规则时抛出
            ModelHTTPError: 当API请求失败（4xx/5xx状态码）时抛出

        Examples:
            >>> client.submit(
                    benchmark="clickhouse_v1",
                    interval=2,
                    pre_duration=5,
                    specs=[
                        {
                            "children": {
                                "1": {
                                    "children": {
                                        "0": {"value": 1},
                                        "1": {"value": 0},
                                        "2": {"value": 42},
                                    }
                                },
                            },
                            "value": 1,
                        }
                    ]
                )
        """
        url = f"{self.url_prefix}{self.URL_ENDPOINTS['submit']}"

        payload = {
            "benchmark": benchmark,
            "interval": interval,
            "pre_duration": pre_duration,
            "specs": specs,
        }
        return self.client.post(url, json=payload)
