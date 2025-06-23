#!/usr/bin/env -S uv run -s
import typer
from pathlib import Path
import subprocess
import datetime
import os
import shutil
import platform
from dataclasses import dataclass

app = typer.Typer()
BACKUP_DIR = Path("./temp/backup_psql")
BACKUP_DIR.mkdir(parents=True, exist_ok=True)

DEFAULT_PG_HOST = "10.10.10.220"
DEFAULT_PG_PORT = "32432"
DEFAULT_PG_USER = "postgres"
DEFAULT_PG_PASSWORD = "yourpassword"
DEFAULT_PG_DB = "rcabench"

REQUIRED_BINARIES = ["psql", "pg_dump", "pg_restore"]

# TODO 做一个Docker镜像


@dataclass
class DatabaseConfig:
    host: str
    port: str
    user: str
    password: str
    database: str


def run_command(cmd: list, env: dict):
    try:
        subprocess.run(cmd, check=True, env=env or os.environ.copy())
    except subprocess.CalledProcessError as e:
        typer.secho(
            f"[✘] Command failed: {' '.join(cmd)}\nError: {e}", fg=typer.colors.RED
        )
        raise typer.Exit(code=1)


def get_server_version(db_config: DatabaseConfig):
    env = os.environ.copy()
    env["PGPASSWORD"] = db_config.password

    try:
        result = subprocess.run(
            [
                "psql",
                "-h",
                db_config.host,
                "-p",
                db_config.port,
                "-U",
                db_config.user,
                "-d",
                db_config.database,
                "-t",
                "-c",
                "SELECT version();",
            ],
            capture_output=True,
            text=True,
            check=True,
            env=env,
        )
        return result.stdout.strip()
    except subprocess.CalledProcessError:
        return "Unknown"


def check_binaries():
    missing = []
    for binary in REQUIRED_BINARIES:
        if not shutil.which(binary):
            missing.append(binary)
    return missing


