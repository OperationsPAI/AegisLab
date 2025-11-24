import os
from enum import Enum
from pathlib import Path

import toml
from dotenv import load_dotenv
from git import Repo
from rich.console import Console

__all__ = [
    "ENV",
    "console",
    "repo",
    "DEFAULT_SDK_DIR",
    "INITIAL_DATA_PATH",
    "RELEASE_NAME",
    "PROJECT_ROOT",
    "HELM_CHART_PATH",
]


DEFAULT_SDK_DIR = "sdk/python"
DEFAULT_REPO_URL = "10.10.10.240/library"
DEFAULT_PORT = 30080

PROJECT_ROOT = Path(__file__).parent.parent.parent.parent.parent
HELM_CHART_PATH = PROJECT_ROOT / "helm"
INITIAL_DATA_PATH = PROJECT_ROOT / "helm" / "files" / "initial_data" / "data.json"

K8S_NAMESPACE = "exp"
RELEASE_NAME = "rcabench"
TIME_FORMAT = "%Y%m%d_%H%M%S"


class ENV(str, Enum):
    DEV = "dev"
    PROD = "prod"
    TEST = "test"


class SourceType(str, Enum):
    LOCAL = "local"
    REMOTE = "remote"


load_dotenv(PROJECT_ROOT / ".env")

with open(PROJECT_ROOT / "src" / "config.dev.toml") as f:
    dev_config = toml.load(f)

console = Console()  # Initialize a global console object for rich output
repo = Repo(PROJECT_ROOT)  # Initialize the git repository at the current directory

k8s_context_mapping: dict[ENV, str] = {
    ENV.DEV: os.getenv("DEV_CONTEXT", ""),
    ENV.TEST: os.getenv("TEST_CONTEXT", ""),
}
