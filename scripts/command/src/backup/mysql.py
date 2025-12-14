import json
import os
import platform
import shutil
from datetime import datetime
from pathlib import Path
from typing import Any

from pydantic import BaseModel, Field, field_validator, model_validator
from sqlalchemy import create_engine, text
from sqlalchemy.orm import Session, sessionmaker

from src.common.command import run_command, run_pipeline
from src.common.common import (
    DATAPACK_ROOT_PATH,
    ENV,
    PROJECT_ROOT,
    console,
    settings,
)

BACKUP_DIR = PROJECT_ROOT / "scripts" / "command" / "temp" / "backup_mysql"
REQUIRED_BINARIES = ["mysql", "mysqldump", "mysqlpump"]

__all__ = ["mysql_configs", "MysqlClient", "MysqlBackupManager"]


class MysqlConfig(BaseModel):
    user: str = Field(default="root")
    password: str = Field(default="yourpassword")
    host: str = Field(default="127.0.0.1")
    port: str = Field(default="3306")
    db: str = Field(default="rcabench")

    @field_validator("host", mode="after")
    def resolve_localhost(cls, v: str) -> str:
        if v.lower() == "localhost":
            return "127.0.0.1"
        return v

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


mysql_configs: dict[ENV, MysqlConfig] = {}


def _init_mysql_configs():
    for env in ENV:
        settings.setenv(env.value)
        config = MysqlConfig.model_validate(settings.database.mysql.to_dict())
        mysql_configs[env] = config


_init_mysql_configs()


