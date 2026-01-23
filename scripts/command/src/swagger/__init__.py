from src.common.common import LanguageType
from src.swagger.common import Generator
from src.swagger.init import init
from src.swagger.python import PythonSDK
from src.swagger.typescript import TypeScriptClient

__all__ = ["init", "Generator"]

Generator.register_client(LanguageType.TYPESCRIPT, TypeScriptClient)

Generator.register_sdk(LanguageType.PYTHON, PythonSDK)