@app.command()
def install_tools(
    version: str = typer.Option(
        "16", "--version", "-v", help="PostgreSQL version (e.g.: 16, 15, 14)"
    ),
):
    """
    Install PostgreSQL client tools (psql, pg_dump, pg_restore)
    """
    missing = check_binaries()
    if not missing:
        typer.secho(
            "✅ PostgreSQL client tools are already installed.", fg=typer.colors.GREEN
        )

        # Check version compatibility
        try:
            result = subprocess.run(
                ["pg_dump", "--version"], capture_output=True, text=True
            )
            current_version = result.stdout.strip()
            typer.echo(f"📋 Current version: {current_version}")

            # Extract major version
            import re

            version_match = re.search(r"pg_dump.*?(\d+)", current_version)
            if version_match:
                current_major = version_match.group(1)
                if current_major != version:
                    typer.secho(
                        f"⚠️  Version mismatch! Current: {current_major}, Target: {version}",
                        fg=typer.colors.YELLOW,
                    )
                    typer.secho(
                        f"💡 To install PostgreSQL {version}, please uninstall current version first and re-run this command",
                        fg=typer.colors.YELLOW,
                    )
                else:
                    typer.secho(
                        f"✅ Version matches: PostgreSQL {version}",
                        fg=typer.colors.GREEN,
                    )
        except Exception:
            pass

        return

    typer.secho(
        f"🔍 Detected missing tools: {', '.join(missing)}", fg=typer.colors.YELLOW
    )
    typer.secho(f"🎯 Target version: PostgreSQL {version}", fg=typer.colors.BLUE)

    os_name = platform.system()
    env = os.environ.copy()

    if os_name == "Darwin":
        if not shutil.which("brew"):
            typer.echo("Please install Homebrew first: https://brew.sh/")
            raise typer.Exit()

        # Install specific version on macOS
        if version == "16":
            package = "postgresql@16"
        elif version == "15":
            package = "postgresql@15"
        elif version == "14":
            package = "postgresql@14"
        else:
            package = f"postgresql@{version}"

        typer.echo(f"📦 Package: {package}")
        run_command(["brew", "install", package], env)

        # Add to PATH
        typer.echo("🔗 Adding to PATH...")
        path_cmd = (
            f"echo 'export PATH=\"/opt/homebrew/opt/{package}/bin:$PATH\"' >> ~/.zshrc"
        )
        typer.echo(f"💡 Please run: {path_cmd}")
        typer.echo("💡 Then run: source ~/.zshrc")

    elif os_name == "Linux":
        distro = ""
        try:
            with open("/etc/os-release", "r") as f:
                os_release = f.read().lower()
                if "ubuntu" in os_release or "debian" in os_release:
                    distro = "debian"
        except FileNotFoundError:
            pass

        if distro == "debian":
            # Install specific version on Debian/Ubuntu
            typer.echo("📥 Configuring PostgreSQL official APT repository...")

            # Method 1: Use automated script (recommended)
            typer.echo("🚀 Trying automated configuration script...")
            try:
                # Install postgresql-common (contains automated script)
                run_command(["sudo", "apt", "update"], env)
                run_command(["sudo", "apt", "install", "-y", "postgresql-common"], env)

                # Run automated configuration script
                run_command(
                    ["sudo", "/usr/share/postgresql-common/pgdg/apt.postgresql.org.sh"],
                    env,
                )

                typer.secho(
                    "✅ Automated configuration completed!", fg=typer.colors.GREEN
                )

            except Exception as e:
                typer.secho(
                    f"⚠️  Automated configuration failed: {e}", fg=typer.colors.YELLOW
                )
                typer.echo("📋 Trying manual configuration...")

                # Method 2: Manual configuration (fallback)
                try:
                    # Install dependencies
                    run_command(
                        ["sudo", "apt", "install", "-y", "curl", "ca-certificates"], env
                    )

                    # Create directory and download key
                    run_command(
                        ["sudo", "install", "-d", "/usr/share/postgresql-common/pgdg"],
                        env,
                    )

                    key_cmd = [
                        "sudo",
                        "curl",
                        "-o",
                        "/usr/share/postgresql-common/pgdg/apt.postgresql.org.asc",
                        "--fail",
                        "https://www.postgresql.org/media/keys/ACCC4CF8.asc",
                    ]
                    run_command(key_cmd, env)

                    # Get system codename
                    try:
                        codename_result = subprocess.run(
                            ["sh", "-c", ". /etc/os-release && echo $VERSION_CODENAME"],
                            capture_output=True,
                            text=True,
                        )
                        codename = (
                            codename_result.stdout.strip()
                            if codename_result.returncode == 0
                            else "bookworm"
                        )
                    except Exception:
                        codename = "bookworm"

                    typer.echo(f"📋 Detected system version: {codename}")

                    # Create repository configuration file
                    repo_content = f"deb [signed-by=/usr/share/postgresql-common/pgdg/apt.postgresql.org.asc] https://apt.postgresql.org/pub/repos/apt {codename}-pgdg main"
                    repo_cmd = [
                        "sudo",
                        "sh",
                        "-c",
                        f"echo '{repo_content}' > /etc/apt/sources.list.d/pgdg.list",
                    ]
                    run_command(repo_cmd, env)

                    typer.secho(
                        "✅ Manual configuration completed!", fg=typer.colors.GREEN
                    )

                except Exception as manual_error:
                    typer.secho(
                        f"❌ Manual configuration also failed: {manual_error}",
                        fg=typer.colors.RED,
                    )
                    raise typer.Exit(code=1)

            # Update package list
            typer.echo("🔄 Updating package list...")
            run_command(["sudo", "apt", "update"], env)

            # Install specific version client tools
            package = f"postgresql-client-{version}"
            typer.echo(f"📦 Package: {package}")
            run_command(["sudo", "apt", "install", "-y", package], env)

            # Check installation success
            pg_dump_path = f"/usr/lib/postgresql/{version}/bin/pg_dump"
            if os.path.exists(pg_dump_path):
                typer.echo(
                    f"✅ PostgreSQL {version} client tools installed to: /usr/lib/postgresql/{version}/bin/"
                )
                typer.echo("💡 You can create symbolic links:")
                typer.echo(f"   ./cli.py link-tools --version {version}")
                typer.echo("💡 Or run manually:")
                typer.echo(
                    f"   sudo ln -sf /usr/lib/postgresql/{version}/bin/* /usr/local/bin/"
                )
            else:
                typer.secho(
                    f"⚠️  Installation may be incomplete, not found: {pg_dump_path}",
                    fg=typer.colors.YELLOW,
                )

        else:
            typer.echo(
                "Current system not supported for automatic installation, please install PostgreSQL client tools manually."
            )
            typer.echo(
                f"💡 Please visit https://www.postgresql.org/download/ to download PostgreSQL {version}"
            )
            raise typer.Exit()
    else:
        typer.echo(
            f"Current system not supported for automatic installation: {os_name}"
        )
        typer.echo(
            f"💡 Please visit https://www.postgresql.org/download/ to download PostgreSQL {version}"
        )
        raise typer.Exit()

    typer.secho("✅ Installation completed!", fg=typer.colors.GREEN)

    # Verify installation
    if shutil.which("pg_dump"):
        try:
            result = subprocess.run(
                ["pg_dump", "--version"], capture_output=True, text=True
            )
            typer.echo(f"📋 Installed version: {result.stdout.strip()}")
        except Exception:
            pass


