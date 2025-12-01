import typer

app = typer.Typer(
    help="Main application for command-line interface.",
    pretty_exceptions_show_locals=False,
)


def main():
    from src.cli import backup, chaos, formatter, pedestal, rcabench_, swagger

    app.add_typer(backup.app, name="backup", help="Backup and migration utilities.")
    app.add_typer(chaos.app, name="chaos", help="Chaos engineering")
    app.add_typer(formatter.app, name="formatter", help="Code formatting utilities.")
    app.add_typer(pedestal.app, name="pedestal", help="Pedestal utilities.")
    app.add_typer(rcabench_.app, name="rcabench", help="RCABench utilities.")
    app.add_typer(swagger.app, name="swagger", help="Swagger/OpenAPI utilities.")

    app()
