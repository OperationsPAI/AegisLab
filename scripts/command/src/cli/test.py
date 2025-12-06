import typer

from src.common.common import ENV
from src.common.minikube_cli import MinikubeCLI
from src.test import setup_env, teardown_env

app = typer.Typer()


@app.command()
def setup(
    nodes: int = typer.Option(3, "--nodes", "-n", help="Number of Minikube nodes"),
    cpus: int = typer.Option(2, "--cpus", "-c", help="CPUs per node"),
    memory: str = typer.Option("4g", "--memory", "-m", help="Memory per node"),
    skip_minikube: bool = typer.Option(
        False, "--skip-minikube", help="Skip Minikube cluster creation"
    ),
    is_ci: bool = typer.Option(
        False, "--is-ci", help="Whether running in a CI environment"
    ),
):
    """Set up the local test environment."""
    setup_env(nodes, cpus=cpus, memory=memory, skip_minikube=skip_minikube, is_ci=is_ci)


@app.command()
def teardown(
    delete_cluster: bool = typer.Option(
        False, "--delete-cluster", help="Delete Minikube cluster"
    ),
):
    """Tear down the test environment."""
    teardown_env(delete_cluster)


@app.command()
def status():
    """Check the status of the test environment."""
    minikube = MinikubeCLI()
    minikube.status()
