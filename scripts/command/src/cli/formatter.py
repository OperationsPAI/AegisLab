import typer

from src.common.common import LanguageType, ScopeType, settings
from src.formatter import Formatter

app = typer.Typer()


@app.command(name="go")
def format_go(
    scope: ScopeType = typer.Option(
        ScopeType.STAGED,
        "--scope",
        "-s",
        case_sensitive=False,
        help="Scope of files to format: 'staged' for staged files, 'modified' for all uncommited files, 'all' for all working files.",
    ),
):
    """Formats Go files based on the specified scope."""

    settings.reload()

    Formatter.get_formatter(LanguageType.GO, scope).run()


@app.command(name="python")
def format_python(
    scope: ScopeType = typer.Option(
        ScopeType.STAGED,
        "--scope",
        "-s",
        case_sensitive=False,
        help="Scope of files to format: 'staged' for staged files, 'modified' for all uncommited files, 'all' for all working files, 'sdk' for SDK files only.",
    ),
):
    """Formats Python files based on the specified scope."""

    settings.reload()

    Formatter.get_formatter(LanguageType.PYTHON, scope).run()
