import csv
import os
import pandas as pd
from datetime import datetime


def devide_by_pod_name(file_path):
    processed_file_path = os.path.join(file_path, "processed_metrics")

    if not os.path.exists(processed_file_path):
        os.makedirs(processed_file_path)

    input_file = os.path.join(file_path, "metrics.csv")

    with open(input_file, mode="r", encoding="utf-8") as file:
        reader = csv.DictReader(file)

        pod_metrics = {}

        for row in reader:
            pod_name = row["k8s_pod_name"]
            timestamp = row["TimeUnix"]
            metric_name = row["MetricName"]
            value = row["Value"]
            direction = row.get("direction", "")
            if not pod_name:
                continue
            if metric_name == "k8s.pod.network.io":
                if direction == "transmit":
                    metric_name = "transmit_bytes"
                elif direction == "receive":
                    metric_name = "receive_bytes"

            if pod_name not in pod_metrics:
                pod_metrics[pod_name] = []
            pod_metrics[pod_name].append((timestamp, metric_name, value))

    def parse_timestamp(ts):
        try:
            return datetime.strptime(ts, "%Y-%m-%d %H:%M:%S.%f")
        except ValueError:
            return datetime.strptime(ts.split(".")[0], "%Y-%m-%d %H:%M:%S")

    exclude_columns = {
        "k8s.container.ready",
        "k8s.container.restarts",
        "k8s.deployment.available",
        "k8s.deployment.desired",
        "k8s.pod.phase",
        "k8s.replicaset.available",
        "k8s.replicaset.desired",
        "k8s.statefulset.current_pods",
        "k8s.statefulset.desired_pods",
        "k8s.statefulset.ready_pods",
        "k8s.statefulset.updated_pods",
        "k8s.namespace.phase",
        "k8s.container.cpu_limit",
        "k8s.container.cpu_request",
        "k8s.container.memory_limit",
        "k8s.container.memory_request",
        "k8s.pod.memory.page_faults",
        "k8s.pod.memory.major_page_faults",
        "k8s.pod.memory.rss",
        "k8s.pod.memory.working_set",
        "k8s.pod.memory.node.utilization",
        "k8s.pod.memory.available",
        "k8s.pod.cpu.utilization",
        "k8s.pod.cpu.node.utilization",
        "k8s.pod.filesystem.available",
        "k8s.pod.filesystem.usage",
        "k8s.pod.filesystem.capacity",
        "container.memory.page_faults",
        "container.memory.major_page_faults",
        "container.memory.working_set",
        "container.memory.rss",
        "container.memory.available",
        "container.memory.usage",
        "container.cpu.utilization",
        "container.filesystem.available",
        "container.filesystem.capacity",
        "container.filesystem.usage",
    }

    for pod_name, logs in pod_metrics.items():
        if (
            "redis-cart" in pod_name
            or "loadgenerator" in pod_name
            or "mysql" in pod_name
        ):
            continue

        # 按时间戳排序
        logs.sort(key=lambda x: parse_timestamp(x[0]))

        # 获取所有的 MetricName，排除不需要的列
        metric_names = sorted(set(log[1] for log in logs) - exclude_columns)

        # 创建 metric 文件
        file_name = f"{pod_name}.csv"
        with open(
            os.path.join(processed_file_path, file_name), mode="w", encoding="utf-8"
        ) as file:
            # 写入列名：TimeUnix 加上所有 MetricName
            file.write("TimeUnix," + ",".join(metric_names) + "\n")

            # 按时间戳分组
            grouped_logs = {}
            for timestamp, metric_name, value in logs:
                if timestamp not in grouped_logs:
                    grouped_logs[timestamp] = {metric: "" for metric in metric_names}
                if metric_name in metric_names:
                    grouped_logs[timestamp][metric_name] = value

            for timestamp in sorted(grouped_logs.keys(), key=parse_timestamp):
                values = [grouped_logs[timestamp][metric] for metric in metric_names]
                if any(values):
                    file.write(f"{timestamp}," + ",".join(values) + "\n")

    # print("Metrics have been successfully separated, sorted by timestamp, and saved.")


def process_metrics(file_path):
    processed_file_path = os.path.join(file_path, "processed_metrics")
    # 定义网络指标和文件系统指标需要处理的列
    network_columns = ["k8s.pod.network.errors", "receive_bytes", "transmit_bytes"]
    # filesystem_columns = ['k8s.pod.filesystem.capacity', 'k8s.pod.filesystem.usage']

    # 遍历输入文件夹中的 CSV 文件
    for filename in os.listdir(processed_file_path):
        if filename.endswith(".csv"):
            input_path = os.path.join(processed_file_path, filename)
            df = pd.read_csv(input_path)

            # 处理网络指标：差分计算
            if all(col in df.columns for col in network_columns):
                if len(df) > 1:
                    df[network_columns] = df[network_columns].diff()
                    df[network_columns] = df[network_columns].fillna(0)  # 填充NaN值
                    # print(f'Network metrics processed for: {filename}')
                else:
                    print(f"Skipped network metrics (not enough rows): {filename}")
            else:
                print(f"Skipped network metrics (missing columns): {filename}")

            # 保存处理后的数据
            output_path = os.path.join(processed_file_path, filename)
            df.to_csv(output_path, index=False)

            # print(f'Processed: {filename}')


# main
def process_metric_data(file_path):
    # process normal metrics
    normal_file_path = os.path.join(file_path, "normal")
    devide_by_pod_name(normal_file_path)
    process_metrics(normal_file_path)

    # process abnormal metrics
    abnormal_file_path = os.path.join(file_path, "abnormal")
    devide_by_pod_name(abnormal_file_path)
    process_metrics(abnormal_file_path)


# Example usage:
# file_path = R'E:\OneDrive - CUHK-Shenzhen\RCA_Dataset\test_new_datasets\onlineboutique\pod_failure\paymentservice-1011-1525'
# process_metric_data(file_path)
