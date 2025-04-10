# Run this file:
# uv run pytest -s tests/test_dataset_api.py
from pprint import pprint
from uuid import UUID
import os
import pytest


@pytest.mark.parametrize(
    "group_ids, output_path",
    [
        ([UUID("20eab8e5-2c1b-4119-9343-de931bf839a5")], os.getcwd()),
        ([UUID("1316756f-0a00-45f1-8707-ff1d1846b8e3")], os.getcwd()),
    ],
)
def test_download_datasets(sdk, group_ids, output_path):
    """测试批量下载数据集"""
    file_path = sdk.dataset.download(group_ids, output_path)
    pprint(file_path)


@pytest.mark.parametrize("page_num, page_size", [(1, 10), (1, 11)])
def test_list_datasets(sdk, page_num, page_size):
    """测试数据集分页列表查询"""
    data = sdk.dataset.list(page_num, page_size)
    pprint(data)


@pytest.mark.parametrize(
    "name, sort",
    [
        ("ts-ts-ui-dashboard-pod-failure-mngdrf", "desc"),
        ("ts-ts-ui-dashboard-pod-failure-ngtpvl", "desc"),
    ],
)
def test_query_dataset(sdk, name, sort):
    """测试指定数据集详细信息查询"""
    data = sdk.dataset.query(name, sort)
    pprint(data)


@pytest.mark.parametrize(
    "names",
    [(["ts-ts-ui-dashboard-pod-failure-mngdrf"])],
)
def test_delete_datatests(sdk, names):
    """测试批量删除数据集"""
    data = sdk.dataset.delete(names)
    pprint(data)
