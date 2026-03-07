"""
CLI commands for etcd configuration management.

This module provides Typer CLI commands for managing etcd configurations.
"""

import typer
from rich.panel import Panel

from src.common.common import ENV, console, settings
from src.etcd import clear_etcd_configs, init_etcd_configs, list_etcd_configs

app = typer.Typer(help="etcd configuration management utilities")


@app.command()
def init(
    env: ENV = typer.Option(ENV.DEV, "--env", "-e", help="Environment: dev or staging"),
    force: bool = typer.Option(
        False, "--force", "-f", help="Force overwrite existing values in etcd"
    ),
    dry_run: bool = typer.Option(
        False, "--dry-run", help="Show what would be done without making changes"
    ),
    values_file: str = typer.Option(
        None,
        "--values-file",
        "-v",
        help='JSON file with custom values (format: {"config_key": "value"})',
    ),
):
    """Initialize etcd with consumer configurations from data.yaml"""

    if env == ENV.PROD:
        console.print(
            f"[bold red]❌ Initialization in {env.value} environment is not allowed.[/bold red]"
        )
        raise typer.Exit(1)

    settings.reload()

    console.print(
        Panel.fit(
            "[bold cyan]etcd Configuration Initializer[/bold cyan]",
            border_style="cyan",
        )
    )

    try:
        init_etcd_configs(env, force=force, dry_run=dry_run, values_file=values_file)
    except FileNotFoundError as e:
        console.print(f"[bold red]❌ {e}[/bold red]")
        raise typer.Exit(1)
    except Exception as e:
        console.print(f"[bold red]❌ Failed to initialize etcd: {e}[/bold red]")
        raise typer.Exit(1)


@app.command()
def list(
    env: ENV = typer.Option(ENV.DEV, "--env", "-e", help="Target Environment"),
):
    """List all consumer configurations in etcd"""

    settings.reload()

    console.print(
        Panel.fit(
            "[bold cyan]Current Consumer Configurations[/bold cyan]",
            border_style="cyan",
        )
    )

    console.print()

    try:
        list_etcd_configs(env)
    except Exception as e:
        console.print(f"[bold red]❌ Failed to list configurations: {e}[/bold red]")
        raise typer.Exit(1)


@app.command()
def clear(
    env: ENV = typer.Option(ENV.DEV, "--env", "-e", help="Environment: dev or staging"),
    yes: bool = typer.Option(False, "--yes", "-y", help="Skip confirmation prompt"),
):
    """Clear all consumer configurations from etcd"""

    if env == ENV.PROD:
        console.print(
            f"[bold red]❌ Clearing configurations in {env.value} environment is not allowed.[/bold red]"
        )
        raise typer.Exit(1)

    settings.reload()

    console.print(
        Panel.fit(
            "[bold red]⚠️  Clear etcd Configurations[/bold red]",
            border_style="red",
        )
    )

    etcd_prefix = settings.etcd.prefix
    console.print(
        f"[bold yellow]This will delete all configurations under {etcd_prefix}[/bold yellow]"
    )

    if not yes:
        confirm = typer.confirm("Are you sure you want to proceed?")
        if not confirm:
            console.print("[bold]Operation cancelled[/bold]")
            raise typer.Exit(0)

    try:
        clear_etcd_configs(env)
    except Exception as e:
        console.print(f"[bold red]❌ Failed to clear configurations: {e}[/bold red]")
        raise typer.Exit(1)
