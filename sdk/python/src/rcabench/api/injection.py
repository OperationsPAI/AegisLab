from typing import Any, Dict, List, Union
from ..const import InjectionConfModes, Pagination
from ..model.common import SubmitResult
from ..model.injection import ListResult, SpecNode, SubmitReq

__all__ = ["Injection"]


class Injection:
    URL_PREFIX = "/injections"

    URL_ENDPOINTS = {
        "get_conf": "/conf",
        "list": "",
        "submit": "",
    }

    def __init__(self, client):
        self.client = client

    def get_conf(self, mode: str) -> Union[Dict[str, Any], SpecNode]:
        """
        获取指定模式的注入配置信息

        Args:
            mode: 配置模式，必须存在于 InjectionConfModes 中，可选值：["display", "engine"]

        Returns:
            SpecNode | dict: 配置数据对象，模式为 'engine' 时返回 SpecNode 模型实例，
                            其他模式返回原始响应字典

        Raises:
            TypeError: 参数类型非字符串时抛出
            ValueError: 参数值不在允许范围内时抛出

        Example:
            >>> engine_config = client.get_conf(mode="engine")
            >>> print(engine_config.model_dump())
        """
        url = f"{self.URL_PREFIX}{self.URL_ENDPOINTS['get_conf']}"

        if not isinstance(mode, str):
            raise TypeError("Mode must be string")
        if mode not in InjectionConfModes:
            raise ValueError(f"Injection conf mode must be one of {InjectionConfModes}")

        result = self.client.get(url, params={"mode": mode})
        if mode == "engine":
            return SpecNode.model_validate(result)
        else:
            return result

    def list(self, page_num: int, page_size: int) -> ListResult:
        """
        分页查询注入记录

        Args:
            page_num:  页码（从1开始的正整数），默认为1
            page_size: 每页数据量，仅允许 10/20/50 三种取值，默认为10

        Returns:
            ListResult: 包含数据集基本信息和分页结果的数据模型实例

        Raises:
            TypeError: 参数类型错误时抛出
            ValueError: 参数值不符合要求时抛出

        Example:
            >>> injections = client.list(page_num=1, page_size=10)
        """
        url = f"{self.URL_PREFIX}{self.URL_ENDPOINTS['list']}"

        if not isinstance(page_num, int):
            raise TypeError("Page number must be integer")
        if not isinstance(page_size, int):
            raise TypeError("Page size must be integer")

        if page_num < Pagination.DEFAULT_PAGE_NUM:
            raise ValueError(f"Page number must be ≥ {Pagination.DEFAULT_PAGE_NUM}")
        if page_size not in Pagination.ALLOWED_PAGE_SIZES:
            raise ValueError(
                f"Page size must be one of {Pagination.ALLOWED_PAGE_SIZES}"
            )

        params = {"page_num": page_num, "page_size": page_size}
        return ListResult.model_validate(self.client.get(url, params=params))

    def submit(
        self,
        benchmark: str,
        interval: int,
        pre_duration: int,
        specs: List[Dict[str, Any]],
    ) -> SubmitResult:
        """
        批量注入新的故障

        Args:
            benchmark: 基准测试数据库
            interval: 故障注入间隔（分钟），必须为 ≥1 的整数
            pre_duration: 注入前的正常时间（分钟），必须为 ≥1 的整数
            specs: 分层参数配置，每个元素需符合 SpecNode 结构

        Returns:
            SubmitResult: 包含任务提交结果的数据模型实例

        Raises:
            TypeError: 参数类型错误时抛出（如非整型间隔时间）
            ValueError: 参数值非法时抛出（如零或负数的间隔时间）

        Example:
            >>> task = client.submit(
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
            >>> print(task.task_id)
        """
        url = f"{self.URL_PREFIX}{self.URL_ENDPOINTS['submit']}"

        payload = {
            "benchmark": benchmark,
            "interval": interval,
            "pre_duration": pre_duration,
            "specs": specs,
        }
        payload = SubmitReq.model_validate(payload).model_dump()
        return SubmitResult.model_validate(self.client.post(url, payload=payload))
