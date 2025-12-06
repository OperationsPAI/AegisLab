from pathlib import Path

from pydantic import BaseModel, Field

from src.common.command import run_command
from src.common.common import PROJECT_ROOT, console

__all__ = ["MinikubeCLI", "MinikubeConfig"]


class MinikubeConfig(BaseModel):
    """Configuration for Minikube cluster."""

    nodes: int = 1
    driver: str = "docker"
    cpus: int = 2
    memory: str = "4g"
    profile: str | None = None
    extra_args: list[str] = Field(default_factory=list)


class MinikubeCLI:
    """Manager for Minikube operations."""

    def __init__(self, cwd: Path = PROJECT_ROOT):
        self.cwd = cwd

    def start(self, config: MinikubeConfig) -> bool:
        """Start Minikube cluster."""
        console.print("[bold blue]🚀 Starting Minikube cluster...[/bold blue]")

        cmd = [
            "minikube",
            "start",
            "--nodes",
            str(config.nodes),
            "--driver",
            config.driver,
            "--cpus",
            str(config.cpus),
            "--memory",
            config.memory,
        ]

        if config.profile:
            cmd.extend(["--profile", config.profile])

        cmd.extend(config.extra_args)

        try:
            run_command(cmd, cwd=self.cwd, check=True)
            console.print("[bold green]✅ Minikube cluster started[/bold green]")
            return True
        except SystemExit:
            console.print("[bold red]❌ Failed to start Minikube[/bold red]")
            return False

    def stop(self, profile: str | None = None) -> bool:
        """Stop Minikube cluster."""
        console.print("[bold blue]🛑 Stopping Minikube cluster...[/bold blue]")

        cmd = ["minikube", "stop"]
        if profile:
            cmd.extend(["--profile", profile])

        try:
            run_command(cmd, cwd=self.cwd, check=True)
            return True
        except SystemExit:
            return False

    def delete(self, profile: str | None = None) -> bool:
        """Delete Minikube cluster."""
        console.print("[bold blue]🗑️ Deleting Minikube cluster...[/bold blue]")

        cmd = ["minikube", "delete"]
        if profile:
            cmd.extend(["--profile", profile])

        try:
            run_command(cmd, cwd=self.cwd, check=True)
            return True
        except SystemExit:
            return False

    def status(self, profile: str | None = None) -> bool:
        """Get Minikube cluster status."""
        cmd = ["minikube", "status"]
        if profile:
            cmd.extend(["--profile", profile])

        try:
            run_command(cmd, cwd=self.cwd, check=True)
            return True
        except SystemExit:
            return False
