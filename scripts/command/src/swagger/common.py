from abc import ABC
from enum import Enum

from src.common.common import PROJECT_ROOT, LanguageType

SWAGGER_ROOT = PROJECT_ROOT / "src" / "docs"
OPENAPI2_DIR = SWAGGER_ROOT / "openapi2"
OPENAPI3_DIR = SWAGGER_ROOT / "openapi3"
CONVERTED_DIR = SWAGGER_ROOT / "converted"


class RunMode(str, Enum):
    CLIENT = "client"
    SDK = "sdk"


class Generator(ABC):
    """Base generator class with factory pattern."""

    _client_registry: dict[LanguageType, type["Generator"]] = {}
    _sdk_registry: dict[LanguageType, type["Generator"]] = {}

    @classmethod
    def register_client(
        cls, name: LanguageType, generator_class: type["Generator"]
    ) -> None:
        """Register a client generator class with a name."""
        cls._client_registry[name] = generator_class

    @classmethod
    def register_sdk(
        cls, name: LanguageType, generator_class: type["Generator"]
    ) -> None:
        """Register a sdk generator class with a name."""
        cls._sdk_registry[name] = generator_class

    @staticmethod
    def get_client_generator(generator_type: LanguageType, version: str) -> "Generator":
        """Factory method to get a client generator instance based on type."""
        generator_class = Generator._client_registry.get(generator_type)
        if not generator_class:
            available = ", ".join(Generator._client_registry.keys())
            raise ValueError(
                f"Unknown client generator type: {generator_type}. Available: {available}"
            )

        return generator_class(version)

    @staticmethod
    def get_sdk_generator(generator_type: LanguageType, version: str) -> "Generator":
        """Factory method to get a sdk generator instance based on type."""
        generator_class = Generator._sdk_registry.get(generator_type)
        if not generator_class:
            available = ", ".join(Generator._sdk_registry.keys())
            raise ValueError(
                f"Unknown sdk generator type: {generator_type}. Available: {available}"
            )

        return generator_class(version)

    def __init__(self, version: str) -> None:
        self.version = version

    def generate(self) -> None:
        """Generate the client or SDK."""
        raise NotImplementedError
