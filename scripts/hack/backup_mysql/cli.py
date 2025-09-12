#!/usr/bin/env -S uv run -s
import datetime
import os
import platform
import shutil
import subprocess
from dataclasses import dataclass
from pathlib import Path
from typing import Any

import typer
from rich.console import Console

BACKUP_DIR = Path("./temp/backup_mysql")
BACKUP_DIR.mkdir(parents=True, exist_ok=True)


DEFAULT_MYSQL_HOST = "10.10.10.220"
DEFAULT_MYSQL_PORT = "32206"
DEFAULT_MYSQL_USER = "root"
DEFAULT_MYSQL_PASSWORD = "yourpassword"
DEFAULT_MYSQL_DB = "rcabench"

REQUIRED_BINARIES = ["mysql", "mysqldump", "mysqlpump"]


@dataclass
class DatabaseConfig:
    host: str
    port: str
    user: str
    password: str
    database: str


app = typer.Typer(help="MySQL Backup Tool")
console = Console()


def check_binaries():
    """
    Check if required MySQL client binaries are available in PATH

    Returns:
        List of missing binary names
    """
    missing = []
    for binary in REQUIRED_BINARIES:
        if not shutil.which(binary):
            missing.append(binary)
    return missing


def database_exists(db_config: DatabaseConfig) -> bool:
    """
    Check if database exists on the server.

    Args:
        db_config: Database connection configuration

    Returns:
        True if database exists, False otherwise
    """
    try:
        result = subprocess.run(
            [
                "mysql",
                "-h",
                db_config.host,
                "-P",
                db_config.port,
                "-u",
                db_config.user,
                f"-p{db_config.password}",
                "-e",
                f"SELECT SCHEMA_NAME FROM INFORMATION_SCHEMA.SCHEMATA WHERE SCHEMA_NAME = '{db_config.database}';",
                "--silent",
                "--raw",
            ],
            capture_output=True,
            text=True,
            check=True,
        )
        return db_config.database in result.stdout.strip()
    except subprocess.CalledProcessError:
        return False


def get_latest_backup_file() -> Path | None:
    """
    Get the most recent backup file from the backup directory

    Returns:
        Path to the latest backup file or None if no backups found
    """
    backups = []
    for backup_path in BACKUP_DIR.glob(f"{DEFAULT_MYSQL_DB}_mysql_backup_*"):
        try:
            if backup_path.is_file() and backup_path.stat().st_size > 1024:
                backups.append(
                    (
                        backup_path,
                        backup_path.stat().st_size,
                        os.path.getmtime(backup_path),
                    )
                )
        except OSError as e:
            console.print(f"[yellow]⚠️ Cannot access {backup_path}: {e}[/yellow]")

    if not backups:
        return None

    # Sort by modification time, return the latest
    latest_backup = sorted(backups, key=lambda x: x[2])[-1]
    backup_path, size, mtime = latest_backup
    size_mb = size / (1024 * 1024)
    timestamp = datetime.datetime.fromtimestamp(mtime).strftime("%Y-%m-%d %H:%M:%S")

    console.print(f"📁 Found latest backup: {backup_path.name}")
    console.print(f"   Size: {size_mb:.2f} MB")
    console.print(f"   Created: {timestamp}")

    return backup_path


def get_server_version(db_config: DatabaseConfig) -> str:
    """
    Get MySQL server version by connecting to the database

    Args:
        db_config: Database connection configuration

    Returns:
        MySQL server version string or "Unknown" if connection fails
    """
    try:
        result = subprocess.run(
            [
                "mysql",
                "-h",
                db_config.host,
                "-P",
                db_config.port,
                "-u",
                db_config.user,
                f"-p{db_config.password}",
                "-e",
                "SELECT VERSION();",
                "--silent",
                "--raw",
            ],
            capture_output=True,
            text=True,
            check=True,
        )
        return result.stdout.strip()
    except subprocess.CalledProcessError:
        return "Unknown"


def run_command(cmd: list, env: dict[str, Any] | None = None):
    """
    Execute a shell command with error handling

    Args:
        cmd: Command to execute as a list of arguments
        env: Environment variables dictionary (optional)

    Raises:
        typer.Exit: If command execution fails
    """
    try:
        subprocess.run(cmd, check=True, env=env or os.environ.copy())
    except subprocess.CalledProcessError as e:
        console.print(f"[red]❌ Command failed: {' '.join(cmd)}\nError: {e}[/red]")
        raise typer.Exit(code=1)


