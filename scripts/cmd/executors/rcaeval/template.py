from typing import List, Dict
from RCAEval.e2e import nsigma
from pprint import pprint
import json
import numpy as np
import os
import pandas as pd


def clean_data(ori_df: pd.DataFrame) -> pd.DataFrame:
    # handle inf
    df = ori_df.replace([np.inf, -np.inf], np.nan)

    # handle na
    df = df.fillna(method="ffill")
    df = df.fillna(0)

    return df


def process_metric_csv(file_path: str) -> pd.DataFrame:
    """
    从 metrics.csv 文件中读取和处理指标数据。

    参数:
    - file_path (str): metrics.csv 文件路径。

    返回:
    - pd.DataFrame: 以时间戳和 Pod 作为索引，指标作为列名的数据框。
    """
    used_columns = ["TimeUnix", "MetricName", "Value", "ResourceAttributes"]
    selected_columns = ["TimeStamp", "MetricName", "Value", "PodName"]
    chunk_size = 10**6
    metric_columns = None
    selected_data = []

    for chunk in pd.read_csv(file_path, chunksize=chunk_size, usecols=used_columns):
        print(f"Initial chunk size: {chunk.shape}")

        if not chunk.empty:
            chunk = chunk.copy()
            chunk["TimeStamp"] = pd.to_datetime(
                chunk["TimeUnix"], unit="ns"
            ).dt.tz_localize(os.environ["TIMEZONE"])
            chunk["PodName"] = chunk["ResourceAttributes"].apply(
                lambda x: json.loads(x).get("k8s.deployment.name", None)
            )

            chunk = chunk[selected_columns].pivot_table(
                index="TimeStamp",
                columns=["PodName", "MetricName"],
                values="Value",
                aggfunc="mean",
            )
            chunk.columns = ["_".join(col).strip() for col in chunk.columns.values]
            chunk = chunk.reset_index()

            metric_columns = chunk.columns.tolist()
            selected_data.append(chunk)

    if selected_data:
        df = pd.concat(selected_data, ignore_index=True)

        return clean_data(df)
    else:
        return pd.DataFrame(columns=metric_columns)


def diagnose_faults(
    fault_list: List[Dict], metric_data: pd.DataFrame, top_n: int = 5
) -> pd.DataFrame:
    """
    诊断故障并返回每个时间戳的 Top N 服务。

    参数:
        fault_list (List[Dict]): 故障事件列表，每个事件包含 'inject_time'。
        metric_data (pd.DataFrame): 指标数据，包含 'TimeStamp', 'PodName' 和其他指标列。
        top_n (int, optional): 每个时间戳选择的 Top N 服务。默认为5。

    返回:
        pd.DataFrame: 包含 'timestamp' 和 'pod' 的扩展数据框。
    """
    results = []

    metric_data["TimeStamp"] = pd.to_numeric(metric_data["TimeStamp"], errors="coerce")
    metric_data.rename(columns={"TimeStamp": "time"}, inplace=True)

    for event in fault_list:
        inject_time = pd.to_datetime(event["inject_time"], unit="s").value

        out = nsigma(metric_data, inject_time, dataset="train-ticket", anomalies=None)

        unique_set = set()
        top_services = []
        for item in out["ranks"]:
            service = item.split("_")[0]
            if service not in unique_set:
                unique_set.add(service)
                top_services.append(service)

        top_services = top_services[:top_n]
        results.append({"timestamp": inject_time, "top_services": top_services})

    top_services_df = pd.DataFrame(results)

    # 扩展每个时间戳的 Top 服务为单独的行
    expanded_rows = []
    for _, row in top_services_df.iterrows():
        services = row["top_services"]
        for idx, service in enumerate(services):
            expanded_rows.append(
                {
                    "level": "service",
                    "result": service,
                    "rank": idx + 1,
                    "confidence": 0,
                }
            )

    return pd.DataFrame(expanded_rows)


# IMPORTANT: do not change the function signature!!
def start_rca(params: Dict):
    pprint(params)
    directory = "/app/output"
    if not os.path.exists(directory):
        os.makedirs(directory)

    metric_file = params["metric_file"]
    inject_time_range = [normal_time[-1] for normal_time in params["normal_time_range"]]

    metric_data = process_metric_csv(metric_file)

    fault_list = []
    for i in range(len(inject_time_range)):
        fault_list.append(
            {
                "inject_time": inject_time_range[i],
            }
        )

    result = diagnose_faults(fault_list, metric_data)
    if not result.empty:
        result.to_csv("./output/result.csv", index=False)
