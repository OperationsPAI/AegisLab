from rcabench.rcabench import RCABenchSDK
import pytest

BASE_URL = "http://localhost:8082"


@pytest.fixture
def sdk() -> RCABenchSDK:
    """
    初始化 RCABenchSDK 并返回实例
    """
    return RCABenchSDK(BASE_URL)
