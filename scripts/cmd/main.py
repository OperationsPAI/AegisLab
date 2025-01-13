from typing import Any, Dict
from rich.console import Console
from rich.prompt import Prompt
from utils import Executor


def get_input(prompt: str, default: Any = None) -> Any:
    """获取用户输入"""
    return Prompt.ask(prompt, default=default)


def select_and_execute(config: Dict):
    keys = list(config.keys())
    if "func" not in keys:
        selected_func_config = Prompt.ask(
            "选择要执行的函数",
            choices=keys,
            default=keys[0],
            show_choices=True,
        )

        select_and_execute(config[selected_func_config])
    else:
        func = config["func"]
        kwargs = {}
        params = config.get("params", None)

        if params:
            for param in params:
                name = param["name"]
                kwargs[name] = Prompt.ask(
                    f"请输入 {name} 的值",
                    choices=param.get("choices", None),
                    default=param.get("default", None),
                    show_choices="choices" in list(param.keys()),
                )

        func(**kwargs)


if __name__ == "__main__":
    console = Console()

    algo_library = get_input("请输入算法库名称", default="rcaeval")
    src_dir = get_input(
        "请输入算法库源位置", default="/home/nn/workspace/lib/RCAEval/RCAEval/e2e"
    )
    executor = Executor(console, algo_library, src_dir)

    config = {
        "Create": {
            "func": executor.create,
            "params": [
                {
                    "name": "mode",
                    "help": "算法类别",
                    "choices": ["cpu", "gpu", "both"],
                    "default": "cpu",
                }
            ],
        },
        "Delete": {"func": executor.delete},
        "ToPostman": {
            "func": executor.to_postman,
            "params": [
                {
                    "name": "benchmark",
                    "help": "基准",
                    "default": "clickhouse",
                },
                {
                    "name": "dataset",
                    "help": "数据集",
                    "default": "ts-ts-preserve-service-cpu-exhaustion-mvlf65",
                },
            ],
        },
    }

    select_and_execute(config)
