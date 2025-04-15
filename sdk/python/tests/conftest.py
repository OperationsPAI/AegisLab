from rcabench.rcabench import RCABenchSDK
import pytest

#   BASE_URL = "http://10.10.10.220:32080"
BASE_URL = "http://127.0.0.1:8082"


@pytest.fixture
def sdk() -> RCABenchSDK:
    """
    初始化 RCABenchSDK 并返回实例
    """
    return RCABenchSDK(BASE_URL)
