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
        # PodFailure-dev
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
                                "2": {"value": 29},
                            }
                        },
                    },
                    "value": 1,
                }
            ],
        ),
        # CPUStress-prod
        (
            "clickhouse",
            30,
            10,
            [
                {
                    "children": {
                        "4": {
                            "children": {
                                "0": {"value": 15},
                                "1": {"value": 0},
                                "2": {"value": 32},
                                "3": {"value": 10},
                                "4": {"value": 2},
                            }
                        },
                    },
                    "value": 4,
                },
            ],
        ),
        # JVMMemoryStress-dev
        (
            "clickhouse",
            3,
            1,
            [
                {
                    "children": {
                        "28": {
                            "children": {
                                "0": {"value": 1},
                                "1": {"value": 0},
                                "2": {"value": 611},
                                "3": {"value": 1},
                            }
                        },
                    },
                    "value": 28,
                },
            ],
        ),
        # NetworkDuplicate-dev
        (
            "clickhouse",
            3,
            1,
            [
                {
                    "children": {
                        "19": {
                            "children": {
                                "0": {"value": 1},
                                "1": {"value": 0},
                                "2": {"value": 43},
                                "3": {"value": 23},
                                "4": {"value": 14},
                                "5": {"value": 1},
                            }
                        },
                    },
                    "value": 19,
                },
            ],
        ),
        (
            "clickhouse",
            3,
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
                },
                {
                    "children": {
                        "1": {
                            "children": {
                                "0": {"value": 1},
                                "1": {"value": 0},
                                "2": {"value": 35},
                            }
                        },
                    },
                    "value": 1,
                },
            ],
        ),
    ],
)
def test_submit_injections(sdk, benchmark, interval, pre_duration, specs):
    """测试批量注入故障"""
    data = sdk.injection.submit(benchmark, interval, pre_duration, specs)
    pprint(data)
