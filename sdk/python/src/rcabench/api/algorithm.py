from typing import Dict, List, Optional
from ..model.task import SubmitResult
import itertools

__all__ = ["Algorithm"]


class Algorithm:
    URL_PREFIX = "/algorithms"

    URL_ENDPOINTS = {
        "list": "",
        "submit": "",
    }

    def __init__(self, client):
        self.client = client

    def submit(
        self,
        algorithms: List[str],
        datasets: List[str],
        payload: Optional[List[Dict[str, str]]] = None,
    ) -> SubmitResult:
        """
        批量提交算法任务

        提供两种互斥的参数模式：

        模式1: 自动生成任务组合 (当payload为None时)
        - 生成算法(algorithm)与数据集(dataset)的笛卡尔积
        - 示例: algorithms=["a1"], datasets=["d1", "d2"] → 生成2个任务

        模式2: 直接提交预定义任务 (当payload非None时)
        - 需确保每个任务字典包含完整字段
        - 示例: [{"algorithm": "a1", "dataset": "d1"}]

        Args:
            algorithms: 算法标识列表，仅在模式1时生效，至少包含1个元素
            datasets: 数据集标识列表，仅在模式1时生效，至少包含1个元素
            payload: 预定义任务字典列表，仅在模式2时使用，与上述列表参数互斥

        Returns:
            SubmitResult: 包含任务组ID和追踪链信息的结构化响应对象，
            可通过属性直接访问数据，如:
            >>> result = submit(...)
            >>> print(result.group_id)
            UUID('e7cbb5b8-554e-4c82-a018-67f626fc12c6')

        Raises:
            ValueError: 以下情况抛出：
                - 混用payload与algorithms/datasets参数 (模式冲突)
                - 模式1参数存在空列表 (需algorithms和datasets非空)
                - 模式2的payload中缺少algorithm/dataset字段

        Example:
            # 模式1: 自动生成2个任务 (a1×d1, a1×d2)
            submit(
                algorithms=["a1"],
                datasets=["d1", "d2"]
            )

            # 模式2: 直接提交预定义任务
            submit(
                payload=[
                    {"algorithm": "a1", "dataset": "d1"},
                    {"algorithm": "a2", "dataset": "d3"}
                ]
            )
        """
        url = f"{self.URL_PREFIX}{self.URL_ENDPOINTS['submit']}"

        required_keys = {"algorithm", "dataset"}
        if payload is not None:
            if algorithms or datasets:
                raise ValueError("Cannot mix payload with algorithms/datasets")

            for i, item in enumerate(payload):
                if not required_keys.issubset(item.keys()):
                    raise ValueError(f"Payload item{i} missing required keys: {item}")
        else:
            param_lists = [algorithms, datasets]
            if any(not param_list for param_list in param_lists):
                raise ValueError(
                    "Must provide either payload or all three list parameters"
                )

            combinations = itertools.product(*param_lists)
            payload = [
                dict(zip(required_keys, combination)) for combination in combinations
            ]

        return SubmitResult.model_validate(self.client.post(url, payload))

    def list(self) -> List[str]:
        """
        获取当前可用的算法列表

        通过GET请求获取服务端预定义的算法集合，返回算法标识符的列表。
        该列表通常用于后续任务提交时指定算法参数。

        Returns:
            List[str]: 算法名称字符串列表，按服务端返回顺序排列

            示例: ["detector", "e-diagnose"]
        """
        url = f"{self.URL_PREFIX}{self.URL_ENDPOINTS['list']}"
        return self.client.get(url)["algorithms"]
