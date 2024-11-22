import csv
import os
import shutil
from datetime import datetime

import pandas as pd


def find_time_range(file_path):
    # 初始化变量，用于存储最早和最晚的时间
    earliest_start = None
    latest_end = None

    try:
        # 读取CSV文件
        with open(file_path, mode="r") as file:
            csv_reader = csv.DictReader(file)
            for row in csv_reader:
                # 找到第三个短横线的位置，并从此处分割字符串
                third_dash_index = row["TimeRanges"].find(
                    "-",
                    row["TimeRanges"].find("-", row["TimeRanges"].find("-") + 1) + 1,
                )
                if third_dash_index == -1:
                    # print("Skipping invalid time range:", row['TimeRanges'])
                    continue

                start_time_str = row["TimeRanges"][:third_dash_index]
                end_time_str = row["TimeRanges"][third_dash_index + 1 :]

                # 将时间字符串转换为datetime对象
                start_time = datetime.fromisoformat(start_time_str)
                end_time = datetime.fromisoformat(end_time_str)

                # 更新最早开始时间和最晚结束时间
                if earliest_start is None or start_time < earliest_start:
                    earliest_start = start_time
                if latest_end is None or end_time > latest_end:
                    latest_end = end_time

        # 返回结果
        return earliest_start, latest_end

    except FileNotFoundError:
        print(f"Error: The file '{file_path}' does not exist.")
        return None, None
    except Exception as e:
        print(f"An error occurred: {e}")
        return None, None


def read_service_names(csv_file):
    """从CSV文件中读取ServiceName"""
    service_names = []
    with open(csv_file, mode="r") as file:
        csv_reader = csv.DictReader(file)
        for row in csv_reader:
            service_names.append(row["ServiceName"])
    return service_names


def keep_abnormal_service_metrics(directory, service_names, filter_path):
    """遍历目录，将符合ServiceName的文件复制到指定目录"""
    files_to_keep = []

    # 如果目标路径不存在，则创建它
    if not os.path.exists(filter_path):
        os.makedirs(filter_path)

    # 遍历文件夹中的文件
    for file in os.listdir(directory):
        # 如果文件名匹配任一ServiceName，将其添加到保留列表并复制到目标目录
        if any(file.startswith(service_name) for service_name in service_names):
            files_to_keep.append(file)
            src_file_path = os.path.join(directory, file)
            dst_file_path = os.path.join(filter_path, file)
            shutil.copy2(src_file_path, dst_file_path)
            # print(f"Copied file: {file} to {filter_path}")

    return files_to_keep


def parse_custom_datetime(date_str):
    """解析包含纳秒的时间字符串，转换为datetime对象，截断到微秒。"""
    if "." in date_str:
        date_str, ns = date_str.split(".")
        ns = ns.ljust(6, "0")  # 确保至少有6位数字
        date_str = f"{date_str}.{ns[:6]}"  # 截取前6位数字
    return datetime.strptime(date_str, "%Y-%m-%d %H:%M:%S.%f")


def keep_abnormal_period(directory, earliest, latest):
    # 遍历文件夹中的CSV文件
    for file_name in os.listdir(directory):
        if file_name.endswith(".csv"):
            file_path = os.path.join(directory, file_name)
            with open(file_path, mode="r") as file:
                csv_reader = csv.reader(file)
                header = next(csv_reader)  # 读取表头
                filtered_data = [header]  # 初始化筛选后的数据包含表头

                for row in csv_reader:
                    try:
                        # 解析时间，适应多种格式
                        if "T" in row[0]:
                            row_time = datetime.fromisoformat(row[0])
                        else:
                            row_time = parse_custom_datetime(row[0])

                        if earliest <= row_time <= latest:
                            filtered_data.append(row)
                    except ValueError:
                        print(f"Skipping row with invalid date format: {row[0]}")

            # 将筛选后的数据写回到文件（或写入新文件）
            with open(file_path, "w", newline="") as file:
                csv_writer = csv.writer(file)
                csv_writer.writerows(filtered_data)
                # print(f"Filtered and updated file: {file_name}")


def filter_outliers(df, column, lower_percentile=0.005, upper_percentile=0.995):
    lower_bound = df[column].quantile(lower_percentile)
    upper_bound = df[column].quantile(upper_percentile)
    return df[(df[column] >= lower_bound) & (df[column] <= upper_bound)]


def filter_csv_columns(df):
    filtered_df = df[["TimeUnix"]].copy()  # 先保留 'TimeUnix' 列

    # 遍历所有数值列进行筛选
    for column in df.columns:
        if column != "TimeUnix":  # 忽略 'TimeUnix' 列
            filtered_column = filter_outliers(df, column)  # 过滤正态分布外 1% 的数据
            filtered_df[column] = filtered_column[column]  # 将筛选后的列添加到新数据框

    return filtered_df


