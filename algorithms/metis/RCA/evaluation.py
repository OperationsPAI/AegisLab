import asyncio
import os
from contextlib import contextmanager
from pathlib import Path

import numpy as np
import pandas as pd
from aiomultiprocess import Pool
from sklearn.preprocessing import MinMaxScaler,StandardScaler

from AD.log.log_ad import detect as log_ad
from AD.log.log_processing import parse as log_parse
from AD.metric.metric_ad import metric_ad
from AD.metric.metric_preprocessing import process_metric_data
from AD.trace.trace_ad import detect as trace_ad
from AD.trace.trace_processing import parse as trace_parse
from RCL.log_rcl import (
    add_template_counts_to_csv,
    process_and_output_log_data,
    process_log_files,
)
from RCL.metric_rcl import metric_rcl
from RCL.trace_rcl import (
    calculate_changes,
    print_top_spans,
    read_and_merge_data,
    sort_and_filter_data,
    weighted_change,
)


@contextmanager
def pushd(path):
    prev = os.getcwd()
    os.chdir(path)
    try:
        yield
    finally:
        os.chdir(prev)


async def trace_rcl(base_dir):
    base_dir = Path(base_dir)
    with pushd(base_dir):
        merged_df = read_and_merge_data("trace_ad_output/normal.csv", "trace_ad_output/abnormal.csv")
        merged_df = calculate_changes(merged_df)
        result_df, filtered_df = sort_and_filter_data(merged_df)
        weighted_change_result = weighted_change(filtered_df)
        # print(weighted_change_result)
        print_top_spans("rcl_output/trace_rcl_results.csv")


async def log_rcl(base_dir):
    base_dir = Path(base_dir)
    input_folder = base_dir / 'log_ad_output/parsed/abnormal'
    file_path = base_dir / 'rcl_output'
    csv_file_path = base_dir / 'rcl_output/log_rcl_results.csv'
    directory = base_dir / 'log_ad_output/parsed/normal'
    process_log_files(input_folder, file_path)
    add_template_counts_to_csv(csv_file_path, csv_file_path, directory)
    process_and_output_log_data(csv_file_path)


async def metric_rcl_async(base_dir):
    base_dir = Path(base_dir)
    with pushd(base_dir):
        metric_rcl(base_dir)


