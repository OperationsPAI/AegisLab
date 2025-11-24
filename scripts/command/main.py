import typer

from src.common.common import DEFAULT_SDK_DIR, ENV, PROJECT_ROOT, SourceType, console

main_app = typer.Typer(help="Main application for command-line interface.")


chaos_app = typer.Typer(help="Chaos engineering utilities.")
main_app.add_typer(chaos_app, name="chaos")

formatter_app = typer.Typer(help="Code formatting utilities.")
main_app.add_typer(formatter_app, name="formatter")

pedestal_app = typer.Typer(help="Pedestal management utilities.")
main_app.add_typer(pedestal_app, name="pedestal")

swagger_app = typer.Typer(help="Swagger documentation utilities.")
main_app.add_typer(swagger_app, name="swagger")

rcabench_app = typer.Typer(help="RCABench utilities.")
main_app.add_typer(rcabench_app, name="rcabench")

swagger_app = typer.Typer(help="Swagger documentation utilities.")
main_app.add_typer(swagger_app, name="swagger")

mysql_app = typer.Typer(help="MySQL backup and restore utilities.")
main_app.add_typer(mysql_app, name="mysql")

reids_app = typer.Typer(help="Redis restore utilities.")
main_app.add_typer(reids_app, name="redis")


@chaos_app.command(name="clean-finalizers")
def clean_chaos_finalizers(
    env: ENV = typer.Option(
        ENV.DEV,
        "--env",
        "-e",
        help="Target environment (e.g., dev, test).",
    ),
    ns_prefix: str = typer.Option(
        "aegislab-chaos-",
        "--ns-prefix",
        "-p",
        help="Namespace prefix to filter target namespaces.",
    ),
    ns_count: int = typer.Option(
        10,
        "--ns-count",
        "-c",
        help="Number of namespaces to process.",
    ),
):
    """Cleans finalizers from chaos resources in specified namespaces."""
    import src.chaos as chaos_module

    chaos_module.clean_finalziers(env, ns_prefix, ns_count)


@chaos_app.command(name="delete-chaos")
def delete_chaos_resources(
    env: ENV = typer.Option(
        ENV.DEV,
        "--env",
        "-e",
        help="Target environment (e.g., dev, test).",
    ),
    ns_prefix: str = typer.Option(
        "aegislab-chaos-",
        "--ns-prefix",
        "-p",
        help="Namespace prefix to filter target namespaces.",
    ),
    ns_count: int = typer.Option(
        10,
        "--ns-count",
        "-c",
        help="Number of namespaces to process.",
    ),
):
    """Deletes chaos resources in specified namespaces."""
    import src.chaos as chaos_module

    chaos_module.delete_chaos_resources(env, ns_prefix, ns_count)


@formatter_app.command(name="python")
def format_python(
    # Option 1: Staged Files (Flag)
    staged: bool = typer.Option(
        False,
        "--staged",
        "-s",
        help="Format only files currently staged in Git (for pre-commit).",
    ),
    # Option 2: All Working Files (Flag)
    all_files: bool = typer.Option(
        False,
        "--all",
        "-a",
        help="Format all Python files in the repository working directory.",
    ),
):
    """Formats Python files based on the specified scope."""

    from src.formatter import PythonFormatter

    # Mutual Exclusivity Check
    active_scopes = sum([staged, all_files])

    if active_scopes > 1:
        console.print(
            "[bright_red]❌ Error: Only one scope can be specified: --staged, --all.[/bright_red]",
        )
        raise typer.Exit(code=1)

    files_to_format: list[str] = []
    if staged:
        files_to_format = PythonFormatter.get_staged_files()
        console.print("[yellow]Scope:[/yellow] Formatting STAGED files...")

    elif all_files:
        files_to_format = PythonFormatter.get_all_files()
        console.print("[yellow]Scope:[/yellow] Formatting ALL working files...")

    else:
        files_to_format = PythonFormatter.get_sdk_files(DEFAULT_SDK_DIR)
        console.print(
            f"[yellow]Scope:[/yellow] Formatting specific directory: {DEFAULT_SDK_DIR}"
        )

    console.print(
        f"\n[bright_cyan]Files to format (total: {len(files_to_format)}):[/bright_cyan]"
    )
    for file in files_to_format:
        console.print(f"  [dim]-[/dim] {file}")
    console.print()

    formatter = PythonFormatter()
    formatter.run(files_to_format)


