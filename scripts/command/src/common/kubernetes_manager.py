import time
from collections.abc import Callable
from functools import wraps
from pathlib import Path
from typing import Any, cast

import kubernetes
from kubernetes import config, watch
from kubernetes.client.api import AppsV1Api, BatchV1Api, CoreV1Api
from kubernetes.client.models import V1Deployment
from kubernetes.client.rest import ApiException
from pydantic import BaseModel

from src.common.command import run_command
from src.common.common import ENV, console, settings

__all__ = ["KubernetesManager", "with_k8s_manager"]


class K8sSessionData(BaseModel):
    """Session data for Kubernetes manager."""

    apps_api: AppsV1Api | None = None
    batch_api: BatchV1Api | None = None
    core_api: CoreV1Api | None = None
    context_name: str | None = None

    model_config = {"arbitrary_types_allowed": True}


class KubernetesManager:
    """Kubernetes API manager with safe initialization and singleton pattern per environment.

    Usage:
    with KubernetesManager(env=ENV.DEV) as k8s_manager:
        namespaces = k8s_manager.list_namespaces()
        print(f"Namespaces: {namespaces}")
    """

    _context_mapping: dict[ENV, str] = {}
    _instances: dict[ENV, "KubernetesManager"] = {}
    _sessions: dict[ENV, K8sSessionData] = {}

    def __new__(cls, env: ENV | None = None):
        """Create or return existing singleton instance for the given environment."""
        if env is None:
            # If no env provided, create a non-singleton instance (backward compatibility)
            instance = super().__new__(cls)
            instance._is_singleton = False
            return instance

        # Use env as unique identifier for singleton instances
        if env not in cls._instances:
            instance = super().__new__(cls)
            cls._instances[env] = instance
            instance._initialized = False
            instance._is_singleton = True

        return cls._instances[env]

    def __init__(self, env: ENV | None = None):
        """Initialize Kubernetes configuration and API client."""
        # Avoid duplicate initialization of the same instance
        if hasattr(self, "_initialized") and self._initialized:
            return

        self.env = env
        self._apps_api: AppsV1Api | None = None
        self._batch_api: BatchV1Api | None = None
        self._core_api: CoreV1Api | None = None

        # For non-singleton instances (backward compatibility)
        if not hasattr(self, "_is_singleton"):
            self._is_singleton = False
            self._initialize()
        else:
            self._initialized = True

    def __enter__(self):
        """Context manager entry: ensure session is initialized and switched to correct context."""
        if not self._is_singleton or self.env is None:
            # For non-singleton instances, just return self
            return self

        # Check if there is already a valid session
        if self.env not in self._sessions or not self._is_session_valid():
            self._initialize_session()

        # Load session data
        session_data = self._sessions[self.env]
        self._apps_api = session_data.apps_api
        self._batch_api = session_data.batch_api
        self._core_api = session_data.core_api

        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        """Context manager exit: maintain singleton state, do not close session."""
        pass

    def _is_session_valid(self) -> bool:
        """Check if the current session is valid."""
        if self.env is None:
            return False

        session_data = self._sessions.get(self.env)
        if not session_data:
            return False

        # Check if API clients are initialized and context is correct
        return (
            session_data.core_api is not None
            and session_data.context_name == self.get_context_mapping().get(self.env)
        )

    def _initialize_session(self):
        """Initialize a new session with proper context switching."""
        if self.env is None:
            raise ValueError(
                "Environment (env) must be provided for session initialization"
            )

        # Initialize configuration
        self._initialize()

        # Switch to the correct context for this environment
        target_context = self.get_context_mapping().get(self.env)
        if target_context:
            console.print(
                f"[bold blue]Switching to context: {target_context}[/bold blue]..."
            )
            if not self.switch_context(target_context):
                console.print(
                    f"[bold red]Failed to switch to context: {target_context}[/bold red]"
                )
                raise RuntimeError(f"Failed to switch to context: {target_context}")

            current_ctx = self.get_current_context()
            console.print(f"[bold green]Current context: {current_ctx}\n[/bold green]")

        # Store session information
        self._sessions[self.env] = K8sSessionData(
            apps_api=self._apps_api,
            batch_api=self._batch_api,
            core_api=self._core_api,
            context_name=target_context,
        )

    def _initialize(self):
        """Try to initialize Kubernetes configuration."""
        try:
            config.load_kube_config()
            self._apps_api = AppsV1Api()
            self._batch_api = BatchV1Api()
            self._core_api = CoreV1Api()
        except config.ConfigException:
            try:
                config.load_incluster_config()
                self._apps_api = AppsV1Api()
                self._batch_api = BatchV1Api()
                self._core_api = CoreV1Api()
            except config.ConfigException:
                # No valid config available, _api remains None
                console.print(
                    "[bold yellow]Warning: No Kubernetes config found. K8s operations will be unavailable.[/bold yellow]"
                )

    @classmethod
    def get_context_mapping(cls) -> dict[ENV, str]:
        """Load Kubernetes context mapping from settings."""
        if cls._context_mapping:
            return cls._context_mapping

        for env_member in ENV:
            setting_key = f"{env_member.name}_CONTEXT"

            context_name = settings.get(setting_key)

            if context_name:
                cls._context_mapping[env_member] = str(context_name)

        return cls._context_mapping

    @classmethod
    def clear_sessions(cls):
        """Clear all cached sessions and instances."""
        cls._sessions.clear()
        cls._instances.clear()

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

    def check_pod(
        self,
        name: str,
        namespace: str,
        label_selector: str,
        field_selector: str,
        output_error: bool = False,
    ) -> bool:
        """Check if a pod with specific criteria is running in the cluster."""
        assert self._core_api is not None, "Kubernetes API is not initialized"

        try:
            pods = self._core_api.list_namespaced_pod(
                namespace=namespace,
                label_selector=label_selector,
                field_selector=field_selector,
            )

            for pod in pods.items:
                if pod.name == name:
                    return True

            return False

        except ApiException as e:
            if output_error:
                console.print(f"[bold red]Error checking running pods: {e}[/bold red]")
            return False

    def get_current_context_cluster(self) -> str:
        """Get the current Kubernetes context's cluster name."""
        try:
            contexts, active_context = config.list_kube_config_contexts()
            if active_context:
                return active_context["context"]["cluster"]
            return ""
        except config.ConfigException:
            return "in-cluster"

    def get_node_access_url(self, port: int) -> str:
        assert self._core_api is not None, "Kubernetes API is not initialized"

        try:
            nodes = self._core_api.list_node()
            if not nodes.items:
                return ""

            # Get the first node's addresses
            node = nodes.items[0]
            for address in node.status.addresses:
                if address.type == "ExternalIP":
                    return f"{address.address}:{port}"
            for address in node.status.addresses:
                if address.type == "InternalIP":
                    return f"{address.address}:{port}"

            return ""
        except ApiException as e:
            console.print(f"[bold red]Error retrieving node access URL: {e}[/bold red]")
            return ""

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
            self._core_api = CoreV1Api()

            return True

        except config.ConfigException as e:
            console.print(f"[bold red]Error switching context: {e}[/bold red]")
            return False

    def delete_namespace(self, namespace: str) -> bool:
        """Delete a specific Kubernetes namespace."""
        assert self._core_api is not None, "Kubernetes API is not initialized"

        try:
            self._core_api.delete_namespace(name=namespace)
            return True

        except ApiException as e:
            console.print(
                f"[bold red]Error deleting namespace {namespace}: {e}[/bold red]"
            )
            return False

    def list_chaos_resources(self, namespace: str, chaos_type: str) -> list[str]:
        """List all chaos resources of a specific type in a given namespace."""
        assert self._core_api is not None, "Kubernetes API is not initialized"

        try:
            custom_api = kubernetes.client.CustomObjectsApi()
            group = "chaos-mesh.org"
            version = "v1alpha1"
            plural = chaos_type + "s"  # e.g., podchaos -> podchaoss

            resources = custom_api.list_namespaced_custom_object(
                group=group,
                version=version,
                namespace=namespace,
                plural=plural,
            )

            resource_names = [
                item["metadata"]["name"] for item in resources.get("items", [])
            ]

            return resource_names

        except ApiException:
            return []

    def list_contexts(self) -> list[str]:
        """List all available Kubernetes contexts."""
        try:
            contexts, _ = config.list_kube_config_contexts()
            return [ctx["name"] for ctx in contexts]  # type: ignore
        except config.ConfigException:
            return []

    def list_namespaces(
        self, prefix: str | None = None, limit: int | None = None
    ) -> list[str]:
        """List all Kubernetes Namespaces, optionally filtered by prefix and limited in number."""
        assert self._core_api is not None, "Kubernetes API is not initialized"

        try:
            namespaces = self._core_api.list_namespace()
            namespace_names = [
                ns.metadata.name
                for ns in namespaces.items
                if prefix is None or ns.metadata.name.startswith(prefix)
            ]

            if limit is not None:
                namespace_names = namespace_names[:limit]

            return namespace_names

        except ApiException as e:
            console.print(f"[bold red]Error listing namespaces: {e}[/bold red]")
            return []

    def list_secrets(self, namespace: str) -> list[str]:
        """List all Secrets in a given namespace."""
        assert self._core_api is not None, "Kubernetes API is not initialized"

        try:
            secrets = self._core_api.list_namespaced_secret(namespace=namespace)
            secret_names = [secret.metadata.name for secret in secrets.items]
            return secret_names
        except ApiException as e:
            console.print(
                f"[bold red]Error listing secrets in namespace {namespace}: {e}[/bold red]"
            )
            return []

    def check_and_create_namespace(self, namespace_name: str) -> bool:
        """Check if a Kubernetes Namespace exists; if not, create it."""
        assert self._core_api is not None, "Kubernetes API is not initialized"

        try:
            self._core_api.read_namespace(name=namespace_name)
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
                    self._core_api.create_namespace(body=namespace_body)  # type: ignore
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

    def remove_finalizers(
        self, namespace: str, chaos_type: str, resource_name: str
    ) -> bool:
        """Remove finalizers from a specific chaos resource."""
        assert self._core_api is not None, "Kubernetes API is not initialized"

        try:
            custom_api = kubernetes.client.CustomObjectsApi()
            group = "chaos-mesh.org"
            version = "v1alpha1"
            plural = chaos_type + "s"  # e.g., podchaos -> podchaoss

            body = {"metadata": {"finalizers": [], "resourceVersion": ""}}

            custom_api.patch_namespaced_custom_object(
                group=group,
                version=version,
                namespace=namespace,
                plural=plural,
                name=resource_name,
                body=body,
            )

            return True

        except ApiException as e:
            console.print(
                f"[bold red]Error removing finalizers from {chaos_type} '{resource_name}' in namespace '{namespace}': {e}[/bold red]"
            )
            return False

    def delete_chaos_resource(
        self,
        namespace: str,
        chaos_type: str,
        resource_name: str,
        output_err: bool = False,
    ) -> bool:
        """Delete a specific chaos resource."""
        assert self._core_api is not None, "Kubernetes API is not initialized"

        try:
            custom_api = kubernetes.client.CustomObjectsApi()
            group = "chaos-mesh.org"
            version = "v1alpha1"
            plural = chaos_type + "s"  # e.g., podchaos -> podchaoss

            custom_api.delete_namespaced_custom_object(
                group=group,
                version=version,
                namespace=namespace,
                plural=plural,
                name=resource_name,
                body=kubernetes.client.V1DeleteOptions(),
            )

            console.print(
                f"[gray]Successfully deleted {chaos_type} '{resource_name}' in namespace '{namespace}'[/gray]"
            )
            return True

        except ApiException as e:
            if output_err:
                console.print(
                    f"[bold red]Error deleting {chaos_type} '{resource_name}' in namespace '{namespace}': {e}[/bold red]"
                )
            return False

    def delete_jobs(self, namespace: str, output_err: bool = False) -> bool:
        """Delete all Jobs in a given namespace."""
        assert self._batch_api is not None, "Kubernetes Batch API is not initialized"

        try:
            jobs = self._batch_api.list_namespaced_job(namespace=namespace)

            if len(jobs.items) == 0:
                console.print(
                    f"[bold yellow]No Jobs found in namespace '{namespace}' to delete.[/bold yellow]"
                )
                return True

            for job in jobs.items:
                job_name = job.metadata.name
                self._batch_api.delete_namespaced_job(
                    name=job_name,
                    namespace=namespace,
                    body=kubernetes.client.V1DeleteOptions(
                        propagation_policy="Foreground"
                    ),
                )

                console.print(
                    f"[gray]Successfully deleted Job '{job_name}' in namespace '{namespace}'[/gray]"
                )

            console.print(
                f"[bold green]All {len(jobs.items)} Jobs deleted in namespace '{namespace}'[/bold green]"
            )

            return True

        except ApiException as e:
            if output_err:
                console.print(
                    f"[bold red]Error deleting jobs in namespace '{namespace}': {e}[/bold red]"
                )
            return False

    def watch_deployments_ready(
        self,
        names: list[str],
        namespace: str,
        timeout_seconds: int = 300,
    ) -> bool:
        """Watch and wait for multiple Deployments to become ready using Kubernetes Watch API.

        Args:
            names: List of Deployment names to watch.
            namespace: The namespace where the Deployments are located.
            timeout_seconds: Maximum time to wait for all Deployments to become ready.

        Returns:
            True if all Deployments become ready within the timeout, False otherwise.
        """
        assert self._apps_api is not None, "Kubernetes Apps API is not initialized"

        if not names:
            return True

        pending_deployments = set(names)
        w = watch.Watch()

        try:
            stream = w.stream(
                self._apps_api.list_namespaced_deployment,
                namespace=namespace,
                timeout_seconds=timeout_seconds,
            )

            for event in stream:
                event = cast(dict[str, Any], event)
                deployment = cast(V1Deployment, event["object"])

                metadata = deployment.metadata
                status = deployment.status
                spec = deployment.spec
                if metadata is None or status is None or spec is None:
                    continue

                deployment_name = metadata.name

                # Skip if not in our watch list
                if deployment_name not in pending_deployments:
                    continue

                spec_replicas = spec.replicas or 0

                # Check if this Deployment is ready
                if (
                    status.replicas is not None
                    and status.ready_replicas is not None
                    and status.available_replicas is not None
                    and status.replicas == spec_replicas
                    and status.ready_replicas == spec_replicas
                    and status.available_replicas == spec_replicas
                ):
                    pending_deployments.discard(deployment_name)
                    console.print(
                        f"[bold gray]   - Deployment '{deployment_name}' is ready. "
                        f"Remaining: {len(pending_deployments)}[/bold gray]"
                    )

                    if not pending_deployments:
                        w.stop()
                        console.print(
                            f"[bold green]All {len(names)} Deployments in namespace '{namespace}' are ready.[/bold green]"
                        )
                        return True

        except ApiException as e:
            console.print(
                f"[bold red]API error watching Deployments: {e.reason}[/bold red]"
            )
            return False
        finally:
            w.stop()

        console.print(
            f"[bold red]Timeout waiting for Deployments in namespace '{namespace}'. "
            f"Still pending: {pending_deployments}[/bold red]"
        )
        return False

    def watch_all_deployments_ready(
        self,
        namespace: str,
        timeout_seconds: int = 300,
    ) -> bool:
        """Watch and wait for ALL Deployments in a namespace to become ready.

        This method first lists all Deployments in the namespace, then watches
        until all of them are ready or timeout occurs.

        Args:
            namespace: The namespace to watch all Deployments in.
            timeout_seconds: Maximum time to wait for all Deployments to become ready.

        Returns:
            True if all Deployments become ready within the timeout, False otherwise.
        """
        assert self._apps_api is not None, "Kubernetes Apps API is not initialized"

        try:
            # First, get all deployment names in the namespace
            deployments = self._apps_api.list_namespaced_deployment(namespace=namespace)
            deployment_names = [
                d.metadata.name
                for d in deployments.items
                if d.metadata is not None and d.metadata.name is not None
            ]

            if not deployment_names:
                console.print(
                    f"[bold yellow]No Deployments found in namespace '{namespace}'.[/bold yellow]"
                )
                return True

            console.print(
                f"[bold blue]Watching {len(deployment_names)} Deployments in namespace '{namespace}': "
                f"{', '.join(deployment_names)}[/bold blue]"
            )

            # Use the existing method to watch these deployments
            return self.watch_deployments_ready(
                names=deployment_names,
                namespace=namespace,
                timeout_seconds=timeout_seconds,
            )

        except ApiException as e:
            console.print(
                f"[bold red]API error listing Deployments in namespace '{namespace}': {e.reason}[/bold red]"
            )
            return False