async def ranking(base_dir):
    base_dir = Path(base_dir)
    files_info = {
        base_dir / 'rcl_output/log_service_scores.csv': ['Service Name', 'Score'],
        base_dir / 'metric_ad_output/service_list.csv': ['ServiceName', 'AnomalyScore', 'TimeRanges'],
        base_dir / 'rcl_output/trace_service_scores.csv': ['ServiceName', 'ProportionalWeightedChange'],
    }

    # Check each file
    for file, columns in files_info.items():
        if not os.path.exists(file):
            df = pd.DataFrame({col: [np.nan] for col in columns})
            df.to_csv(file, index=False)
    log_scores_path = base_dir / 'rcl_output/log_service_scores.csv'
    service_scores_path = base_dir / 'metric_ad_output/service_list.csv'
    trace_changes_path = base_dir / 'rcl_output/trace_service_scores.csv'
    log_scores = pd.read_csv(log_scores_path)
    service_scores = pd.read_csv(service_scores_path)
    trace_changes = pd.read_csv(trace_changes_path)
    service_name = base_dir.name
    last_dir = base_dir.name.split('-')[0]
    second_last_dir = base_dir.parts[-2]

    # 归一化函数
    def normalize(df, column):
        if df.shape[0] <= 1 :
            df[f'{column}_normalized'] = 1
        else:
            scaler = MinMaxScaler()
            df[f'{column}_normalized'] = scaler.fit_transform(df[[column]])
        return df

    # 归一化处理
    log_scores = normalize(log_scores, 'Score')
    service_scores = normalize(service_scores, 'AnomalyScore')
    trace_changes = normalize(trace_changes, 'ProportionalWeightedChange')

    combined_scores = {}

    for _, row in log_scores.iterrows():
        if row is not None and pd.notna(row['Service Name']) and pd.notna(row['Score_normalized']):
            combined_scores[row['Service Name']] = row['Score_normalized'] / 3

    for _, row in service_scores.iterrows():
        if row is not None and pd.notna(row['ServiceName']) and pd.notna(row['AnomalyScore_normalized']):
            name = row['ServiceName']
            if pd.notna(name):
                name = name.split('-')[0]
            combined_scores[name] = combined_scores.get(name, 0) + row['AnomalyScore_normalized'] / 3

    for _, row in trace_changes.iterrows():
        if row is not None and pd.notna(row['ServiceName']) and pd.notna(row['ProportionalWeightedChange_normalized']):
            combined_scores[row['ServiceName']] = (
                combined_scores.get(row['ServiceName'], 0) + row['ProportionalWeightedChange_normalized'] / 3
            )


    result_df = pd.DataFrame(list(combined_scores.items()), columns=['ServiceName', 'CombinedScore'])
    result_df = result_df.sort_values(by='CombinedScore', ascending=False)
    result_df.to_csv(base_dir / 'final_ranking.csv', index=False)
    csv_file = base_dir / 'final_ranking.csv'
    df = pd.read_csv(csv_file)
    first_column = df.columns[0]

    # 获取前5个条目，如果不足5个则获取实际的条目数
    top_5 = df[first_column].head(5).tolist()

    ranking = None
    # print(top_5)
    # 根据 top_5 的长度动态判断排名
    if len(top_5) >= 1 and last_dir == top_5[0]:
        ranking = 1
    elif len(top_5) >= 2 and last_dir == top_5[1]:
        ranking = 2
    elif len(top_5) >= 3 and last_dir == top_5[2]:
        ranking = 3
    elif len(top_5) >= 4 and last_dir == top_5[3]:
        ranking = 4
    elif len(top_5) >= 5 and last_dir == top_5[4]:
        ranking = 5

    # result_df = pd.DataFrame({'service_name': [service_name], 'Anomaly_type': [second_last_dir], 'Ranking': [ranking]})
    # if os.path.exists('result.csv'):
    #     existing_df = pd.read_csv('result.csv')
    #     existing_df = existing_df[existing_df['service_name'] != service_name]
    #     existing_df = pd.concat([existing_df, result_df], ignore_index=True)
    # else:
    #     existing_df = result_df
    # existing_df.to_csv('result.csv', index=False)


async def process_case(base_dir, pool: Pool, file_type="log"):
    system_anomalous = False
    if file_type == "metric":
        process_metric_data(base_dir)
        system_anomalous = metric_ad(base_dir)
        if system_anomalous:
            # print("system_anomalous:", system_anomalous)
            # print("dir_path", base_dir)
            await pool.apply(metric_rcl_async, args=(base_dir,))
            await pool.apply(metric_rcl_async, args=(base_dir,))
        else:
            print(f"No metric-related anomalies detected in case {base_dir}.")
    else:
        data = await eval(f"{file_type}_parse")(base_dir, pool, f"{file_type}s.csv")
        system_anomalous, anomalies = await pool.apply(eval(f"{file_type}_ad"), args=(base_dir, data))
        if system_anomalous:
            await pool.apply(eval(f"{file_type}_rcl"), args=(base_dir,))
    return system_anomalous


async def evaluate(base_dir, pool):
    file_types = ["log", "trace", "metric"]
    system_anomalous = any(await asyncio.gather(*(process_case(base_dir, pool, file_type) for file_type in file_types)))
    if system_anomalous:
        await ranking(base_dir)
    else:
        print(f"No anomalies detected in case {base_dir}.")


async def main():
    # Define paths
    base_dirs = [
        R"E:\OneDrive - CUHK-Shenzhen\RCA_Dataset\test_new_datasets\onlineboutique\pod_failure\checkoutservice-1012-1643",
    ]

    async with Pool() as pool:
        await asyncio.gather(*(evaluate(base_dir, pool) for base_dir in base_dirs))


if __name__ == "__main__":
    asyncio.run(main())
