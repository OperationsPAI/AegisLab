# Run this file:
# uv run pytest -s tests/test_injection_api.py
from pprint import pprint
from rcabench import rcabench
import pytest

BASE_URL = "http://127.0.0.1:8082"


@pytest.mark.parametrize("mode", ["display", "engine"])
def test_get_injection_conf(mode):
    """测试获取注入配置信息"""

    sdk = rcabench.RCABenchSDK(BASE_URL)

    data = sdk.injection.get_conf(mode)
    pprint(data)


@pytest.mark.parametrize("page_num, page_size", [(1, 10)])
def test_list_injections(page_num, page_size):
    """测试分页查询注入记录"""

    sdk = rcabench.RCABenchSDK(BASE_URL)

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
def test_submit_injections(benchmark, interval, pre_duration, specs):
    """测试批量注入故障"""

    sdk = rcabench.RCABenchSDK(BASE_URL)

    data = sdk.injection.submit(benchmark, interval, pre_duration, specs)
    pprint(data)


# def fill_node(node: rcabench.Node):
#     if "children" in node:
#         for children, sub_node in node["children"].items():
#             fill_node(sub_node)
#     if "children" not in node:
#         node["value"] = random.randint(node["range"][0], node["range"][1])


# def test_get_para():
#     sdk = rcabench.RCABenchSDK(BASE_URL)
#     data = sdk.injection.get_parameters()
#     pprint(data)
#     assert data is not None
#     chosen_key = random.choice(list(data["children"].keys()))
#     fill_node(data["children"][chosen_key])
#     pprint(data["children"][chosen_key])

#     data["value"] = chosen_key
#     res = sdk.injection.submit(1, 2, "ts", [data])
#     print(res)
