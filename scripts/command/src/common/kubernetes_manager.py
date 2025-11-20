import kubernetes
from kubernetes import config
from kubernetes.client.api import CoreV1Api
from kubernetes.client.rest import ApiException

from src.common import ENV, console

__all__ = ["KubernetesManager"]


class KubernetesManager:
    """Kubernetes API manager with safe initialization."""

    CONTEXT_MAPPING: dict[ENV, str] = {
        ENV.DEV: "kubernetes-admin@kubernetes",
        ENV.TEST: "k3d-test-cluster",
    }

    def __init__(self):
        """Initialize Kubernetes configuration and API client."""
        self._api: CoreV1Api | None = None
        self._initialize()

    def _initialize(self):
        """Try to initialize Kubernetes configuration."""
        try:
            config.load_kube_config()
            self._api = CoreV1Api()
        except config.ConfigException:
            try:
                config.load_incluster_config()
                self._api = CoreV1Api()
            except config.ConfigException:
                # No valid config available, _api remains None
                console.print(
                    "[bold yellow]Warning: No Kubernetes config found. K8s operations will be unavailable.[/bold yellow]"
                )

    @property
    def api(self) -> CoreV1Api | None:
        """Get the CoreV1Api instance, or None if not available."""
        return self._api

    def is_available(self) -> bool:
        """Check if Kubernetes API is available."""
        return self._api is not None

    def get_current_context(self) -> str:
        """Get the current Kubernetes context name."""
        try:
            contexts, active_context = config.list_kube_config_contexts()
            if active_context:
                return active_context["name"]
            return ""
        except config.ConfigException:
            # Running in-cluster, no context concept
            return "in-cluster"

    def get_current_context_cluster(self) -> str:
        """Get the current Kubernetes context's cluster name."""
        try:
            contexts, active_context = config.list_kube_config_contexts()
            if active_context:
                return active_context["context"]["cluster"]
            return ""
        except config.ConfigException:
            return "in-cluster"

    def switch_context(self, context_name: str) -> bool:
        """Switch to a specific Kubernetes context."""
        try:
            contexts, _ = config.list_kube_config_contexts()
            context_names = [ctx["name"] for ctx in contexts]  # type: ignore

            if context_name not in context_names:
                console.print(
                    f"[bold red]Context '{context_name}' not found[/bold red]"
                )
                console.print(f"Available contexts: {', '.join(context_names)}")
                return False

            # Switch context by reloading config with new context
            config.load_kube_config(context=context_name)

            # Reinitialize API client with new context
            self._api = CoreV1Api()

            console.print(
                f"[bold green]Switched to context: {context_name}[/bold green]"
            )
            return True

        except config.ConfigException as e:
            console.print(f"[bold red]Error switching context: {e}[/bold red]")
            return False

    def list_contexts(self) -> list[str]:
        """List all available Kubernetes contexts."""
        try:
            contexts, _ = config.list_kube_config_contexts()
            return [ctx["name"] for ctx in contexts]  # type: ignore
        except config.ConfigException:
            return []

    def check_and_create_namespace(self, namespace_name: str) -> bool:
        """Check if a Kubernetes Namespace exists; if not, create it."""
        if not self.is_available():
            console.print(
                "[bold red]Kubernetes API is not available. Cannot manage namespaces.[/bold red]"
            )
            return False

        try:
            self._api.read_namespace(name=namespace_name)  # type: ignore
            console.print(
                f"[bold green]Namespace {namespace_name} already exists.[/bold green]"
            )
            return True

        except ApiException as e:
            if e.status == 404:
                console.print(
                    f"[bold yellow]Namespace {namespace_name} does not exist. Creating it now...[/bold yellow]"
                )

                namespace_body = kubernetes.client.V1Namespace(
                    metadata=kubernetes.client.V1ObjectMeta(name=namespace_name)
                )

                try:
                    self._api.create_namespace(body=namespace_body)  # type: ignore
                    console.print(
                        f"[bold green]Successfully created Namespace: {namespace_name}[/bold green]"
                    )
                    return True
                except ApiException as create_e:
                    console.print(
                        f"[bold red]Error creating namespace {namespace_name}: {create_e}[/bold red]"
                    )
                    return False

            else:
                console.print(
                    f"[bold red]Error checking namespace {namespace_name}: {e}[/bold red]"
                )
                return False

        except Exception as e:
            console.print(f"[bold red]An unexpected error occurred: {e}[/bold red]")
            return False
