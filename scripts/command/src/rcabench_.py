import re
import time

import tomlkit
from python_on_whales import DockerClient
from ruamel.yaml import YAML

from src.common.common import ENV, PROJECT_ROOT, console, settings
from src.common.kubernetes_manager import KubernetesManager, with_k8s_manager

DOCKER_COMPOSE_FILE = PROJECT_ROOT / "docker-compose.yaml"
SKAFFOLD_FILE = PROJECT_ROOT / "skaffold.yaml"


@with_k8s_manager()
def check_db(env: ENV, k8s_manager: KubernetesManager):
    """Checks the health of the RCABench database."""
    _check_pod_health(k8s_manager, "RCABench MySQL Database", "rcabench-mysql")


@with_k8s_manager()
def check_redis(env: ENV, k8s_manager: KubernetesManager):
    """Checks the health of the RCABench Redis cache."""
    _check_pod_health(k8s_manager, "RCABench Redis Cache", "rcabench-redis")


def _check_pod_health(
    k8s_manager: KubernetesManager, service_name: str, pod_name: str
) -> None:
    """Checks the health of a given pod in Kubernetes."""
    console.print(f"[bold blue]🔍 Checking {service_name} health...[/bold blue]")

    is_running = k8s_manager.check_pod(
        pod_name,
        namespace=settings.k8s_namespace,
        label_selector=f"app={pod_name}",
        field_selector="status.phase=Running",
        prefix_match=True,
    )

    if is_running:
        console.print(f"[bold green]✅ {service_name} is running[/bold green]")
        return

    console.print(
        f"[bold red]❌ {service_name} is NOT running[/bold red] "
        f"in namespace [yellow]'{settings.k8s_namespace}'[/yellow]."
    )
    raise SystemExit(1)


def _wait_for_healthy(docker, timeout=120):
    """
    Waits for all Docker Compose services to become healthy.
    """
    console.print("[bold blue]⏳ Waiting for services to be healthy...[/bold blue]")
    start_time = time.time()

    while True:
        containers = docker.compose.ps()
        all_healthy = True
        states = []

        for container in containers:
            health_status = (
                container.state.health.status
                if container.state.health
                else "no healthcheck"
            )
            is_running = container.state.running

            states.append(f"{container.name}: [bold]{health_status}[/bold]")

            if not is_running or (
                container.state.health and health_status != "healthy"
            ):
                all_healthy = False

        if all_healthy:
            console.print("[bold green]✅ All services are healthy![/bold green]")
            return True

        if time.time() - start_time > timeout:
            console.print(
                "[bold red]❌ Timeout: Services failed to become healthy.[/bold red]"
            )
            for s in states:
                console.print(f"  - {s}")
            raise SystemExit(1)

        time.sleep(2)


@with_k8s_manager()
def local_deploy(env: ENV, k8s_manager: KubernetesManager):
    assert env == ENV.DEV, "Local deploy is only supported for DEV environment."

    console.print("[bold blue]🚀 Starting local RCAbench deployment...[/bold blue]")

    docker = DockerClient(compose_files=[DOCKER_COMPOSE_FILE])
    try:
        docker.compose.down(remove_orphans=True)
        console.print("[bold green]✅ Cleaned up existing containers[/bold green]")
    except Exception:
        console.print(
            "[bold yellow]⚠️ No existing containers to clean up.[/bold yellow]"
        )

    console.print()

    try:
        docker.compose.up(detach=True)
        console.print("[bold green]✅ Started required services[/bold green]\n")
        _wait_for_healthy(docker)
    except Exception as e:
        console.print(
            f"[bold red]⚠️ Some services may have failed to start: {e}[/bold red]"
        )
        raise SystemExit(1)

    console.print()
    k8s_manager.delete_jobs(settings.k8s_namespace, output_err=True)


def update_version(version: str):
    """Updates the version number across all relevant files."""
    # Update version in src/main.go Swagger annotation
    main_go_path = PROJECT_ROOT / "src" / "main.go"
    content = main_go_path.read_text(encoding="utf-8")
    content = re.sub(r"(@version\s+)\S+", rf"\g<1>{version}", content)
    main_go_path.write_text(content, encoding="utf-8")

    # Update version in src/config.dev.toml
    dev_config_path = PROJECT_ROOT / "src" / "config.dev.toml"
    with open(dev_config_path, encoding="utf-8") as f:
        dev_config = tomlkit.load(f)

    dev_config["version"] = version
    with open(dev_config_path, "w", encoding="utf-8") as f:
        tomlkit.dump(dev_config, f)

    # Update version in helm/values.yaml
    yaml = YAML()
    yaml.preserve_quotes = True
    yaml.indent(mapping=2, sequence=4, offset=2)

    helm_values_path = PROJECT_ROOT / "helm" / "values.yaml"
    with open(helm_values_path, encoding="utf-8") as f:
        helm_values = yaml.load(f)

    helm_values["configmap"]["version"] = version
    with open(helm_values_path, "w", encoding="utf-8") as f:
        yaml.dump(helm_values, f)

    # Update appVersion in helm/Chart.yaml
    chart_yaml_path = PROJECT_ROOT / "helm" / "Chart.yaml"
    with open(chart_yaml_path, encoding="utf-8") as f:
        chart = yaml.load(f)

    chart["appVersion"] = version
    with open(chart_yaml_path, "w", encoding="utf-8") as f:
        yaml.dump(chart, f)
