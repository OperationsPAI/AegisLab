from abc import ABC, abstractmethod
from pathlib import Path


class Parser(ABC):
    def __init__(self, base_dir):
        self.base_dir = Path(base_dir)

    @abstractmethod
    async def parse(self, pool): ...
