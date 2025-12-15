from pathlib import Path

import typer

from src.common.common import ENV, settings
from src.pedestal import install_pedestals

app = typer.Typer()


@app.command()
def install(
    env: ENV = typer.Option(
        ENV.DEV,
        "--env",
        "-e",
        help="Target environment (e.g., dev, test).",
    ),
    name: str = typer.Option(
        ...,
        "--name",
        "-n",
        help="Pedestal container name to install",
    ),
    count: int = typer.Option(
        ...,
        "--count",
        "-c",
        help="Number of pedestal releases to install.",
    ),
    values_file: str | None = typer.Option(
        None,
        "--values-file",
        "-v",
        help="Path to a custom values.yaml file for the Helm chart.",
    ),
    dry_run: bool = typer.Option(
        False,
        "--dry-run",
        "-d",
        help="Simulate the installation without making any changes.",
    ),
    force: bool = typer.Option(
        False,
        "--force",
        "-f",
        help="Force reinstall even if the release already exists.",
    ),
):
    """Installs multiple pedestal Helm releases based on the specified container name and count."""

    settings.reload()

    values_file_path = Path(values_file) if values_file else None
    install_pedestals(
        env,
        name=name,
        count=count,
        values_file=values_file_path,
        dry_run=dry_run,
        force=force,
    )