@app.command()
def check_version(
    host: str = typer.Option(DEFAULT_MYSQL_HOST, "--host", "-h", help="MySQL host address"),
    port: str = typer.Option(DEFAULT_MYSQL_PORT, "--port", "-P", help="MySQL port"),
    user: str = typer.Option(DEFAULT_MYSQL_USER, "--user", "-u", help="MySQL username"),
    password: str = typer.Option(DEFAULT_MYSQL_PASSWORD, "--password", "-p", help="MySQL password"),
    database: str = typer.Option(DEFAULT_MYSQL_DB, "--database", "-d", help="Database name"),
):
    """
    Check MySQL server version and connection status

    Connects to the specified MySQL server and displays version information.
    Useful for verifying connectivity and server details.
    """
    db_config = DatabaseConfig(host, port, user, password, database)

    console.print("🔍 Checking MySQL server version...")
    version = get_server_version(db_config)

    if version == "Unknown":
        console.print("[red]❌ Unable to connect to MySQL server[/red]")
        console.print("[yellow]💡 Please check connection parameters[/yellow]")
    else:
        console.print(f"[green]📋 MySQL Server Version: {version}[/green]")


@app.command()
def install_tools(
    version: str = typer.Option("8.0", "--version", "-v", help="MySQL version (e.g.: 8.0, 5.7)"),
):
    """
    Install MySQL client tools (mysql, mysqldump, mysqlpump)

    Downloads and installs the specified version of MySQL client tools
    for the current operating system (macOS with Homebrew or Debian/Ubuntu).
    """
    missing = check_binaries()
    if not missing:
        console.print("[green]✅ MySQL client tools are already installed.[/green]")
        return

    console.print(f"[yellow]🔍 Detected missing tools: {', '.join(missing)}[/yellow]")
    console.print(f"[blue]🎯 Target version: MySQL {version}[/blue]")

    os_name = platform.system()
    env = os.environ.copy()

    if os_name == "Darwin":
        if not shutil.which("brew"):
            console.print("Please install Homebrew first: https://brew.sh/")
            raise typer.Exit()

        # Install MySQL on macOS
        run_command(["brew", "install", "mysql"], env)

    elif os_name == "Linux":
        distro = ""
        try:
            with open("/etc/os-release") as f:
                os_release = f.read().lower()
                if "ubuntu" in os_release or "debian" in os_release:
                    distro = "debian"
        except FileNotFoundError:
            pass

        if distro == "debian":
            # Install MySQL APT repository
            console.print("📥 Configuring MySQL official APT repository...")

            try:
                # Download MySQL APT config package
                run_command(
                    [
                        "wget",
                        "https://dev.mysql.com/get/mysql-apt-config_0.8.29-1_all.deb",
                        "-O",
                        "/tmp/mysql-apt-config.deb",
                    ],
                    env,
                )

                # Install the package (this will add MySQL repository)
                run_command(["sudo", "dpkg", "-i", "/tmp/mysql-apt-config.deb"], env)

                # Update package list
                run_command(["sudo", "apt", "update"], env)

                # Install MySQL client
                run_command(["sudo", "apt", "install", "-y", "mysql-client"], env)

            except Exception as e:
                console.print(f"[red]❌ Installation failed: {e}[/red]")
                console.print("[yellow]💡 Try alternative installation:[/yellow]")
                console.print("   sudo apt install mysql-client-core-8.0")
                raise typer.Exit(code=1)
        else:
            console.print("Current system not supported for automatic installation")
            raise typer.Exit()
    else:
        console.print(f"Current system not supported: {os_name}")
        raise typer.Exit()

    console.print("[green]✅ Installation completed![/green]")


@app.command()
def list():
    """List backup files"""
    backups = sorted(BACKUP_DIR.glob("*_mysql_backup_*"))
    if not backups:
        console.print("No MySQL backup files found.")
        return

    console.print("📋 Available MySQL backup files:")
    for backup in backups:
        if backup.is_file():
            size_mb = backup.stat().st_size / (1024 * 1024)
            format_type = "compressed" if backup.suffix == ".gz" else "plain"
            console.print(f"  🗂 {backup.name} ({format_type}, {size_mb:.2f} MB)")


