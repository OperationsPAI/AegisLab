import typer

from src.common.common import console, settings
from src.formatter import PythonFormatter

app = typer.Typer()


@app.command(name="python")
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

    settings.reload()

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
        files_to_format = PythonFormatter.get_sdk_files(settings.python_sdk_dir)
        console.print(
            f"[yellow]Scope:[/yellow] Formatting specific directory: {settings.python_sdk_dir}"
        )

    console.print(
        f"\n[bright_cyan]Files to format (total: {len(files_to_format)}):[/bright_cyan]"
    )
    for file in files_to_format:
        console.print(f"  [dim]-[/dim] {file}")
    console.print()

    formatter = PythonFormatter()
    formatter.run(files_to_format)
