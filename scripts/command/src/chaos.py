from collections.abc import Callable

from src.common.common import ENV, console
from src.common.kubernetes_manager import KubernetesManager, with_k8s_manager

__all__ = ["clean_finalziers", "delete_chaos_resources"]

CHAOS_TYPES = [
    "dnschaos",
    "httpchaos",
    "jvmchaos",
    "networkchaos",
    "podchaos",
    "stresschaos",
    "timechaos",
]


@with_k8s_manager
def _execute_k8s_task(
    env: ENV,
    ns_prefix: str,
    ns_count: int,
    k8s_manager: KubernetesManager,
    action_func: Callable[[KubernetesManager, str, str, str], bool],
    success_msg: str,
    failure_msg: str,
    start_msg: str,
    end_msg: str,
):
    """Generic executor for Kubernetes chaos resource tasks."""
    console.print(start_msg)
    console.print(
        f"[gray]Dynamically getting namespaces with prefix {ns_prefix}...[/gray]"
    )

    namespaces = k8s_manager.list_namespaces(ns_prefix, limit=ns_count)
    console.print("[cyan]Found the following namespaces:[/cyan]")
    for ns in namespaces:
        console.print(f"  - [gray]{ns}[/gray]")
    console.print(f"[cyan]Total: {len(namespaces)} namespaces[/cyan]")

    # 3. Process each namespace and chaos resource
    for ns in namespaces:
        console.print(f"[bold blue]Processing namespace: {ns}[/bold blue]")
        for chaos_type in CHAOS_TYPES:
            try:
                resources = k8s_manager.list_chaos_resources(ns, chaos_type=chaos_type)
                if not resources:
                    console.print(
                        f"[bold yellow]No {chaos_type} resources found in namespace '{ns}'[/bold yellow]"
                    )
                    continue

                for res in resources:
                    if action_func(k8s_manager, ns, chaos_type, res):
                        console.print(
                            f"[gray]{success_msg.format(chaos_type=chaos_type, res=res, ns=ns)}[/gray]"
                        )
                    else:
                        console.print(
                            f"[bold yellow]{failure_msg.format(chaos_type=chaos_type, res=res, ns=ns)}[/bold yellow]"
                        )

            except Exception as e:
                console.print(
                    f"[red]Error processing {chaos_type} in namespace '{ns}': {e}[/red]"
                )

    console.print(end_msg)


def clean_chaos_finalziers(env: ENV, ns_prefix: str, ns_count: int):
    """Clean all chaos resource finalizers in namespaces with specific prefix"""

    def remove_finalizers_action(
        k8s_manager: KubernetesManager, ns: str, chaos_type: str, res: str
    ) -> bool:
        return k8s_manager.remove_finalizers(
            ns, chaos_type=chaos_type, resource_name=res
        )

    _execute_k8s_task(
        env,
        ns_prefix,
        ns_count,
        action_func=remove_finalizers_action,
        success_msg="Successfully removed finalizers from {chaos_type} '{res}' in namespace '{ns}'",
        failure_msg="No finalizers to remove for {chaos_type} '{res}' in namespace '{ns}'",
        start_msg="[bold blue]🧹 Cleaning chaos finalizers...[/bold blue]",
        end_msg="[bold green]✅ Finalizer cleanup completed![/bold green]",
    )


def delete_chaos_resources(env: ENV, ns_prefix: str, ns_count: int):
    """Delete chaos resources in namespaces with specific prefix"""

    def delete_resource_action(
        k8s_manager: KubernetesManager, ns: str, chaos_type: str, res: str
    ) -> bool:
        return k8s_manager.delete_chaos_resource(
            ns, chaos_type=chaos_type, resource_name=res
        )

    _execute_k8s_task(
        env,
        ns_prefix,
        ns_count,
        action_func=delete_resource_action,
        success_msg="Successfully deleted {chaos_type} '{res}' in namespace '{ns}'",
        failure_msg="Failed to delete {chaos_type} '{res}' in namespace '{ns}'",
        start_msg="[bold blue]🗑️ Deleting all chaos resources...[/bold blue]",
        end_msg="[bold green]✅ Chaos resource deletion completed![/bold green]",
    )