def kubectl_apply(manifest: str | Path) -> bool:
    """Apply a Kubernetes manifest using kubectl."""
    manifest_str = str(manifest)
    console.print(f"[bold blue]Applying manifest: {manifest_str}[/bold blue]")
    try:
        run_command(["kubectl", "apply", "-f", manifest_str], check=True)
        return True
    except SystemExit:
        return False


def with_k8s_manager(func: Callable) -> Callable:
    """Decorator to ensure KubernetesManager is available for the decorated function.

    The decorator will automatically create or reuse a singleton KubernetesManager instance
    for the given environment and pass it to the decorated function.
    """

    @wraps(func)
    def wrapper(*args, **kwargs) -> Any:
        env = None
        if "env" in kwargs:
            env = kwargs["env"]
        elif args:
            env = args[0]

        if env is None:
            console.print(
                "[red]❌ Decorator error: Function must accept 'env' argument.[/red]"
            )
            raise SystemExit(1)

        # If k8s_manager is already provided, use it directly
        if "k8s_manager" in kwargs:
            return func(*args, **kwargs)

        try:
            with KubernetesManager(env=env) as k8s_manager:
                if k8s_manager is None:
                    console.print(
                        "[red]❌ Kubernetes is not available or not configured properly. (Check context/config)[/red]"
                    )
                    raise SystemExit(1)

                return func(*args, k8s_manager=k8s_manager, **kwargs)
        except RuntimeError as e:
            console.print(
                f"[bold red]Error initializing KubernetesManager: {e}[/bold red]"
            )
            raise SystemExit(1)

    return wrapper
