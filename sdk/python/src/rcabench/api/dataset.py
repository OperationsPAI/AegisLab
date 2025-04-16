from typing import List, Optional, Union
from .validation import validate_request_response
from ..const import Pagination, Dataset as DatasetConst
from ..client.http_client import HTTPClient
from ..model.common import PaginationReq, SubmitResult
from ..model.dataset import (
    DeleteReq,
    DeleteResult,
    DownloadReq,
    DownloadResult,
    ListResult,
    QueryReq,
    QueryResult,
    SubmitReq,
)
from ..model.error import ModelHTTPError
from tqdm import tqdm
from urllib.parse import unquote
from uuid import UUID
import os


class Dataset:
    URL_PREFIX = "/datasets"

    URL_ENDPOINTS = {
        "delete": "",
        "download": "/download",
        "list": "",
        "query": "/query",
        "submit": "",
    }

    def __init__(self, client: HTTPClient, api_version: str):
        self.client = client
        self.url_prefix = f"{api_version}{self.URL_PREFIX}"

    @validate_request_response(DeleteReq, DeleteResult)
    def delete(self, names: List[str]) -> Union[DeleteResult, ModelHTTPError]:
        """
        批量删除数据集

        Args:
            names (List[str]): 要删除的数据集名称列表，需满足：
                - 非空列表
                - 每个元素为字符串类型
                - 字符串非空且长度在1-64字符之间
                示例: ["ts-ts-preserve-service-cpu-exhaustion-znzxcn"]

        Returns:
            DeleteResult: 包含成功删除数目和未删除数据集的结构化响应对象，

        Raises:
            ModelValidationError: 当输入参数不符合Pydantic模型验证规则时抛出
            ModelHTTPError: 当API请求失败（4xx/5xx状态码）时抛出

        Example:
            >>> delete(["ts-ts-preserve-service-cpu-exhaustion-znzxcn"])  # 提交删除请求，返回操作结果
            >>> delete([123])  # 错误示例
            TypeError: Dataset name must be string
        """
        url = f"{self.url_prefix}{self.URL_ENDPOINTS['delete']}"

        return self.client.delete(url, params={"names": names})

    @validate_request_response(DownloadReq, DownloadResult)
    def download(
        self,
        group_ids: Optional[List[UUID]],
        names: Optional[List[str]],
        output_path: str,
    ) -> Union[DownloadResult, ModelHTTPError]:
        """
        批量下载数据集文件组

        通过流式下载将多个数据集打包为 ZIP 文件，自动处理以下功能：
        - 显示实时下载进度条
        - 从 Content-Disposition 解析原始文件名
        - 分块写入避免内存溢出

        Args:
            group_ids (List[UUID] | None): 任务组标识列表，为 UUID 格式
            names (List[str] | None): 数据集名称列表
            output_path (str): 文件保存目录路径，需确保有写权限

        Returns:
            str: 下载文件的完整保存路径
                示例: "/data/downloads/package.zip"

        Raises:
            ModelValidationError: 当输入参数不符合Pydantic模型验证规则时抛出
            ModelHTTPError: 当API请求失败（4xx/5xx状态码）时抛出

        Example:
            >>> download(
                group_ids=["550e8400-e29b-41d4-a716-446655440000"],
                output_path="/data/downloads"
            )
            '/data/downloads/package.zip'
        """
        url = f"{self.url_prefix}{self.URL_ENDPOINTS['download']}"

        response = self.client.get(
            url, params={"group_ids": group_ids, "names": names}, stream=True
        )
        if isinstance(response, ModelHTTPError):
            return response

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

        return {"file_path": file_path}

    @validate_request_response(PaginationReq, ListResult)
    def list(
        self,
        page_num: int = Pagination.DEFAULT_PAGE_NUM,
        page_size: int = Pagination.DEFAULT_PAGE_SIZE,
    ) -> Union[ListResult, ModelHTTPError]:
        """
        分页查询数据集

        Args:
            page_num:  页码（从1开始的正整数）
            page_size: 每页数据量，仅允许 10/20/50 三种取值

        Returns:
            ListResult: 包含数据集基本信息和分页结果的数据模型实例

        Raises:
            ModelValidationError: 当输入参数不符合Pydantic模型验证规则时抛出
            ModelHTTPError: 当API请求失败（4xx/5xx状态码）时抛出

        Example:
            >>> dataset = client.list(page_num=1, page_size=10)
        """
        url = f"{self.url_prefix}{self.URL_ENDPOINTS['list']}"

        params = {"page_num": page_num, "page_size": page_size}
        return self.client.get(url, params=params)

    @validate_request_response(QueryReq, QueryResult)
    def query(
        self, name: str, sort: str = DatasetConst.DEFAULT_SORT
    ) -> Union[QueryResult, ModelHTTPError]:
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

        Raises:
            ModelValidationError: 当输入参数不符合Pydantic模型验证规则时抛出
            ModelHTTPError: 当API请求失败（4xx/5xx状态码）时抛出

        Example:
            >>> dataset = client.query("order-service-latency")
            142.3
        """
        url = f"{self.url_prefix}{self.URL_ENDPOINTS['query']}"

        params = {"name": name, "sort": sort}
        return self.client.get(url, params=params)

    @validate_request_response(SubmitReq, SubmitResult)
    def submit(self, payloads):
        """查询单个数据集"""
        url = f"{self.url_prefix}{self.URL_ENDPOINTS['submit']}"

        return self.client.post(url, json=payloads)
