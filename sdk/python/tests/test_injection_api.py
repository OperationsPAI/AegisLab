# Run this file:
# uv run pytest -s tests/test_injection_api.py
from pprint import pprint
import pytest


@pytest.mark.parametrize("mode", ["display", "engine"])
def test_get_injection_conf(sdk, mode):
    """测试获取注入配置信息"""
    data = sdk.injection.get_conf(mode)
    pprint(data)


@pytest.mark.parametrize("page_num, page_size", [(1, 10), (0, 10)])
def test_list_injections(sdk, page_num, page_size):
    """测试分页查询注入记录"""
    data = sdk.injection.list(page_num, page_size)
    pprint(data)


@pytest.mark.parametrize(
    "benchmark, interval, pre_duration, specs",
    [
        (
            "clickhouse",
            2,
            1,
            [
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
            ],
        )
    ],
)
def test_submit_injections(sdk, benchmark, interval, pre_duration, specs):
    """测试批量注入故障"""
    data = sdk.injection.submit(benchmark, interval, pre_duration, specs)
    pprint(data)
