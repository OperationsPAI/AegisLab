from enum import Enum

from src.common.common import PROJECT_ROOT

SWAGGER_ROOT = PROJECT_ROOT / "src" / "docs"
OPENAPI2_DIR = SWAGGER_ROOT / "openapi2"
OPENAPI3_DIR = SWAGGER_ROOT / "openapi3"
CONVERTED_DIR = SWAGGER_ROOT / "converted"


class RunMode(str, Enum):
    SDK = "sdk"
    RUNTIME = "runtime"
    PORTAL = "portal"
    ADMIN = "admin"