@app.command()
def check_version(
    host: str = typer.Option(
        DEFAULT_PG_HOST, "--host", "-h", help="PostgreSQL host address"
    ),
    port: str = typer.Option(DEFAULT_PG_PORT, "--port", "-p", help="PostgreSQL port"),
    user: str = typer.Option(
        DEFAULT_PG_USER, "--user", "-U", help="PostgreSQL username"
    ),
    password: str = typer.Option(
        DEFAULT_PG_PASSWORD, "--password", "-W", help="PostgreSQL password"
    ),
    database: str = typer.Option(
        DEFAULT_PG_DB, "--database", "-d", help="Database name"
    ),
):
    """
    Check PostgreSQL server version
    """
    db_config = DatabaseConfig(host, port, user, password, database)

    server_version = get_server_version(db_config)

    typer.echo("🔍 Version information:")
    typer.echo(f"  Server: {server_version}")

    # Extract version number
    try:
        import re

        server_match = re.search(r"(\d+)\.(\d+)", server_version)

        if server_match:
            server_major = int(server_match.group(1))
            typer.secho(
                f"✅ Server major version: {server_major}", fg=typer.colors.GREEN
            )
        else:
            typer.echo("Unable to parse server version information")
    except Exception:
        typer.echo("Unable to parse version information")


