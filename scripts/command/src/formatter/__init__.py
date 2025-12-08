from src.common.common import LanguageType
from src.formatter.common import Formatter
from src.formatter.go import GoFormatter
from src.formatter.python import PythonFormatter

Formatter.register(LanguageType.GO, GoFormatter)
Formatter.register(LanguageType.PYTHON, PythonFormatter)
