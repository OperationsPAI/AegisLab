# Run this file:
# uv run pytest -s tests/test_dataset_api.py
from typing import List, Optional
from pprint import pprint
import pytest


@pytest.mark.parametrize(
    "execution_ids, metrics, rank",
    [([470], None, None)],
)
def test_execute(
    sdk,
    execution_ids: List[int],
    metrics: Optional[List[str]],
    rank: Optional[int],
):
    """测试算法评估"""
    file_path = sdk.evaluation.execute(execution_ids, metrics, rank)
    pprint(file_path)
