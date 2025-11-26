import os
import re
import subprocess
from collections import Counter

from rich.table import Table

from src.common.common import DEFAULT_SDK_DIR, console, repo


class PythonFormatter:
    """Handles Python code linting and formatting for staged files."""

    IGNORED_DIRS = {
        ".git",
        ".venv",
        "__pycache__",
        "dist",
        "build",
        "node_modules",
        "python-gen",
    }

    EXTRA_ARGS = ["--config", os.path.join(DEFAULT_SDK_DIR, "pyproject.toml")]

    def __init__(self):
        self.has_errors = False

    @staticmethod
    def get_staged_files() -> list[str]:
        """Get all staged Python files from git."""
        try:
            staged_changes = repo.index.diff(repo.head.commit, staged=True)

            python_files = []

            for diff_item in staged_changes:
                file_path = diff_item.b_path if diff_item.b_path else diff_item.a_path

                if file_path and file_path.endswith(".py"):
                    python_files.append(file_path)

            return python_files

        except Exception as e:
            console.print(f"[bold red]❌ Failed to get staged files: {e}[/bold red]")
            return []

    @staticmethod
    def get_all_files() -> list[str]:
        """
        Recursively finds all Python (.py) files in the current working directory,
        excluding common ignored directories like .git and .venv.
        """
        all_python_files = []

        for root, dirs, files in os.walk("."):
            dirs[:] = [d for d in dirs if d not in PythonFormatter.IGNORED_DIRS]
            for file in files:
                if file.endswith(".py"):
                    full_path = os.path.join(root, file)

                    if full_path.startswith("./"):
                        full_path = full_path[2:]

                    all_python_files.append(full_path)

        return all_python_files

    @staticmethod
    def get_sdk_files(sdk_dir: str) -> list[str]:
        """
        Recursively finds all Python (.py) files within the specified SDK directory,
        excluding common ignored directories.
        """
        sdk_python_files = []

        if not os.path.isdir(sdk_dir):
            console.print(
                f"[bold red]❌ Error: Directory not found at path: {sdk_dir}[/bold red]"
            )
            return []

        for root, dirs, files in os.walk(sdk_dir):
            dirs[:] = [d for d in dirs if d not in PythonFormatter.IGNORED_DIRS]

            for file in files:
                if file.endswith(".py"):
                    full_path = os.path.join(root, file)
                    normalized_path = os.path.normpath(full_path)
                    sdk_python_files.append(normalized_path)

        return sdk_python_files

    def _categorize_files(self, files: list[str]) -> dict[str, list[str]]:
        """Categorize files into sdk/python and other files."""
        sdk_files = []
        other_files = []

        for file in files:
            if file.startswith(DEFAULT_SDK_DIR):
                sdk_files.append(file)
            elif not file.startswith("src/"):
                other_files.append(file)

        return {
            "sdk": sdk_files,
            "other": other_files,
        }

    def _run_ruff_check(self, category: str, files: list[str]) -> bool:
        """Run ruff check --fix on files."""
        cmd = ["ruff", "check", "--fix", "--unsafe-fixes"]
        cmd.extend(files)
        if category == "sdk":
            cmd.extend(self.EXTRA_ARGS)

        try:
            process = subprocess.run(
                cmd, cwd=os.getcwd(), check=False, capture_output=True
            )
            if process.stderr:
                console.print(
                    f"[bold yellow]⚠️  Ruff check encountered issues: {process.stderr.decode()}[/bold yellow]"
                )
                return False

            return True
        except Exception as e:
            console.print(
                f"[bold yellow]⚠️  Ruff check encountered issues: {e}[/bold yellow]"
            )
            return False

    def _check_remaining_errors(self, category: str, files: list[str]) -> str | None:
        """Check for remaining errors after fix."""
        cmd = ["ruff", "check"]
        cmd.extend(files)
        if category == "sdk":
            cmd.extend(self.EXTRA_ARGS)

        try:
            result = subprocess.run(
                cmd, cwd=os.getcwd(), capture_output=True, text=True, check=False
            )
            return result.stdout if result.stdout else None
        except Exception as e:
            console.print(f"[bold yellow]⚠️  Failed to check errors: {e}[/bold yellow]")
            return None

    def _display_error_statistics(self, output: str) -> None:
        """Display error statistics in a formatted table."""
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

        # for line in output_contents[:20]:
        #     console.print(line)

        # if len(output_contents) > 20:
        #     console.print(
        #         f"\n[bright cyan]... (showing first 20 results of {len(content)} total)[/bright cyan]"
        #     )

        console.print(output)

    def _run_ruff_format(self, category: str, files: list[str]) -> bool:
        """Run ruff format on files."""
        cmd = ["ruff", "format"] + files
        cmd.extend(files)
        if category == "sdk":
            cmd.extend(self.EXTRA_ARGS)

        try:
            subprocess.run(cmd, cwd=os.getcwd(), check=True, capture_output=True)
            return True
        except subprocess.CalledProcessError as e:
            console.print(
                f"[bold yellow]⚠️  Ruff format encountered issues: {e}[/bold yellow]"
            )
            return False

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

    def run(self, files_to_format: list[str]) -> int:
        """Main execution flow."""
        console.print("[bold blue]🎨 Formatting Python files with ruff...[/bold blue]")

        # Process each category
        for category, files in self._categorize_files(files_to_format).items():
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
