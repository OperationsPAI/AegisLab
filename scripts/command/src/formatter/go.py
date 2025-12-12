import os

from src.common.command import run_command
from src.common.common import PROJECT_ROOT, ScopeType, console
from src.formatter.common import (
    Formatter,
    get_modified_files_helper,
    get_staged_files_helper,
)


class GoFormatter(Formatter):
    """Handles Go code linting and formatting."""

    IGNORED_DIRS = []
    REQUIRED_BINARIES = ["go", "golangci-lint"]
    SUFFIX = ".go"

    def __init__(self, scope: ScopeType = ScopeType.STAGED):
        """
        Initialize GoFormatter with a specific scope.

        Args:
            scope: The scope type determining which files to format
        """
        self.scope = scope
        self.files_to_format = self._get_files()

    def _get_files(self) -> list[str]:
        """
        Get files to format based on the configured scope.

        Returns:
            List of file paths to format
        """
        files_to_format: list[str] = []
        if self.scope == ScopeType.STAGED:
            files_to_format = get_staged_files_helper(suffies=[self.SUFFIX])
        elif self.scope == ScopeType.Modified:
            files_to_format = get_modified_files_helper(
                suffies=[self.SUFFIX], include_untracked=True
            )
        elif self.scope == ScopeType.ALL:
            files_to_format = self._get_all_files()
        else:
            console.print(f"[bold yellow]⚠️  Unknown scope: {self.scope}[/bold yellow]")
            raise ValueError("Unknown scope type")

        console.print(
            f"\n[cyan]Files to format (total: {len(files_to_format)}):[/cyan]"
        )

        for file in files_to_format:
            console.print(f"    [dim]-[/dim] {file}")
        console.print()

        return files_to_format

    def _get_all_files(self) -> list[str]:
        """
        Recursively finds all Go (.go) files in the current working directory,
        excluding common ignored directories.
        """
        all_go_files = []

        for root, dirs, files in os.walk(PROJECT_ROOT):
            dirs[:] = [d for d in dirs if d not in self.IGNORED_DIRS]
            for file in files:
                if file.endswith(self.SUFFIX):
                    full_path = os.path.join(root, file)

                    if full_path.startswith("./"):
                        full_path = full_path[2:]

                    all_go_files.append(full_path)

        return all_go_files

    def run(self) -> int:
        """Main execution flow."""
        if not self.files_to_format:
            console.print("[bold yellow]No Go files to format.[/bold yellow]")
            return 0

        console.print(
            "[bold blue]🎨 Formatting Go files with golangci-lint...[/bold blue]"
        )

        try:
            if self.scope == ScopeType.STAGED:
                run_command(
                    [
                        "golangci-lint",
                        "run",
                        "--issues-exit-code=1",
                        "--path-prefix=src",
                        "--whole-files",
                        "--new-from-rev=HEAD~1",
                    ],
                    cwd=PROJECT_ROOT / "src",
                    check=True,
                )
                console.print(
                    "[bold green]✅ Go formatting completed successfully[/bold green]"
                )
                return 0
            else:
                raise NotImplementedError(
                    "Only STAGED scope is implemented for Go formatting."
                )
        except Exception as e:
            console.print(
                "\n[bold red]❌ Go formatting completed with errors[/bold red]"
            )
            console.print(f"[bold red]Error details: {e}[/bold red]")
            return 1
