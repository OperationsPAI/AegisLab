import os
import platform
import shutil
from datetime import datetime
from pathlib import Path
from typing import Any

from pydantic import BaseModel, Field, field_validator, model_validator

from src.common.command import run_command, run_pipeline
from src.common.common import PROJECT_ROOT, SourceType, console, settings

BACKUP_DIR = PROJECT_ROOT / "scripts" / "command" / "temp" / "backup_mysql"
REQUIRED_BINARIES = ["mysql", "mysqldump", "mysqlpump"]

__all__ = ["MysqlClient"]


class MysqlConfig(BaseModel):
    user: str = Field(default="root")
    password: str = Field(default="yourpassword")
    host: str = Field(default="127.0.0.1")
    port: str = Field(default="3306")
    db: str = Field(default="rcabench")

    @model_validator(mode="before")
    @classmethod
    def strip_prefix(cls, data: dict[str, Any]) -> Any:
        if isinstance(data, dict):
            cleaned_data = {}
            prefix = "mysql_"

            for key, value in data.items():
                if key.startswith(prefix):
                    new_key = key[len(prefix) :]
                    cleaned_data[new_key] = value
                else:
                    cleaned_data[key] = value
            return cleaned_data

        return data

    @field_validator("host", mode="after")
    def resolve_localhost(cls, v: str) -> str:
        if v.lower() == "localhost":
            return "127.0.0.1"
        return v

    def check_database_exists(self) -> bool:
        """Check if the specified database exists."""
        result = run_command(
            [
                "mysql",
                "-h",
                self.host,
                "-P",
                self.port,
                "-u",
                self.user,
                f"-p{self.password}",
                "-e",
                f"SELECT SCHEMA_NAME FROM INFORMATION_SCHEMA.SCHEMATA WHERE SCHEMA_NAME = '{self.db}';",
                "--silent",
                "--raw",
            ],
            capture_output=True,
        )
        return self.db in result.stdout.strip()

    def get_connection_cmd(self) -> list[str]:
        """Get the MySQL connection command."""
        return [
            "mysql",
            "-h",
            self.host,
            "-P",
            self.port,
            "-u",
            self.user,
            f"-p{self.password}",
            self.db,
        ]


local_mysql_config = MysqlConfig.model_validate(settings["database"])
settings.setenv("remote")
remote_mysql_config = MysqlConfig.model_validate(settings.database.to_dict())


