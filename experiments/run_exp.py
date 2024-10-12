# This file will be moved into image workspace, and start the evaluation by running the script

from rca import start_rca
import os
from pathlib import Path


if __name__ == "__main__":
    workspace = os.environ['WORKSPACE']
    if workspace == "":
        print("ERROR: the WORKSPACE environ is not defined.")
    base_path = Path(workspace)
    start_rca(base_path/"input"/"logs.csv", 
              base_path/"input"/"traces.csv", 
              base_path/"input"/"metrics.csv", 
              base_path/"input"/"events.csv", 
              base_path/"input"/"profilings.csv", 
              [(int(os.environ['NORMAL_START']), int(os.environ['NORMAL_END']))] if os.environ['NORMAL_START']!='' and os.environ['NORMAL_END']!='' else [], 
              [(int(os.environ['ABNORMAL_START']), int(os.environ['ABNORMAL_END']))]  if os.environ['ABNORMAL_START']!='' and os.environ['ABNORMAL_END']!='' else []
              )