@app.command()
def pg_backup(
    host: str = typer.Option(
        DEFAULT_PG_HOST, "--host", "-h", help="PostgreSQL host address"
    ),
    port: str = typer.Option(DEFAULT_PG_PORT, "--port", "-p", help="PostgreSQL port"),
    user: str = typer.Option(
        DEFAULT_PG_USER, "--user", "-U", help="PostgreSQL username"
    ),
    password: str = typer.Option(
        DEFAULT_PG_PASSWORD, "--password", "-W", help="PostgreSQL password"
    ),
    database: str = typer.Option(
        DEFAULT_PG_DB, "--database", "-d", help="Database name"
    ),
    format: str = typer.Option(
        "custom", "--format", "-F", help="Backup format: plain, custom, directory, tar"
    ),
    compress: int = typer.Option(6, "--compress", "-Z", help="Compression level (0-9)"),
):
    """
    Backup database using pg_dump official tool (recommended)
    """
    # Check if pg_dump exists
    if not shutil.which("pg_dump"):
        typer.secho(
            "❌ pg_dump not found, please install PostgreSQL client tools first",
            fg=typer.colors.RED,
        )
        typer.secho("💡 Run: ./cli.py install-tools", fg=typer.colors.YELLOW)
        raise typer.Exit(code=1)

    timestamp = datetime.datetime.now().strftime("%Y%m%d_%H%M%S")

    if format == "directory":
        backup_file = BACKUP_DIR / f"{database}_pg_backup_{timestamp}"
        backup_file.mkdir(exist_ok=True)
    else:
        extension = {"plain": "sql", "custom": "dump", "tar": "tar"}.get(format, "dump")
        backup_file = BACKUP_DIR / f"{database}_pg_backup_{timestamp}.{extension}"

    env = os.environ.copy()
    env["PGPASSWORD"] = password

    cmd = [
        "pg_dump",
        "-h",
        host,
        "-p",
        port,
        "-U",
        user,
        "-d",
        database,
        "-F",
        format,
        "-f",
        str(backup_file),
        "--verbose",
        "--no-password",
    ]

    if format != "plain":
        cmd.extend(["-Z", str(compress)])

    typer.echo(f"🔄 Starting database backup {database}...")
    typer.echo(f"📁 Backup format: {format}")
    typer.echo(f"📦 Compression level: {compress}")

    try:
        run_command(cmd, env)
        typer.secho(
            f"✅ Database backup completed: {backup_file}", fg=typer.colors.GREEN
        )

        # Display backup file size
        if format == "directory":
            total_size = sum(
                os.path.getsize(os.path.join(dirpath, filename))
                for dirpath, dirnames, filenames in os.walk(backup_file)
                for filename in filenames
            )
        else:
            total_size = backup_file.stat().st_size

        size_mb = total_size / (1024 * 1024)
        typer.echo(f"📊 Backup size: {size_mb:.2f} MB")

    except Exception as e:
        typer.secho(f"❌ Backup failed: {e}", fg=typer.colors.RED)
        raise typer.Exit(code=1)


@app.command()
def pg_restore(
    backup_file: str,
    host: str = typer.Option(
        DEFAULT_PG_HOST, "--host", "-h", help="PostgreSQL host address"
    ),
    port: str = typer.Option(DEFAULT_PG_PORT, "--port", "-p", help="PostgreSQL port"),
    user: str = typer.Option(
        DEFAULT_PG_USER, "--user", "-U", help="PostgreSQL username"
    ),
    password: str = typer.Option(
        DEFAULT_PG_PASSWORD, "--password", "-W", help="PostgreSQL password"
    ),
    database: str = typer.Option(
        DEFAULT_PG_DB, "--database", "-d", help="Database name"
    ),
    clean: bool = typer.Option(
        False, "--clean", help="Clean database objects before restore"
    ),
    create: bool = typer.Option(False, "--create", help="Create database"),
):
    """
    Restore database using pg_restore official tool (recommended)
    """
    backup_path = BACKUP_DIR / backup_file
    if not backup_path.exists():
        typer.secho(
            f"❌ Backup file does not exist: {backup_path}", fg=typer.colors.RED
        )
        raise typer.Exit()

    # Detect backup format
    if backup_path.is_dir():
        format_type = "directory"
        restore_tool = "pg_restore"
    elif backup_file.endswith(".sql"):
        format_type = "plain"
        restore_tool = "psql"
    else:
        format_type = "custom/tar"
        restore_tool = "pg_restore"

    if not shutil.which(restore_tool):
        typer.secho(
            f"❌ {restore_tool} not found, please install PostgreSQL client tools first",
            fg=typer.colors.RED,
        )
        typer.secho("💡 Run: ./cli.py install-tools", fg=typer.colors.YELLOW)
        raise typer.Exit(code=1)

    env = os.environ.copy()
    env["PGPASSWORD"] = password

    typer.echo(f"🔄 Starting database restore {database}...")
    typer.echo(f"📁 Backup format: {format_type}")
    typer.echo(f"🛠️  Using tool: {restore_tool}")

    try:
        if restore_tool == "psql":
            # Use psql for plain format
            cmd = [
                "psql",
                "-h",
                host,
                "-p",
                port,
                "-U",
                user,
                "-d",
                database,
                "-f",
                str(backup_path),
                "--no-password",
                "-v",
                "ON_ERROR_STOP=1",
            ]
        else:
            # Use pg_restore for custom/tar/directory formats
            cmd = [
                "pg_restore",
                "-h",
                host,
                "-p",
                port,
                "-U",
                user,
                "-d",
                database,
                "--verbose",
                "--no-password",
                str(backup_path),
            ]

            if clean:
                cmd.append("--clean")
            if create:
                cmd.append("--create")

        run_command(cmd, env)
        typer.secho("✅ Database restore completed!", fg=typer.colors.GREEN)

    except Exception as e:
        typer.secho(f"❌ Restore failed: {e}", fg=typer.colors.RED)
        typer.secho(
            "💡 Tip: If it's a permission issue, try adding --clean or --create options",
            fg=typer.colors.YELLOW,
        )
        raise typer.Exit(code=1)


