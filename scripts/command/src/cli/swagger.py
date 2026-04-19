import typer

from src.common.common import settings
from src.swagger import init
from src.swagger.apifox import ApifoxTarget

app = typer.Typer(help="Swagger/OpenAPI generation utilities.")


@app.command(name="init")
def swagger_init(
    version: str = typer.Option(..., "--version", "-v", help="API version."),
    apifox_targets: list[ApifoxTarget] | None = typer.Option(
        None,
        "--apifox-target",
        "-t",
        help="Optional Apifox upload targets: sdk, portal, admin, or all. Omit to skip upload.",
    ),
):
    """Generate normalized OpenAPI artifacts from Go Swagger annotations."""
    settings.reload()
    init(version, apifox_targets=apifox_targets)
