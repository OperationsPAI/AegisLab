from typing import Dict, List, Optional
from ..model.common import SubmitResult

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
        algorithms: List[List[str]],
        datasets: List[str],
        payload: Optional[List[Dict[str, str]]] = None,
    ) -> SubmitResult:
        """
        批量提交算法任务

        提供两种互斥的参数模式：

        模式1: 自动生成任务组合 (当payload为None时)
        - 生成算法(algorithm)与数据集(dataset)的笛卡尔积
        - 示例: algorithms=[["a1", "tag"]], datasets=["d1", "d2"] → 生成2个任务 tag 为空表示默认使用最新的 tag

        模式2: 直接提交预定义任务 (当payload非None时)
        - 需确保每个任务字典包含完整字段
        - 示例: [{"algorithm": ["a1", "tag"], "dataset": "d1"}]

        Args:
            algorithms: 算法标识列表，仅在模式1时生效，至少包含1个元素 例如 ["a1", "tag"]表示为a1:tag的镜像
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
            TypeError: 参数类型不符合预期时抛出

        Example:
            # 模式1: 自动生成2个任务 (a1×d1, a1×d2)
            submit(
                algorithms=[["a1", "tag"]],
                datasets=["d1", "d2"]
            )

            # 模式2: 直接提交预定义任务
            submit(
                payload=[
                    {"algorithm": ["a1", "tag"], "dataset": "d1"},
                    {"algorithm": ["a2", "tag"], "dataset": "d3"}
                ]
            )
        """
        # 类型检查
        if not isinstance(algorithms, list) and algorithms is not None:
            raise TypeError(
                f"algorithms must be a list, got {type(algorithms).__name__}"
            )

        if not isinstance(datasets, list) and datasets is not None:
            raise TypeError(f"datasets must be a list, got {type(datasets).__name__}")

        if payload is not None and not isinstance(payload, list):
            raise TypeError(f"payload must be a list, got {type(payload).__name__}")

        # 更详细的类型检查 - 仅在非None值时进行检查
        if algorithms:
            for i, algo in enumerate(algorithms):
                if not isinstance(algo, list):
                    raise TypeError(
                        f"Algorithm item {i} must be a tuple, got {type(algo).__name__}"
                    )
                if len(algo) < 1 or len(algo) > 2:
                    raise TypeError(
                        f"Algorithm tuple {i} must have 1 or 2 elements, got {len(algo)}"
                    )
                if not all(isinstance(item, str) for item in algo):
                    raise TypeError(
                        f"All elements in algorithm tuple {i} must be strings"
                    )

        if datasets and not all(isinstance(ds, str) for ds in datasets):
            raise TypeError("All elements in datasets must be strings")

        if payload:
            for i, item in enumerate(payload):
                if not isinstance(item, dict):
                    raise TypeError(
                        f"Payload item {i} must be a dict, got {type(item).__name__}"
                    )

        url = f"{self.URL_PREFIX}{self.URL_ENDPOINTS['submit']}"

        # 模式2: 使用预定义任务
        if payload is not None:
            # 检查是否混用了两种模式的参数
            if algorithms or datasets:
                raise ValueError("Cannot mix payload with algorithms/datasets")

            # 验证每个任务项是否包含必要的键
            for i, item in enumerate(payload):
                if "algorithm" not in item or "dataset" not in item:
                    raise ValueError(f"Payload item{i} missing required keys: {item}")

        # 模式1: 自动生成任务组合
        else:
            # 检查参数列表是否为空
            if not algorithms or not datasets:
                raise ValueError(
                    "Must provide either payload or all three list parameters"
                )

            # 创建算法和数据集的笛卡尔积组合
            payload = []
            for algorithm in algorithms:
                for dataset in datasets:
                    # 为每个组合创建一个任务字典
                    task = {
                        "algorithm": algorithm[0],
                        "dataset": dataset,
                        "tag": algorithm[1] if len(algorithm) > 1 else None,
                    }
                    payload.append(task)

        return SubmitResult.model_validate(self.client.post(url, payload=payload))

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
        response = self.client.get(url)

        # 类型检查响应数据
        if not isinstance(response, dict):
            raise TypeError(f"Expected dict response, got {type(response).__name__}")

        if "algorithms" not in response:
            raise ValueError("Response missing 'algorithms' key")

        algorithms = response["algorithms"]
        if not isinstance(algorithms, list):
            raise TypeError(
                f"Expected list of algorithms, got {type(algorithms).__name__}"
            )

        if not all(isinstance(algo, str) for algo in algorithms):
            raise TypeError("All algorithm names must be strings")

        return algorithms
