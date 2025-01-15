from typing import Callable, Dict, List
from functools import wraps
from rich.console import Console
from rich.prompt import Prompt


def select_choice(console: Console, prompt: str, choices: List):
    while True:
        choice = Prompt.ask(
            f"选择{prompt}",
            choices=choices,
            default=choices[0],
            show_choices=True,
        )
        if choice in choices:
            break

        console.print(f"输入{prompt} [bold red]{choice}[/bold red] 不存在")

    return choice


def select_function(func: Callable):
    @wraps(func)
    def wrapper(config: Dict):
        keys = list(config.keys())
        if "func" not in keys:
            selected_func_config = Prompt.ask(
                "选择要执行的函数",
                choices=keys,
                default=keys[0],
                show_choices=True,
            )
            return wrapper(config[selected_func_config])
        else:
            return func(config)

    return wrapper
