import csv
from datetime import datetime
import pandas as pd
import os
import shutil

def find_time_range(file_path):
    # 初始化变量，用于存储最早和最晚的时间
    earliest_start = None
    latest_end = None

    try:
        # 读取CSV文件
        with open(file_path, mode='r') as file:
            csv_reader = csv.DictReader(file)
            for row in csv_reader:
                # 找到第三个短横线的位置，并从此处分割字符串
                third_dash_index = row['TimeRanges'].find('-', row['TimeRanges'].find('-', row['TimeRanges'].find(
                    '-') + 1) + 1)
                if third_dash_index == -1:
                    print("Skipping invalid time range:", row['TimeRanges'])
                    continue

                start_time_str = row['TimeRanges'][:third_dash_index]
                end_time_str = row['TimeRanges'][third_dash_index + 1:]

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
    """ 从CSV文件中读取ServiceName """
    service_names = []
    with open(csv_file, mode='r') as file:
        csv_reader = csv.DictReader(file)
        for row in csv_reader:
            service_names.append(row['ServiceName'])
    return service_names


# def filter_and_delete_other_files(directory, service_names):
    # """ 遍历目录，删除不匹配ServiceName的文件 """
    # files_to_keep = []
    # files_to_delete = []

    # # 遍历文件夹中的文件
    # for file in os.listdir(directory):
    #     # 如果文件名匹配任一ServiceName，将其添加到保留列表
    #     if any(file.startswith(service_name) for service_name in service_names):
    #         files_to_keep.append(file)
    #     else:
    #         files_to_delete.append(file)

    # # 删除不需要的文件
    # for file in files_to_delete:
    #     os.remove(os.path.join(directory, file))
    #     print(f"Deleted file: {file}")

    # return files_to_keep



def filter_files_by_ad_result(directory, service_names, filter_path):
    """ 遍历目录，将符合ServiceName的文件复制到指定目录，不删除不匹配的文件 """
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
            print(f"Copied file: {file} to {filter_path}")

    return files_to_keep


def parse_custom_datetime(date_str):
    """解析包含纳秒的时间字符串，转换为datetime对象，截断到微秒。"""
    if '.' in date_str:
        date_str, ns = date_str.split('.')
        ns = ns.ljust(6, '0')  # 确保至少有6位数字
        date_str = f"{date_str}.{ns[:6]}"  # 截取前6位数字
    return datetime.strptime(date_str, "%Y-%m-%d %H:%M:%S.%f")


def filter_data_by_ad_time_window(directory, earliest, latest):
    # 遍历文件夹中的CSV文件
    for file_name in os.listdir(directory):
        if file_name.endswith('.csv'):
            file_path = os.path.join(directory, file_name)
            with open(file_path, mode='r') as file:
                csv_reader = csv.reader(file)
                header = next(csv_reader)  # 读取表头
                filtered_data = [header]  # 初始化筛选后的数据包含表头

                for row in csv_reader:
                    try:
                        # 解析时间，适应多种格式
                        if 'T' in row[0]:
                            row_time = datetime.fromisoformat(row[0])
                        else:
                            row_time = parse_custom_datetime(row[0])

                        if earliest <= row_time <= latest:
                            filtered_data.append(row)
                    except ValueError:
                        print(f"Skipping row with invalid date format: {row[0]}")

            # 将筛选后的数据写回到文件（或写入新文件）
            with open(file_path, 'w', newline='') as file:
                csv_writer = csv.writer(file)
                csv_writer.writerows(filtered_data)
                print(f"Filtered and updated file: {file_name}")


def filter_outliers(df, column, lower_percentile=0.005, upper_percentile=0.995):

    lower_bound = df[column].quantile(lower_percentile)
    upper_bound = df[column].quantile(upper_percentile)
    return df[(df[column] >= lower_bound) & (df[column] <= upper_bound)]


def filter_csv_columns(df):

    filtered_df = df[['TimeUnix']].copy()  # 先保留 'TimeUnix' 列

    # 遍历所有数值列进行筛选
    for column in df.columns:
        if column != 'TimeUnix':  # 忽略 'TimeUnix' 列
            filtered_column = filter_outliers(df, column)  # 过滤正态分布外 1% 的数据
            filtered_df[column] = filtered_column[column]  # 将筛选后的列添加到新数据框

    return filtered_df



def filter_data(input_folder, output_folder):

    if not os.path.exists(output_folder):
        os.makedirs(output_folder)

    # 遍历输入文件夹中的每个 CSV 文件
    for filename in os.listdir(input_folder):
        if filename.endswith('.csv'):  # 只处理 CSV 文件
            file_path = os.path.join(input_folder, filename)  # 获取输入文件的完整路径
            # 加载 CSV 文件
            df = pd.read_csv(file_path)
            # 对数据进行筛选
            filtered_data = filter_csv_columns(df)
            # 将筛选出的数据保存到新的 CSV 文件中
            output_file_path = os.path.join(output_folder, filename)  # 输出文件路径
            filtered_data.to_csv(output_file_path, index=False)  # 保存为 CSV 文件，不包含索引
            # 输出处理进度信息
            print(f'处理文件：{filename}，筛选并保存至：{output_file_path}')

file_path = R'E:\OneDrive - CUHK-Shenzhen\RCA_Dataset\test_new_datasets\onlineboutique\cpu\checkoutservice-1011-1441'


normal_file_path = os.path.join(file_path, "normal/processed_metrics")
detect_file_path = os.path.join(file_path, "abnormal/processed_metrics")

normal_filtered_path = os.path.join(file_path, "normal/processed_metrics/ad_filtered_metrics")
detect_filtered_path = os.path.join(file_path, "abnormal/processed_metrics/ad_filtered_metrics")

service_list_path = os.path.join(file_path, "metric_ad_output/service_list.csv")

earliest, latest = find_time_range(service_list_path)
print(f"Earliest start time: {earliest.isoformat()}")
print(f"Latest end time: {latest.isoformat()}")

service_names = read_service_names(service_list_path)

filter_files_by_ad_result(normal_file_path, service_names, normal_filtered_path)
filter_files_by_ad_result(detect_file_path, service_names, detect_filtered_path)
filter_data_by_ad_time_window(detect_file_path, earliest, latest)

filter_data(normal_filtered_path,normal_filtered_path)
filter_data(detect_filtered_path,detect_filtered_path)