import time
from collections.abc import Generator
from pathlib import Path

import pytest
from dotenv import load_dotenv
from python_on_whales import DockerClient
from rcabench.client import RCABenchClient
from rcabench.openapi.api_client import ApiClient

load_dotenv(dotenv_path=Path(__file__).parent.parent / ".env.test")

# Get the project root directory (AegisLab)
PROJECT_ROOT = Path(__file__).parent.parent.parent.parent
DOCKER_COMPOSE_FILE = PROJECT_ROOT / "docker-compose.test.yaml"

DEFAULT_PAGE = 1
DEFAULT_PAGE_SIZE = 20


@pytest.fixture(scope="module")
def docker_compose():
    """Start Docker Compose services for testing and tear them down after tests."""
    docker = DockerClient(compose_files=[DOCKER_COMPOSE_FILE])

    print("\n🧹 Cleaning up existing containers...")
    try:
        docker.compose.down(volumes=True, remove_orphans=True)
    except Exception:
        pass

    print("\n🚀 Starting Docker Compose services...")
    try:
        docker.compose.up(detach=True, build=True)
    except Exception as e:
        print(f"⚠️  Some services may have failed to start: {e}")

    # Wait for services to be ready
    print("⏳ Waiting for services to be ready...")
    max_retries = 60
    retry_count = 0

    while retry_count < max_retries:
        try:
            services = docker.compose.ps()
            running_services = [s for s in services if s.state.status == "running"]
            print(f"   {len(running_services)}/{len(services)} services running", end="\r")

            if len(running_services) >= 4:
                print(f"\n✅ {len(running_services)} services are running successfully")
                time.sleep(10)
                break
        except Exception as e:
            print(f"\n⏳ Services not ready yet: {e}")

        time.sleep(2)
        retry_count += 1

    if retry_count >= max_retries:
        print("\n❌ Services failed to start in time")
        print("\n📋 Service status:")
        try:
            services = docker.compose.ps()
            for service in services:
                print(f"  - {service.name}: {service.state.status}")
        except Exception:
            pass

        print("\n📋 Logs:")
        try:
            docker.compose.logs(tail="50")
        except Exception:
            pass

        docker.compose.down(volumes=True)
        raise RuntimeError("Docker Compose services failed to start")

    yield docker

    # print("\n🛑 Stopping Docker Compose services...")
    # docker.compose.down(volumes=True)
    # print("✅ Docker Compose services stopped")


@pytest.fixture(scope="module")
def rcabench_client(docker_compose) -> Generator[ApiClient, None, None]:
    """Create RCABench client that connects to test environment."""
    time.sleep(3)
    with RCABenchClient().get_client() as client:
        yield client