def filter_metrics(input_folder, output_folder):
    if not os.path.exists(output_folder):
        os.makedirs(output_folder)

    # 遍历输入文件夹中的每个 CSV 文件
    for filename in os.listdir(input_folder):
        if filename.endswith(".csv"):
            file_path = os.path.join(input_folder, filename)

            df = pd.read_csv(file_path)

            filtered_data = filter_csv_columns(df)

            # 将筛选出的数据保存到新的 CSV 文件中
            output_file_path = os.path.join(output_folder, filename)  # 输出文件路径
            filtered_data.to_csv(
                output_file_path, index=False
            )  # 保存为 CSV 文件，不包含索引
            # 输出处理进度信息
            # print(f'处理文件：{filename}，筛选并保存至：{output_file_path}')


def calculate_metric_statistics(input_path, output_path):
    # 确保输出目录存在
    if not os.path.exists(output_path):
        os.makedirs(output_path)

    # 遍历输入目录中的所有文件
    for root, dirs, files in os.walk(input_path):
        for file in files:
            if file.endswith(".csv"):
                input_file_path = os.path.join(root, file)
                try:
                    # 读取CSV文件
                    data = pd.read_csv(input_file_path)

                    # 忽略第一列（假设是时间戳）
                    data = data.iloc[:, 1:]

                    # 计算每列的最大值、最小值和平均值
                    summary_df = pd.DataFrame(
                        {
                            "Metric": data.columns,
                            "Max": data.max().round(3),
                            "Min": data.min().round(3),
                            "Mean": data.mean().round(3),
                        }
                    )

                    # 构建输出文件路径
                    output_file_path = os.path.join(output_path, f"{file}")
                    # 保存结果到CSV文件
                    summary_df.to_csv(output_file_path, index=False)
                    # print(f"Summary statistics saved to {output_file_path}")

                except Exception as e:
                    print(f"Error processing file {input_file_path}: {str(e)}")


def calculate_overlap(normal_path, abnormal_path, output_path, k=0.8):
    # Ensure the output directory exists
    if not os.path.exists(output_path):
        os.makedirs(output_path)

    overlap_results = []
    # Process each file in the normal directory
    for filename in os.listdir(normal_path):
        normal_file = os.path.join(normal_path, filename)
        abnormal_file = os.path.join(abnormal_path, filename)

        # Check if the corresponding file exists in the abnormal directory
        if os.path.exists(abnormal_file):
            # Load the data
            normal_data = pd.read_csv(normal_file)
            abnormal_data = pd.read_csv(abnormal_file)

            # Iterate through each metric in the normal file
            for i, row in normal_data.iterrows():
                metric = row["Metric"]
                if (row["Min"] == 0 and row["Max"] == 0) and row["Mean"] == 0:
                    continue
                normal_range = (row["Min"], row["Max"])
                normal_average = row[
                    "Mean"
                ]  # Use the Mean value directly from the data

                # Find the same metric in the abnormal file
                abnormal_row = abnormal_data[abnormal_data["Metric"] == metric]
                if not abnormal_row.empty:
                    abnormal_range = (
                        abnormal_row.iloc[0]["Min"],
                        abnormal_row.iloc[0]["Max"],
                    )
                    abnormal_average = abnormal_row.iloc[0][
                        "Mean"
                    ]  # Use the Mean value directly from the data

                    # Check if the ranges are exactly the same
                    if abnormal_range == normal_range:
                        overlap_percentage = 100  # Full overlap
                    else:
                        # Calculate the overlap and avg_change
                        max_lower = max(normal_range[0], abnormal_range[0])
                        min_upper = min(normal_range[1], abnormal_range[1])
                        overlap = max(0, min_upper - max_lower)
                        total_range = normal_range[1] - normal_range[0]
                        overlap_percentage = (
                            (overlap / total_range) * 100 if total_range > 0 else 0
                        )

                    avg_change = (
                        abs(abnormal_average - normal_average) / normal_average * 100
                        if normal_average != 0
                        else 0
                    )
                    change_score = k * overlap_percentage + (1 - k) * avg_change

                    # Store the results
                    result = pd.DataFrame(
                        {
                            "service_name": [
                                "-".join(filename.split(".")[0].split("-")[:-2])
                            ],
                            "metric_name": [metric],
                            "change_score": [change_score],
                            "overlap_percentage": [overlap_percentage],
                            "avg_change": [avg_change],
                            "abnormal_max": [abnormal_range[1]],
                            "abnormal_min": [abnormal_range[0]],
                            "abnormal_average": [abnormal_average],
                            "normal_max": [normal_range[1]],
                            "normal_min": [normal_range[0]],
                            "normal_average": [normal_average],
                        }
                    )
                    overlap_results.append(result)
    overlap_results = pd.concat(overlap_results, ignore_index=True)

    # Sort the results by change_score in descending order
    results = overlap_results.sort_values(by="change_score", ascending=False)
    results = results[results["change_score"].abs() > 0]

    # Save the results to a CSV file
    results_file = os.path.join(output_path, "metric_rcl_results.csv")
    results.to_csv(results_file, index=False)
    # print(f"Results saved to {results_file}")


