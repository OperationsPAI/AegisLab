import typer

from src.cli.backup import mysql_migrate, redis_migrate
from src.common.common import ENV, PROJECT_ROOT, console, settings
from src.rcabench_ import (
    check_db,
    check_redis,
    local_deploy,
)

app = typer.Typer()


@app.command(name="local-deploy")
def rcabench_local_deploy(
    src: ENV | None = typer.Option(
        None,
        "--src",
        "-s",
        help="Source of the backup to restore from. If not provided, no data migration will be performed.",
    ),
    force: bool = typer.Option(
        False,
        "--force",
        "-f",
        help="Force redeploy even if services are already running.",
    ),
):
    """Deploys RCABench locally using Docker Compose."""

    settings.reload()

    local_deploy(env=ENV.DEV)

    if src is not None:
        if src == ENV.DEV:
            console.print(
                "[red]Source and destination environments cannot be the same.[/red]"
            )
            raise typer.Exit(code=1)

        console.print()
        check_db(src)
        check_redis(src)

        mysql_migrate(src, dst=ENV.DEV, force=force)
        redis_migrate(src, dst=ENV.DEV, force=force, dry_run=False)

    console.print(
        "\n[bold yellow]You can start the application manually later: [/bold yellow]"
    )
    console.print(
        f"[gray]cd {PROJECT_ROOT / 'src'} && go run main.go both --port 8082 [/gray]"
    )
