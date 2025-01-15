from rich.console import Console
from handlers.algo import run_algo
from handlers.harbor import run_harbor
import argparse
import json
import os

CONFIG_PATH = os.path.join(os.path.dirname(os.path.abspath(__file__)), "config.json")


if __name__ == "__main__":
    console = Console()

    with open(CONFIG_PATH, mode="r") as file:
        config = json.load(file)

    parser = argparse.ArgumentParser(description="命令行")
    mutex_group = parser.add_mutually_exclusive_group()

    algo_group = parser.add_argument_group("批量导入算法", description="导入算法配置")
    algo_group.add_argument(
        "-d1", "--default1", action="store_true", help="采用默认配置"
    )

    harbor_group = parser.add_argument_group("批量上传镜像", description="上传镜像配置")
    harbor_group.add_argument(
        "-d2", "--default2", action="store_true", help="采用默认配置"
    )

    mutex_group.add_argument("--algo", action="store_true", help="批量导入算法")
    mutex_group.add_argument("--harbor", action="store_true", help="批量上传镜像")

    args = parser.parse_args()

    if args.algo:
        run_algo(console, args, config["algo"], config["harbor"])
    elif args.harbor:
        run_harbor(console, args, config["harbor"])
