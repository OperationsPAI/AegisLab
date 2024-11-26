from typing import Dict
from pprint import pprint
import os
import pandas as pd
import re

def clean_span_name(span_name):
    span_name = re.sub(r"/[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}", "/<id>", span_name)
    
    if "/verify/" in span_name:
        span_name = re.sub(r"(?<=/verify/)[a-zA-Z0-9]{6}(?=/|$)", "<code>", span_name)
    
    if "/foodservice/foods/" in span_name:
        span_name = re.sub(r"/\d{4}-\d{2}-\d{2}", "/<date>", span_name)
        
        locations = re.findall(r"/([a-zA-Z]+)/([a-zA-Z]+)/[a-zA-Z0-9]+$", span_name)
        if locations:
            span_name = span_name.replace(f"/{locations[0][0]}", "/<location>", 1)
            span_name = span_name.replace(f"/{locations[0][1]}", "/<location>", 1)
        
        span_name = re.sub(r"/[a-zA-Z0-9]+$", "/<train>", span_name)

    return span_name

def process_and_aggregate_csv(file_path: str, service_name: str, start_time: int, end_time: int):
    start_time = pd.to_datetime(start_time, unit='s')
    end_time = pd.to_datetime(end_time, unit='s')
    
    print(start_time, end_time)
    required_columns = ["Timestamp", "SpanAttributes", "Duration", "SpanName", "StatusCode"]
    chunk_size = 10**6
    selected_data = []

    for chunk in pd.read_csv(file_path, chunksize=chunk_size):
        filtered_chunk = chunk[chunk["ServiceName"] == service_name]
        if not filtered_chunk.empty:
            filtered_chunk["Timestamp"] = pd.to_datetime(filtered_chunk["Timestamp"])
            filtered_chunk = filtered_chunk[
                (filtered_chunk["Timestamp"] >= start_time) & 
                (filtered_chunk["Timestamp"] <= end_time)
            ]
            
            filtered_chunk["SpanName"] = filtered_chunk["SpanName"].apply(clean_span_name)
            
            filtered_chunk["Duration"] = filtered_chunk["Duration"] / 1_000_000
            
            selected_data.append(filtered_chunk[required_columns])

    if selected_data:
        data = pd.concat(selected_data, ignore_index=True)
    else:
        return pd.DataFrame(columns=["SpanName", "AvgDuration", "SuccRate", "P90", "P95", "P99"])

    # 聚合计算，包括平均时延、成功率及尾时延
    aggregation = data.groupby("SpanName").agg(
        AvgDuration=("Duration", "mean"),  # 平均时延
        SuccRate=("StatusCode", lambda x: (x == 'Unset').mean()),  # 成功率
        P90=("Duration", lambda x: x.quantile(0.90)),  # P90 时延
        P95=("Duration", lambda x: x.quantile(0.95)),  # P95 时延
        P99=("Duration", lambda x: x.quantile(0.99))   # P99 时延
    ).reset_index()

    print(aggregation)
    return aggregation


def compare_dataframes(normal_df, abnormal_df):
    """
    Compare two DataFrames and identify changes in performance metrics.

    Args:
        normal_df (pd.DataFrame): DataFrame with normal period metrics.
        abnormal_df (pd.DataFrame): DataFrame with abnormal period metrics.

    Returns:
        pd.DataFrame: DataFrame highlighting significant changes.
    """
    # 内连接两个 DataFrame，仅保留同时存在的 SpanName
    merged_df = pd.merge(
        normal_df, abnormal_df, on="SpanName", suffixes=("_normal", "_abnormal")
    )
    
    # 计算差异
    merged_df["AvgDurationChange"] = merged_df["AvgDuration_abnormal"] - merged_df["AvgDuration_normal"]
    merged_df["SuccRateChange"] = merged_df["SuccRate_abnormal"] - merged_df["SuccRate_normal"]
    merged_df["P90Change"] = merged_df["P90_abnormal"] - merged_df["P90_normal"]
    merged_df["P95Change"] = merged_df["P95_abnormal"] - merged_df["P95_normal"]
    merged_df["P99Change"] = merged_df["P99_abnormal"] - merged_df["P99_normal"]

    # 筛选条件：成功率下降或平均时延、尾延迟上升
    filtered_df = merged_df[
        (merged_df["SuccRateChange"] < 0) |  # 成功率下降
        (merged_df["AvgDurationChange"] > 0) |  # 平均时延上升
        (merged_df["P90Change"] > 0) |  # P90 上升
        (merged_df["P95Change"] > 0) |  # P95 上升
        (merged_df["P99Change"] > 0)    # P99 上升
    ]
    
    # 选择需要展示的列
    result_df = filtered_df[[
        "SpanName",
        "AvgDuration_normal", "AvgDuration_abnormal", "AvgDurationChange",
        "SuccRate_normal", "SuccRate_abnormal", "SuccRateChange",
        "P90_normal", "P90_abnormal", "P90Change",
        "P95_normal", "P95_abnormal", "P95Change",
        "P99_normal", "P99_abnormal", "P99Change"
    ]]

    return result_df

# IMPORTANT: do not change the function signature!!
def start_rca(params: Dict):
    pprint(params)
    directory = "/app/output"
    if not os.path.exists(directory):
        os.makedirs(directory)
        
    normal_start, normal_end = params['normal_time_range'][0]
    abnormal_start, abnormal_end = params['abnormal_time_range'][0]
    
    normal_data = process_and_aggregate_csv("/app/input/traces.csv", "ts-ui-dashboard", normal_start, normal_end)
    abnormal_data = process_and_aggregate_csv("/app/input/traces.csv", "ts-ui-dashboard", abnormal_start, abnormal_end)
    compare_result = compare_dataframes(normal_data, abnormal_data)
    
    normal_data.to_csv("/app/output/normal_data.csv", index=False)
    abnormal_data.to_csv("/app/output/abnormal_data.csv", index=False)
    compare_result.to_csv("/app/output/compare.csv", index=False)
 
 
if __name__ == "__main__":
    # data = process_and_aggregate_csv("/home/nn/workspace/rcabench/input/traces.csv", "ts-ui-dashboard", "2024-11-25 10:05:13.627123723349", "2024-11-27 10:05:13.627123723349")
    # output_file = "output.csv"
    # data.to_csv(output_file, index=False)  # index=False 不保存索引列
    # print(f"DataFrame saved to {output_file}")
    


    normal_data = process_and_aggregate_csv("/home/nn/workspace/rcabench/input/traces.csv", "ts-ui-dashboard", 1732585695, 1732586895)
    abnormal_data = process_and_aggregate_csv("/home/nn/workspace/rcabench/input/traces.csv", "ts-ui-dashboard", 1732586895, 1732586946)
    compare_result = compare_dataframes(normal_data, abnormal_data)
    
    # normal_data.to_csv("/app/output/normal_data.csv", index=False)
    # abnormal_data.to_csv("/app/output/abnormal_data.csv", index=False)
    # compare_result.to_csv("/app/output/compare.csv", index=False)