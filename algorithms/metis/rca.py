import ast
import asyncio
import json
import os
import shutil
from functools import partial
from pathlib import Path
from typing import Dict, List, Tuple

import pandas as pd
from aiomultiprocess import Pool
from evaluation import evaluate


def extract_resource_attribute(attributes: str, key: str):
    """Extract a specific key's value from the ResourceAttributes string."""
    try:
        attributes = json.loads(attributes)
        return attributes.get(key, pd.NA)
    except:
        return pd.NA


async def parselog(df: pd.DataFrame, start_time, end_time):
    df = df[(df["Timestamp"] >= start_time) & (df["Timestamp"] <= end_time)]
    return df[
        [
            "Timestamp",
            "TraceId",
            "SpanId",
            "SeverityText",
            "SeverityNumber",
            "ServiceName",
            "Body",
        ]
    ]


def extract_metric(df):
    ResourceAttributes = df.ResourceAttributes
    Attributes = df.Attributes
    r = json.loads(ResourceAttributes.replace("'", '"'))
    a = json.loads(Attributes.replace("'", '"'))
    return dict(
        k8s_namespace_name=r.get("k8s.namespace.name", pd.NA),
        k8s_pod_uid=r.get("k8s.pod.uid", pd.NA),
        k8s_pod_name=r.get("k8s.pod.name", pd.NA),
        k8s_container_name=r.get("k8s.container.name", pd.NA),
        direction=a.get("direction", pd.NA),
    )


async def parsemetric(df: pd.DataFrame, start_time, end_time):
    # Parse the ResourceAttributes and Attributes to extract needed values
    df[
        [
            "k8s_namespace_name",
            "k8s_pod_uid",
            "k8s_pod_name",
            "k8s_container_name",
            "direction",
        ]
    ] = df.apply(extract_metric, axis="columns", result_type="expand")
    df.dropna(subset=["k8s_namespace_name"], inplace=True)
    df = df[(df["TimeUnix"] >= start_time) & (df["TimeUnix"] <= end_time)]
    return df[
        [
            "k8s_namespace_name",
            "k8s_pod_uid",
            "k8s_pod_name",
            "k8s_container_name",
            "MetricName",
            "MetricDescription",
            "TimeUnix",
            "Value",
            "MetricUnit",
            "direction",
        ]
    ]


async def concat(results):
    return await asyncio.to_thread(pd.concat, results)


async def filter_csv_data(
    query_type, start_time, end_time, input_file_paths, pool: Pool
):
    """Filter CSV data for logs, metrics, and traces based on the query type."""
    batch_size = 10000
    if query_type == "log":
        log_df = pd.read_csv(input_file_paths["log_file"], parse_dates=["Timestamp"])
        batches = [
            log_df.iloc[start : start + batch_size].copy()
            for start in range(0, len(log_df), batch_size)
        ]
        results = await pool.starmap(
            parselog,
            zip(batches, [start_time] * len(batches), [end_time] * len(batches)),
        )
        filtered_df = await pool.apply(concat, (results,))
        return filtered_df

    elif query_type == "metric":
        df_gauge = pd.read_csv(
            input_file_paths["metric_file"], parse_dates=["TimeUnix"]
        )
        df_sum = pd.read_csv(
            input_file_paths["metric_sum_file"], parse_dates=["TimeUnix"]
        )

        gauge_batches = [
            df_gauge.iloc[start : start + batch_size].copy()
            for start in range(0, len(df_gauge), batch_size)
        ]
        gauge_results = await pool.starmap(
            parsemetric,
            zip(
                gauge_batches,
                [start_time] * len(gauge_batches),
                [end_time] * len(gauge_batches),
            ),
        )
        filtered_gauge = await pool.apply(concat, (gauge_results,))

        sum_batches = [
            df_sum.iloc[start : start + batch_size].copy()
            for start in range(0, len(df_sum), batch_size)
        ]
        sum_results = await pool.starmap(
            parsemetric,
            zip(
                sum_batches,
                [start_time] * len(sum_batches),
                [end_time] * len(sum_batches),
            ),
        )
        filtered_sum = await pool.apply(concat, (sum_results,))

        filtered_sum = filtered_sum[
            (
                filtered_sum["MetricName"].isin(
                    ["k8s.pod.network.io", "k8s.pod.network.errors"]
                )
            )
        ]

        # Combine the results
        filtered_df = pd.concat([filtered_gauge, filtered_sum])
        return filtered_df

    elif query_type == "trace":
        trace_df = pd.read_csv(
            input_file_paths["trace_file"], parse_dates=["Timestamp"]
        )
        # Apply filters
        filtered_df = trace_df[
            (trace_df["Timestamp"] >= start_time) & (trace_df["Timestamp"] <= end_time)
        ]
        parent_df = trace_df[["SpanId", "ServiceName"]].rename(
            columns={"SpanId": "ParentSpanId", "ServiceName": "ParentServiceName"}
        )
        filtered_df = filtered_df.merge(parent_df, on="ParentSpanId", how="left")[
            [
                "Timestamp",
                "TraceId",
                "SpanId",
                "ParentSpanId",
                "SpanName",
                "ServiceName",
                "Duration",
                "ParentServiceName",
            ]
        ]
        return filtered_df
    else:
        raise ValueError("Invalid query type")


