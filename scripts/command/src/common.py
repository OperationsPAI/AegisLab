import os
from enum import Enum

import kubernetes.client
from git import Repo
from kubernetes import config
from kubernetes.client.rest import ApiException
from rich.console import Console

DEFAULT_SDK_DIR = "sdk/python"

INITIAL_DATA_PATH = os.path.join(os.getcwd(), "helm", "files/initial_data", "data.json")

console = Console()  # Initialize a global console object for rich output
repo = Repo(".")  # Initialize the git repository at the current directory

try:
    config.load_kube_config()
except config.ConfigException:
    config.load_incluster_config()

api = kubernetes.client.CoreV1Api()


def get_current_context() -> str:
    """Get the current Kubernetes context name."""
    try:
        contexts, active_context = config.list_kube_config_contexts()
        if active_context:
            return active_context["name"]
        return ""
    except config.ConfigException:
        # Running in-cluster, no context concept
        return "in-cluster"


def get_current_context_cluster() -> str:
    """Get the current Kubernetes context's cluster name."""
    try:
        contexts, active_context = config.list_kube_config_contexts()
        if active_context:
            return active_context["context"]["cluster"]
        return ""
    except config.ConfigException:
        return "in-cluster"


def switch_context(context_name: str) -> bool:
    """Switch to a specific Kubernetes context."""
    try:
        contexts, _ = config.list_kube_config_contexts()
        context_names = [ctx["name"] for ctx in contexts]  # type: ignore

        if context_name not in context_names:
            console.print(f"[bold red]Context '{context_name}' not found[/bold red]")
            console.print(f"Available contexts: {', '.join(context_names)}")
            return False

        # Switch context by reloading config with new context
        config.load_kube_config(context=context_name)

        # Reinitialize API client with new context
        global api
        api = kubernetes.client.CoreV1Api()

        console.print(f"[bold green]Switched to context: {context_name}[/bold green]")
        return True

    except config.ConfigException as e:
        console.print(f"[bold red]Error switching context: {e}[/bold red]")
        return False


def list_contexts() -> list[str]:
    """List all available Kubernetes contexts."""
    try:
        contexts, _ = config.list_kube_config_contexts()
        return [ctx["name"] for ctx in contexts]  # type: ignore
    except config.ConfigException:
        return []


def check_and_create_namespace(namespace_name: str):
    """Check if a Kubernetes Namespace exists; if not, create it."""
    try:
        api.read_namespace(name=namespace_name)
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
                api.create_namespace(body=namespace_body)
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
        print(f"[bold red]An unexpected error occurred: {e}[/bold red]")
        return False


class ENV(str, Enum):
    DEV = "dev"
    PROD = "prod"
    TEST = "test"