@app.command()
def pg_list():
    """
    List backup files created with pg_dump
    """
    backups = sorted(BACKUP_DIR.glob("*_pg_backup_*"))
    if not backups:
        typer.echo("No pg_dump backup files found.")
        return

    typer.echo("📋 Available pg_dump backup files:")
    for backup in backups:
        if backup.is_dir():
            # Directory format
            total_size = sum(
                os.path.getsize(os.path.join(dirpath, filename))
                for dirpath, dirnames, filenames in os.walk(backup)
                for filename in filenames
            )
            size_mb = total_size / (1024 * 1024)
            typer.echo(f"  📁 {backup.name}/ (directory format, {size_mb:.2f} MB)")
        else:
            # File format
            size_mb = backup.stat().st_size / (1024 * 1024)
            format_type = "plain" if backup.suffix == ".sql" else "custom/tar"
            typer.echo(f"  🗂 {backup.name} ({format_type}, {size_mb:.2f} MB)")


@app.command()
def help_backup():
    """
    Show PostgreSQL official backup tools description and recommendations
    """
    typer.echo("🛠️  PostgreSQL Official Backup and Restore Tools Guide\n")

    typer.secho("📦 Available Commands:", fg=typer.colors.GREEN, bold=True)
    typer.echo("  pg-backup     - Backup database using pg_dump")
    typer.echo("  pg-restore    - Restore database using pg_restore/psql")
    typer.echo("  pg-list       - List available backup files")
    typer.echo(
        "  auto-install  - Auto-detect server version and install matching client tools"
    )
    typer.echo(
        "  install-tools - Manually install specific version PostgreSQL client tools"
    )
    typer.echo("  link-tools    - Create symbolic links for PostgreSQL tools")
    typer.echo("  check-version - Check database version")
    typer.echo("")

    typer.secho("🚀 Quick Start:", fg=typer.colors.BLUE, bold=True)
    typer.echo("1. Auto install: ./cli.py auto-install")
    typer.echo("2. Start backup: ./cli.py pg-backup")
    typer.echo("3. View backups: ./cli.py pg-list")
    typer.echo("")

    typer.secho("🔧 Manual Version Management:", fg=typer.colors.BLUE, bold=True)
    typer.echo("1. Install PG16: ./cli.py install-tools --version 16")
    typer.echo("2. Create links: ./cli.py link-tools --version 16")
    typer.echo("3. Verify version: pg_dump --version")
    typer.echo("")

    typer.secho("💡 Recommended Usage:", fg=typer.colors.YELLOW, bold=True)
    typer.echo("1. Full backup: ./cli.py pg-backup --format custom --compress 9")
    typer.echo("2. Quick restore: ./cli.py pg-restore <backup_file> --clean")
    typer.echo("3. Text backup: ./cli.py pg-backup --format plain (readable SQL)")
    typer.echo("4. Directory backup: ./cli.py pg-backup --format directory (parallel)")
    typer.echo("")

    typer.secho("📋 Format Comparison:", fg=typer.colors.CYAN, bold=True)
    typer.echo("  custom   - Compressed binary, small size, fast restore (recommended)")
    typer.echo("  plain    - Pure SQL text, readable, large size")
    typer.echo(
        "  directory- Directory format, supports parallel, good for large databases"
    )
    typer.echo("  tar      - TAR archive, convenient for transfer")
    typer.echo("")

    typer.secho("🔧 Common Usage Examples:", fg=typer.colors.MAGENTA, bold=True)
    typer.echo("# Backup to remote server")
    typer.echo("./cli.py pg-backup --host 10.10.10.220 --port 32432")
    typer.echo("")
    typer.echo("# Restore to local database")
    typer.echo("./cli.py pg-restore backup.dump --host 127.0.0.1 --port 5432 --clean")
    typer.echo("")
    typer.echo("# Cross-server migration")
    typer.echo("./cli.py pg-backup --host old-server")
    typer.echo("./cli.py pg-restore backup.dump --host new-server --clean")


