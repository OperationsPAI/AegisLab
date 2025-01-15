from typing import Dict, List, Union
import ast
import io
import os
import sys
import tokenize

__all__ = ["extract_algos", "replace_algo_dockerfile", "replace_algo_py"]


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


def extract_algos(
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


def replace_algo_dockerfile(content: str, old_algo: str, new_algo: str) -> str:
    return content.replace(old_algo, new_algo)


def replace_algo_py(content: str, old_algo: str, new_algo: str) -> str:
    """
    替换模版中的算法

    参数:
    - content  (str): python 模版文件内容。
    - old_algo (str): 替换前算法
    - new_algo (str): 替换后算法

    返回:
        str: 替换算法后的文件内容
    """
    tokens = tokenize.generate_tokens(io.StringIO(content).readline)

    modified_tokens = []
    for token_type, token_string, _, _, _ in tokens:
        if token_string == old_algo:
            modified_tokens.append((token_type, new_algo))
        else:
            modified_tokens.append((token_type, token_string))

    modified_code = tokenize.untokenize(modified_tokens)

    return modified_code
