import typer

from src.backup.mysql import MysqlBackupManager
from src.backup.redis_ import RedisClient
from src.common.common import SourceType, console, settings

app = typer.Typer()

mysql_app = typer.Typer(help="MySQL backup and restore utilities.")
app.add_typer(mysql_app, name="mysql")

reids_app = typer.Typer(help="Redis restore utilities.")
app.add_typer(reids_app, name="redis")


@mysql_app.command(name="migrate")
def mysql_migrate(
    src: str = typer.Option(
        "remote",
        "--src",
        "-s",
        help="Source of the backup to restore from (local or remote).",
    ),
    force: bool = typer.Option(
        False,
        "--force",
        "-f",
        help="Force restore even if the database is not empty.",
    ),
):
    """Restores MySQL database from backup."""

    settings.reload()

    console.print("[bold blue]Starting database migration...[/bold blue]")

    console.print("[bold blue]Step 1: Installing necessary tools...[/bold blue]")
    MysqlBackupManager.install_tools()
    console.print()

    source_type = SourceType.LOCAL if src.lower() == "local" else SourceType.REMOTE
    client = MysqlBackupManager(source_type)

    console.print(
        f"[bold blue]Step 2: Creating backup from {source_type} server...[/bold blue]"
    )
    client.backup()
    console.print()

    console.print(
        f"[bold blue]Step 3: Restoring backup to {client.dst} server...[/bold blue]"
    )
    client.restore(force)
    console.print()

    console.print("[bold green]✅ MySQL migration completed successfully![/bold green]")


@reids_app.command(name="migrate")
def redis_migrate(
    src: SourceType = typer.Option(
        "local",
        "--src",
        "-s",
        help="Source of the backup to restore from (local or remote).",
    ),
    exact_match: bool = typer.Option(
        False,
        "--exact_match",
        help="Source Redis stream exact matching (default: False)",
    ),
    force: bool = typer.Option(
        False,
        "--force",
        "-f",
        help="Force restore even if the redis is not empty.",
    ),
    dry_run: bool = typer.Option(
        False,
        "--dry_run",
        help="Perform a dry run without making any changes.",
    ),
):
    """Restores Redis database from backup."""

    settings.reload()

    console.print("[bold blue]Starting Redis migration...[/bold blue]")

    source_type = SourceType.LOCAL if src.lower() == "local" else SourceType.REMOTE
    client = RedisClient(source_type)

    console.print(
        f"[bold blue]Step 1: Restoring hash data from {source_type} server...[/bold blue]"
    )
    client.copy_hashes(force, dry_run=dry_run)
    console.print()

    console.print(
        f"[bold blue]Step 2: Restoring stream data to {client.dst} server...[/bold blue]"
    )
    client.copy_streams(
        exact_match,
        force=force,
        dry_run=dry_run,
    )
    console.print()

    console.print("[bold green]✅ Redis migration completed successfully![/bold green]")
