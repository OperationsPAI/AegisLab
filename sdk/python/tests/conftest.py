import pytest
from dotenv import load_dotenv

from rcabench.client import RCABenchClient

load_dotenv()


@pytest.fixture(scope="session")
def rcabench_client():
    with RCABenchClient() as client:
        yield client
