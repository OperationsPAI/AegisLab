from typing import List
from tqdm import tqdm
from urllib.parse import unquote
from ..const import Pagination
from ..model.dataset import DeleteResult, ListResult, QueryResult
import os
import uuid


class Dataset:
    URL_PREFIX = "/datasets"

    URL_ENDPOINTS = {
        "delete": "",
        "download": "/download",
        "list": "",
        "query": "/query",
        "submit": "",
    }

    def __init__(self, client):
        self.client = client

    def delete(self, names: List[str]) -> DeleteResult:
        """
        批量删除指定数据集

        Args:
            names (List[str]): 要删除的数据集名称列表，需满足：
                - 非空列表
                - 每个元素为字符串类型
                - 字符串非空且长度在1-64字符之间
                示例: ["ts-ts-preserve-service-cpu-exhaustion-znzxcn"]

        Returns:
            DeleteResult: 包含成功删除数目和未删除数据集的结构化响应对象，
            可通过属性直接访问数据，如:
            >>> result = submit(...)
            >>> print(result.success_count)
            2

        Raises:
            TypeError: 当名称列表包含非字符串元素时
            ValueError: 当名称列表为空时

        Example:
            >>> delete(["ts-ts-preserve-service-cpu-exhaustion-znzxcn"])  # 提交删除请求，返回操作结果

            >>> delete([123])  # 错误示例
            TypeError: Dataset name must be string
        """
        url = f"{self.URL_PREFIX}{self.URL_ENDPOINTS['delete']}"

        if not names:
            raise ValueError("Dataset names list cannot be empty")

        if not all(isinstance(name, str) for name in names):
            invalid = [name for name in names if not isinstance(name, str)]
            raise TypeError(f"Dataset names must be string. Invalid: {invalid}")

        for name in names:
            if not name:
                raise ValueError("Dataset name cannot be empty string")
            if len(name) > 64:
                raise ValueError(f"Name too long (max 64 chars): {name}")

        return DeleteResult.model_validate(
            self.client.delete(url, params={"names": names})
        )

    def download(self, group_ids: List[str], output_path: str) -> str:
        """
        批量下载数据集文件组

        通过流式下载将多个数据集打包为ZIP文件，自动处理以下功能：
        - 显示实时下载进度条
        - 从Content-Disposition解析原始文件名
        - 分块写入避免内存溢出

        Args:
            group_ids (List[str]): 任务组标识列表，通常为UUID格式
                示例: ["550e8400-e29b-41d4-a716-446655440000"]
            output_path (str): 文件保存目录路径，需确保有写权限
                示例: "/data/downloads"

        Returns:
            str: 下载文件的完整保存路径
                示例: "/data/downloads/package.zip"

        Raises:
            FileNotFoundError: 当output_path不存在时触发
            PermissionError: 当output_path无写权限时触发

        Example:
            >>> download(
                group_ids=["550e8400-e29b-41d4-a716-446655440000"],
                output_path="/data/downloads"
            )
            '/data/downloads/package.zip'
        """
        url = f"{self.URL_PREFIX}{self.URL_ENDPOINTS['download']}"

        if not group_ids:
            raise ValueError("group_ids cannot be empty list")

        for gid in group_ids:
            if not isinstance(gid, str):
                raise ValueError(f"Non-string value in group_ids: {repr(gid)}")

            try:
                uuid.UUID(gid)
            except ValueError:
                raise ValueError(f"Invalid UUID format: {gid}")

        if not os.path.isdir(output_path):
            raise FileNotFoundError(f"Output path does not exist: {output_path}")

        if not os.access(output_path, os.W_OK):
            raise PermissionError(f"No write permission on path: {output_path}")

        response = self.client.get(url, params={"group_ids": group_ids}, stream=True)

        total_size = int(response.headers.get("content-length", 0))
        progress_bar = tqdm(total=total_size, unit="B", unit_scale=True)

        filename = unquote(url.split("/")[-1])
        if "Content-Disposition" in response.headers:
            content_disposition = response.headers["Content-Disposition"]
            filename = unquote(content_disposition.split("filename=")[-1].strip('"'))

        file_path = os.path.join(output_path, filename)

        try:
            with open(os.path.join(output_path, filename), "wb") as f:
                for chunk in response.iter_content(chunk_size=1024):
                    if chunk:
                        f.write(chunk)
                        progress_bar.update(len(chunk))
        except IOError as e:
            raise RuntimeError(f"Failed to write file: {str(e)}") from e
        finally:
            progress_bar.close()

        return file_path

    def list(
        self,
        page_num: int = Pagination.DEFAULT_PAGE_NUM,
        page_size: int = Pagination.DEFAULT_PAGE_SIZE,
    ) -> ListResult:
        """
        分页查询数据集

        Args:
            page_num:  页码（从1开始的正整数），默认为1
            page_size: 每页数据量，仅允许 10/20/50 三种取值，默认为10

        Returns:
            dict: 包含数据集和分页信息的字典，结构示例：
                {"data": [...], "total": 100, "page": 1}

        Raises:
            TypeError: 参数类型错误时抛出
            ValueError: 参数值不符合要求时抛出

        Example:
            >>> dataset = client.list(page_num=1, page_size=10)
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

    def query(self, name: str, sort: str = "desc") -> QueryResult:
        """查询指定名称的数据集详细信息

        获取指定数据集的完整分析记录，包括检测结果和执行记录

        Args:
            name (str): 数据集名称（必填）
                  - 类型：字符串
                  - 字符串非空且长度在1-64字符之间
                  示例：["ts-ts-preserve-service-cpu-exhaustion-znzxcn"]
            sort (str): 排序方式（可选）
                  - 允许值：desc（降序）/ asc（升序）
                  - 默认值：desc

        Returns:
            dict: 数据集完整信息，结构参考 QueryResult 模型定义
                  - 包含基础信息、检测指标、算法执行结果等
                  - 查询失败时返回 None

        Raises:
            TypeError: 参数类型错误时抛出
            ValueError: 参数值不合法时抛出）

        Example:
            >>> dataset = client.query("order-service-latency")
            >>> print(dataset["detector_result"]["p99"])
            142.3
        """
        url = f"{self.URL_PREFIX}{self.URL_ENDPOINTS['query']}"

        if not isinstance(name, str):
            raise TypeError("Dataset names must be string")
        if not name:
            raise ValueError("Dataset name cannot be empty string")
        if len(name) > 64:
            raise ValueError(f"Name too long (max 64 chars): {name}")

        if sort not in {"desc", "asc"}:
            raise ValueError(f"Invalid sort value: {sort}. Must be 'desc' or 'asc'")

        params = {"name": name, "sort": sort}
        return QueryResult.model_validate(self.client.get(url, params=params))

    def submit(self, payload):
        """查询单个数据集"""
        url = f"{self.URL_PREFIX}{self.URL_ENDPOINTS['submit']}"
        return self.client.post(url, payload=payload)
