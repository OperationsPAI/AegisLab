import typer

from src.common.common import LanguageType, console, settings
from src.swagger import Generator, init

app = typer.Typer()


@app.command(name="init")
def swagger_init(
    version: str = typer.Option(..., "--version", "-v", help="API version."),
):
    """Initializes Swagger documentation setup."""
    init(version)


@app.command()
def generate_client(
    language: LanguageType = typer.Option(
        LanguageType.TYPESCRIPT,
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
    """Generates Swagger client documentation."""

    settings.reload()

    if language != LanguageType.TYPESCRIPT:
        console.print(
            f"[bold red]❌ Client generation for {language} is not supported yet.[/bold red]"
        )
        raise typer.Exit(code=1)

    Generator.get_client_generator(language, version).generate()


@app.command()
def generate_sdk(
    language: LanguageType = typer.Option(
        LanguageType.PYTHON,
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

    Generator.get_sdk_generator(language, version).generate()
