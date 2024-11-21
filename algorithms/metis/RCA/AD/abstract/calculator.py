from abc import ABC, abstractmethod


class Calculator(ABC):
    @staticmethod
    @abstractmethod
    async def calculate(self, data, normal): ...
