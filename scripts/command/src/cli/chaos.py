import typer

from src.chaos import clean_chaos_finalziers, delete_chaos_resources
from src.common.common import ENV, settings

app = typer.Typer()


@app.command()
def clean_finalizers(
    env: ENV = typer.Option(
        ENV.DEV,
        "--env",
        "-e",
        help="Target environment (e.g., dev, test).",
    ),
    ns_prefix: str = typer.Option(
        "aegislab-chaos-",
        "--ns-prefix",
        "-p",
        help="Namespace prefix to filter target namespaces.",
    ),
    ns_count: int = typer.Option(
        10,
        "--ns-count",
        "-c",
        help="Number of namespaces to process.",
    ),
):
    """Cleans finalizers from chaos resources in specified namespaces."""

    settings.reload()

    clean_chaos_finalziers(env, ns_prefix, ns_count)


@app.command()
def delete_resources(
    env: ENV = typer.Option(
        ENV.DEV,
        "--env",
        "-e",
        help="Target environment (e.g., dev, test).",
    ),
    ns_prefix: str = typer.Option(
        "aegislab-chaos-",
        "--ns-prefix",
        "-p",
        help="Namespace prefix to filter target namespaces.",
    ),
    ns_count: int = typer.Option(
        10,
        "--ns-count",
        "-c",
        help="Number of namespaces to process.",
    ),
):
    """Deletes chaos resources in specified namespaces."""

    settings.reload()

    delete_chaos_resources(env, ns_prefix, ns_count)
