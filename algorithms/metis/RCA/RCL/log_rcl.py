import csv
import os
from collections import defaultdict
from datetime import datetime

import pandas as pd


def process_log_files(input_folder_path, output_folder_path):
    # 确保输出文件夹存在
    os.makedirs(output_folder_path, exist_ok=True)

    # 初始化计数器和存储结构
    total_logs = 0
    error_template_counts = defaultdict(lambda: defaultdict(int))
    warn_template_counts = defaultdict(lambda: defaultdict(int))
    time_stamps = []
    log_bodies = defaultdict(lambda: defaultdict(list))

    # 遍历文件夹中的CSV文件
    for filename in os.listdir(input_folder_path):
        if filename.endswith('.csv'):
            file_path = os.path.join(input_folder_path, filename)
            service_name = filename.replace('.csv', '')

            # 读取CSV文件
            with open(file_path, 'r', encoding='utf-8') as file:
                reader = csv.DictReader(file)
                for row in reader:
                    temp_id = row['temp_id']
                    log_level = row['log_level']
                    timestamp_str = row['Timestamp']
                    log_body = row['Body']

                    # 解析时间戳，处理不同长度的小数秒
                    try:
                        timestamp = datetime.strptime(timestamp_str[:26], "%Y-%m-%d %H:%M:%S.%f")
                    except ValueError:
                        timestamp = datetime.strptime(timestamp_str[:19], "%Y-%m-%d %H:%M:%S")

                    time_stamps.append(timestamp)

                    # 更新日志总数
                    total_logs += 1

                    # 更新错误和警告模板计数，并记录日志体
                    if log_level == 'ERROR':
                        error_template_counts[service_name][temp_id] += 1
                        log_bodies[service_name][temp_id].append(log_body)
                    elif log_level == 'WARN':
                        warn_template_counts[service_name][temp_id] += 1
                        log_bodies[service_name][temp_id].append(log_body)

    # 计算时间跨度
    total_minutes = (max(time_stamps) - min(time_stamps)).total_seconds() / 60 if time_stamps else 1

    # 准备输出数据
    output_data = []
    for service_name in error_template_counts.keys():
        for temp_id in set(error_template_counts[service_name].keys()).union(warn_template_counts[service_name].keys()):
            error_count = error_template_counts[service_name][temp_id]
            warn_count = warn_template_counts[service_name][temp_id]
            error_rate = error_count / total_logs if total_logs > 0 else 0
            warn_rate = warn_count / total_logs if total_logs > 0 else 0
            avg_errors_per_minute = error_count / total_minutes
            avg_warns_per_minute = warn_count / total_minutes
            most_common_body = max(set(log_bodies[service_name][temp_id]), key=log_bodies[service_name][temp_id].count)
            output_data.append(
                [
                    service_name,
                    temp_id,
                    error_rate,
                    warn_rate,
                    avg_errors_per_minute,
                    avg_warns_per_minute,
                    most_common_body,
                ]
            )

    # 按错误率降序排序
    output_data.sort(key=lambda x: x[2], reverse=True)

    # 写入输出CSV文件
    output_file = 'log_rcl_results.csv'
    output_file_path = os.path.join(output_folder_path, output_file)
    with open(output_file_path, 'w', newline='', encoding='utf-8') as file:
        writer = csv.writer(file)
        writer.writerow(
            [
                'Service Name',
                'Template ID',
                'Error Rate',
                'Warn Rate',
                'Avg Errors/Min',
                'Avg Warns/Min',
                'Most Common Log Body',
            ]
        )
        for row in output_data:
            writer.writerow(row)

    # print(f"Summary saved to {output_file_path}")


def count_template_occurrences(service_name, template_id, directory):
    # 构建服务对应的日志文件路径
    log_file_path = os.path.join(directory, f"{service_name}.csv")

    # 检查文件是否存在
    if not os.path.exists(log_file_path):
        # print(f"Log file not found: {log_file_path}")
        return 0

    # 读取CSV文件
    try:
        log_data = pd.read_csv(log_file_path)
        # 计算包含指定Template ID的条目数量
        count = log_data['temp_id'].value_counts().get(template_id, 0)
        return count
    except Exception as e:
        print(f"Error reading {log_file_path}: {e}")
        return 0


def add_template_counts_to_csv(csv_file_path, output_csv_path, directory):
    # 读取CSV文件
    df = pd.read_csv(csv_file_path)

    # 为新的计数列初始化列
    df['Normal Rate/Min'] = 0

    # 遍历每一行数据
    for index, row in df.iterrows():
        # 获取服务名称和模板ID
        service_name = row['Service Name'].replace("_filtered", "")
        template_id = row['Template ID']

        # 统计模板出现的次数
        count = count_template_occurrences(service_name, template_id, directory)

        # 计算出现次数除以10的结果，并更新到新列中
        df.at[index, 'Normal Rate/Min'] = count / 10

    # 将更新后的DataFrame保存到新的CSV文件中
    df.to_csv(output_csv_path, index=False)
    # print(f"Updated CSV has been saved to {output_csv_path}")


def process_and_output_log_data(csv_file_path):
    # 读取 CSV 文件
    df = pd.read_csv(csv_file_path)

    # 确保 "Normal Rate/Min" 字段存在并计算
    if 'Normal Rate/Min' not in df.columns:
        df['Normal Rate/Min'] = [0] * len(df)  # 示例中的预设值，具体实现需根据实际情况来计算或修改

    # 计算增长或降低的比例
    def calculate_pattern(row):
        change_percentage = row['Error Rate'] * 100
        if change_percentage > 0:
            return f"increase {change_percentage:.2f}%"
        elif change_percentage < 0:
            return f"decrease {-change_percentage:.2f}%"
        else:
            return "no change"

    df['Pattern'] = df.apply(calculate_pattern, axis=1)

    # 创建输出格式
    output = []
    for index, row in df.head(5).iterrows():
        print("log event - top ", (index + 1))
        output.append("- log temp: {}".format(row['Most Common Log Body']))
        output.append("- service: {}".format(row['Service Name']))
        output.append("- normal_frequence (per min): {:.2f}".format(row['Normal Rate/Min']))
        output.append("- observed_frequence: {:.2f}".format(row['Avg Errors/Min']))
        output.append("- pattern: {}".format(row['Pattern']))
        output.append("")
    print("\n".join(output))

    # 确保所有相关列都正确处理，特别是数值列
    df['Avg Errors/Min'] = pd.to_numeric(df['Avg Errors/Min'], errors='coerce')
    df['Normal Rate/Min'] = pd.to_numeric(df['Normal Rate/Min'], errors='coerce')

    # 计算得分
    df['Score'] = (df['Avg Errors/Min'] - df['Normal Rate/Min']) * (1 / (df.index + 1))

    # 按照服务名称累加得分
    service_scores = df.groupby('Service Name')['Score'].sum()

    # 按得分降序排序
    sorted_scores = service_scores.sort_values(ascending=False)

    # 输出保存到CSV文件
    sorted_scores.to_csv(os.path.join(os.path.dirname(csv_file_path), "log_service_scores.csv"), header=['Score'])

    # print("Service scores saved to log_service_scores.csv")


# 使用示例
if __name__ == '__main__':
    input_folder = 'log_ad_output/parsed/abnormal'
    file_path = 'rcl_output'
    csv_file_path = 'rcl_output/log_rcl_results.csv'
    directory = 'log_ad_output/parsed/normal'
    process_log_files(input_folder, file_path)
    add_template_counts_to_csv(csv_file_path, csv_file_path, directory)
    process_and_output_log_data(csv_file_path)
