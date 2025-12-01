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
    force: bool = typer.Option(
        False,
        "--force",
        "-f",
        help="Force reinstall even if the release already exists.",
    ),
):
    """Installs multiple pedestal Helm releases based on the specified container name and count."""

    settings.reload()

    install_pedestals(env, name=name, count=count, force=force)