@app.command()
def link_tools(
    version: str = typer.Option(
        "16", "--version", "-v", help="PostgreSQL version (e.g.: 16, 15, 14)"
    ),
):
    """
    Create symbolic links for specific PostgreSQL version tools
    """
    os_name = platform.system()

    if os_name == "Linux":
        pg_bin_path = f"/usr/lib/postgresql/{version}/bin"

        if not os.path.exists(pg_bin_path):
            typer.secho(
                f"❌ PostgreSQL {version} path not found: {pg_bin_path}",
                fg=typer.colors.RED,
            )
            typer.secho(
                f"💡 Please run first: ./cli.py install-tools --version {version}",
                fg=typer.colors.YELLOW,
            )
            raise typer.Exit(code=1)

        typer.echo(f"🔗 Creating symbolic links for PostgreSQL {version}...")

        # Command to create symbolic links
        link_cmd = f"sudo ln -sf {pg_bin_path}/* /usr/local/bin/"

        try:
            subprocess.run(link_cmd, shell=True, check=True)
            typer.secho(
                "✅ Symbolic links created successfully!", fg=typer.colors.GREEN
            )

            # Verify
            try:
                result = subprocess.run(
                    ["pg_dump", "--version"], capture_output=True, text=True
                )
                typer.echo(f"📋 Current version: {result.stdout.strip()}")
            except Exception:
                pass

        except subprocess.CalledProcessError as e:
            typer.secho(f"❌ Failed to create symbolic links: {e}", fg=typer.colors.RED)
            typer.echo(f"💡 Please run manually: {link_cmd}")
            raise typer.Exit(code=1)

    elif os_name == "Darwin":
        typer.echo(
            "For macOS, please use Homebrew to manage PATH, or manually add to ~/.zshrc"
        )
        typer.echo(
            f'💡 Add this line to ~/.zshrc: export PATH="/opt/homebrew/opt/postgresql@{version}/bin:$PATH"'
        )
    else:
        typer.echo(f"Unsupported system: {os_name}")
        raise typer.Exit(code=1)


