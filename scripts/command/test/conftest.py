import json
from collections.abc import Generator
from pathlib import Path

import pytest
from rcabench.client import RCABenchClient
from rcabench.openapi.api.traces_api import TracesApi
from rcabench.openapi.api_client import ApiClient

from src.common.common import ENV, PROJECT_ROOT, console, settings
from src.common.helm_cli import HelmCLI, HelmRelease
from src.common.kubernetes_manager import KubernetesManager, with_k8s_manager
from src.port_manager import PortForwardManager

DEFAULT_PAGE = 1
DEFAULT_PAGE_SIZE = 20


cli = HelmCLI()


@pytest.fixture(scope="session", autouse=True)
def setup_environment() -> Generator[ApiClient, None, None]:
    settings.reload()

    port_manager = PortForwardManager(env=ENV.TEST)
    _install_rcabench(env=ENV.TEST, port_manager=port_manager)

    with RCABenchClient(
        base_url=port_manager.get_service_url("rcabench-exp"),
        # base_url="http://localhost:28080",
        username=settings.get("RCABENCH_USERNAME"),
        password=settings.get("RCABENCH_PASSWORD"),
    ).get_client() as client:
        yield client

    _uninstall_rcabench(env=ENV.TEST, port_manager=port_manager)


def pytest_generate_tests(metafunc):
    """
    Standard pytest hook to dynamically generate test cases (Parametrization).

    This implementation follows the 'Sibling JSON' pattern:
    - Looks for a JSON file with the same name as the test module (e.g., test_foo.json)
    - Maps the test function name to a key in that JSON file
    - Each JSON key should contain an array of test case objects

    Example JSON structure:
    {
        "test_function_name": [
            {"name": "case1", "input": ..., "expected": ...},
            {"name": "case2", "input": ..., "expected": ...}
        ]
    }

    Usage in test:

    def test_function_name(self, case: dict[str, Any]):
        # case will be {"name": "case1", "input": ..., "expected": ...}

        input_data = case["input"]

        expected = case["expected"]
    """
    # Only inject data if the test function has a parameter named 'case'
    if "case" not in metafunc.fixturenames:
        return

    module_path = Path(metafunc.module.__file__)
    json_path = module_path.with_suffix(".json")

    # If JSON file doesn't exist, skip parametrization (test will fail with missing fixture)
    if not json_path.exists():
        pytest.skip(f"Test data file not found: {json_path}")
        return

    try:
        with open(json_path, encoding="utf-8") as f:
            all_data = json.load(f)
    except (OSError, json.JSONDecodeError) as e:
        pytest.exit(f"Failed to load or parse test data file {json_path}: {e}")

    func_name = metafunc.function.__name__
    cases_data = all_data.get(func_name, [])

    if not isinstance(cases_data, list):
        pytest.exit(
            f"Test data for '{func_name}' in {json_path} must be a list, "
            f"got {type(cases_data).__name__}"
        )

    if not cases_data:
        pytest.skip(f"No test cases found for '{func_name}' in {json_path}")
        return

    # Generate test IDs from 'name' field or use index
    ids = [
        case.get("name", f"{func_name}_{i}")
        if isinstance(case, dict)
        else f"{func_name}_{i}"
        for i, case in enumerate(cases_data)
    ]

    metafunc.parametrize("case", cases_data, ids=ids, scope="function")


@with_k8s_manager()
def _install_rcabench(
    env: ENV, k8s_manager: KubernetesManager, port_manager: PortForwardManager
):
    release = HelmRelease(
        name="rcabench",
        chart=(PROJECT_ROOT / "helm").as_posix(),
        namespace="exp",
        values_file=PROJECT_ROOT / "manifests" / "test" / "rcabench.yaml",
        create_namespace=True,
        extra_args=[
            "--kube-context",
            k8s_manager.get_current_context(),
        ],
    )

    installed = cli.install(release, verbose=True, wait=True, timeout="10m0s")
    if not installed:
        raise RuntimeError("Failed to install RCABench Helm chart")

    port_manager.start_forwarding(
        env=ENV.TEST, namespace="exp", k8s_manager=k8s_manager
    )


@with_k8s_manager()
def _uninstall_rcabench(
    env: ENV, k8s_manager: KubernetesManager, port_manager: PortForwardManager
):
    console.print("\n")
    port_manager.stop_all_forwards()

    uninstalled = cli.uninstall(
        name="rcabench",
        namespace="exp",
        verbose=True,
        extra_args=["--kube-context", k8s_manager.get_current_context()],
    )
    if not uninstalled:
        console.print("[bold yellow]⚠️ Failed to uninstall RCABench[/bold yellow]")
