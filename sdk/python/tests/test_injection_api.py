# Run this file:
# uv run pytest -s tests/test_dataset_api.py
from rcabench import rcabench
from pprint import pprint
import random

BASE_URL = "http://127.0.0.1:8082"


def fill_node(node: rcabench.Node):
    if "children" in node:
        for children, sub_node in node["children"].items():
            fill_node(sub_node)
    if "children" not in node:
        node["value"] = random.randint(node["range"][0], node["range"][1])


def test_get_para():
    sdk = rcabench.RCABenchSDK(BASE_URL)
    data = sdk.injection.get_parameters()
    pprint(data)
    assert data is not None
    chosen_key = random.choice(list(data["children"].keys()))
    fill_node(data["children"][chosen_key])
    pprint(data["children"][chosen_key])
    res = sdk.injection.submit(1, 2, "ts", [data])
    print(res)
