from typing import Dict
from pprint import pprint
import json
import numpy as np
import os
import pandas as pd


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
            chunk = (
                chunk[selected_columns]
                .pivot_table(
                    index=["TimeStamp", "PodName"], columns="MetricName", values="Value"
                )
                .reset_index()
            )
            metric_columns = chunk.columns.tolist()
            selected_data.append(chunk)

    if selected_data:
        return pd.concat(selected_data, ignore_index=True)
    else:
        return pd.DataFrame(columns=metric_columns)


def calculate_rho_squared(
    df_normal: pd.DataFrame, df_abnormal: pd.DataFrame
) -> Dict[str, float]:
    """
    计算每个指标的 rho_squared 值。

    参数:
        df_normal (pd.DataFrame): 正常数据范围的数据框。
        df_abnormal (pd.DataFrame): 异常数据范围的数据框。

    返回:
        Dict[str, float]: 每个指标的 rho_squared 值。
    """
    results = {}
    for column in df_normal.columns:
        if column in df_abnormal.columns and column not in ["TimeStamp", "PodName"]:
            normal_data = df_normal[column].dropna()
            abnormal_data = df_abnormal[column].dropna()

            # 截断法对齐数据
            min_length = min(len(normal_data), len(abnormal_data))

            if min_length < 2:
                continue

            normal_data = normal_data[:min_length]
            abnormal_data = abnormal_data[:min_length]

            try:
                cov_matrix = np.cov(normal_data, abnormal_data)
                cov = cov_matrix[0, 1]
                var_normal = normal_data.var()
                var_abnormal = abnormal_data.var()

                if var_normal > 0 and var_abnormal > 0:
                    rho_squared = (cov**2) / (var_normal * var_abnormal)
                    results[column] = rho_squared

            except Exception:
                continue

    return results


def diagnose_faults(
    fault: Dict, metric_data: pd.DataFrame, top_n: int = 5
) -> pd.DataFrame:
    """
    诊断故障并返回每个时间戳的 Top N 服务。

    参数:
        fault (Dict): 故障事件。
        metric_data (pd.DataFrame): 指标数据，包含 'TimeStamp', 'PodName' 和其他指标列。
        top_n (int, optional): 每个时间戳选择的 Top N 服务。默认为5。

    返回:
        pd.DataFrame: 包含 'timestamp' 和 'pod' 的扩展数据框。
    """
    all_services = {}

    metric_data["TimeStamp"] = pd.to_numeric(metric_data["TimeStamp"], errors="coerce")
    service_names = metric_data["PodName"].unique()

    normal_start = pd.to_datetime(
        pd.to_numeric(fault["normal_range"][0], errors="coerce"), unit="s"
    ).value
    normal_end = pd.to_datetime(
        pd.to_numeric(fault["normal_range"][1], errors="coerce"), unit="s"
    ).value
    abnormal_start = pd.to_datetime(
        pd.to_numeric(fault["abnormal_range"][0], errors="coerce"), unit="s"
    ).value
    abnormal_end = pd.to_datetime(
        pd.to_numeric(fault["abnormal_range"][1], errors="coerce"), unit="s"
    ).value

    for service in service_names:
        service_df = metric_data[metric_data["PodName"] == service].copy()
        service_df = service_df.dropna(subset=["TimeStamp"]).sort_values("TimeStamp")

        normal_range = service_df[
            (service_df["TimeStamp"] >= normal_start)
            & (service_df["TimeStamp"] < normal_end)
        ]
        abnormal_range = service_df[
            (service_df["TimeStamp"] >= abnormal_start)
            & (service_df["TimeStamp"] <= abnormal_end)
        ]

        metric_scores = calculate_rho_squared(normal_range, abnormal_range)
        if metric_scores:
            max_score = max(metric_scores.values())
            all_services[service] = max_score

    # 为每个时间戳选择 Top N 服务
    sorted_services = sorted(all_services.items(), key=lambda x: x[1], reverse=True)
    top_services = [service for service, _ in sorted_services[:top_n]]

    expanded_rows = []
    # 扩展每个时间戳的 Top 服务为单独的行
    for idx, service in enumerate(top_services):
        expanded_rows.append(
            {
                "level": "service",
                "result": service,
                "rank": idx + 1,
                "confidence": 0,
            }
        )

    expanded_df = pd.DataFrame(expanded_rows)
    return expanded_df


# IMPORTANT: do not change the function signature!!
def start_rca(params: Dict):
    pprint(params)

    normal_metric_data = process_metric_csv(params["normal_metric_file"])
    abnormal_metric_data = process_metric_csv(params["abnormal_metric_file"])
    metric_data = pd.concat(
        [normal_metric_data, abnormal_metric_data], ignore_index=True
    )

    fault = {
        "normal_range": [os.environ["NORMAL_START"], os.environ["NORMAL_END"]],
        "abnormal_range": [os.environ["ABNORMAL_START"], os.environ["ABNORMAL_END"]],
    }
    result = diagnose_faults(fault, metric_data)
    if not result.empty:
        result.to_csv(
            os.path.join(os.environ["OUTPUT_PATH"], "result.csv"), index=False
        )
