import typer

from src.backup.mysql import MysqlClient, mysql_configs
from src.common.common import ENV, settings

app = typer.Typer()


@app.command(name="sync-db")
def sync_to_database(
    dst: ENV = typer.Option(
        ENV.DEV,
        "--dst",
        "-d",
        help="Destination environment to restore the backup to.",
    ),
):
    """Synchronize local datapack files to the database."""

    settings.reload()

    mysql_config = mysql_configs[dst]
    mysql_client = MysqlClient(mysql_config)
    mysql_client.sync_datapacks_to_database()