@app.command()
def backup(
    host: str = typer.Option(DEFAULT_MYSQL_HOST, "--host", "-h", help="MySQL host address"),
    port: str = typer.Option(DEFAULT_MYSQL_PORT, "--port", "-P", help="MySQL port"),
    user: str = typer.Option(DEFAULT_MYSQL_USER, "--user", "-u", help="MySQL username"),
    password: str = typer.Option(DEFAULT_MYSQL_PASSWORD, "--password", "-p", help="MySQL password"),
    database: str = typer.Option(DEFAULT_MYSQL_DB, "--database", "-d", help="Database name"),
    compress: bool = typer.Option(False, "--compress", help="Enable compression"),
    single_transaction: bool = typer.Option(False, "--single-transaction", help="Use single transaction (InnoDB)"),
    routines: bool = typer.Option(False, "--routines", help="Include routines"),
    triggers: bool = typer.Option(False, "--triggers", help="Include triggers"),
):
    """
    Create a database backup using mysqldump with various options.

    Performs a logical backup of the specified MySQL database with options for
    compression, transaction consistency, and including database objects like
    routines and triggers.

    Args:
        host: MySQL server hostname or IP address
        port: MySQL server port number
        user: MySQL username for authentication
        password: MySQL password for authentication
        database: Name of the database to backup
        compress: Enable compression during data transfer
        single_transaction: Use single transaction for consistent backup
        routines: Include stored procedures and functions in backup
        triggers: Include triggers in backup
    """
    if not shutil.which("mysqldump"):
        console.print("[red]❌ mysqldump not found, please install MySQL client tools first[/red]")
        console.print("[yellow]💡 Run: ./cli.py install-tools[/yellow]")
        raise typer.Exit(code=1)

    timestamp = datetime.datetime.now().strftime("%Y%m%d_%H%M%S")
    backup_file = BACKUP_DIR / f"{database}_mysql_backup_{timestamp}.sql"

    cmd = [
        "mysqldump",
        "-h",
        host,
        "-P",
        port,
        "-u",
        user,
        f"-p{password}",
        "--result-file",
        str(backup_file),
        "--verbose",
    ]

    if single_transaction:
        cmd.append("--single-transaction")
    if routines:
        cmd.append("--routines")
    if triggers:
        cmd.append("--triggers")
    if compress:
        cmd.append("--compress")

    cmd.append(database)

    console.print(f"🔄 Starting database backup {database}...")
    console.print(f"📁 Output file: {backup_file}")
    console.print(f"📦 Single transaction: {single_transaction}")
    console.print(f"🔧 Include routines: {routines}")
    console.print(f"⚡ Include triggers: {triggers}")
    console.print(f"🗜️ Compression: {compress}")

    try:
        run_command(cmd)

        # Compress the backup file
        if backup_file.exists():
            console.print("🗜️ Compressing backup file...")
            run_command(["gzip", str(backup_file)])
            backup_file = backup_file.with_suffix(".sql.gz")

        console.print(f"[green]✅ Database backup completed: {backup_file}[/green]")

        # Display backup file size
        if backup_file.exists():
            size_mb = backup_file.stat().st_size / (1024 * 1024)
            console.print(f"📊 Backup size: {size_mb:.2f} MB")

    except Exception as e:
        console.print(f"[red]❌ Backup failed: {e}[/red]")
        raise typer.Exit(code=1)


