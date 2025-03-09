# Run this file:
# uv run pytest -s tests/test_connection.py
from rcabench import rcabench

from pprint import pprint

import pytest


@pytest.mark.asyncio
async def test_connection():
    sdk = rcabench.RCABenchSDK(base_url="http://10.10.10.220:32080")
    pprint(sdk.injection.get_namespace_pod_info())