async def collect_and_save_data(
    folder, start_time, end_time, data_type, input_file_paths, pool
):
    """Collect and save data in batches."""
    filepath = Path(folder) / f"{data_type}s.csv"
    filtered_df = await filter_csv_data(
        data_type, start_time, end_time, input_file_paths, pool
    )
    await asyncio.to_thread(filtered_df.to_csv, filepath, index=False, mode="w")


def create_folders():
    """Create normal and abnormal folders for storing data."""
    normal_folder = Path("data") / "normal"
    abnormal_folder = Path("data") / "abnormal"
    normal_folder.mkdir(parents=True, exist_ok=True)
    abnormal_folder.mkdir(parents=True, exist_ok=True)
    return normal_folder, abnormal_folder


def parse_time(unix_time: int, delta=None):
    """Parse the Unix timestamp to a human-readable format."""
    dt = pd.to_datetime(unix_time, utc=True, unit="s")
    if delta:
        dt += delta
    return dt.strftime("%Y-%m-%d %H:%M:%S")


async def process_case(normal_range, abnormal_range, input_file_paths, pool):
    """Process a single chaos event."""

    abnormal_start = parse_time(abnormal_range[0], pd.Timedelta(minutes=-5))
    abnormal_end = parse_time(abnormal_range[1])
    normal_end = parse_time(normal_range[1], pd.Timedelta(minutes=-5))
    normal_start = parse_time(normal_range[1], pd.Timedelta(minutes=-15))

    print(f"Processing normal range: {normal_start} - {normal_end}")
    print(f"Processing abnormal range: {abnormal_start} - {abnormal_end}")
    normal_folder, abnormal_folder = create_folders()
    tasks = [
        collect_and_save_data(
            folder, start_time, end_time, data_type, input_file_paths, pool
        )
        for folder, start_time, end_time in [
            (normal_folder, normal_start, normal_end),
            (abnormal_folder, abnormal_start, abnormal_end),
        ]
        for data_type in ["log", "metric", "trace"]
    ]
    await asyncio.gather(*tasks)


params = {
    "log_file": "/home/nn/workspace/rcabench/benchmarks/clickhouse/input/logs.csv",
    "trace_file": "/home/nn/workspace/rcabench/benchmarks/clickhouse/input/traces.csv",
    "trace_id_ts_file": "trace.csv",
    "metric_file": "/home/nn/workspace/rcabench/benchmarks/clickhouse/input/metrics.csv",
    "metric_sum_file": "/home/nn/workspace/rcabench/benchmarks/clickhouse/input/metric_sum.csv",
    "metric_summary_file": "metric.csv",
    "metric_histogram_file": "metrics_histogram.csv",
    "event_file": "event.csv",
    "profiling_file": "profile.csv",
    "normal_time_range": [(1728826808, 1728827108)],
    "abnormal_time_range": [(1728827108, 1728827408)],
    "output_file_path": "/app/output/result.csv",
}


async def start_rca(params: Dict):
    normal_time_range = params["normal_time_range"][0]
    abnormal_time_range = params["abnormal_time_range"][0]
    childconcurrency = 20
    processes = os.cpu_count()
    queuecount = processes // 4
    async with Pool(
        processes=processes, childconcurrency=childconcurrency, queuecount=queuecount
    ) as pool:
        await process_case(
            normal_range=normal_time_range,
            abnormal_range=abnormal_time_range,
            input_file_paths=params,
            pool=pool,
        )
        print("Data collection completed.")
        try:
            await asyncio.wait_for(evaluate("data", pool), timeout=30.0)
            print("Evaluation completed.")
        except asyncio.TimeoutError:
            raise asyncio.TimeoutError("Evaluation timed out.")
        except Exception as e:
            print(f"An error occurred: {e}")
            raise

    result = Path("data") / "final_ranking.csv"
    Path("/app/output").mkdir(exist_ok=1, parents=1)
    if result.exists():
        df = pd.read_csv(result)
        df.rename(
            columns={
                "CombinedScore": "confidence",
                "ServiceName": "result",
                "Index": "rank",
            },
            inplace=True,
        )
        df["level"] = "service"
        df = df[["level", "result", "rank", "confidence"]]
        df.to_csv("/app/output/result.csv", index=False)


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


if __name__ == "__main__":
    print_directory_tree("/app")
    asyncio.run(start_rca(params))
