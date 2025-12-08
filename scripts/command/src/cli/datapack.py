import typer

from src.backup.mysql import MysqlClient, mysql_configs
from src.common.common import SourceType, settings

app = typer.Typer()


@app.command(name="sync-db")
def sync_to_database(
    src: SourceType = typer.Option(
        SourceType.REMOTE,
        "--src",
        "-s",
        help="Source of the backup to restore from (local or remote).",
    ),
):
    """Synchronize local datapack files to the database."""

    settings.reload()

    mysql_config = mysql_configs[src]
    mysql_client = MysqlClient(mysql_config)
    mysql_client.sync_datapacks_to_database()
