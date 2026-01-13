import typer

from src.common.common import settings
from src.swagger import generate_python_sdk, generate_typescript_sdk, init

app = typer.Typer()


@app.command(name="init")
def swagger_init(
    version: str = typer.Option(..., "--version", "-v", help="API version."),
):
    """Initializes Swagger documentation setup."""
    init(version)


@app.command()
def generate_sdk(
    language: str = typer.Option(
        "python",
        "--language",
        "-l",
        help="SDK language.",
    ),
    version: str = typer.Option(
        "1.0.0",
        "--version",
        "-v",
        help="API version.",
    ),
):
    """Generates SDK Swagger documentation."""

    settings.reload()

    if language.lower() == "python":
        generate_python_sdk(version)
    elif language.lower() == "typescript":
        generate_typescript_sdk(version)
