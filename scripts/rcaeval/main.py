from typing import Dict, List, Union
import argparse
import ast
import io
import os
import shutil
import subprocess
import sys
import tokenize

FILE_PATH = "/home/nn/workspace/lib/RCAEval/RCAEval/e2e/__init__.py"

DIR = os.path.dirname(os.path.abspath(__file__))
ALGORITHMS_DIR = os.path.join(os.getcwd(), "algorithms")

TEMPLATE_DOCKERFILE = os.path.join(DIR, "resources/template.Dockerfile")
TEMPLATE_PY = os.path.join(DIR, "resources/template.py")

PY_NAME = "rca.py"
DOCKERFILE_NAME = "builder.Dockerfile"


def check_py310() -> bool:
    """
    检查 python 版本是否为 3.10.*
    """
    return sys.version_info.major == 3 and sys.version_info.minor == 10


def check_torch_import(content: str) -> bool:
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


def extract_relative_imports(
    init_file: str, flag: bool
) -> Dict[str, Union[bool, List[str]]]:
    """
    根据筛选条件获取所有算法

    参数:
    - init_file (str): RCAEval/RCAEval/e2e/__init__.py 文件路径。
    - flag (bool): 判断 python 版本是否为 3.10.* 。
    """
    with open(init_file, "r") as file:
        content = file.read()

    tree = ast.parse(content)
    results = {}

    def handle_import_from(node):
        if node.level > 0:
            with open(
                os.path.join(os.path.dirname(init_file), f"{node.module}.py")
            ) as file:
                content = file.read()

            is_torch_import = check_torch_import(content)

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
            stmts = node.body if flag else node.orelse

            for stmt in stmts:
                if isinstance(stmt, ast.ImportFrom):
                    handle_import_from(stmt)

                if isinstance(stmt, ast.Try):
                    for child in stmt.body:
                        if isinstance(child, ast.ImportFrom):
                            handle_import_from(child)

    return results


def replace_algo(content: str, new_algo: str, old_algo: str = "nsigma"):
    tokens = tokenize.generate_tokens(io.StringIO(content).readline)

    modified_tokens = []
    for token_type, token_string, _, _, _ in tokens:
        if token_string == old_algo:
            modified_tokens.append((token_type, new_algo))
        else:
            modified_tokens.append((token_type, token_string))

    modified_code = tokenize.untokenize(modified_tokens)

    return modified_code


def format_code(library: str) -> None:
    """
    格式化生成的代码

    参数:
    - library (str): 格式化库。
    """
    if library == "ruff":
        command = ["ruff", "format", ALGORITHMS_DIR]

    result = subprocess.run(command, capture_output=True, text=True)

    if result.returncode == 0:
        print("格式化成功!")
    else:
        print(f"格式化失败: {result.stderr}")


def gen_code(
    results: Dict[str, Union[bool, List[str]]], library: str, mode: str, content: str
) -> List[str]:
    """
    生成算法代码

    参数:
    - results (Dict[str, Union[bool, List[str]]]): 解析导入结果。
    - library (str): 格式化库。
    - mode (str): 算法调用资源类别。
    - content (str): python 模版文件内容。

    返回:
        List[str]: 导入的算法列表。
    """
    import_algos = []

    for _, result in results.items():
        if mode != ("cpu" if result["torch"] else "gpu"):
            for algo in result["algos"]:
                algo_dir = os.path.join(ALGORITHMS_DIR, f"{algo}_rcaeval")
                if not os.path.exists(algo_dir):
                    os.makedirs(algo_dir)

                with open(os.path.join(algo_dir, PY_NAME), mode="w") as file:
                    file.write(replace_algo(content, algo))

                shutil.copy(
                    TEMPLATE_DOCKERFILE, os.path.join(algo_dir, DOCKERFILE_NAME)
                )

                import_algos.append(algo)

    format_code(library)

    return import_algos


def delete_algos(algos: List[str]) -> None:
    for algo in algos:
        try:
            shutil.rmtree(os.path.join(ALGORITHMS_DIR, f"{algo}_rcaeval"))
        except Exception as e:
            print(f"删除算法{algo}错误: {e}")


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="批量导入RCAEval算法")

    parser.add_argument(
        "-f", "--file", type=str, help="__init__.py 文件路径", default=FILE_PATH
    )
    parser.add_argument(
        "-l", "--format_library", type=str, help="格式化库", default="ruff"
    )
    parser.add_argument(
        "-m",
        "--mode",
        type=str,
        choices=["cpu", "gpu", "both"],
        help="算法类别",
        default="cpu",
    )
    parser.add_argument(
        "--is_py310",
        action="store_true",
        help=" python 版本是否为 3.10.* ",
        default=check_py310(),
    )

    args = parser.parse_args()

    file_path, library, mode, is_py310 = (
        args.file,
        args.format_library,
        args.mode,
        args.is_py310,
    )
    results = extract_relative_imports(file_path, is_py310)

    with open(TEMPLATE_PY, mode="r") as file:
        content = file.read()

    imported_algos = gen_code(results, library, mode, content)
    print(f"导入算法数目: {len(imported_algos)}")
    print(f"导入算法名称: {imported_algos}")
