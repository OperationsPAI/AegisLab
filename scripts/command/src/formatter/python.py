import os
import re
from collections import Counter

from rich.table import Table

from src.common.command import run_command
from src.common.common import PROJECT_ROOT, ScopeType, console, settings
from src.formatter.common import (
    Formatter,
    get_modified_files_helper,
    get_staged_files_helper,
)


class PythonFormatter(Formatter):
    """Handles Python code linting and formatting."""

    IGNORED_DIRS = {
        ".git",
        ".venv",
        "__pycache__",
        "dist",
        "build",
        "node_modules",
        "python-gen",
    }
    REQUIRED_BINARIES = ["ruff"]
    SUFFIX = ".py"

    def __init__(self, scope: ScopeType = ScopeType.STAGED, sdk_dir: str | None = None):
        """
        Initialize PythonFormatter with a specific scope.

        Args:
            scope: The scope type determining which files to format
            sdk_dir: Custom SDK directory path (only used when scope is SDK)
        """
        self.scope = scope
        self.sdk_dir = sdk_dir or settings.python_sdk_dir
        self.has_errors = False
        self.extra_args = ["--config", os.path.join(self.sdk_dir, "pyproject.toml")]
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
        elif self.scope == ScopeType.SDK:
            files_to_format = self._get_sdk_files()
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
        Recursively finds all Python (.py) files in the current working directory,
        excluding common ignored directories.
        """
        all_python_files = []

        for root, dirs, files in os.walk(PROJECT_ROOT):
            dirs[:] = [d for d in dirs if d not in self.IGNORED_DIRS]
            for file in files:
                if file.endswith(self.SUFFIX):
                    full_path = os.path.join(root, file)

                    if full_path.startswith("./"):
                        full_path = full_path[2:]

                    all_python_files.append(full_path)

        return all_python_files

    def _get_sdk_files(self) -> list[str]:
        """
        Recursively finds all Python (.py) files within the specified SDK directory,
        excluding common ignored directories.
        """
        sdk_python_files = []

        if not os.path.isdir(self.sdk_dir):
            console.print(
                f"[bold red]❌ Error: Directory not found at path: {self.sdk_dir}[/bold red]"
            )
            return []

        for root, dirs, files in os.walk(self.sdk_dir):
            dirs[:] = [d for d in dirs if d not in self.IGNORED_DIRS]

            for file in files:
                if file.endswith(self.SUFFIX):
                    full_path = os.path.join(root, file)
                    normalized_path = os.path.normpath(full_path)
                    sdk_python_files.append(normalized_path)

        return sdk_python_files

    @staticmethod
    def install_tools():
        pass

    def _categorize_files(self) -> dict[str, list[str]]:
        """Categorize files into sdk/python and other files."""
        sdk_files = []
        other_files = []

        for file in self.files_to_format:
            if file.startswith(self.sdk_dir):
                sdk_files.append(file)
            elif not file.startswith("src/"):
                other_files.append(file)

        return {
            ScopeType.SDK.value: sdk_files,  # Use ScopeType enum value
            "other": other_files,
        }

    def _run_ruff_check(self, category: str, files: list[str]) -> bool:
        """Run ruff check --fix on files."""
        cmd = ["ruff", "check", "--fix", "--unsafe-fixes"]
        cmd.extend(files)
        if category == ScopeType.SDK.value:
            cmd.extend(self.extra_args)

        try:
            process = run_command(
                cmd, check=False, capture_output=True, cmd_print=False
            )
            if process.stderr:
                console.print(
                    f"[bold yellow]    ⚠️  Ruff check encountered issues: {process.stderr}[/bold yellow]"
                )
                return False

            return True
        except Exception as e:
            console.print(
                f"[bold yellow]    ⚠️  Ruff check encountered issues: {e}[/bold yellow]"
            )
            return False

    def _check_remaining_errors(self, category: str, files: list[str]) -> str | None:
        """Check for remaining errors after fix."""
        cmd = ["ruff", "check"]
        cmd.extend(files)
        if category == ScopeType.SDK.value:
            cmd.extend(self.extra_args)

        try:
            result = run_command(
                cmd, check=False, capture_output=True, text=True, cmd_print=False
            )
            if result.stdout:
                # Ensure we have a string, not bytes
                if isinstance(result.stdout, bytes):
                    return result.stdout.decode("utf-8")
                return result.stdout
            return None
        except Exception as e:
            console.print(
                f"[bold yellow]    ⚠️  Failed to check errors: {e}[/bold yellow]"
            )
            return None

    def _display_error_statistics(self, output: str) -> None:
        """Display error statistics in a formatted table."""
        # Ensure output is a string
        if not isinstance(output, str):
            console.print(
                f"[bold yellow]    ⚠️  Invalid output type: {type(output)}[/bold yellow]"
            )
            return

        pattern = r"([A-Z]{1,2}\d{3,4})"
        codes = re.findall(pattern, output)
        stats = dict(Counter(codes))
        if not stats:
            return

        total = sum(stats.values())
        self.has_errors = True

        console.print(f"\n[bold red]⚠️  Remaining issues: {total}[/bold red]")

        table = Table(
            title="📊 Issue Statistics by Type",
            show_header=True,
            header_style="cyan",
        )
        table.add_column("Error Code", style="cyan", width=12)
        table.add_column("Count", justify="right", style="red", width=8)

        for code, count in sorted(stats.items(), key=lambda x: x[1], reverse=True):
            table.add_row(code, str(count))

        console.print(table)

        console.print("\n[cyan]Details:[/cyan]")
        codes = [code for code in stats.keys()]
        code_pattern = r"(" + "|".join(re.escape(code) for code in codes) + r")"
        split_results = re.split(code_pattern, output.strip())

        current_code = ""
        output_contents: list[str] = []
        for i, item in enumerate(split_results):
            if i == 0:
                continue
            if i % 2 != 0:
                current_code = item
            else:
                content = item.strip()
                output_contents.append(f"{current_code} {content}")

        console.print(output)

    def _run_ruff_format(self, category: str, files: list[str]) -> bool:
        """Run ruff format on files."""
        cmd = ["ruff", "format"] + files
        cmd.extend(files)
        if category == ScopeType.SDK.value:
            cmd.extend(self.extra_args)

        run_command(cmd, check=True, capture_output=True, cmd_print=False)
        return True

    def _process_files(self, category: str, files: list[str]) -> None:
        """Process files in a given category."""
        if not files:
            return

        console.print("[gray]    Step 1/3: Running ruff check --fix...[/gray]")
        self._run_ruff_check(category, files=files)

        console.print("[gray]    Step 2/3: Checking remaining errors...[/gray]")
        output = self._check_remaining_errors(category, files=files)
        if output:
            self._display_error_statistics(output)

        console.print("[gray]    Step 3/3: Running ruff format...[/gray]")
        self._run_ruff_format(category, files=files)

    def run(self) -> int:
        """Main execution flow."""
        if not self.files_to_format:
            console.print("[bold yellow]No Python files to format.[/bold yellow]")
            return 0

        console.print("[bold blue]🎨 Formatting Python files with ruff...[/bold blue]")

        # Process each category
        for category, files in self._categorize_files().items():
            self._process_files(category, files=files)

        if self.has_errors:
            console.print(
                "\n[bold red]❌ Python formatting completed with errors[/bold red]"
            )
            console.print(
                "[bold yellow]💡 Please fix the remaining issues manually[/bold yellow]"
            )
            return 1
        else:
            console.print(
                "\n[bold green]✅ Python formatting completed successfully[/bold green]"
            )
            return 0
