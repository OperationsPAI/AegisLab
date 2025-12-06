import time
from collections.abc import Generator

import pytest
from python_on_whales import DockerClient
from rcabench.client import RCABenchClient
from rcabench.openapi.api_client import ApiClient

from src.common.common import PROJECT_ROOT, settings
from src.common.docker_manager import DockerManager

# Get the project root directory (AegisLab)
DOCKER_COMPOSE_FILE = PROJECT_ROOT / "docker-compose.test.yaml"

DEFAULT_PAGE = 1
DEFAULT_PAGE_SIZE = 20


@pytest.fixture(scope="module")
def docker_client() -> Generator[DockerClient, None, None]:
    """Start Docker Compose services for testing and tear them down after tests."""
    with DockerManager(
        compose_file=DOCKER_COMPOSE_FILE,
        max_retries=60,
        startup_wait=10,
    ) as dm:
        if dm.is_running():
            yield dm.get_client()

    dm.stop_services()
    dm.clear_sessions()


@pytest.fixture(scope="module")
def rcabench_client(docker_client) -> Generator[ApiClient, None, None]:
    """Create RCABench client that connects to test environment."""
    time.sleep(3)
    with RCABenchClient(
        base_url=settings.get("RCABENCH_BASE_URL"),
        username=settings.get("RCABENCH_USERNAME"),
        password=settings.get("RCABENCH_PASSWORD"),
    ).get_client() as client:
        yield client
