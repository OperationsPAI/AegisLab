from typing import Dict, List, Optional
from types import ModuleType
from rich.console import Console
from rich.prompt import Prompt
from . import ALGORITHMS_DIR, EXECUTORS_DIR
from .console import select_choice, select_function
from .harbor import Executor as HarborExecutor
import importlib.util
import json
import os
import subprocess
import shutil
import sys
import toml

__all__ = ["run_algo"]


MODES = ["cpu", "gpu", "both"]
TEMPLATE_DOCKERFILE = "template.Dockerfile"
TEMPLATE_PY = "template.py"
DOCKERFILE_NAME = "builder.Dockerfile"
PY_NAME = "rca.py"
IMPORTED_FILE = "imported_algos.txt"


def _read_algos(file_path: str) -> List[str]:
    algos = []
    with open(file_path, mode="r") as file:
        for line in file.readlines():
            algos.append(line.replace("\n", ""))

    return algos


def _write_algos(file_path: str, algos: List[str]) -> None:
    with open(file_path, mode="w") as file:
        for algo in algos:
            file.write(f"{algo}\n")


def _load_module_from_file(
    console: Console, module_path: str, module_name: str
) -> Optional[ModuleType]:
    """
    根据文件路径动态加载模块。

    参数:
    - module_name (str): 模块的名称
    - module_path (str): 模块文件的路径。

    返回:
        Optional(ModuleType) 加载的模块对象
    """
    try:
        spec = importlib.util.spec_from_file_location(module_name, module_path)
        module = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(module)
        sys.modules[module_name] = module

        console.print(f"成功加载模块: {module_name}")
        return module
    except Exception as e:
        console.print(f"加载模块时出错: {e}")
        return None


class Executor:
    FUNCTIONS = ["extract_algos", "replace_algo_dockerfile", "replace_algo_py"]

    def __init__(
        self,
        console: Console,
        algo_library: str,
        module_name: str,
        toml_name: str,
        algo_config: Dict,
        harbor_config: Dict,
    ):
        self.console = console
        self.algo_library = algo_library
        self.toml_name = toml_name
        self.algo_config = algo_config
        self.harbor_config = harbor_config

        self.executor_dir = os.path.join(EXECUTORS_DIR, algo_library)
        self.module_path = os.path.join(self.executor_dir, module_name)
        self.imported_file = os.path.join(self.executor_dir, IMPORTED_FILE)

        self.extract_algos = None
        self.replace_algo_dockerfile = None
        self.replace_algo_py = None

        self._import_functions()

    def _import_functions(self):
        module = _load_module_from_file(
            self.console, self.module_path, self.algo_library
        )
        if module:
            for attr_name in Executor.FUNCTIONS:
                if hasattr(module, attr_name):
                    setattr(self, attr_name, getattr(module, attr_name))
                else:
                    self.console.print(f"模块中没有找到属性: {attr_name}")

    def create(self, mode: str) -> List[str]:
        """
        生成算法代码

        参数:
        - mode (str): 导入算法的资源类别。

        返回:
            List[str]: 导入的算法列表。
        """
        assert mode in MODES
        imported_algos = self.extract_algos(self.algo_config["path"])

        with open(os.path.join(self.executor_dir, TEMPLATE_DOCKERFILE), mode="r") as f:
            dockerfile_content = f.read()

        with open(os.path.join(self.executor_dir, TEMPLATE_PY), mode="r") as f:
            py_content = f.read()

        results = []
        for _, item in imported_algos.items():
            if mode != ("cpu" if item["torch"] else "gpu"):
                for algo in item["algos"]:
                    algo_dir = os.path.join(
                        ALGORITHMS_DIR, f"{algo}_{self.algo_library}"
                    )
                    if not os.path.exists(algo_dir):
                        os.makedirs(algo_dir)

                    with open(os.path.join(algo_dir, DOCKERFILE_NAME), mode="w") as f:
                        f.write(
                            self.replace_algo_dockerfile(
                                dockerfile_content,
                                self.algo_config["algo"],
                                f"{algo}_{self.algo_library}",
                            )
                        )

                    with open(os.path.join(algo_dir, PY_NAME), mode="w") as f:
                        f.write(
                            self.replace_algo_py(
                                py_content, self.algo_config["algo"], algo
                            )
                        )

                    info_data = {"name": algo}
                    with open(os.path.join(algo_dir, self.toml_name), mode="w") as f:
                        toml.dump(info_data, f)

                    results.append(algo)

        _write_algos(self.imported_file, results)

        command = ["ruff", "format", ALGORITHMS_DIR]
        result = subprocess.run(command, capture_output=True, text=True)
        if result.returncode == 0:
            self.console.print("格式化成功!")
        else:
            self.console.print(f"格式化失败: {result.stderr}")

        image_configs = [
            {"name": f"{algo}_{self.algo_library}", "tag": "latest"} for algo in results
        ]
        harbor_executor = HarborExecutor(self.console, self.harbor_config)
        harbor_executor.push_images("algorithm", image_configs[:1])

    def delete(self) -> None:
        """删除算法代码"""
        try:
            for algo in _read_algos(self.imported_file):
                try:
                    shutil.rmtree(
                        os.path.join(ALGORITHMS_DIR, f"{algo}_{self.algo_library}")
                    )
                    self.console.print(f"算法 [bold green]{algo}[/bold green] 删除成功")
                except Exception as e:
                    self.console.print(
                        f"删除算法 [bold red]{algo}[/bold red] 错误: {e}"
                    )

            os.remove(self.imported_file)
        except FileNotFoundError:
            self.console.print("请先使用函数 [bold red]Create[/bold red] 导入算法")

    def to_postman(self, benchmark: str, dataset: str) -> None:
        content = []
        try:
            for algo in _read_algos(self.imported_file):
                content.append(
                    {
                        "algorithm": f"{algo}_{self.algo_library}",
                        "benchmark": benchmark,
                        "dataset": dataset,
                    }
                )

            with open(
                os.path.join(self.executor_dir, "postman.json"), mode="w"
            ) as file:
                json.dump(content, file, indent=4)
            self.console.print("数据已成功写入 [bold green]postman.json[/bold green]")
        except FileNotFoundError:
            self.console.print("请先使用函数 [bold red]Create[/bold red] 导入算法")


def run_algo(console: Console, args, algo_config: Dict, harbor_config: Dict):
    @select_function
    def main(config: Dict):
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

    module_name = algo_config["module_name"]
    toml_name = algo_config["toml_name"]

    if args.default1:
        algo_library = "rcaeval"
        executor = Executor(
            console, algo_library, toml_name, algo_config["library"][algo_library]
        )
        executor.create(mode="cpu")
        sys.exit()

    # 获取导入的算法库
    algo_library = select_choice(console, "算法库", list(algo_config["library"].keys()))

    executor = Executor(
        console,
        algo_library,
        module_name,
        toml_name,
        algo_config["library"][algo_library],
        harbor_config,
    )

    runner_config = {
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

    main(runner_config)
