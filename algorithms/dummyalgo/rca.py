from typing import Dict
from pprint import pprint
import os
import time

def print_directory_tree(start_path, prefix=""):
    # 获取当前目录下的所有文件和子目录
    items = os.listdir(start_path)
    items.sort()  # 可选：按字母排序
    for i, item in enumerate(items):
        # 确定是否是最后一个元素，用于绘制正确的符号
        is_last = i == len(items) - 1
        # 拼接路径
        path = os.path.join(start_path, item)
        # 根据是否是最后一个元素选择符号
        connector = "└── " if is_last else "├── "
        print(f"{prefix}{connector}{item}")
        # 如果是目录，递归调用
        if os.path.isdir(path):
            new_prefix = prefix + ("    " if is_last else "│   ")
            print_directory_tree(path, new_prefix)


# IMPORTANT: do not change the function signature!!
def start_rca(params: Dict):
    pprint(params)
    directory = "/app/output"
    print_directory_tree("./")

    if not os.path.exists(directory):
        os.makedirs(directory)

    with open("/app/input/logs.csv") as f:
        data = f.readlines()
        print(data[:3])

    file_path = os.path.join(directory, "my_file.txt")

    with open(file_path, "w") as file:
        file.write("hello world")