class MysqlClient:
    def __init__(self, src: SourceType):
        if src == SourceType.LOCAL:
            self.dst = SourceType.REMOTE
            self.src_mysql_config = local_mysql_config
            self.dst_mysql_config = remote_mysql_config
        elif src == SourceType.REMOTE:
            self.dst = SourceType.LOCAL
            self.src_mysql_config = remote_mysql_config
            self.dst_mysql_config = local_mysql_config

        self.target_dir = BACKUP_DIR / src.value / self.src_mysql_config.db
        if not self.target_dir.exists():
            self.target_dir.mkdir(parents=True, exist_ok=True)

    @staticmethod
    def install_tools() -> None:
        """
        Install MySQL client tools (mysql, mysqldump, mysqlpump)

        Downloads and installs the specified version of MySQL client tools
        for the current operating system (macOS with Homebrew or Debian/Ubuntu).
        """
        missing = []
        for binary in REQUIRED_BINARIES:
            if not shutil.which(binary):
                missing.append(binary)

        if not missing:
            console.print(
                "[bold green]✅ MySQL client tools are already installed.[/bold green]"
            )
            return

        console.print(
            f"[bold yellow]🔍 Detected missing tools: {', '.join(missing)}[/bold yellow]"
        )

        os_name = platform.system()
        if os_name == "Linux":
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
                console.print(
                    "[bold blue]📥 Configuring MySQL official APT repository...[/bold blue]"
                )

                try:
                    # Download MySQL APT config package
                    run_command(
                        [
                            "wget",
                            "https://dev.mysql.com/get/mysql-apt-config_0.8.29-1_all.deb",
                            "-O",
                            "/tmp/mysql-apt-config.deb",
                        ],
                    )

                    # Install the package (this will add MySQL repository)
                    run_command(["sudo", "dpkg", "-i", "/tmp/mysql-apt-config.deb"])

                    # Update package list
                    run_command(["sudo", "apt", "update"])

                    # Install MySQL client
                    run_command(["sudo", "apt", "install", "-y", "mysql-client"])

                except Exception as e:
                    console.print(f"[bold red]❌ Installation failed: {e}[/bold red]")
                    console.print()
                    console.print(
                        "[bold yellow]💡 Try alternative installation:[/bold yellow]"
                    )
                    console.print("sudo apt install mysql-client-core-8.0")
                    raise SystemExit(1)
            else:
                console.print("Current system not supported for automatic installation")
                raise SystemExit(1)
        else:
            console.print(
                f"[bold red]Current system not supported: {os_name}[/bold red]"
            )
            raise SystemExit(1)

        console.print(
            "[bold green]✅ MySQL client tools installation completed![/bold green]"
        )

    def backup(self):
        """Backup remote MySQL database to local file."""
        backup_file = (
            self.target_dir
            / f"mysql_backup_{datetime.now().strftime(settings.time_format)}.sql"
        )

        console.print(
            f"[bold blue]🔄 Starting database backup {self.src_mysql_config.db}...[/bold blue]"
        )
        console.print(f"[gray]    Output file: {backup_file}[/gray]")

        run_command(
            [
                "mysqldump",
                "-h",
                self.src_mysql_config.host,
                "-P",
                self.src_mysql_config.port,
                "-u",
                self.src_mysql_config.user,
                f"-p{self.src_mysql_config.password}",
                self.src_mysql_config.db,
                "--result-file",
                str(backup_file),
                "--verbose",
                "--compression-algorithms=zlib",
                "--single-transaction",
                "--routines",
                "--triggers",
            ]
        )

        if backup_file.exists():
            size_mb = backup_file.stat().st_size / (1024 * 1024)
            console.print(f"[gray]📊 Backup size: {size_mb:.2f} MB[/gray]")
            console.print()

            console.print("[bold blue]🗜️ Compressing backup file...[/bold blue]")
            run_command(["gzip", backup_file.as_posix()])
            backup_file = backup_file.with_suffix(backup_file.suffix + ".gz")
            shutil.rmtree(backup_file, ignore_errors=True)

            console.print(
                f"[bold green]✅ Database backup completed: {backup_file}[/bold green]"
            )

    def restore(self, force: bool = False):
        """Restore MySQL database from backup file."""
        backup_file = self._get_latest_backup_file()
        if not backup_file:
            console.print(
                f"[bold red]❌ No valid backup files found in {self.target_dir}[/bold red]"
            )
            raise SystemExit(1)

        if not force:
            if self.dst_mysql_config.check_database_exists():
                console.print(
                    f"[bold yellow]⚠️ Target database '{self.dst_mysql_config.db}' already exists on {self.dst.name}.[/bold yellow]"
                )
                console.print(
                    "[gray]Use [yellow]--force[/yellow] option to overwrite the existing database.[gray]"
                )
                return

        console.print("[bold blue]🔄 Starting database restore...[/bold blue]")
        console.print(f"[gray]    Backup file: {backup_file}[/gray]")

        mysql_cmd = self.dst_mysql_config.get_connection_cmd()

        try:
            console.print("[bold blue]📦 Decompressing backup file...[/bold blue]")
            decompress_cmd = [
                "zcat" if shutil.which("zcat") else "gunzip -c",
                str(backup_file),
            ]

            run_pipeline(cmd1=decompress_cmd, cmd2=mysql_cmd)
            console.print("[bold green]✅ Database restore completed![/bold green]")

        except Exception as e:
            console.print(f"[bold red]❌ Restore failed: {e}[/bold red]")
            raise SystemExit(1)

    def _get_latest_backup_file(self) -> Path | None:
        """Get the latest backup file in the target directory."""
        backups = []
        for backup_path in self.target_dir.glob("mysql_backup_*"):
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
                console.print(
                    f"[bold yellow]⚠️ Cannot access {backup_path}: {e}[/bold yellow]"
                )

        if not backups:
            return None

        # Sort by modification time, return the latest
        latest_backup = sorted(backups, key=lambda x: x[2])[-1]
        backup_path, size, mtime = latest_backup
        size_mb = size / (1024 * 1024)
        timestamp = datetime.fromtimestamp(mtime).strftime(settings.time_format)

        console.print(f"[cyan]📁 Found latest backup: {backup_path.name}[/cyan]")
        console.print(f"[gray]    Size: {size_mb:.2f} MB[/gray]")
        console.print(f"[gray]    Created: {timestamp}[/gray]")

        return backup_path