class MysqlClient:
    def __init__(self, config: MysqlConfig):
        self._config = config
        self._session = self._connect_database()

    def _connect_database(self) -> Session:
        """
        Connect to the MySQL database and return a session.

        Returns:
            Session: SQLAlchemy database session

        Raises:
            SystemExit: If connection fails
        """
        try:
            db_url = f"mysql+pymysql://{self._config.user}:{self._config.password}@{self._config.host}:{self._config.port}/{self._config.db}"
            engine = create_engine(db_url, echo=False, pool_pre_ping=True)

            Session = sessionmaker(bind=engine)
            session = Session()

            # Test connection
            result = session.execute(text("SELECT version()"))
            version = result.scalar()
            console.print("[bold green]✅ Database connection successful[/bold green]")
            console.print(f"[dim]Version: {version[:50]}...[/dim]")  # type: ignore

            return session

        except Exception as e:
            console.print(f"[bold red]❌ Database connection failed: {e}[/bold red]")
            raise SystemExit(1)

    def get_session(self) -> Session:
        """Get the current database session."""
        return self._session

    def sync_datapacks_to_database(self):
        """
        Synchronize local datapack files to database.

        Scans local datapack directories and imports missing fault_injection records
        into the database based on injection.json files.
        """
        if not DATAPACK_ROOT_PATH.exists:
            console.print(
                f"[bold yellow]⚠️ Datapack root path does not exist: {DATAPACK_ROOT_PATH}[/bold yellow]"
            )
            return

        local_datapacks = [
            entry for entry in DATAPACK_ROOT_PATH.iterdir() if entry.is_dir()
        ]
        console.print(
            f"[bold green]📁 Found {len(local_datapacks)} local datapack directories[/bold green]"
        )
        console.print()

        console.print("[bold blue]🔄 Starting database synchronization...[/bold blue]")
        dir = (
            DATAPACK_ROOT_PATH
            / "sync_logs"
            / datetime.now().strftime(settings.time_format)
        )
        if not dir.exists():
            dir.mkdir(parents=True, exist_ok=True)

        # Dictionary to store datapack_name -> database_id mapping
        datapack_id_mapping = {}
        mapping_file = dir / "datapack_id_mapping.json"

        try:
            # Get all fault_injections from database
            query = text("SELECT id, name FROM fault_injections ORDER BY id DESC")
            result = self._session.execute(query)
            db_records = result.fetchall()
            console.print(
                f"[bold green]✅ Found {len(db_records)} records in database[/bold green]"
            )

            db_datapack_names = []
            for row in db_records:
                injection_name = row[1]
                db_datapack_names.append(injection_name)

            # Add missing database records from local files
            added_count = 0
            skipped_count = 0

            benchmark_results = self._session.execute(
                text("""
                    SELECT cv.id, c.name FROM container_versions cv
                    JOIN containers c ON cv.container_id = c.id
                    WHERE c.type = 1 AND cv.status >= 0
                    ORDER BY cv.id ASC LIMIT 1
                """)
            ).fetchall()
            benchmark_dict = {item[1]: item[0] for item in benchmark_results}
            if not benchmark_dict:
                console.print(
                    "[bold red]❌ No benchmark containers found in database[/bold red]"
                )
                return

            pedestal_results = self._session.execute(
                text("""
                    SELECT cv.id, c.name FROM container_versions cv
                    JOIN containers c ON cv.container_id = c.id
                    WHERE c.type = 2 AND cv.status >= 0
                    ORDER BY cv.id ASC LIMIT 1
                """)
            ).fetchall()
            pedestal_dict = {item[1]: item[0] for item in pedestal_results}
            if not pedestal_dict:
                console.print(
                    "[bold red]❌ No pedestal containers found in database[/bold red]"
                )
                return

            def _safe_get(data: dict, key: str, default: Any = None) -> Any:
                value = data.get(key, default)
                return default if value is None else value

            def _parse_timestamp(timestamp_str: Any) -> Any:
                if timestamp_str is None:
                    return None
                try:
                    if isinstance(timestamp_str, str):
                        return datetime.fromisoformat(
                            timestamp_str.replace("Z", "+00:00")
                        )
                    return timestamp_str
                except Exception:
                    return None

            for local_datapack in local_datapacks:
                datapack_name = local_datapack.name

                if datapack_name not in db_datapack_names:
                    injection_json_path = local_datapack / "injection.json"

                    if not injection_json_path.exists():
                        console.print(
                            f"[gray]    ⚠️ Missing injection.json: {datapack_name}[/gray]"
                        )
                        skipped_count += 1
                        continue

                    with open(injection_json_path, encoding="utf-8") as f:
                        injection_data = json.load(f)

                    # Process engine_config
                    engine_config = _safe_get(injection_data, "engine_config")
                    if engine_config is None:
                        console.print(
                            f"[gray]    ⚠️ Skipping {datapack_name}: missing engine_config[/gray]"
                        )
                        skipped_count += 1
                        continue

                    # Process groundtruth - convert to JSON string
                    groundtruth_data = _safe_get(injection_data, "ground_truth")
                    groundtruth_json = None
                    if groundtruth_data:
                        try:
                            groundtruth_json = json.dumps(groundtruth_data)
                        except Exception as e:
                            console.print(
                                f"[gray]    ⚠️ Skipping {datapack_name}: failed to serialize groundtruth: {e}[/gray]"
                            )

                    # Process duration fields
                    pre_duration = _safe_get(injection_data, "pre_duration")
                    if pre_duration is None:
                        console.print(
                            f"[gray]    ⚠️ Skipping {datapack_name}: missing pre_duration[/gray]"
                        )
                        skipped_count += 1
                        continue

                    # Get benchmark_id
                    benchmark_name = _safe_get(injection_data, "benchmark")
                    benchmark_id = benchmark_dict.get(benchmark_name)
                    if benchmark_id is None:
                        console.print(
                            f"[gray]    ⚠️ Skipping {datapack_name}: benchmark '{benchmark_name}' not found[/gray]"
                        )
                        skipped_count += 1
                        continue

                    # Get pedestal_id from datapack name (format: namespace-pedestal-service-...)
                    parts = datapack_name.split("-")
                    pedestal_name = parts[1] if len(parts) > 1 else None
                    if pedestal_name is None:
                        console.print(
                            f"[gray]    ⚠️ Skipping {datapack_name}: cannot determine pedestal name[/gray]"
                        )
                        skipped_count += 1
                        continue
                    elif pedestal_name == "mysql":
                        pedestal_name = "ts"

                    pedestal_id = pedestal_dict.get(pedestal_name)
                    if pedestal_id is None:
                        console.print(
                            f"[gray]    ⚠️ Skipping {datapack_name}: pedestal '{pedestal_name}' not found[/gray]"
                        )
                        skipped_count += 1
                        continue

                    try:
                        # Insert fault_injection
                        result = self._session.execute(
                            text("""
                                INSERT INTO fault_injections (
                                    name, fault_type, display_config, engine_config, groundtruth,
                                    pre_duration, start_time, end_time, state, status,
                                    description, benchmark_id, pedestal_id,
                                    created_at, updated_at
                                ) VALUES (
                                    :name, :fault_type, :display_config, :engine_config, :groundtruth,
                                    :pre_duration, :start_time, :end_time, :state, :status,
                                    :description, :benchmark_id, :pedestal_id,
                                    :created_at, :updated_at
                                )
                            """),
                            {
                                "name": _safe_get(injection_data, "name")
                                or _safe_get(injection_data, "injection_name"),
                                "fault_type": _safe_get(injection_data, "fault_type"),
                                "display_config": _safe_get(
                                    injection_data, "display_config"
                                ),
                                "engine_config": _safe_get(
                                    injection_data, "engine_config"
                                ),
                                "groundtruth": groundtruth_json,
                                "pre_duration": _safe_get(
                                    injection_data, "pre_duration"
                                ),
                                "start_time": _parse_timestamp(
                                    _safe_get(injection_data, "start_time")
                                ),
                                "end_time": _parse_timestamp(
                                    _safe_get(injection_data, "end_time")
                                ),
                                "state": 4,  # default: completed
                                "status": 1,  # enabled
                                "description": _safe_get(injection_data, "description"),
                                "benchmark_id": benchmark_id,
                                "pedestal_id": pedestal_id,
                                "created_at": _parse_timestamp(
                                    _safe_get(injection_data, "created_at")
                                ),
                                "updated_at": _parse_timestamp(
                                    _safe_get(injection_data, "updated_at")
                                ),
                            },
                        )

                        inserted_id = result.lastrowid  # type: ignore
                        self._session.commit()

                        datapack_id_mapping[datapack_name] = inserted_id

                        console.print(
                            f"[gray]    📝 Added record: {datapack_name} (ID: {inserted_id})[/gray]"
                        )
                        added_count += 1

                    except Exception as e:
                        self._session.rollback()
                        console.print(
                            f"[bold red]❌ Failed to add {datapack_name}: {e}[/bold red]"
                        )
                        skipped_count += 1

            console.print()
            console.print(
                "[bold green]✅ Datapack synchronization completed![/bold green]"
            )
            console.print(f"[gray]    Added: {added_count} records[/gray]")
            console.print(f"[gray]    Skipped: {skipped_count} records[/gray]")

            # Save datapack ID mapping to JSON file
            if datapack_id_mapping:
                try:
                    with open(mapping_file, "w", encoding="utf-8") as f:
                        json.dump(datapack_id_mapping, f, indent=2, ensure_ascii=False)
                    console.print()
                    console.print(
                        f"[bold green]💾 Saved ID mapping to: {mapping_file}[/bold green]"
                    )
                    console.print(
                        f"[gray]    Total mappings: {len(datapack_id_mapping)}[/gray]"
                    )
                except Exception as e:
                    console.print(
                        f"[bold yellow]⚠️ Failed to save ID mapping: {e}[/bold yellow]"
                    )

            with open(dir / "metadata.json", "w", encoding="utf-8") as f:
                metadata = {
                    "host": self._config.host,
                    "finished_at": datetime.now().strftime(settings.time_format),
                }
                json.dump(metadata, f, indent=2, ensure_ascii=False)

        except Exception as e:
            console.print(
                f"[bold red]❌ Datapack synchronization failed: {e}[/bold red]"
            )
            self._session.rollback()
            dir.rmdir()
        finally:
            self._session.close()


class MysqlBackupManager:
    """Manage MySQL database backup and restore operations."""

    def __init__(self, src: ENV, dst: ENV):
        self.src = src
        self.dst = dst

        self.src_mysql_config = mysql_configs[self.src]
        self.dst_mysql_config = mysql_configs[self.dst]

        self.target_dir = BACKUP_DIR / self.src.value / self.src_mysql_config.db
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
        """Backup MySQL database from src to dst local files."""
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

        db_exists = False
        if not force:
            try:
                _ = MysqlClient(self.dst_mysql_config)
            except SystemExit:
                db_exists = True

            if db_exists:
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
