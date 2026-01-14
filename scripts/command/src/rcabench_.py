from python_on_whales import DockerClient

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


@with_k8s_manager()
def local_deploy(env: ENV, k8s_manager: KubernetesManager):
    services = ["redis", "mysql", "jaeger", "buildkitd"]

    console.print("[bold blue]🚀 Starting local RCAbench deployment...[/bold blue]")

    docker = DockerClient(compose_files=[DOCKER_COMPOSE_FILE])
    try:
        docker.compose.down(remove_orphans=True)
        console.print("[bold green]✅ Cleaned up existing containers.[/bold green]")
    except Exception:
        console.print(
            "[bold yellow]⚠️ No existing containers to clean up.[/bold yellow]"
        )

    try:
        docker.compose.up(services=services, detach=True)
        console.print("[bold green]✅ Started required services.[/bold green]")
    except Exception as e:
        console.print(
            f"[bold red]⚠️ Some services may have failed to start: {e}[/bold red]"
        )
        raise SystemExit(1)

    console.print()
    k8s_manager.delete_jobs(settings.k8s_namespace, output_err=True)
