from typing import Dict, List, Union
import argparse
import ast
import os
import shutil
import sys


FILE_PATH = "/home/nn/workspace/lib/RCAEval/RCAEval/e2e/__init__.py"
PWD = os.getcwd()
TEMPLATE_DOCKERFILE = os.path.join(PWD, "scripts/rcaeval/resources/template.Dockerfile")
TEMPLATE_PY = os.path.join(PWD, "scripts/rcaeval/resources/template.py")

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
            temp = node.body if flag else node.orelse

            for child in temp:
                if isinstance(child, ast.ImportFrom):
                    handle_import_from(child)

    return results


class ReplaceAlgo(ast.NodeTransformer):
    def __init__(self, new_algo: str, old_algo: str) -> None:
        super().__init__()

        self.new_algo = new_algo
        self.old_algo = old_algo

    def visit_ImportFrom(self, node):
        if node.module == "RCAEval.e2e":
            node.names = [
                ast.alias(name=self.new_algo, asname=None)
                if alias.name == self.old_algo
                else alias
                for alias in node.names
            ]

        return node

    def visit_Name(self, node):
        if node.id == self.old_algo:
            node.id = self.new_algo

        return node


def replace_algo(content: str, new_algo: str, old_algo: str = "nsigma"):
    tree = ast.parse(content)

    transformer = ReplaceAlgo(new_algo, old_algo)
    modified_tree = transformer.visit(tree)
    ast.fix_missing_locations(modified_tree)

    modified_code = ast.unparse(modified_tree)

    return modified_code


def gen_code(results, mode: str, content: str) -> List[str]:
    import_algos = []

    for _, result in results.items():
        if mode != ("cpu" if result["torch"] else "gpu"):
            for algo in result["algos"]:
                algo_dir = os.path.join(PWD, f"algorithms/{algo}_rcaeval")
                if not os.path.exists(algo_dir):
                    os.makedirs(algo_dir)

                with open(os.path.join(algo_dir, PY_NAME), mode="w") as file:
                    file.write(replace_algo(content, algo))

                shutil.copy(
                    TEMPLATE_DOCKERFILE, os.path.join(algo_dir, DOCKERFILE_NAME)
                )

                import_algos.append(algo)

    return import_algos


def delete_algos(algos: List[str]) -> None:
    for algo in algos:
        try:
            shutil.rmtree(os.path.join(PWD, f"algorithms/{algo}_rcaeval"))
        except Exception as e:
            print(f"删除算法{algo}错误: {e}")


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="批量导入RCAEval算法")

    parser.add_argument(
        "-f", "--file", type=str, help="__init__.py 文件路径", default=FILE_PATH
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

    file_path, mode, is_py310 = args.file, args.mode, args.is_py310
    results = extract_relative_imports(file_path, is_py310)

    with open(TEMPLATE_PY, mode="r") as file:
        content = file.read()

    import_algos = gen_code(results, mode, content)
    print(f"导入算法数目: {len(import_algos)}")
    print(f"导入算法名称: {import_algos}")
