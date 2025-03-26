from typing import Any, Dict, Optional
from datetime import datetime
import logging
import sys


class _ColorCodes:
    """ANSI 颜色代码"""

    DEBUG = "\033[36m"
    INFO = "\033[32m"
    WARNING = "\033[33m"
    ERROR = "\033[31m"
    RESET = "\033[0m"  # 重置颜色


class _ColoredFormatter(logging.Formatter):
    """按日志级别着色的格式化器"""

    def __init__(self, fmt: str, datefmt: str = None):
        super().__init__(fmt, datefmt)

    def format(self, record: logging.LogRecord) -> str:
        ori_message = super().format(record)

        color = getattr(_ColorCodes, record.levelname, _ColorCodes.RESET)
        levelname = record.levelname
        colored_level = f"{color}{levelname}{_ColorCodes.RESET}"

        return ori_message.replace(levelname, colored_level)


class CustomLogger:
    _instance = None

    def __new__(
        cls,
        name: str = "RCABench SDK",
        log_level: str = "INFO",
        enable_color: bool = True,
        enable_file: bool = False,
        filename: Optional[str] = None,
    ):
        if cls._instance is None:
            cls._instance = super().__new__(cls)
            cls._instance.__init_logger(
                name, log_level, enable_color, enable_file, filename
            )
        return cls._instance

    @staticmethod
    def _get_console_formatter(enable_color) -> logging.Formatter:
        """控制台专用格式化器（支持颜色）"""
        base_fmt = "[%(asctime)s] [%(levelname)s] [%(module)s:%(lineno)d] - %(message)s"
        if enable_color:
            return _ColoredFormatter(base_fmt, datefmt="%Y-%m-%d %H:%M:%S")
        else:
            return logging.Formatter(base_fmt, datefmt="%Y-%m-%d %H:%M:%S")

    def __init_logger(
        self,
        name: str,
        log_level: str,
        enable_color: bool,
        enable_file: bool,
        filename: Optional[str],
    ) -> None:
        self.logger = logging.getLogger(name)
        self.logger.setLevel(log_level.upper())
        self.logger.propagate = False

        for handler in self.logger.handlers[:]:
            self.logger.removeHandler(handler)

        console_handler = logging.StreamHandler(sys.stdout)
        console_handler.setFormatter(self._get_console_formatter(enable_color))
        self.logger.addHandler(console_handler)

        if enable_file and filename:
            file_handler = logging.FileHandler(filename, encoding="utf-8")
            self.logger.addHandler(file_handler)

    def log(
        self,
        level: str,
        message: str,
        extra: Dict[str, Any] = None,
        exc_info: bool = False,
    ) -> None:
        log_method = getattr(self.logger, level.lower(), self.logger.info)
        extra = extra or {}
        extra.update({"timestamp": datetime.isoformat()})
        log_method(message, extra=extra, exc_info=exc_info)

    def info(self, message: str, **kwargs) -> None:
        self.log("INFO", message, **kwargs)

    def error(self, message: str, **kwargs) -> None:
        self.log("ERROR", message, **kwargs)


logger = CustomLogger().logger
