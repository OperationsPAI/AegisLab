from io import StringIO

import yaml
from python_on_whales import DockerClient

from src.common.command import run_command
from src.common.common import (
    DEFAULT_PORT,
    DEFAULT_REPO_URL,
    ENV,
    HELM_CHART_PATH,
    K8S_NAMESPACE,
    PROJECT_ROOT,
    RELEASE_NAME,
    console,
)
from src.common.kubernetes_manager import KubernetesManager, with_k8s_manager

__all__ = ["check_secrets", "execute_release_workflow"]

DOCKER_COMPOSE_FILE = PROJECT_ROOT / "docker-compose.yaml"


@with_k8s_manager
def check_db(env: ENV, k8s_manager: KubernetesManager):
    """Checks the health of the RCABench database."""
    _check_pod_health(k8s_manager, "RCABench MySQL Database", "rcabench-mysql")


@with_k8s_manager
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
        namespace=K8S_NAMESPACE,
        label_selector=f"app={pod_name}",
        field_selector="status.phase=Running",
    )

    if is_running:
        console.print(f"[bold green]✅ {service_name} is running[/bold green]")
        return

    console.print(
        f"[bold red]❌ {service_name} is NOT running[/bold red] "
        f"in namespace [yellow]'{K8S_NAMESPACE}'[/yellow]."
    )
    raise SystemExit(1)


@with_k8s_manager
def check_secrets(env: ENV, k8s_manager: KubernetesManager):
    """Checks for the presence of required Kubernetes secrets."""
    console.print(
        "🔍 [bold blue]Checking for required Kubernetes secrets...[/bold blue]"
    )
    console.print("[gray]Extracting Secret names from Helm templates...[/gray]")

    helm_cmd = [
        "helm",
        "template",
        RELEASE_NAME,
        HELM_CHART_PATH.as_posix(),
        "-n",
        K8S_NAMESPACE,
        "-s",
        "templates/secret.yaml",
    ]

    yaml_output = run_command(helm_cmd, capture_output=True).stdout
    if not yaml_output:
        console.print(
            "[red]❌ No Secret templates found in Helm chart. Skipping secret check.[/red]"
        )
        raise SystemExit(1)

    secret_name_set: set[str] = set()
    try:
        for doc in yaml.load_all(StringIO(yaml_output), Loader=yaml.SafeLoader):  # type: ignore
            if isinstance(doc, dict) and doc.get("kind") == "Secret":
                metadata = doc.get("metadata", {})
                name = metadata.get("name")
                if name:
                    secret_name_set.add(name)
    except yaml.YAMLError as e:
        console.print(
            f"[bold yellow]Failed to parse Helm template output as YAML: {e}[/bold yellow]"
        )
        return []

    required_secrets = sorted(list(secret_name_set))
    if not required_secrets:
        console.print(
            "[red]❌ No Secret templates found in Helm chart. Skipping secret check.[/red]"
        )
        raise SystemExit(1)

    console.print("[cyan]Required Secrets:[/cyan]")
    for secret in required_secrets:
        console.print(f"[gray]  - {secret}[/gray]")
    console.print()

    console.print(
        "[bold blue]🔍 Checking existing Secrets in the cluster...[/bold blue]"
    )

    all_ok = True
    existing_secrets = k8s_manager.list_secrets(K8S_NAMESPACE)
    for secret in required_secrets:
        if secret in existing_secrets:
            console.print(f"[gray]✅Found Secret: {secret}[/gray]")
        else:
            console.print(f"[gray]❌ Missing Secret: {secret}[/gray]")
            all_ok = False

    if not all_ok:
        console.print(
            "[red]❌ Some unrequired Secrets were found. Please run 'make install-secrets' first.[/red]"
        )
        raise SystemExit(1)


@with_k8s_manager
def local_deploy(env: ENV, k8s_manager: KubernetesManager):
    console.print("[bold blue]🚀 Starting local RCAbench deployment...[/bold blue]")

    docker = DockerClient(compose_files=[DOCKER_COMPOSE_FILE])
    try:
        docker.compose.down(remove_orphans=True)
    except Exception:
        console.print(
            "[bold yellow]⚠️ No existing containers to clean up.[/bold yellow]"
        )

    try:
        docker.compose.up(
            services=["redis", "mysql", "jaeger", "buildkitd"], detach=True
        )
    except Exception as e:
        console.print(
            f"[bold red]⚠️ Some services may have failed to start: {e}[/bold red]"
        )
        raise SystemExit(1)

    console.print()
    k8s_manager.delete_jobs(K8S_NAMESPACE, output_err=True)


@with_k8s_manager
def execute_release_workflow(env: ENV, k8s_manager: KubernetesManager):
    console.print("[bold blue]🚀 Executing RCAbench release workflow...[/bold blue]")

    console.print("[bold blue]Step 1: Verifying Secrets...[/bold blue]")
    check_secrets(env, k8s_manager=k8s_manager)

    console.print()
    console.print("[bold blue]Step 2: Deploying with Skaffold...[/bold blue]")
    run_command(["skaffold", "run", f"--default-repo={DEFAULT_REPO_URL}"])

    console.print()
    console.print("[bold blue]Step 3: Waiting for deployment...[/bold blue]")
    k8s_manager.wait_for_all_deployments_available(K8S_NAMESPACE)

    console.print()
    console.print(
        "[bold green]✅ RCAbench release workflow completed successfully![/bold green]"
    )
    console.print("[cyan]Deployment Summary:[/cyan]")
    console.print(f"  - Namespace: {K8S_NAMESPACE}")
    console.print(f"  - Release Name: {RELEASE_NAME}")
    console.print(f"  - Access URL: {k8s_manager.get_node_access_url(DEFAULT_PORT)}")
