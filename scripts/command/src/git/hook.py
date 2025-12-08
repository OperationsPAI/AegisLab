from src.common.common import LanguageType, ScopeType, console
from src.formatter.common import Formatter


def pre_commit():
    """Run pre-commit checks for Go and Python formatting."""
    console.print("[bold blue]Running pre-commit checks...[/bold blue]")
    console.print()

    console.print("[bold blue]Checking Go formatting...[/bold blue]")
    if Formatter.get_formatter(LanguageType.GO, ScopeType.STAGED).run() != 0:
        console.print(
            "[bold red]❌ Go formatting failed. Please fix the issues before committing.[/bold red]"
        )
        exit(1)
    console.print()

    console.print("[bold blue]Checking Python formatting...[/bold blue]")
    if Formatter.get_formatter(LanguageType.PYTHON, ScopeType.STAGED).run() != 0:
        console.print(
            "[bold red]❌ Python formatting failed. Please fix the issues before committing.[/bold red]"
        )
        exit(1)
    console.print()

    console.print("[bold green]✅ Pre-commit checks completed.[/bold green]")
