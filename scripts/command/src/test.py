from concurrent.futures import ThreadPoolExecutor, as_completed

from src.common.common import ENV, PROJECT_ROOT, console
from src.common.helm_cli import HelmCLI, HelmRelease
from src.common.kubernetes_manager import (
    KubernetesManager,
    kubectl_apply,
    with_k8s_manager,
)
from src.common.minikube_cli import MinikubeCLI, MinikubeConfig

__all__ = ["setup_env", "teardown_env"]

MANIFESTS_DIR = PROJECT_ROOT / "manifests"
LOCAL_DEV_DIR = MANIFESTS_DIR / "local-dev"


def setup_env(
    nodes: int = 3,
    cpus: int = 2,
    memory: str = "4g",
    skip_minikube: bool = False,
    is_ci: bool = False,
):
    """Set up test environment.

    Args:
        nodes: Number of Minikube nodes
        cpus: CPUs per node
        memory: Memory per node
        skip_minikube: Skip Minikube cluster creation
        is_ci: Whether running in a CI environment
    """
    console.print("[bold blue]🚀 Setting up test environment...[/bold blue]\n")

    # Step 1: Start Minikube
    _install_minikube(nodes, cpus, memory, skip_minikube)

    console.print()

    # Step 2: Install Helm releases
    _install_helm_releases(ENV.TEST, is_ci=is_ci)

    console.print("[bold green]✅ Test environment setup complete![/bold green]")


def _install_minikube(nodes: int, cpus: int, memory: str, skip_minikube: bool):
    minikube_cli = MinikubeCLI()
    if skip_minikube:
        console.print("[bold yellow]Skipping Minikube startup...[/yellow]")
    else:
        config = MinikubeConfig(
            nodes=nodes,
            driver="docker",
            cpus=cpus,
            memory=memory,
        )
        if not minikube_cli.start(config):
            console.print("[bold red]❌ Failed to start Minikube[/bold red]")
            raise SystemExit(1)

    with KubernetesManager(env=ENV.TEST) as k8s_manager:
        ready = k8s_manager.watch_deployments_ready(
            ["coredns"], namespace="kube-system"
        )
        if not ready:
            console.print(
                "[bold red]❌ Minikube deployments failed to become ready in time[/bold red]"
            )
            raise SystemExit(1)
        console.print("[bold green]✅ Minikube is up and running[/bold green]")


@with_k8s_manager()
def _install_helm_releases(env: ENV, k8s_manager: KubernetesManager, is_ci: bool):
    helm_cli = HelmCLI()

    # Install cert-manager (prerequisite for otel-kube-stack)
    console.print("[bold blue]📦 Installing cert-manager...[/bold blue]")
    kubectl_apply(
        "https://github.com/cert-manager/cert-manager/releases/latest/download/cert-manager.yaml"
    )

    if not k8s_manager.watch_deployments_ready(
        ["cert-manager"], namespace="cert-manager", timeout_seconds=300
    ):
        console.print(
            "[bold red]❌ cert-manager deployments failed to become ready in time[/bold red]"
        )
        raise SystemExit(1)

    console.print()

    releases = _get_helm_releases()
    not_installed_releases: list[HelmRelease] = []

    for release in releases:
        if release.create_namespace:
            k8s_manager.check_and_create_namespace(release.namespace)

        if not helm_cli.install(release):
            not_installed_releases.append(release)

        if release.name == "cilium":
            # Apply Cilium metrics
            kubectl_apply(MANIFESTS_DIR / "cilium" / "metrics.yaml")

        console.print()

    if not_installed_releases:
        if is_ci:
            console.print(
                f"[bold red]❌ The following releases failed to install: "
                f"{', '.join([r.name for r in not_installed_releases])}[/bold red]"
            )
            raise SystemExit(1)
        else:
            console.print(
                "[bold yellow]⚠️ Some releases failed to install: "
                f"{', '.join([r.name for r in not_installed_releases])}[/bold yellow]"
            )

    console.print(
        "[bold blue]⏳ Waiting for all deployments to be ready...[/bold blue]"
    )

    namespaces_to_check: list = []
    for release in releases:
        if release.namespace not in namespaces_to_check:
            namespaces_to_check.append(release.namespace)

    def watch_namespace(ns: str, timeout_seconds: int) -> tuple[str, bool]:
        """Watch a single namespace and return (namespace, success)."""
        console.print(f"[dim]Watching namespace: {ns}[/dim]")
        success = k8s_manager.watch_all_deployments_ready(
            ns, timeout_seconds=timeout_seconds
        )
        return (ns, success)

    failed_namespaces: list[str] = []

    with ThreadPoolExecutor(max_workers=len(namespaces_to_check)) as executor:
        futures = {
            executor.submit(watch_namespace, ns, 600): ns for ns in namespaces_to_check
        }

        for future in as_completed(futures):
            ns, success = future.result()
            if not success:
                failed_namespaces.append(ns)

    if failed_namespaces:
        console.print(
            f"[bold red]❌ Deployments failed in namespaces: {', '.join(failed_namespaces)}[/bold red]"
        )
        raise SystemExit(1)

    if not not_installed_releases:
        console.print("[bold green]✅ All deployments are ready![/bold green]")

    console.print("\n[cyan]Installed components:[/cyan]")
    for release in releases:
        if release not in not_installed_releases:
            console.print(
                f"[bold gray]    - {release.name} in namespace '{release.namespace}'[bold gray]"
            )


