import asyncio
import csv
import os
from contextlib import contextmanager
from pathlib import Path
from typing import Sequence

import aiofiles
import numpy as np
import pandas as pd
import toml
from aiomultiprocess import Pool
from sklearn.cluster import KMeans
from sklearn.preprocessing import MinMaxScaler, StandardScaler

from AD.abstract import Detector, Parser
from AD.log import LogDetector, LogParser
from AD.metric import MetricParser
from AD.metric.metric_ad import MetricDetector
from AD.trace import TraceDetector, TraceParser
from RCL.log_rcl import log_rcl
from RCL.metric_rcl import metric_rcl
from RCL.trace_rcl import (
    calculate_changes,
    get_top_spans,
    read_and_merge_data,
    sort_and_filter_data,
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
        merged_df = read_and_merge_data(
            "trace_ad_output/normal.csv", "trace_ad_output/abnormal.csv"
        )
        merged_df = calculate_changes(merged_df)
        weighted_change_result = sort_and_filter_data(merged_df)
        # weighted_change_result = weighted_change(filtered_df)
        # print(weighted_change_result)
        return get_top_spans("rcl_output/trace_rcl_results.csv"), weighted_change_result


def ranking(base_dir: Path, anomalous: Sequence, score: Sequence[pd.DataFrame]):
    scores = []
    for i, (anomalous, score) in enumerate(zip(anomalous, score)):
        if anomalous:
            scores.append(score)

    combined_scores = (
        pd.concat(scores, ignore_index=True)
        .groupby("ServiceName", as_index=False)["anomaly_score"]
        .sum()
    )
    combined_scores = combined_scores[combined_scores["anomaly_score"].abs() > 0]
    combined_scores = combined_scores.rename(columns={"anomaly_score": "CombinedScore"})
    combined_scores = combined_scores.sort_values(by="CombinedScore", ascending=False)

    # Add an index column starting from 1
    combined_scores = combined_scores.reset_index(drop=True)
    combined_scores["Index"] = combined_scores.index + 1

    combined_scores.to_csv(base_dir / "final_ranking.csv", index=False)
    # print(combined_scores)

    return combined_scores


def fault_type_infer(score_list, event_list):
    """基于故障库预测fault type"""
    fault_library_path = root_base_dir / "fault_library.csv"
    fault_library_df = pd.read_csv(fault_library_path)

    # Define sources and initialize scores and events list
    sources = ["log", "trace", "metric"]
    scores = [0, 0, 0]
    most_common_events = []
    # Extract anomaly scores and most common events
    for i, source in enumerate(sources):
        score_df = score_list[i]
        event = event_list[i]
        if not score_df.empty:
            scores[i] = score_df["anomaly_score"][0]
            # print(f"{source} score:", scores[i])
        if event:
            # print(event)
            if i == 0:
                most_common_events.append(event[0]["log_temp"])
            elif i == 1:
                most_common_events.append(event[0]["span_name"])
            elif i == 2:
                most_common_events.append(event[0]["metric_name"])

    print(scores)
    # print("most common events:", most_common_events)

    log_pre_score, trace_pre_score, metric_pre_score = scores
    print("log_pre_score:", log_pre_score)
    print("trace_pre_score:", trace_pre_score)
    print("metric_pre_score:", metric_pre_score)

    # Filter fault types based on score ranges
    filtered_df = fault_library_df[
        (
            fault_library_df["log_anomaly_score"]
            .apply(lambda x: x.split("-"))
            .apply(lambda rng: int(rng[0]) <= log_pre_score <= int(rng[1]))
        )
        & (
            fault_library_df["trace_anomaly_score"]
            .apply(lambda x: x.split("-"))
            .apply(lambda rng: int(rng[0]) <= trace_pre_score <= int(rng[1]))
        )
        & (
            fault_library_df["metric_anomaly_score"]
            .apply(lambda x: x.split("-"))
            .apply(lambda rng: int(rng[0]) <= metric_pre_score <= int(rng[1]))
        )
    ]

    # Further filter based on most common events
    def has_common_events(fault_events):
        return any(event in fault_events for event in most_common_events)

    filtered_df = filtered_df[
        filtered_df["most_common_event"].apply(lambda x: has_common_events(eval(x)))
    ]

    # If multiple fault types remain, prioritize by event match count and score proximity
    if len(filtered_df) > 1:
        filtered_df["event_match_count"] = filtered_df["most_common_event"].apply(
            lambda x: len(set(most_common_events) & set(eval(x)))
        )
        max_event_match = filtered_df["event_match_count"].max()
        filtered_df = filtered_df[filtered_df["event_match_count"] == max_event_match]

    if len(filtered_df) > 1:

        def score_diff(row):
            log_range = eval(row["log_anomaly_score"])
            trace_range = eval(row["trace_anomaly_score"])
            metric_range = eval(row["metric_anomaly_score"])
            return (
                abs(log_pre_score - sum(log_range) / 2)
                + abs(trace_pre_score - sum(trace_range) / 2)
                + abs(metric_pre_score - sum(metric_range) / 2)
            )

        filtered_df["score_proximity"] = filtered_df.apply(score_diff, axis=1)
        filtered_df = filtered_df.sort_values("score_proximity").iloc[:1]

    # Return recommended fault type
    if not filtered_df.empty:
        recommended_fault_type = filtered_df.iloc[0]["fault_type"]
    else:
        recommended_fault_type = "No matching fault type found"

    print("Recommended Fault Type:", recommended_fault_type)
    return recommended_fault_type


def get_multi_score(score_list):
    # Define sources and initialize scores list
    sources = ["log", "trace", "metric"]
    scores = [0, 0, 0]
    # Extract anomaly scores and most common events

    for i, source in enumerate(sources):
        score_df = score_list[i]
        if not score_df.empty:
            scores[i] = score_df["anomaly_score"][0]
            # print(f"{source} score:", scores[i])
    return scores


def fault_infer(prediction_data):
    """基于K-Means 聚类算法预测fault type"""
    # get score_list
    score_data = [item[2] for item in prediction_data]
    case_name_list = [item[0] for item in prediction_data]
    print(type(case_name_list))
    print(case_name_list)
    print(score_data)

    # score_data = [[0.43490573620211487, 0.56789, 31897.9947],[0.56789, 0.6789, 12345.6789]]

    df = pd.DataFrame(
        score_data,
        columns=["log_anomaly_score", "trace_anomaly_score", "metric_anomaly_score"],
    )

    # get features
    X = df[["log_anomaly_score", "trace_anomaly_score", "metric_anomaly_score"]]

    # K-Means algo
    kmeans = KMeans(n_clusters=3, random_state=42)
    clusters = kmeans.fit_predict(X)
    draw_cluster(X, clusters)
    # get cluster centers
    cluster_centers = kmeans.cluster_centers_

    # print cluster centers
    for i, center in enumerate(cluster_centers):
        print(
            f"Cluster {i}: log_anomaly_score={center[0]:.3f}, trace_anomaly_score={center[1]:.3f}, metric_anomaly_score={center[2]:.3f}"
        )

    # 计算每个点到其所属簇中心的距离
    distances = np.linalg.norm(X.values - cluster_centers[clusters], axis=1)
    df["distance_to_center"] = distances

    # 离群点检测（设置阈值，比如距离超过 90% 分位数的点为离群点）
    threshold = np.percentile(distances, 90)
    df["is_outlier"] = df["distance_to_center"] > threshold

    # 打印离群点
    outliers = df[df["is_outlier"]]
    print("\nOutliers:")
    print(outliers)

    # 假设有部分已知的标签
    known_labels = {2: "cpu-exhaustion", 1: "memory-exhaustion", 0: "pod-failure"}

    # 添加手动标签
    print(type(clusters))
    print(clusters)
    df["cluster"] = clusters
    df["case"] = case_name_list
    df["fault_type_prediction"] = df["cluster"].map(known_labels)

    # 检查每个簇的标签分布
    print(df.groupby("cluster")["fault_type_prediction"].value_counts())

    print(df)
    # 保存为 CSV 文件
    output_file = root_base_dir / "cluster_results.csv"
    df.to_csv(output_file, index=False)
    print(f"\nDataFrame saved to {output_file}")
    return df


def draw_cluster(X, clusters):
    import matplotlib.pyplot as plt
    from mpl_toolkits.mplot3d import Axes3D

    # 创建 3D 图形
    fig = plt.figure()
    ax = fig.add_subplot(111, projection="3d")

    # 绘制 3D 散点图
    ax.scatter(
        X.iloc[:, 0], X.iloc[:, 1], X.iloc[:, 2], c=clusters, cmap="viridis", s=50
    )

    # 设置标题和轴标签
    ax.set_title("Clustering Results (3D Data)")
    ax.set_xlabel("log_anomaly_score")
    ax.set_ylabel("trace_anomaly_score")
    ax.set_zlabel("metric_anomaly_score")

    plt.show()


def calculate_metrics_bak(prediction_data, type_prediction, gt_list):
    # 这块支持的是聚类算法返回类型的eval
    AC_at_1 = 0
    AC_at_3 = 0
    avg_at_5 = 0
    type_acc = 0
    total_cases = len(gt_list)

    for gt in gt_list:
        case_name = gt["case"]
        gt_service = gt["service"]
        gt_type = gt["chaos_type"]

        # fault-type evaluation
        for _, row in type_prediction.iterrows():
            fault_case = row["case"]
            fault_type_prediction = row["fault_type_prediction"]

            if fault_case == case_name:
                if fault_type_prediction == gt_type:
                    type_acc += 1

        # for fault_case, fault_type_prediction in type_prediction:
        #     if fault_case == case_name:
        #         if fault_type_prediction == gt_type:
        #             type_acc += 1

        # service evaluation
        for case, combined_ranking, _ in prediction_data:
            if case == case_name:
                # 找到 ground truth service 在 combined_ranking 中的 index
                ranking_df = combined_ranking
                if gt_service in ranking_df["ServiceName"].values:
                    service_index = ranking_df[ranking_df["ServiceName"] == gt_service][
                        "Index"
                    ].values[0]

                    # AC@1
                    if service_index == 1:
                        AC_at_1 += 1

                    # AC@3
                    if service_index <= 3:
                        AC_at_3 += 1

                    # Avg@5
                    if service_index <= 5:
                        avg_at_5 += (5 - service_index + 1) / 5.0

                break

    # ranking
    AC_at_1 /= total_cases
    AC_at_3 /= total_cases
    avg_at_5 /= total_cases
    type_acc /= total_cases

    return {"AC@1": AC_at_1, "AC@3": AC_at_3, "Avg@5": avg_at_5, "Type_AC": type_acc}


def calculate_metrics(prediction_data, gt_list):
    AC_at_1 = 0
    AC_at_3 = 0
    avg_at_5 = 0
    total_cases = len(gt_list)

    for gt in gt_list:
        case_name = gt["case"]
        gt_service = gt["service"]

        # 在 prediction_data 中找到对应的 case 和 combined_ranking
        for case, combined_ranking in prediction_data:
            if case == case_name:
                # 找到 ground truth service 在 combined_ranking 中的 index
                ranking_df = combined_ranking
                if gt_service in ranking_df["ServiceName"].values:
                    service_index = ranking_df[ranking_df["ServiceName"] == gt_service][
                        "Index"
                    ].values[0]

                    # AC@1
                    if service_index == 1:
                        AC_at_1 += 1

                    # AC@3
                    if service_index <= 3:
                        AC_at_3 += 1

                    # Avg@5
                    if service_index <= 5:
                        avg_at_5 += (5 - service_index + 1) / 5.0

                break

    # ranking
    AC_at_1 /= total_cases
    AC_at_3 /= total_cases
    avg_at_5 /= total_cases

    return {"AC@1": AC_at_1, "AC@3": AC_at_3, "Avg@5": avg_at_5}


async def process_case(base_dir, pool: Pool, file_type="log"):
    system_anomalous = False
    if file_type == "metric":
        parser = MetricParser(base_dir)
        data = await parser.parse(pool)
        detector = MetricDetector(base_dir)
        system_anomalous, anomalies = await detector.detect(data, pool)
        if system_anomalous:
            return system_anomalous, *metric_rcl(base_dir)
    else:
        parser: Parser = {"log": LogParser, "trace": TraceParser}[file_type](base_dir)
        data = await parser.parse(pool)
        detector: Detector = {"log": LogDetector, "trace": TraceDetector}[file_type](
            base_dir, data
        )
        system_anomalous, anomalies = await pool.apply(detector.detect)
        if system_anomalous:
            rcl_func = {"log": log_rcl, "trace": trace_rcl}[file_type]
            return system_anomalous, *await pool.apply(rcl_func, args=(base_dir,))
    return system_anomalous, [], pd.DataFrame()


async def evaluate(base_dir, pool):
    base_dir = Path(base_dir)
    case = base_dir.name
    # print(case)
    base_dir = base_dir.absolute()
    file_types = ["log", "trace", "metric"]
    result = await asyncio.gather(
        *(process_case(base_dir, pool, file_type) for file_type in file_types)
    )
    result = list(zip(*result))
    system_anomalous = any(result[0])
    # multi-modal anomaly score
    score_list = list(result[2])
    event_list = list(result[1])
    if system_anomalous:
        combined_events = {}
        for file_type, events in zip(file_types, result[1]):
            combined_events[f"{file_type}_events"] = events
        with open(base_dir / "events.toml", "w") as toml_file:
            toml.dump(combined_events, toml_file)
        # final service ranking prediciton, calculate by multi-modal service ranking
        combined_ranking = ranking(base_dir, result[0], result[2])
        scores = get_multi_score(score_list)
        return case, combined_ranking
        # print(f"Scores: {scores}, Type: {type(scores)}")
        # scores_file_path = root_base_dir / "scores.csv"
        # print(f"Saving to: {scores_file_path}")
        # with open(scores_file_path, "a", newline="") as csv_file:
        #     writer = csv.writer(csv_file)
        #     writer.writerow(scores)
        # return case, combined_ranking, scores

        ## fault type prediction based on fault library comparison
        # fault_type_prediction = fault_type_infer(score_list, event_list)
        # if  fault_type_prediction == "No matching fault type found":
        #     print("No matching:",case)
        # else:
        #     print("matched:",case)
        # return case, combined_ranking, fault_type_prediction
    else:
        print(f"No anomalies detected in case {base_dir}.")
        recall_count += 1
        return case, None, None


######## run one case ########
# root_base_dir = Path(r"E:\OneDrive - CUHK-Shenzhen\RCA_Dataset\test\ts")
# async def main():
#     fault_injection_file = root_base_dir / "fault_injection.toml"
#     data = toml.load(fault_injection_file)
#     gt_list = data["chaos_injection"]
#     # Define paths
#     base_dirs = [
#         # R"E:\OneDrive - CUHK-Shenzhen\RCA_Dataset\1021\ts\request-abort\ts-config-service",
#         # R"E:\OneDrive - CUHK-Shenzhen\RCA_Dataset\1024\ts\pod-failure\ts-consign-service",
#         R"E:\Project\Git\RCA_Dataset\test\ts\ts-food-service-1024-1816",
#         # R"E:\OneDrive - CUHK-Shenzhen\RCA_Dataset\test\ob\adservice-1027-0149",
#     ]
#     async with Pool() as pool:
#         prediction_data = await asyncio.gather(*(evaluate(base_dir, pool) for base_dir in base_dirs))


######## run all cases ########
async def main():
    # root_base_dir = Path(r"E:\OneDrive - CUHK-Shenzhen\RCA_Dataset\test\ts")
    root_base_dir = Path(r"E:\Project\Git\Metis-DataSet\test\ts")

    fault_injection_file = root_base_dir / "fault_injection.toml"
    data = toml.load(fault_injection_file)
    gt_list = data["chaos_injection"]

    case_dirs = [p for p in root_base_dir.iterdir() if p.is_dir()]

    childconcurrency = 20
    processes = os.cpu_count()
    queuecount = processes // 4
    async with Pool(
        processes=processes, childconcurrency=childconcurrency, queuecount=queuecount
    ) as pool:
        prediction_data = await asyncio.gather(
            *(evaluate(case_dir, pool) for case_dir in case_dirs)
        )

    service_evaluation_results = calculate_metrics(prediction_data, gt_list)
    print(service_evaluation_results)

    # type_prediction_df = fault_infer(prediction_data)
    # service_evaluation_results = calculate_metrics_bak(prediction_data, type_prediction_df, gt_list)
    # print(service_evaluation_results)


import time


@contextmanager
def timer():
    start_time = time.time()
    yield
    end_time = time.time()
    print(f"Execution time: {end_time - start_time:.4f} seconds")


if __name__ == "__main__":
    with timer():
        asyncio.run(main())