@app.command()
def restore(
    backup_file: str = typer.Option(None, "--backup-file", "-f", help="Backup file path"),
    host: str = typer.Option(DEFAULT_MYSQL_HOST, "--host", "-h", help="MySQL host address"),
    port: str = typer.Option(DEFAULT_MYSQL_PORT, "--port", "-P", help="MySQL port"),
    user: str = typer.Option(DEFAULT_MYSQL_USER, "--user", "-u", help="MySQL username"),
    password: str = typer.Option(DEFAULT_MYSQL_PASSWORD, "--password", "-p", help="MySQL password"),
    database: str = typer.Option(DEFAULT_MYSQL_DB, "--database", "-d", help="Database name"),
    force: bool = typer.Option(False, "--force", help="Force overwrite existing database without confirmation"),
):
    """
    Restore a database from a backup file using mysql client.

    Restores a MySQL database from a backup file created by mysqldump.
    Supports both compressed (.gz) and plain SQL files. If no backup file
    is specified, automatically uses the most recent backup.

    Args:
        backup_file: Path to the backup file to restore
        host: MySQL server hostname or IP address
        port: MySQL server port number
        user: MySQL username for authentication
        password: MySQL password for authentication
        database: Target database name for restoration
        force: Force overwrite without confirmation
    """
    if backup_file is None:
        backup_path = get_latest_backup_file()
        if backup_path is None:
            console.print("[red]❌ No backup file found, please specify --backup-file[/red]")
            raise typer.Exit(code=1)
    else:
        backup_path = Path(backup_file)

    if not backup_path.exists():
        console.print(f"[red]❌ Backup file does not exist: {backup_path}[/red]")
        raise typer.Exit(code=1)

    if not shutil.which("mysql"):
        console.print("[red]❌ mysql client not found, please install MySQL client tools first[/red]")
        console.print("[yellow]💡 Run: ./cli.py install-tools[/yellow]")
        raise typer.Exit(code=1)

    # Check if database exists and ask for confirmation
    if not force:
        db_config = DatabaseConfig(host, port, user, password, database)
        if database_exists(db_config):
            console.print(f"[yellow]⚠️  Database '{database}' already exists on {host}:{port}[/yellow]")
            console.print("[yellow]This operation will overwrite existing data![/yellow]")

            if not typer.confirm("Continue with restore?"):
                console.print("[yellow]Restore cancelled by user[/yellow]")
                raise typer.Exit(code=0)

    console.print(f"🔄 Starting database restore {database}...")
    console.print(f"📁 Backup file: {backup_path}")
    console.print(f"🎯 Target: {host}:{port}")
    console.print(f"💪 Force mode: {force}")

    try:
        # Handle compressed files
        if backup_path.suffix == ".gz":
            console.print("📦 Decompressing backup file...")
            cmd = ["zcat" if shutil.which("zcat") else "gunzip -c", str(backup_path)]
            if shutil.which("zcat"):
                cmd = ["zcat", str(backup_path)]
            else:
                cmd = ["gunzip", "-c", str(backup_path)]

            mysql_cmd = [
                "mysql",
                "-h",
                host,
                "-P",
                port,
                "-u",
                user,
                f"-p{password}",
                database,
            ]

            # Use shell pipeline
            full_cmd = f"{' '.join(cmd)} | {' '.join(mysql_cmd)}"
            subprocess.run(full_cmd, shell=True, check=True)
        else:
            cmd = [
                "mysql",
                "-h",
                host,
                "-P",
                port,
                "-u",
                user,
                f"-p{password}",
                database,
            ]

            with open(backup_path) as f:
                subprocess.run(cmd, stdin=f, check=True)

        console.print("[green]✅ Database restore completed![/green]")

    except Exception as e:
        console.print(f"[red]❌ Restore failed: {e}[/red]")
        console.print("[yellow]💡 Tip: Make sure the target database exists[/yellow]")
        raise typer.Exit(code=1)


@app.command()
def help():
    """
    Display comprehensive help information for MySQL backup and restore operations.

    Shows available commands, usage examples, and best practices for database
    backup and restoration operations.
    """
    console.print("🛠️ [bold green]MySQL Official Backup and Restore Tools Guide[/bold green]\n")

    console.print("[bold green]📦 Available Commands:[/bold green]")
    console.print("  check-version    - Check database version")
    console.print("  install-tools    - Install MySQL client tools")
    console.print("  list       - List available backup files")
    console.print("  backup     - Backup database using mysqldump")
    console.print("  restore    - Restore database using mysql client")
    console.print("")

    console.print("[bold blue]🚀 Quick Start:[/bold blue]")
    console.print("1. Install tools: uv run python cli.py install-tools")
    console.print("2. Start backup: uv run python cli.py backup")
    console.print("3. View backups: uv run python cli.py list")
    console.print("4. Restore: uv run python cli.py restore")
    console.print("")

    console.print("[bold yellow]💡 Recommended Usage:[/bold yellow]")
    console.print("1. Full backup: uv run python cli.py backup --compress --single-transaction")
    console.print("2. Quick restore: uv run python cli.py restore <backup_file>")
    console.print("3. No compression: uv run python cli.py backup --no-compress")
    console.print("")

    console.print("[bold cyan]📋 Feature Comparison:[/bold cyan]")
    console.print("  --compress           - Enable compression during transfer")
    console.print("  --single-transaction - Consistent backup for InnoDB tables")
    console.print("  --routines          - Include stored procedures and functions")
    console.print("  --triggers          - Include triggers")
    console.print("")

    console.print("[bold magenta]🔧 Common Usage Examples:[/bold magenta]")
    console.print("# Backup from remote server")
    console.print("uv run python cli.py backup cli.py backup --host 10.10.10.220 --port 32206")
    console.print("")
    console.print("# Restore to local database")
    console.print("uv run python cli.py restore --host 127.0.0.1 --port 3306")
    console.print("")
    console.print("# Cross-server migration")
    console.print("uv run python cli.py backup --host old-server")
    console.print("uv run python cli.py restore --host new-server")


if __name__ == "__main__":
    app()
