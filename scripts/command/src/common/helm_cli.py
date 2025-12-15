import time
from pathlib import Path

from pydantic import BaseModel, Field

from src.common.command import run_command
from src.common.common import PROJECT_ROOT, console

__all__ = ["HelmCLI", "HelmRelease"]

# Retry configuration
DEFAULT_MAX_RETRIES = 3
DEFAULT_RETRY_DELAY = 5  # seconds


class HelmRelease(BaseModel):
    """Represents a Helm release configuration."""

    name: str
    chart: str
    namespace: str
    repo_name: str | None = None
    repo_url: str | None = None
    version: str | None = None
    values_file: Path | None = None
    create_namespace: bool = False
    extra_args: list[str] = Field(default_factory=list)

    model_config = {"arbitrary_types_allowed": True}


class HelmCLI:
    """Manager for Helm operations."""

    def __init__(self, cwd: Path = PROJECT_ROOT):
        self.cwd = cwd
        self._repos_added: set[str] = set()

    def add_repo(self, name: str, url: str) -> bool:
        """Add a Helm repository."""
        if name in self._repos_added:
            console.print(f"[dim]Repo {name} already added, skipping...[/dim]")
            return True

        console.print(f"[bold blue]Adding Helm repo: {name}[/bold blue]")
        try:
            run_command(["helm", "repo", "add", name, url], cwd=self.cwd, check=True)
            self._repos_added.add(name)
            return True
        except SystemExit:
            console.print(
                f"[yellow]Repo {name} may already exist, continuing...[/yellow]"
            )
            self._repos_added.add(name)
            return True

    def repo_update(self) -> bool:
        """Update Helm repositories."""
        console.print("[bold blue]Updating Helm repos...[/bold blue]")
        try:
            run_command(["helm", "repo", "update"], cwd=self.cwd, check=True)
            return True
        except SystemExit:
            return False

    def is_release_exist(self, name: str, namespace: str) -> bool:
        """Check if a Helm release already exists."""
        try:
            run_command(
                ["helm", "status", name, "--namespace", namespace],
                cwd=self.cwd,
                check=True,
                capture_output=True,
            )
            return True
        except SystemExit:
            return False

    def install(
        self,
        release: HelmRelease,
        verbose: bool = False,
        wait: bool = False,
        timeout: str = "5m0s",
        max_retries: int = DEFAULT_MAX_RETRIES,
        retry_delay: int = DEFAULT_RETRY_DELAY,
        dry_run: bool = False,
    ) -> bool:
        """Install a Helm release with retry mechanism.

        Args:
            release: The Helm release configuration.
            max_retries: Maximum number of retry attempts for transient errors.
            retry_delay: Delay in seconds between retries.

        Returns:
            True if installation succeeded, False otherwise.
        """
        # Add repo if specified
        if release.repo_name and release.repo_url:
            self.add_repo(release.repo_name, release.repo_url)

        console.print(
            f"[bold blue]Installing Helm release '{release.name}' in namespace {release.namespace}[/bold blue]"
        )

        cmd = [
            "helm",
            "install",
            release.name,
            release.chart,
            "--namespace",
            release.namespace,
            "--atomic",
        ]

        if release.create_namespace:
            cmd.append("--create-namespace")

        if release.version:
            cmd.extend(["--version", release.version])

        if release.values_file:
            cmd.extend(["-f", str(release.values_file)])

        if verbose:
            cmd.append("--debug")

        if wait:
            cmd.extend(["--wait", "--timeout", timeout])

        if dry_run:
            cmd.append("--dry-run")

        cmd.extend(release.extra_args)

        for attempt in range(1, max_retries + 1):
            try:
                run_command(cmd, cwd=self.cwd, check=True)
                return True
            except SystemExit:
                if attempt < max_retries:
                    console.print(
                        f"[yellow]⚠️ Attempt {attempt}/{max_retries} failed. "
                        f"Retrying in {retry_delay}s...[/yellow]"
                    )
                    time.sleep(retry_delay)
                else:
                    console.print(
                        f"[bold red]❌ Failed to install {release.name} after {max_retries} attempts[/bold red]"
                    )
                    return False

        return False

    def uninstall(
        self,
        name: str,
        namespace: str,
        verbose: bool = False,
        wait: bool = False,
        timeout: str = "5m0s",
        extra_args: list[str] = [],
    ) -> bool:
        """Uninstall a Helm release."""
        console.print(f"[bold blue]Uninstalling Helm release: {name}[/bold blue]")
        cmd = ["helm", "uninstall", name, "--namespace", namespace]

        if verbose:
            cmd.append("--debug")

        if wait:
            cmd.extend(["--wait", "--timeout", timeout])

        if extra_args:
            cmd.extend(extra_args)

        try:
            run_command(cmd, cwd=self.cwd, check=True)
            return True
        except SystemExit:
            return False

    def upgrade(self, release: HelmRelease, install: bool = True) -> bool:
        """Upgrade a Helm release."""
        # Add repo if specified
        if release.repo_name and release.repo_url:
            self.add_repo(release.repo_name, release.repo_url)

        console.print(f"[bold blue]Upgrading Helm release: {release.name}[/bold blue]")

        cmd = [
            "helm",
            "upgrade",
            release.name,
            release.chart,
            "--namespace",
            release.namespace,
        ]

        if install:
            cmd.append("--install")

        if release.create_namespace:
            cmd.append("--create-namespace")

        if release.version:
            cmd.extend(["--version", release.version])

        if release.values_file:
            cmd.extend(["-f", str(release.values_file)])

        cmd.extend(release.extra_args)

        try:
            run_command(cmd, cwd=self.cwd, check=True)
            return True
        except SystemExit:
            return False
