from enum import Enum

import typer

from src.common.common import console, settings
from src.swagger import init
from src.swagger.common import RunMode
from src.swagger.python import PythonSDK
from src.swagger.typescript import TypeScriptSDK

app = typer.Typer(help="Target-specific SDK generation utilities.")


class GenerationEnv(str, Enum):
    LOCAL = "local"
    RELEASE = "release"


class TypeScriptTarget(str, Enum):
    PORTAL = "portal"
    ADMIN = "admin"


class PythonTarget(str, Enum):
    SDK = "sdk"


@app.command(name="typescript")
def generate_typescript_sdk(
    target: TypeScriptTarget = typer.Option(
        ...,
        "--target",
        "-t",
        help="SDK target: portal or admin.",
    ),
    env: GenerationEnv = typer.Option(
        GenerationEnv.LOCAL,
        "--env",
        "-e",
        help="Generation environment: local or release.",
    ),
    version: str = typer.Option(
        "0.0.0",
        "--version",
        "-v",
        help="SDK package version.",
    ),
):
    """Generate one TypeScript SDK package."""

    settings.reload()
    init(version)
    TypeScriptSDK(version, target=RunMode(target.value)).generate()

    if env == GenerationEnv.RELEASE:
        console.print(
            "[dim]Release-ready TypeScript package generated. Publish with your registry step when needed.[/dim]"
        )


@app.command(name="python")
def generate_python_sdk(
    target: PythonTarget = typer.Option(
        ...,
        "--target",
        "-t",
        help="SDK target: sdk.",
    ),
    env: GenerationEnv = typer.Option(
        GenerationEnv.LOCAL,
        "--env",
        "-e",
        help="Generation environment: local or release.",
    ),
    version: str = typer.Option(
        "0.0.0",
        "--version",
        "-v",
        help="SDK package version.",
    ),
):
    """Generate the Python SDK package."""

    del target

    settings.reload()
    init(version)
    PythonSDK(version).generate()

    if env == GenerationEnv.RELEASE:
        console.print(
            "[dim]Release-ready Python package generated. Publish with your registry step when needed.[/dim]"
        )
