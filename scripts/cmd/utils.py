from typing import Dict, List, Union
from rich.console import Console
import ast
import io
import json
import os
import subprocess
import shutil
import sys
import tokenize
import toml

__all__ = ["Executor"]

DIR = os.path.dirname(os.path.abspath(__file__))
ALGORITHMS_DIR = os.path.join(os.getcwd(), "algorithms")

IMPORTED_FILE = "imported_algos.txt"

TEMPLATE_DOCKERFILE = "template.Dockerfile"
TEMPLATE_PY = "template.py"
TEMPLATE_ALGO_DICT = {"rcaeval": "nsigma"}

DOCKERFILE_NAME = "builder.Dockerfile"
PY_NAME = "rca.py"
TOML_NAME = "info.toml"


def _check_py310() -> bool:
    """
    检查 python 版本是否为 3.10.*
    """
    return sys.version_info.major == 3 and sys.version_info.minor == 10


def _check_torch_import(content: str) -> bool:
    """
    检查是否导入 pytorch
    """
    tree = ast.parse(content)

    for node in ast.walk(tree):
        if isinstance(node, ast.Import):
            for alias in node.names:
                if alias.name == "torch":
                    return True
        elif isinstance(node, ast.ImportFrom):
            if node.module == "torch":
                return True

    return False


def _extract_relative_imports(
    dir: str, file_name: str = "__init__.py"
) -> Dict[str, Union[bool, List[str]]]:
    """
    根据筛选条件获取所有算法

    参数:
    - dir (str): 算法包路径。
    """
    with open(os.path.join(dir, file_name), "r") as file:
        content = file.read()

    tree = ast.parse(content)
    results = {}

    def handle_import_from(node):
        if node.level > 0:
            with open(os.path.join(dir, f"{node.module}.py")) as file:
                content = file.read()

            is_torch_import = _check_torch_import(content)

            for alias in node.names:
                if node.module not in results:
                    results[node.module] = {
                        "torch": is_torch_import,
                        "algos": [alias.name],
                    }
                else:
                    results[node.module]["algos"].append(alias.name)

    # 遍历AST节点
    for node in ast.walk(tree):
        if isinstance(node, ast.If):
            stmts = node.body if _check_py310() else node.orelse

            for stmt in stmts:
                if isinstance(stmt, ast.ImportFrom):
                    handle_import_from(stmt)

                if isinstance(stmt, ast.Try):
                    for child in stmt.body:
                        if isinstance(child, ast.ImportFrom):
                            handle_import_from(child)

    return results


def _replace_algo(content: str, new_algo: str, old_algo: str):
    tokens = tokenize.generate_tokens(io.StringIO(content).readline)

    modified_tokens = []
    for token_type, token_string, _, _, _ in tokens:
        if token_string == old_algo:
            modified_tokens.append((token_type, new_algo))
        else:
            modified_tokens.append((token_type, token_string))

    modified_code = tokenize.untokenize(modified_tokens)

    return modified_code


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


class Executor:
    def __init__(self, console: Console, algo_library: str, src_dir: str):
        self.console = console
        self.algo_library = algo_library
        self.src_dir = src_dir

        self.template_dir = os.path.join(DIR, "resources", algo_library)
        self.imported_file = os.path.join(self.template_dir, IMPORTED_FILE)

    def create(self, mode: str) -> List[str]:
        """
        生成算法代码

        参数:
        - template_dir (str): python 模版文件路径。

        返回:
            List[str]: 导入的算法列表。
        """
        imported_algos = _extract_relative_imports(self.src_dir)

        with open(os.path.join(self.template_dir, TEMPLATE_PY), mode="r") as f:
            content = f.read()

        results = []
        for _, item in imported_algos.items():
            if mode != ("cpu" if item["torch"] else "gpu"):
                for algo in item["algos"]:
                    algo_dir = os.path.join(
                        ALGORITHMS_DIR, f"{algo}_{self.algo_library}"
                    )
                    if not os.path.exists(algo_dir):
                        os.makedirs(algo_dir)

                    with open(os.path.join(algo_dir, PY_NAME), mode="w") as file:
                        file.write(
                            _replace_algo(
                                content, algo, TEMPLATE_ALGO_DICT[self.algo_library]
                            )
                        )

                    info_data = {"name": algo}
                    with open(os.path.join(algo_dir, TOML_NAME), mode="w") as file:
                        toml.dump(info_data, file)

                    shutil.copy(
                        os.path.join(self.template_dir, TEMPLATE_DOCKERFILE),
                        os.path.join(algo_dir, DOCKERFILE_NAME),
                    )

                    results.append(algo)

        _write_algos(self.imported_file, results)

        command = ["ruff", "format", ALGORITHMS_DIR]
        result = subprocess.run(command, capture_output=True, text=True)
        if result.returncode == 0:
            print("格式化成功!")
        else:
            print(f"格式化失败: {result.stderr}")

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

    def to_postman(
        self,
        benchmark: str = "clickhouse",
        dataset: str = "ts-ts-preserve-service-cpu-exhaustion-mvlf65",
    ) -> None:
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

            with open(os.path.join(self.template_dir, "body.json"), mode="w") as file:
                json.dump(content, file, indent=4)
            self.console.print("数据已成功写入 [bold green]body.json[/bold green]")
        except FileNotFoundError:
            self.console.print("请先使用函数 [bold red]Create[/bold red] 导入算法")
