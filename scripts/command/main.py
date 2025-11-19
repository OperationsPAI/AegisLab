import typer

from src.common import DEFAULT_SDK_DIR, ENV, console

main_app = typer.Typer(help="Main application for command-line interface.")

formatter_app = typer.Typer(help="Code formatting utilities.")
main_app.add_typer(formatter_app, name="format")

pedestal_app = typer.Typer(help="Pedestal management utilities.")
main_app.add_typer(pedestal_app, name="pedestal")


@formatter_app.command()
def python(
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


@pedestal_app.command()
def install_releases(
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

    pedestal_module.install_releases(env, name=name, count=count)


if __name__ == "__main__":
    main_app()
