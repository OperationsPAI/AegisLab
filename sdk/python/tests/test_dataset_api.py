# Run this file:
# uv run pytest -s tests/test_dataset_api.py
from rcabench import rcabench
from pprint import pprint
import pytest
import os

BASE_URL = "http://localhost:8082"


@pytest.mark.parametrize(
    "names",
    [(["ts-ts-preserve-service-cpu-exhaustion-znzxcn"])],
)
def test_delete_datatests(names):
    """测试批量删除数据集"""

    sdk = rcabench.RCABenchSDK(BASE_URL)

    data = sdk.dataset.delete(names)
    pprint(data)


@pytest.mark.parametrize(
    "group_ids, output_path",
    [(["afdb9515-d27c-4c42-bc95-be3264dbc094"], os.getcwd())],
)
def test_download_datasets(group_ids, output_path):
    """测试批量下载数据集"""

    sdk = rcabench.RCABenchSDK(BASE_URL)

    file_path = sdk.dataset.download(group_ids, output_path)
    pprint(file_path)


@pytest.mark.parametrize("page_num, page_size", [(1, 10)])
def test_list_datasets(page_num, page_size):
    """测试数据集分页列表查询"""

    sdk = rcabench.RCABenchSDK(BASE_URL)

    data = sdk.dataset.list(page_num, page_size)
    pprint(data)


@pytest.mark.parametrize(
    "name, sort",
    [("ts-ts-preserve-service-cpu-exhaustion-qlwk6f", "desc")],
)
def test_query_dataset(name, sort):
    """测试指定数据集详细信息查询"""

    sdk = rcabench.RCABenchSDK(BASE_URL)

    data = sdk.dataset.query(name, sort)
    pprint(data)
