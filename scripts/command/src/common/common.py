from enum import Enum
from pathlib import Path

from dynaconf import Dynaconf
from rich.console import Console

__all__ = [
    "ENV",
    "console",
    "PROJECT_ROOT",
    "HELM_CHART_PATH",
    "INITIAL_DATA_PATH",
    "settings",
]


PROJECT_ROOT = Path(__file__).parent.parent.parent.parent.parent
COMMAND_ROOT_PATH = Path(__file__).parent.parent.parent

DOTENV_PATH = PROJECT_ROOT / ".env"
HELM_CHART_PATH = PROJECT_ROOT / "helm"
INITIAL_DATA_PATH = PROJECT_ROOT / "helm" / "files" / "initial_data" / "data.json"

DATAPACK_ROOT_PATH = Path("/mnt/jfs/rcabench_dataset")


class ENV(str, Enum):
    DEV = "dev"
    PROD = "prod"
    STAGING = "staging"
    TEST = "test"


class LanguageType(str, Enum):
    GO = "go"
    PYTHON = "python"


class ScopeType(str, Enum):
    ALL = "all"
    Modified = "modified"
    SDK = "sdk"
    STAGED = "staged"


settings = Dynaconf(
    root_path=COMMAND_ROOT_PATH,
    settings_files=["settings.toml"],
    load_dotenv=True,
    environments=True,
    envvar_prefix=False,
    dotenv_path=DOTENV_PATH,
)


console = Console()  # Initialize a global console object for rich output