@app.command()
def auto_install(
    version: str = typer.Option(
        "16", "--version", "-v", help="PostgreSQL version (e.g.: 16, 15, 14)"
    ),
    server_host: str = typer.Option(
        DEFAULT_PG_HOST, "--server-host", help="Server address (for version detection)"
    ),
    server_port: str = typer.Option(
        DEFAULT_PG_PORT, "--server-port", help="Server port"
    ),
    server_user: str = typer.Option(
        DEFAULT_PG_USER, "--server-user", help="Server username"
    ),
    server_password: str = typer.Option(
        DEFAULT_PG_PASSWORD, "--server-password", help="Server password"
    ),
    server_database: str = typer.Option(
        DEFAULT_PG_DB, "--server-database", help="Database name"
    ),
):
    """
    Auto-detect server version and install matching client tools
    """
    typer.echo("🔍 Auto-installing PostgreSQL client tools...")

    # 1. Detect server version
    typer.echo("📡 Detecting server version...")
    db_config = DatabaseConfig(
        server_host, server_port, server_user, server_password, server_database
    )
    server_version = get_server_version(db_config)

    if server_version == "Unknown":
        typer.secho(
            "⚠️  Unable to detect server version, using specified version",
            fg=typer.colors.YELLOW,
        )
        target_version = version
    else:
        typer.echo(f"📋 Server version: {server_version}")

        # Extract major version
        import re

        version_match = re.search(r"(\d+)\.(\d+)", server_version)

        if version_match:
            target_version = version_match.group(1)
            typer.secho(
                f"🎯 Target version: PostgreSQL {target_version}", fg=typer.colors.GREEN
            )
        else:
            typer.secho(
                "⚠️  Unable to parse server version, using specified version",
                fg=typer.colors.YELLOW,
            )
            target_version = version

    # 2. Check current client version
    current_version = None
    if shutil.which("pg_dump"):
        try:
            result = subprocess.run(
                ["pg_dump", "--version"], capture_output=True, text=True
            )
            current_version_str = result.stdout.strip()
            typer.echo(f"📋 Current client: {current_version_str}")

            version_match = re.search(r"pg_dump.*?(\d+)", current_version_str)
            if version_match:
                current_version = version_match.group(1)

                if current_version == target_version:
                    typer.secho(
                        f"✅ Version already matches: PostgreSQL {current_version}",
                        fg=typer.colors.GREEN,
                    )
                    return
                else:
                    typer.secho(
                        f"⚠️  Version mismatch: Current {current_version}, Need {target_version}",
                        fg=typer.colors.YELLOW,
                    )
        except Exception:
            pass
    else:
        typer.echo("📋 PostgreSQL client tools not detected")

    # 3. Install matching version
    typer.echo(f"🔧 Installing PostgreSQL {target_version} client tools...")

    # Call install_tools command
    try:
        # Directly call install_tools function
        install_tools(target_version)

        # 4. Create symbolic links (Linux only)
        if platform.system() == "Linux":
            typer.echo("🔗 Creating symbolic links...")
            try:
                link_tools(target_version)
            except Exception as e:
                typer.secho(
                    f"⚠️  Symbolic link creation failed: {e}", fg=typer.colors.YELLOW
                )
                typer.echo(
                    f"💡 Please run manually: ./cli.py link-tools --version {target_version}"
                )

        # 5. Verify installation
        typer.echo("✅ Verifying installation...")
        if shutil.which("pg_dump"):
            try:
                result = subprocess.run(
                    ["pg_dump", "--version"], capture_output=True, text=True
                )
                typer.secho(
                    f"🎉 Installation successful: {result.stdout.strip()}",
                    fg=typer.colors.GREEN,
                )
            except Exception:
                pass

        typer.secho(
            "🚀 Auto-installation completed! You can now run backup commands.",
            fg=typer.colors.GREEN,
        )

    except Exception as e:
        typer.secho(f"❌ Auto-installation failed: {e}", fg=typer.colors.RED)
        typer.echo("💡 Please try manual installation:")
        typer.echo(f"   ./cli.py install-tools --version {target_version}")
        typer.echo(f"   ./cli.py link-tools --version {target_version}")
        raise typer.Exit(code=1)


if __name__ == "__main__":
    app()
