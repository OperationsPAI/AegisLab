import typer

app = typer.Typer(
    help="Main application for command-line interface.",
    pretty_exceptions_show_locals=False,
)


def main():
    from src.cli import (
        backup,
        chaos,
        datapack,
        formatter,
        git,
        pedestal,
        port_manager,
        rcabench_,
        swagger,
        test,
    )

    app.add_typer(backup.app, name="backup", help="Backup and migration utilities.")
    app.add_typer(chaos.app, name="chaos", help="Chaos engineering.")
    app.add_typer(datapack.app, name="datapack", help="Datapack management utilities.")
    app.add_typer(formatter.app, name="format", help="Code formatting utilities.")
    app.add_typer(git.app, name="git", help="Git hooks and utilities.")
    app.add_typer(pedestal.app, name="pedestal", help="Pedestal utilities.")
    app.add_typer(
        port_manager.app, name="port", help="Kubernetes port forwarding manager."
    )
    app.add_typer(rcabench_.app, name="rcabench", help="RCABench utilities.")
    app.add_typer(swagger.app, name="swagger", help="Swagger/OpenAPI utilities.")
    app.add_typer(test.app, name="test", help="Test environment utilities.")

    app()
