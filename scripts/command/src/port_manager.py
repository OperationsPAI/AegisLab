"""Kubernetes Port Forwarding Manager

This module provides Kubernetes service port forwarding functionality with support for:
- Auto-discovery and forwarding of all services in a namespace
- Port prefix configuration (prod: 1xxxx, test: 2xxxx)
- Port overflow protection
- Dynamic port mapping retrieval (for testing)
"""

import signal
import subprocess
import time
from enum import Enum
from typing import Any

import psutil
from pydantic import BaseModel

from src.common.common import ENV, console
from src.common.kubernetes_manager import KubernetesManager, with_k8s_manager


class PortPrefix(Enum):
    """Port prefix configuration"""

    PROD = "1"
    TEST = "2"


class PortMapping(BaseModel):
    """Port mapping information"""

    service: str
    namespace: str
    remote_port: int
    local_port: int
    pid: int | None = None

    def get_url(self, protocol: str = "http") -> str:
        """Get local access URL"""
        return f"{protocol}://localhost:{self.local_port}"


class PortForwardManager:
    """Port Forward Manager

    Usage:
        manager = PortForwardManager(env=ENV.TEST)
        manager.start_forwarding()

        url = manager.get_service_url("rcabench", namespace="exp")

        manager.stop_all_forwards()
    """

    def __init__(self, env: ENV):
        """Initialize port forward manager

        Args:
            env: Environment type (TEST or PROD)
        """
        self.env = env
        self.prefix = (
            PortPrefix.TEST.value if env == ENV.TEST else PortPrefix.PROD.value
        )
        self.port_mappings: list[PortMapping] = []
        self.k8s_manager: KubernetesManager | None = None

        console.print(f"Environment: {env.value}")
        console.print(f"Port prefix: {self.prefix}xxxx\n")

    def cleanup_existing_forwards(self):
        """Cleanup existing port forward processes"""
        console.print("[bold blue]🧹 Cleaning up old port forwards...[/bold blue]")

        # Kill kubectl port-forward processes
        killed_count = 0
        for proc in psutil.process_iter(["pid", "name", "cmdline"]):
            try:
                cmdline = proc.info.get("cmdline") or []
                cmdline_str = " ".join(str(c) for c in cmdline)
                if "kubectl" in cmdline_str and "port-forward" in cmdline_str:
                    proc.kill()
                    killed_count += 1
            except (psutil.NoSuchProcess, psutil.AccessDenied, psutil.ZombieProcess):
                continue

        if killed_count > 0:
            console.print(f"   Killed {killed_count} kubectl port-forward process(es)")

        # Kill processes on our port range
        prefix_int = int(self.prefix)
        port_killed_count = 0
        for conn in psutil.net_connections(kind="inet"):
            try:
                if conn.status == "LISTEN" and conn.laddr:
                    port = conn.laddr.port
                    # Check if port is in our range (prefix0000 to prefix9999)
                    if prefix_int * 10000 <= port < (prefix_int + 1) * 10000:
                        proc = psutil.Process(conn.pid)
                        proc.kill()
                        port_killed_count += 1
            except (psutil.NoSuchProcess, psutil.AccessDenied, psutil.ZombieProcess):
                continue

        if port_killed_count > 0:
            console.print(
                f"   Killed {port_killed_count} process(es) on {self.prefix}xxxx ports"
            )

        time.sleep(2)
        console.print("[bold green]✅ Old forwards cleaned[/bold green]\n")

    def _calculate_local_port(self, remote_port: int) -> int:
        """Calculate local port with overflow protection

        Args:
            remote_port: Remote service port

        Returns:
            Local mapped port
        """
        local_port = int(f"{self.prefix}{remote_port}")

        # Handle port number overflow (>65535)
        if local_port > 65535:
            # Map to prefix0000 + (port % 55535)
            local_port = int(self.prefix) * 10000 + (remote_port % 55535)

        return local_port

    @with_k8s_manager()
    def forward_namespace_services(
        self,
        env: ENV,
        namespace: str,
        k8s_manager: KubernetesManager,
        label: str | None = None,
    ) -> dict[str, list[PortMapping]]:
        """Forward all services in a namespace

        Args:
            namespace: Kubernetes namespace
            k8s_manager: Kubernetes manager instance (injected by decorator)
            label: Display label (optional)

        Returns:
            Dictionary mapping service names to port mapping lists
        """
        assert k8s_manager is not None, "Kubernetes manager is required"

        console.print(
            f"[bold blue]🚀 Forwarding all services in {namespace}...[/bold blue]"
        )

        # Use native Kubernetes API instead of kubectl
        services = k8s_manager.get_services_with_ports(namespace)
        service_mappings: dict[str, list[PortMapping]] = {}

        for svc in services:
            svc_name = svc["name"]
            service_mappings[svc_name] = []

            for remote_port in svc["ports"]:
                local_port = self._calculate_local_port(remote_port)

                # Check if port was remapped due to overflow
                expected_port = int(f"{self.prefix}{remote_port}")
                overflow_note = ""
                if local_port != expected_port and expected_port > 65535:
                    overflow_note = " (remapped due to overflow)"

                console.print(
                    f"   {svc_name}:{remote_port} -> "
                    f"localhost:{local_port}{overflow_note}"
                )

                # Start port forwarding
                cmd = [
                    "kubectl",
                    "port-forward",
                    f"svc/{svc_name}",
                    "--address=0.0.0.0",
                    f"{local_port}:{remote_port}",
                    f"--namespace={namespace}",
                ]

                try:
                    proc = subprocess.Popen(
                        cmd,
                        stdout=subprocess.DEVNULL,
                        stderr=subprocess.DEVNULL,
                        preexec_fn=lambda: signal.signal(signal.SIGINT, signal.SIG_IGN),
                    )

                    mapping = PortMapping(
                        service=svc_name,
                        namespace=namespace,
                        remote_port=remote_port,
                        local_port=local_port,
                        pid=proc.pid,
                    )

                    self.port_mappings.append(mapping)
                    service_mappings[svc_name].append(mapping)

                    time.sleep(0.1)  # Avoid starting too many processes at once

                except Exception as e:
                    console.print(
                        f"[bold red]❌ Failed to forward {svc_name}:{remote_port}: {e}[/bold red]"
                    )

        return service_mappings

    @with_k8s_manager()
    def forward_clickhouse(
        self, env: ENV, k8s_manager: KubernetesManager
    ) -> list[PortMapping]:
        """Forward ClickHouse service in monitoring namespace

        Args:
            k8s_manager: Kubernetes manager instance (injected by decorator)

        Returns:
            List of ClickHouse port mappings
        """
        assert k8s_manager is not None, "Kubernetes manager is required"

        console.print("\n[bold blue]🚀 Forwarding ClickHouse...[/bold blue]")

        # Use native Kubernetes API instead of kubectl
        ports = k8s_manager.get_service_ports("clickstack-clickhouse", "monitoring")
        if not ports:
            console.print(
                "[bold yellow]⚠️  ClickHouse service not found or has no ports[/bold yellow]"
            )
            return []

        clickhouse_mappings = []

        for remote_port in ports:
            local_port = self._calculate_local_port(remote_port)

            # Check for overflow
            expected_port = int(f"{self.prefix}{remote_port}")
            overflow_note = ""
            if local_port != expected_port and expected_port > 65535:
                overflow_note = " (remapped due to overflow)"

            console.print(
                f"   clickstack-clickhouse:{remote_port} -> "
                f"localhost:{local_port}{overflow_note}"
            )

            # Start port forwarding
            cmd = [
                "kubectl",
                "port-forward",
                "svc/clickstack-clickhouse",
                "--address=0.0.0.0",
                f"{local_port}:{remote_port}",
                "--namespace=monitoring",
            ]

            try:
                proc = subprocess.Popen(
                    cmd,
                    stdout=subprocess.DEVNULL,
                    stderr=subprocess.DEVNULL,
                    preexec_fn=lambda: signal.signal(signal.SIGINT, signal.SIG_IGN),
                )

                mapping = PortMapping(
                    service="clickstack-clickhouse",
                    namespace="monitoring",
                    remote_port=remote_port,
                    local_port=local_port,
                    pid=proc.pid,
                )

                self.port_mappings.append(mapping)
                clickhouse_mappings.append(mapping)

                time.sleep(0.1)

            except Exception as e:
                console.print(
                    f"[bold red]❌ Failed to forward clickhouse:{remote_port}: {e}[/bold red]"
                )

        return clickhouse_mappings

    def get_service_url(
        self, service_name: str, namespace: str = "exp", protocol: str = "http"
    ) -> str | None:
        """Get local access URL for a service

        Args:
            service_name: Service name
            namespace: Namespace (default: exp)
            protocol: Protocol (default: http)

        Returns:
            Local access URL, or None if not found
        """
        for mapping in self.port_mappings:
            if mapping.service == service_name and mapping.namespace == namespace:
                return mapping.get_url(protocol)
        return None

    def get_port_mapping(
        self, service_name: str, namespace: str = "exp"
    ) -> PortMapping | None:
        """Get port mapping for a service

        Args:
            service_name: Service name
            namespace: Namespace (default: exp)

        Returns:
            Port mapping object, or None if not found
        """
        for mapping in self.port_mappings:
            if mapping.service == service_name and mapping.namespace == namespace:
                return mapping
        return None

    def stop_all_forwards(self):
        """Stop all port forwarding"""
        console.print("[bold blue]🛑 Stopping all port forwards...[/bold blue]")

        stopped_count = 0
        for mapping in self.port_mappings:
            if mapping.pid:
                try:
                    proc = psutil.Process(mapping.pid)
                    proc.terminate()
                    try:
                        proc.wait(timeout=5)
                    except psutil.TimeoutExpired:
                        proc.kill()
                    stopped_count += 1
                except (psutil.NoSuchProcess, psutil.AccessDenied):
                    pass

        self.port_mappings.clear()

        if stopped_count > 0:
            console.print(
                f"[bold green]✅ Stopped {stopped_count} port forward(s)[/bold green]"
            )
        else:
            console.print("[bold green]✅ No active forwards to stop[/bold green]")

    @with_k8s_manager()
    def start_forwarding(
        self, env: ENV, namespace: str, k8s_manager: KubernetesManager
    ) -> dict[str, Any]:
        """Start all port forwarding

        Args:
            k8s_manager: Kubernetes manager instance (injected by decorator)

        Returns:
            Dictionary containing forwarding information
        """
        self.cleanup_existing_forwards()

        # Forward main namespace (pass env for decorator)
        exp_mappings = self.forward_namespace_services(
            env=self.env, namespace=namespace, k8s_manager=k8s_manager
        )

        # Forward ClickHouse (pass env for decorator)
        ch_mappings = self.forward_clickhouse(env=self.env, k8s_manager=k8s_manager)

        console.print(
            f"\n[green]✅ Done! Forwarded:[/green]\n"
            f"   • {namespace} namespace: {len(exp_mappings)} service(s) ({self.prefix}xxxx ports)\n"
            f"   • monitoring namespace: clickstack-clickhouse ({len(ch_mappings)} port(s))"
        )

        return {
            "exp": exp_mappings,
            "clickhouse": ch_mappings,
            "total_mappings": len(self.port_mappings),
        }


def list_active_forwards():
    """List currently active port forward processes"""
    console.print("[cyan]📋 Active kubectl port-forward processes:[/cyan]\n")

    found = False
    for proc in psutil.process_iter(["pid", "name", "cmdline"]):
        try:
            cmdline = proc.info.get("cmdline") or []
            cmdline_str = " ".join(str(c) for c in cmdline)
            if "kubectl" in cmdline_str and "port-forward" in cmdline_str:
                console.print(f"   PID {proc.pid}: {cmdline_str}")
                found = True
        except (psutil.NoSuchProcess, psutil.AccessDenied, psutil.ZombieProcess):
            continue

    if not found:
        console.print("[gray]   No active port forwards[/gray]")
