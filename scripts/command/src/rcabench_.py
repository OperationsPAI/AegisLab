import yaml
from python_on_whales import DockerClient

from src.common.command import run_command
from src.common.common import ENV, PROJECT_ROOT, console, settings
from src.common.kubernetes_manager import KubernetesManager, with_k8s_manager

__all__ = ["execute_release_workflow"]

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


@with_k8s_manager()
def execute_release_workflow(env: ENV, k8s_manager: KubernetesManager):
    console.print("[bold blue]🚀 Executing RCAbench release workflow...[/bold blue]")
    settings.setenv(env.value)

    console.print()
    console.print("[bold blue]Step 1: Deploying with Skaffold...[/bold blue]")

    with open(SKAFFOLD_FILE) as f:
        skaffold_config = yaml.safe_load(f)
    original_build_image = skaffold_config["build"]["artifacts"][0]["image"]
    original_image_repo = skaffold_config["deploy"]["helm"]["releases"][0]["setValues"][
        "image.repository"
    ]

    new_build_image = settings.default_repo
    new_image_repo = settings.default_repo

    skaffold_config["build"]["artifacts"][0]["image"] = new_build_image
    skaffold_config["deploy"]["helm"]["releases"][0]["setValues"][
        "image.repository"
    ] = new_image_repo

    with open("skaffold.yaml", "w", encoding="utf-8") as f:
        yaml.dump(skaffold_config, f, sort_keys=False)

    run_command(["skaffold", "run", f"--default-repo={settings.default_repo}"])

    skaffold_config["build"]["artifacts"][0]["image"] = original_build_image
    skaffold_config["deploy"]["helm"]["releases"][0]["setValues"][
        "image.repository"
    ] = original_image_repo

    with open(SKAFFOLD_FILE, "w", encoding="utf-8") as f:
        yaml.dump(skaffold_config, f, sort_keys=False)

    console.print()
    console.print("[bold blue]Step 2: Waiting for deployment...[/bold blue]")
    k8s_manager.watch_all_deployments_ready(settings.k8s_namespace)

    console.print()
    console.print(
        "[bold green]✅ RCAbench release workflow completed successfully![/bold green]"
    )
    console.print("[cyan]Deployment Summary:[/cyan]")
    console.print(f"  - Namespace: {settings.k8s_namespace}")
    console.print(f"  - Release Name: {settings.release_name}")
    console.print(
        f"  - Access URL: {k8s_manager.get_node_access_url(settings.default_port)}"
    )
