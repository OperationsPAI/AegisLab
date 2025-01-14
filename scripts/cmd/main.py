from typing import Dict
from rich.console import Console
from rich.prompt import Prompt
from utils import Executor
import argparse
import sys


def main(config: Dict):
    keys = list(config.keys())
    if "func" not in keys:
        selected_func_config = Prompt.ask(
            "选择要执行的函数",
            choices=keys,
            default=keys[0],
            show_choices=True,
        )

        main(config[selected_func_config])
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

    parser = argparse.ArgumentParser(description="批量导入算法程序")
    parser.add_argument("-d", "--default", action="store_true", help="采用默认配置")

    args = parser.parse_args()

    if args.default:
        executor = Executor(console, "rcaeval", Executor.config_dict["rcaeval"])
        executor.create(mode="cpu")
        sys.exit()

    algo_libraries = list(Executor.config_dict.keys())
    while True:
        algo_library = Prompt.ask(
            "选择算法库",
            choices=algo_libraries,
            default=algo_libraries[0],
            show_choices=True,
        )
        if algo_library in algo_libraries:
            break

        console.print(f"输入算法库 [bold red]{algo_library}[/bold red] 不存在")

    src_dir = Prompt.ask(
        "请输入算法库源位置", default=Executor.config_dict[algo_library]
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

    main(config)
