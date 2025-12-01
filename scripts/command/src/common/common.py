from enum import Enum
from pathlib import Path

from dynaconf import Dynaconf
from git import Repo
from rich.console import Console

__all__ = [
    "ENV",
    "console",
    "repo",
    "PROJECT_ROOT",
    "HELM_CHART_PATH",
    "INITIAL_DATA_PATH",
    "settings",
]


PROJECT_ROOT = Path(__file__).parent.parent.parent.parent.parent
HELM_CHART_PATH = PROJECT_ROOT / "helm"
INITIAL_DATA_PATH = PROJECT_ROOT / "helm" / "files" / "initial_data" / "data.json"


class ENV(str, Enum):
    DEV = "dev"
    PROD = "prod"
    TEST = "test"


class SourceType(str, Enum):
    LOCAL = "local"
    REMOTE = "remote"


settings = Dynaconf(
    root_path=Path(__file__).parent.parent.parent,
    settings_files=["settings.toml"],
    environments=True,
    envvar_prefix="DYNACONF",
    dotenv_path=Path(__file__).parent.parent.parent / ".env",
)


console = Console()  # Initialize a global console object for rich output
repo = Repo(PROJECT_ROOT)  # Initialize the git repository at the current directory
