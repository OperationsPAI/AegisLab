from enum import Enum
from pathlib import Path

from git import Repo
from rich.console import Console

__all__ = [
    "ENV",
    "console",
    "repo",
    "DEFAULT_SDK_DIR",
    "INITIAL_DATA_PATH",
]

DEFAULT_SDK_DIR = "sdk/python"
INITIAL_DATA_PATH = Path.cwd() / "helm" / "files" / "initial_data" / "data.json"

console = Console()  # Initialize a global console object for rich output
repo = Repo(".")  # Initialize the git repository at the current directory


class ENV(str, Enum):
    DEV = "dev"
    PROD = "prod"
    TEST = "test"