@pedestal_app.command(name="install")
def install_pedestals(
    env: ENV = typer.Option(
        ENV.DEV,
        "--env",
        "-e",
        help="Target environment (e.g., dev, test).",
    ),
    name: str = typer.Option(
        ...,
        "--name",
        "-n",
        help="Pedestal container name to install",
    ),
    count: int = typer.Option(
        ...,
        "--count",
        "-c",
        help="Number of pedestal releases to install.",
    ),
):
    """Installs multiple pedestal Helm releases based on the specified container name and count."""
    import src.pedestal as pedestal_module

    pedestal_module.install_pedestals(env, name=name, count=count)


@rcabench_app.command(name="check-secrets")
def rcabench_check_secrets(
    env: ENV = typer.Option(
        ENV.DEV,
        "--env",
        "-e",
        help="Target environment (e.g., dev, test).",
    ),
):
    """Checks for the presence of required Kubernetes secrets for RCABench."""
    import src.rcabench_ as rcabench_module

    rcabench_module.check_secrets(env)


@rcabench_app.command(name="local-deploy")
def rcabench_local_deploy(
    env: ENV = typer.Option(
        ENV.DEV,
        "--env",
        "-e",
        help="Target environment (e.g., dev, test).",
    ),
):
    """Deploys RCABench locally using Docker Compose."""
    import src.rcabench_ as rcabench_module

    rcabench_module.local_deploy(env)

    if typer.confirm("Do you want to perform data migrations now?"):
        rcabench_module.check_db(env)
        rcabench_module.check_redis(env)

        mysql_migrate(src=SourceType.REMOTE, force=True)
        redis_migrate(src=SourceType.REMOTE, force=True, dry_run=False)

    console.print()
    console.print(
        "[bold yellow]You can start the application manually later: [/bold yellow]"
    )
    console.print(
        f"[gray]cd {PROJECT_ROOT / 'src'} && go run main.go both --port 8082 [/gray]"
    )


@rcabench_app.command(name="run")
def rcabench_execute_release_workflow(
    env: ENV = typer.Option(
        ENV.DEV,
        "--env",
        "-e",
        help="Target environment (e.g., dev, test).",
    ),
):
    """Executes the RCABench release workflow."""
    import src.rcabench_ as rcabench_module

    rcabench_module.execute_release_workflow(env)


@mysql_app.command(name="migrate")
def mysql_migrate(
    src: str = typer.Option(
        "local",
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
    from src.backup.mysql import MysqlClient, install_tools

    console.print("[bold blue]Starting database migration...[/bold blue]")

    console.print("[bold blue]Step 1: Installing necessary tools...[/bold blue]")
    install_tools()
    console.print()

    source_type = SourceType.LOCAL if src.lower() == "local" else SourceType.REMOTE
    client = MysqlClient(source_type)

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
    from src.backup.redis_ import RedisClient

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


@swagger_app.command()
def swagger_init():
    """Initializes Swagger documentation setup."""
    from src.swagger import init

    init()


@swagger_app.command()
def generate_sdk(
    language: str = typer.Option(
        "python",
        "--language",
        "-l",
        help="SDK language.",
    ),
    version: str = typer.Option(
        "1.0.0",
        "--version",
        "-v",
        help="API version.",
    ),
):
    """Generates SDK Swagger documentation."""
    from src.swagger import generate_python_sdk

    if language.lower() == "python":
        generate_python_sdk(version)


if __name__ == "__main__":
    main_app()