def process_and_output_metric_events(file_path):
    metric_rcl_output_path = os.path.join(file_path, "metric_rcl_results.csv")
    # 读取CSV文件

    df = pd.read_csv(metric_rcl_output_path).round(3)
    df_sorted = df.sort_values(by="change_score", ascending=False)
    # print(df)

    metric_events = []
    for index, row in df_sorted.iterrows():
        metric_events.append(
            {
                "top": index + 1,
                "metric_name": row["metric_name"],
                "change_score": row["change_score"],
                "service_name": row["service_name"],
                "overlap_percentage": row["overlap_percentage"],
                "abnormal_min": row["abnormal_min"],
                "abnormal_max": row["abnormal_max"],
                "normal_min": row["normal_min"],
                "normal_max": row["normal_max"],
                "avg_change": row["avg_change"],
                "abnormal_average": row["abnormal_average"],
                "normal_average": row["normal_average"],
            }
        )
    return metric_events


def calculate_service_ranking(file_path):
    metric_rcl_output_path = os.path.join(file_path, "metric_rcl_results.csv")

    df = pd.read_csv(metric_rcl_output_path).round(3)
    filtered_df = df.rename(columns={"service_name": "ServiceName"})

    filtered_df.loc[:, "WeightedChange"] = filtered_df["change_score"] * (
        1 / (filtered_df.index + 1)
    )
    metric_count = filtered_df["ServiceName"].value_counts()
    total_metrics = len(filtered_df)
    filtered_df.loc[:, "anomaly_score"] = filtered_df.apply(
        lambda row: row["WeightedChange"]
        * (metric_count[row["ServiceName"]] / total_metrics),
        axis=1,
    )
    result = filtered_df.groupby("ServiceName")["anomaly_score"].sum().reset_index()
    result = result.sort_values(by="anomaly_score", ascending=False)
    # result.to_csv("rcl_output/metric_service_scores.csv", index=False)
    return result


# main
def metric_rcl(file_path):
    normal_file_path = os.path.join(file_path, "normal/processed_metrics")
    detect_file_path = os.path.join(file_path, "abnormal/processed_metrics")

    normal_filtered_path = os.path.join(
        file_path, "normal/processed_metrics/filtered_scalered_metrics"
    )
    detect_filtered_path = os.path.join(
        file_path, "abnormal/processed_metrics/filtered_scalered_metrics"
    )

    normal_output_path = os.path.join(file_path, "normal/rcl_data")
    detect_output_path = os.path.join(file_path, "abnormal/rcl_data")
    output_file_path = os.path.join(file_path, "rcl_output")

    service_list_path = os.path.join(file_path, "metric_ad_output/service_list.csv")

    earliest, latest = find_time_range(service_list_path)
    # print(f"Earliest start time: {earliest.isoformat()}")
    # print(f"Latest end time: {latest.isoformat()}")

    service_names = read_service_names(service_list_path)

    keep_abnormal_service_metrics(normal_file_path, service_names, normal_filtered_path)
    keep_abnormal_service_metrics(detect_file_path, service_names, detect_filtered_path)
    keep_abnormal_period(detect_filtered_path, earliest, latest)

    filter_metrics(normal_filtered_path, normal_filtered_path)
    filter_metrics(detect_filtered_path, detect_filtered_path)

    calculate_metric_statistics(normal_filtered_path, normal_output_path)
    calculate_metric_statistics(detect_filtered_path, detect_output_path)
    calculate_overlap(normal_output_path, detect_output_path, output_file_path)
    service_ranking_df = calculate_service_ranking(output_file_path)
    return process_and_output_metric_events(output_file_path), service_ranking_df


# Example usage:
# file_path = R'E:\OneDrive - CUHK-Shenzhen\RCA_Dataset\test_new_datasets\ts\cpu\ts-delivery-service-1011-1531'
# metric_rcl(file_path)
