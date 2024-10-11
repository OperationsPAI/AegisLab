import pandas as pd
import numpy as np
import time
from typing import List, Dict
import argparse

def calculate_rho_squared(df_normal: pd.DataFrame, df_abnormal: pd.DataFrame) -> Dict[str, float]:
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
        if column in df_abnormal.columns and column not in ['TimeStamp', 'PodName']:
            normal_data = df_normal[column].dropna()
            abnormal_data = df_abnormal[column].dropna()
            min_length = min(len(normal_data), len(abnormal_data))
            if min_length < 2:
                continue
            try:
                cov_matrix = np.cov(normal_data, abnormal_data)
                cov = cov_matrix[0, 1]
                var_normal = normal_data.var()
                var_abnormal = abnormal_data.var()
                if var_normal > 0 and var_abnormal > 0:
                    rho_squared = (cov ** 2) / (var_normal * var_abnormal)
                    results[column] = rho_squared
            except Exception:
                continue
    return results


def diagnose_faults(
        fault_list: List[Dict],
        metric_data: pd.DataFrame,
        time_range: int = 300,
        top_n: int = 5
) -> pd.DataFrame:
    """
    诊断故障并返回每个时间戳的 Top N 服务。

    参数:
        fault_list (List[Dict]): 故障事件列表，每个事件包含 'inject_timestamp'。
        metric_data (pd.DataFrame): 指标数据，包含 'TimeStamp', 'PodName' 和其他指标列。
        time_range (int, optional): 时间范围（秒）。默认为300秒（5分钟）。
        top_n (int, optional): 每个时间戳选择的 Top N 服务。默认为5。

    返回:
        pd.DataFrame: 包含 'timestamp' 和 'pod' 的扩展数据框。
    """
    results = []
    all_services = {}

    metric_data['TimeStamp'] = pd.to_numeric(metric_data['TimeStamp'], errors='coerce')
    metric_data['TimeStamp'] = metric_data['TimeStamp']
    service_names = metric_data['PodName'].unique()


    for service in service_names:
        service_df = metric_data[metric_data['PodName'] == service].copy()
        service_df = service_df.dropna(subset=['TimeStamp'])
        service_df = service_df.sort_values('TimeStamp')

        for event in fault_list:
            inject_timestamp = int(event['inject_timestamp'])
            if inject_timestamp not in all_services:
                all_services[inject_timestamp] = {}

            normal_start = inject_timestamp - time_range
            normal_end = inject_timestamp
            abnormal_start = inject_timestamp
            abnormal_end = inject_timestamp + time_range

            normal_range = service_df[
                (service_df['TimeStamp']   >= normal_start) &
                (service_df['TimeStamp']  < normal_end)
                ]
            abnormal_range = service_df[
                (service_df['TimeStamp']  >= abnormal_start) &
                (service_df['TimeStamp']   <= abnormal_end)
                ]

            metric_scores = calculate_rho_squared(normal_range, abnormal_range)
            if metric_scores:
                max_score = max(metric_scores.values())
                all_services[inject_timestamp][service] = max_score

    # 为每个时间戳选择 Top N 服务
    for timestamp, services in all_services.items():
        sorted_services = sorted(services.items(), key=lambda x: x[1], reverse=True)
        top_services = [service for service, _ in sorted_services[:top_n]]
        results.append({
            'timestamp': timestamp,
            'top_services': top_services
        })

    top_services_df = pd.DataFrame(results)

    # 扩展每个时间戳的 Top 服务为单独的行
    expanded_rows = []
    for _, row in top_services_df.iterrows():
        timestamp = row['timestamp']
        services = row['top_services']
        for service in services:
            expanded_rows.append({'timestamp': timestamp, 'pod': service})

    expanded_df = pd.DataFrame(expanded_rows)

    return expanded_df



def main(csv_file_path):
    selected_columns = ['TimeStamp', 'MetricName', 'Value', 'PodName']
    df = pd.read_csv(csv_file_path, usecols=selected_columns, parse_dates=['TimeStamp'])

    df_pivot = df.pivot_table(index=['TimeStamp', 'PodName'], columns='MetricName', values='Value').reset_index()

    df_pivot['TimeStamp'] = df_pivot['TimeStamp'].apply(lambda x: int(x.timestamp()))

    sorted_timestamps = df_pivot['TimeStamp'].sort_values()
    
    inject_time = sorted_timestamps.iloc[len(sorted_timestamps) // 2]
    
    fault_list = [
        {'inject_timestamp': inject_time},
    ]

    metric_data = df_pivot
    output_df = diagnose_faults(fault_list, metric_data)
    
    print(output_df)
    
    
if __name__ == "__main__":
    parser = argparse.ArgumentParser(description='read csv')
    parser.add_argument('csv_file', type=str, help='csv path')
    args = parser.parse_args()

    main(args.csv_file)