@with_k8s_manager()
def teardown_env(
    env: ENV, k8s_manager: KubernetesManager, delete_cluster: bool = False
):
    """Tear down test environment.

    Args:
        delete_cluster: Whether to delete the Minikube cluster
    """
    console.print("[bold blue]🧹 Tearing down test environment...[/bold blue]\n")

    helm = HelmCLI()
    minikube = MinikubeCLI()

    # Uninstall Helm releases in reverse order
    releases = _get_helm_releases()
    for release in reversed(releases):
        helm.uninstall(release.name, release.namespace)

    # Delete namespaces using KubernetesManager
    namespaces_to_delete = ["od", "monitoring", "chaos-mesh"]
    for ns in namespaces_to_delete:
        try:
            if k8s_manager.delete_namespace(ns):
                console.print(f"[dim]    - Deleted namespace: {ns}[/dim]")
        except Exception as e:
            console.print(f"[yellow]Failed to delete namespace {ns}: {e}[/yellow]")

    # Delete Minikube cluster if requested
    if delete_cluster:
        minikube.delete()

    console.print("\n[bold green]✅ Teardown complete![/bold green]")


def _get_helm_releases() -> list[HelmRelease]:
    """Define all Helm releases for test development."""
    return [
        # Chaos Mesh
        HelmRelease(
            name="chaos-mesh",
            chart="chaos-mesh/chaos-mesh",
            namespace="chaos-mesh",
            repo_name="chaos-mesh",
            repo_url="https://charts.chaos-mesh.org",
            version="2.8.0",
            create_namespace=True,
        ),
        # Cilium
        HelmRelease(
            name="cilium",
            chart="cilium/cilium",
            namespace="kube-system",
            repo_name="cilium",
            repo_url="https://helm.cilium.io/",
            version="1.18.4",
        ),
        # OpenTelemetry Kube Stack
        HelmRelease(
            name="opentelemetry-kube-stack",
            chart="open-telemetry/opentelemetry-kube-stack",
            namespace="monitoring",
            repo_name="open-telemetry",
            repo_url="https://open-telemetry.github.io/opentelemetry-helm-charts",
            values_file=LOCAL_DEV_DIR / "otel-kube-stack.yaml",
            create_namespace=True,
        ),
        # ClickStack (ClickHouse)
        HelmRelease(
            name="clickstack",
            chart="clickstack/clickstack",
            namespace="monitoring",
            repo_name="clickstack",
            repo_url="https://hyperdxio.github.io/helm-charts",
            values_file=LOCAL_DEV_DIR / "click-stack.yaml",
        ),
        # OpenTelemetry Demo
        HelmRelease(
            name="otel-demo",
            chart="open-telemetry/opentelemetry-demo",
            namespace="od",
            values_file=LOCAL_DEV_DIR / "otel-demo-values.yaml",
            create_namespace=True,
            extra_args=["--set", "prometheus.rbac.create=false"],
        ),
    ]
