import os
from abc import ABC

from git import Repo
from src.common.common import PROJECT_ROOT, LanguageType, ScopeType, console

repo = Repo(PROJECT_ROOT)  # Initialize the git repository at the current directory

__all__ = ["Formatter", "get_staged_files_helper", "get_modified_files_helper"]


class Formatter(ABC):
    """Base formatter class with factory pattern."""

    _registry: dict[LanguageType, type["Formatter"]] = {}

    @classmethod
    def register(cls, name: LanguageType, formatter_class: type["Formatter"]) -> None:
        """Register a formatter class with a name."""
        cls._registry[name] = formatter_class

    @staticmethod
    def get_formatter(formatter_type: LanguageType, scope: ScopeType) -> "Formatter":
        """Factory method to get a formatter instance based on type and scope."""
        formatter_class = Formatter._registry.get(formatter_type)
        if not formatter_class:
            available = ", ".join(Formatter._registry.keys())
            raise ValueError(
                f"Unknown formatter type: {formatter_type}. Available: {available}"
            )

        return formatter_class(scope)  # type: ignore

    def run(self) -> int:
        """Run the formatter on the provided files."""
        raise NotImplementedError


def get_staged_files_helper(suffies: list[str]) -> list[str]:
    """Get staged files with specific suffixes (current behavior)."""
    try:
        staged_changes = repo.head.commit.diff()
        files = []

        for diff_item in staged_changes:
            # Skip deleted files
            if diff_item.change_type == "D":
                continue

            # For all other types (A, M, R, C, T), use b_path (new file in index)
            file_path = diff_item.b_path

            if file_path and os.path.splitext(file_path)[-1] in suffies:
                files.append(file_path)

        return files

    except Exception as e:
        console.print(f"[bold red]❌ Failed to get staged files: {e}[/bold red]")
        return []


def get_modified_files_helper(
    suffies: list[str], include_untracked: bool = False
) -> list[str]:
    """
    Get modified files (staged + unstaged), optionally including untracked.

    Args:
        suffies: File suffixes to filter (e.g., ['.go', '.py'])
        include_untracked: Whether to include untracked files
    """
    try:
        files = []

        staged_changes = repo.head.commit.diff()
        for diff_item in staged_changes:
            if diff_item.change_type == "D":
                continue

            file_path = diff_item.b_path
            if file_path and os.path.splitext(file_path)[-1] in suffies:
                files.append(file_path)

        unstaged_changes = repo.index.diff(None)
        for diff_item in unstaged_changes:
            if diff_item.change_type == "D":
                continue

            file_path = diff_item.b_path
            if file_path and os.path.splitext(file_path)[-1] in suffies:
                if file_path not in files:
                    files.append(file_path)

        # Untracked files (optional)
        if include_untracked:
            untracked_files = repo.untracked_files
            for file_path in untracked_files:
                if os.path.splitext(file_path)[-1] in suffies:
                    if file_path not in files:
                        files.append(file_path)

        return files

    except Exception as e:
        console.print(f"[bold red]❌ Failed to get modified files: {e}[/bold red]")
        return []
