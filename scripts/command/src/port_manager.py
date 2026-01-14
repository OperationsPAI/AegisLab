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


class PortMeta(BaseModel):
    """Port metadata for Kubernetes services

    This class is used to store metadata about Kubernetes service ports,
    including the remote port, and local port.
    """

    remote_port: int
    local_port: int


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

    def __init__(self, env: ENV, namespace: str = "exp"):
        """Initialize port forward manager

        Args:
            env: Environment type (TEST or PROD or STAGING)
            namespace: Kubernetes namespace to forward (default: exp)
        """
        self.env = env
        self.namespace = namespace
        self.prefix = (
            PortPrefix.TEST.value if env == ENV.TEST else PortPrefix.PROD.value
        )

        results = self._calculate_port_mappings(env=env, namespace=namespace)
        self.namespace_services: list[dict[str, Any]] = results[0]
        self.namespace_mappings: list[PortMapping] = results[1]

        results = self._calculate_port_mappings(
            env=env, namespace="monitoring", service_names=["clickstack-clickhouse"]
        )
        self.monitoring_namespaces: list[dict[str, Any]] = results[0]
        self.monitoring_mappings: list[PortMapping] = results[1]

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

        mappings = self.namespace_mappings + self.monitoring_mappings
        local_ports = set(mapping.local_port for mapping in mappings)

        # Kill processes occupying the ports we need
        port_killed_count = 0
        for conn in psutil.net_connections(kind="inet"):
            try:
                if conn.status == "LISTEN" and conn.laddr:
                    port = conn.laddr.port
                    if port in local_ports:
                        proc = psutil.Process(conn.pid)
                        proc.kill()
                        port_killed_count += 1
            except (psutil.NoSuchProcess, psutil.AccessDenied, psutil.ZombieProcess):
                continue

        if port_killed_count > 0:
            console.print(
                f"   Killed {port_killed_count} process(es) on specific ports"
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
    def _calculate_port_mappings(
        self,
        env: ENV,
        k8s_manager: KubernetesManager,
        namespace: str,
        service_names: list[str] = [],
    ) -> tuple[list[dict[str, Any]], list[PortMapping]]:
        """Calculate local ports for all services in the specific namespace

        Args:
            k8s_manager: Kubernetes manager instance (injected by decorator)
            namespace: Kubernetes namespace
            service_names: Optional list of service names to filter

        Returns:
            Tuple of (services list, port mappings list)
        """
        assert k8s_manager is not None, "Kubernetes manager is required"

        console.print("[bold blue]🔍 Calculating local ports...[/bold blue]")

        services = k8s_manager.get_services_with_ports(namespace)
        port_mappings: list[PortMapping] = []

        for svc in services:
            if service_names and svc["name"] not in service_names:
                continue

            svc_name = svc["name"]
            for remote_port in svc["ports"]:
                local_port = self._calculate_local_port(remote_port)
                port_mappings.append(
                    PortMapping(
                        service=svc_name,
                        namespace=namespace,
                        remote_port=remote_port,
                        local_port=local_port,
                    )
                )

        return services, port_mappings

    def forward_services(
        self,
        namespace: str,
        services: list[dict[str, Any]],
        port_mappings: list[PortMapping],
    ) -> dict[str, list[PortMapping]]:
        """Forward all services

        Args:
            namespace: Kubernetes namespace
            services: List of services to forward
            port_mappings: List of port mappings

        Returns:
            Dictionary mapping service names to port mapping lists
        """
        console.print(
            f"[bold blue]🚀 Forwarding all services in namespace {namespace}...[/bold blue]"
        )

        # Use native Kubernetes API instead of kubectl
        service_mappings: dict[str, list[PortMapping]] = {}

        port_mapping_exists: set[str] = set(
            mapping.service for mapping in port_mappings
        )
        port_mapping_dict: dict[int, PortMapping] = {
            mapping.remote_port: mapping for mapping in port_mappings
        }

        for svc in services:
            svc_name = svc["name"]
            if svc_name not in port_mapping_exists:
                continue

            service_mappings[svc_name] = []

            for remote_port in svc["ports"]:
                mapping = port_mapping_dict.get(remote_port)
                if mapping:
                    expected_port = int(f"{self.prefix}{remote_port}")
                    overflow_note = ""
                    if mapping.local_port != expected_port and expected_port > 65535:
                        overflow_note = " (remapped due to overflow)"

                    console.print(
                        f"   {mapping.service}:{remote_port} -> "
                        f"localhost:{mapping.local_port}{overflow_note}"
                    )

                    # Start port forwarding
                    cmd = [
                        "kubectl",
                        "port-forward",
                        f"svc/{svc_name}",
                        "--address=0.0.0.0",
                        f"{mapping.local_port}:{remote_port}",
                        f"--namespace={mapping.namespace}",
                    ]

                    try:
                        proc = subprocess.Popen(
                            cmd,
                            stdout=subprocess.DEVNULL,
                            stderr=subprocess.DEVNULL,
                            preexec_fn=lambda: signal.signal(
                                signal.SIGINT, signal.SIG_IGN
                            ),
                        )

                        mapping.namespace = self.namespace
                        mapping.pid = proc.pid

                        service_mappings[svc_name].append(mapping)

                        time.sleep(0.1)  # Avoid starting too many processes at once

                    except Exception as e:
                        console.print(
                            f"[bold red]❌ Failed to forward {svc_name}:{remote_port}: {e}[/bold red]"
                        )

        return service_mappings

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
        for mapping in self.namespace_mappings:
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
        for mapping in self.namespace_mappings:
            if mapping.service == service_name and mapping.namespace == namespace:
                return mapping
        return None

    def stop_all_forwards(self):
        """Stop all port forwarding"""
        console.print("[bold blue]🛑 Stopping all port forwards...[/bold blue]")

        stopped_count = 0
        for mapping in self.namespace_mappings:
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

        self.namespace_mappings.clear()

        if stopped_count > 0:
            console.print(
                f"[bold green]✅ Stopped {stopped_count} port forward(s)[/bold green]"
            )
        else:
            console.print("[bold green]✅ No active forwards to stop[/bold green]")

    def start_forwarding(self) -> dict[str, Any]:
        """Start all port forwarding

        Returns:
            Dictionary containing forwarding information
        """
        self.cleanup_existing_forwards()

        # Forward main namespace
        namespace_mappings = self.forward_services(
            self.namespace,
            services=self.namespace_services,
            port_mappings=self.namespace_mappings,
        )

        # Forward ClickHouse
        monitoring_mappings = self.forward_services(
            "monitoring",
            services=self.monitoring_namespaces,
            port_mappings=self.monitoring_mappings,
        )

        console.print(
            f"\n[green]✅ Done! Forwarded:[/green]\n"
            f"   • {self.namespace} namespace: {len(namespace_mappings)} service(s) ({self.prefix}xxxx ports)\n"
            f"   • monitoring namespace: clickstack-clickhouse ({len(monitoring_mappings['clickstack-clickhouse'])} port(s))"
        )

        return {
            self.namespace: namespace_mappings,
            "monitoring": monitoring_mappings,
            "total_mappings": len(self.namespace_mappings),
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
