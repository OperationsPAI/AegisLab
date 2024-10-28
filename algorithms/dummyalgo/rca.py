from typing import Dict
from pprint import pprint
import os
# IMPORTANT: do not change the function signature!!
def start_rca(params: Dict):
    pprint(params)
    directory = '/app/output'

    if not os.path.exists(directory):
        os.makedirs(directory)

    file_path = os.path.join(directory, 'my_file.txt')

    with open(file_path, 'w') as file:
        file.write('hello world')