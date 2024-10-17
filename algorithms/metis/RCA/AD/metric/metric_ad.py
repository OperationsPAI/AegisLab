import numpy as np
import pandas as pd
from sklearn.decomposition import PCA
from sklearn.preprocessing import StandardScaler, MinMaxScaler
import matplotlib.pyplot as plt
from typing import Union
import os
from datetime import timedelta
from contextlib import contextmanager


def check_timeseries_shape(timeseries: np.ndarray):
    if timeseries.ndim != 2:
        raise ValueError(
            f"Expected a 2D array with shape (n_samples, n_features), "
            f"but got array with shape {timeseries.shape}"
        )


class TADMethodEstimator:
    def fit(
        self, X: np.ndarray, univariate: bool = False, verbose: bool = False
    ) -> None:
        pass

    def transform(self, X: np.ndarray) -> np.ndarray:
        pass


class PCAError(TADMethodEstimator):
    def __init__(
        self, pca_dim: Union[int, str] = "auto", svd_solver: str = "full"
    ) -> None:
        self.pca_dim = pca_dim
        self.svd_solver = svd_solver
        self.pca = None
        self.scaler = StandardScaler()
        # self.scaler = MinMaxScaler()

    def fit(
        self, X: np.ndarray, univariate: bool = False, verbose: bool = False
    ) -> None:
        check_timeseries_shape(X)

        X_scaled = self.scaler.fit_transform(X)
        n_features = X_scaled.shape[1]

        if self.pca_dim == "auto":
            n_components = min(n_features, max(1, n_features // 2))
            if verbose:
                print(f"Auto-selected number of PCA components: {n_components}")
        else:
            n_components = self.pca_dim

        self.pca = PCA(n_components=n_components, svd_solver=self.svd_solver)
        self.pca.fit(X_scaled)

        reconstruction_error = np.abs(
            X_scaled - self.pca.inverse_transform(self.pca.transform(X_scaled))
        )
        self.threshold = np.percentile(np.mean(reconstruction_error, axis=1), 95)
        if verbose:
            print(f"Anomaly detection threshold set at: {self.threshold}")

    def transform(self, X: np.ndarray) -> np.ndarray:
        check_timeseries_shape(X)

        X_scaled = self.scaler.transform(X)
        reconstructed = self.pca.inverse_transform(self.pca.transform(X_scaled))
        reconstruction_error = np.abs(X_scaled - reconstructed)
        anomaly_scores = np.mean(reconstruction_error, axis=1)

        return anomaly_scores

    def detect(self, X: np.ndarray) -> np.ndarray:
        anomaly_scores = self.transform(X)
        anomalies = anomaly_scores > self.threshold
        return anomalies, anomaly_scores


def detect_continuous_anomalies(
    timestamps, anomalies, anomaly_scores, min_duration_seconds=30
):
    # Identifying continuous anomaly windows that are longer than min_duration_seconds
    continuous_windows = []
    current_window = []
    current_scores = []

    for i in range(len(timestamps)):
        if anomalies[i]:
            current_window.append(timestamps[i])
            current_scores.append(anomaly_scores[i])
        else:
            if len(current_window) > 0:
                # Check if the duration of the window is greater than min_duration_seconds
                if (
                    current_window[-1] - current_window[0]
                ).total_seconds() >= min_duration_seconds:
                    continuous_windows.append((current_window, current_scores))
                current_window = []
                current_scores = []

    if (
        len(current_window) > 0
        and (current_window[-1] - current_window[0]).total_seconds()
        >= min_duration_seconds
    ):
        continuous_windows.append((current_window, current_scores))

    return continuous_windows


def PCA_detection(normal_data_addr, detect_addr, output_file_path):
    service_list = []

    # metric_fig_dir = os.path.join(detect_addr, 'score_fig')
    # os.makedirs(metric_fig_dir, exist_ok=True)

    for file_name in os.listdir(detect_addr):
        if file_name.endswith(".csv"):
            normal_file_path = os.path.join(normal_data_addr, file_name)
            detect_file_path = os.path.join(detect_addr, file_name)

            if not os.path.exists(normal_file_path):
                print(f"Normal data file {file_name} not found.")
                continue

            normal_data_df = pd.read_csv(normal_file_path)
            detect_data_df = pd.read_csv(detect_file_path)

            if normal_data_df.empty or len(normal_data_df) == 0:
                print(
                    f"Normal data file {file_name} is empty or contains only headers."
                )
                continue
            if detect_data_df.empty or len(detect_data_df) == 0:
                print(
                    f"Detect data file {file_name} is empty or contains only headers."
                )
                continue

            normal_data_df["TimeUnix"] = pd.to_datetime(normal_data_df["TimeUnix"])
            detect_data_df["TimeUnix"] = pd.to_datetime(detect_data_df["TimeUnix"])

            # print(normal_data_df)

            # # 如果 'receive_bytes' 列不全为 0，则进行替换 0 为 NaN 和插值
            # if not (normal_data_df['receive_bytes'] == 0).all():
            #     normal_data_df['receive_bytes'].replace(0, np.nan, inplace=True)
            #     normal_data_df['receive_bytes'] = normal_data_df['receive_bytes'].interpolate(method='linear')

            # if not (detect_data_df['receive_bytes'] == 0).all():
            #     detect_data_df['receive_bytes'].replace(0, np.nan, inplace=True)
            #     detect_data_df['receive_bytes'] = detect_data_df['receive_bytes'].interpolate(method='linear')

            # # 如果 'transmit_bytes' 列不全为 0，则进行替换 0 为 NaN 和插值
            # if not (normal_data_df['transmit_bytes'] == 0).all():
            #     normal_data_df['transmit_bytes'].replace(0, np.nan, inplace=True)
            #     normal_data_df['transmit_bytes'] = normal_data_df['transmit_bytes'].interpolate(method='linear')

            # if not (detect_data_df['transmit_bytes'] == 0).all():
            #     detect_data_df['transmit_bytes'].replace(0, np.nan, inplace=True)
            #     detect_data_df['transmit_bytes'] = detect_data_df['transmit_bytes'].interpolate(method='linear')

            # # 替换 NaN 为 0
            normal_data_df.replace(np.nan, 0, inplace=True)
            detect_data_df.replace(np.nan, 0, inplace=True)

            print("file_name:", file_name)
            normal_timestamps = normal_data_df["TimeUnix"]
            normal_features = normal_data_df.drop(columns=["TimeUnix"]).values

            detect_timestamps = detect_data_df["TimeUnix"]
            detect_features = detect_data_df.drop(columns=["TimeUnix"]).values

            # print(normal_data_df)
            # Initialize PCA anomaly detector
            pca_detector = PCAError(pca_dim="auto", svd_solver="full")
            pca_detector.fit(normal_features, verbose=True)

            # Detect anomalies over the entire period
            window_anomalies, complete_anomaly_scores = pca_detector.detect(
                detect_features
            )

            continuous_windows = detect_continuous_anomalies(
                detect_timestamps, window_anomalies, complete_anomaly_scores
            )

            # Check if there are any valid continuous windows
            if len(continuous_windows) > 0:
                earliest_anomaly = continuous_windows[0][0][0]
                latest_anomaly = continuous_windows[-1][0][-1]

                # Use the earliest and latest timestamps as the time range
                earliest_anomaly_str = pd.Timestamp(earliest_anomaly).strftime(
                    "%Y-%m-%dT%H:%M:%S"
                )
                latest_anomaly_str = pd.Timestamp(latest_anomaly).strftime(
                    "%Y-%m-%dT%H:%M:%S"
                )

                # Average anomaly score for all anomaly points
                all_anomaly_scores = np.concatenate(
                    [scores for _, scores in continuous_windows]
                )
                average_anomaly_score = all_anomaly_scores.max()

                # If AnomalyScore is greater than or equal to 0.1, add to service list
                if average_anomaly_score >= 0.1:
                    service_name = file_name.replace(".csv", "")
                    service_list.append(
                        [
                            service_name,
                            round(average_anomaly_score, 3),
                            f"{earliest_anomaly_str}-{latest_anomaly_str}",
                        ]
                    )
            else:
                print(f"No anomalies detected in {file_name}.")

            # Plot and save the graph
            # plt.figure(figsize=(15, 6))
            # plt.plot(detect_timestamps, complete_anomaly_scores, label='Anomaly Score')
            # plt.hlines(pca_detector.threshold, xmin=detect_timestamps.min(), xmax=detect_timestamps.max(), colors='r', linestyles='dashed', label='Threshold')
            # plt.xlabel('TimeUnix')
            # plt.ylabel('Anomaly Score')
            # plt.title(f"{file_name} - Reconstruction Error Over Full Time Period")
            # plt.legend()
            # fig_save_path = os.path.join(metric_fig_dir, f"{file_name}.png")
            # plt.savefig(fig_save_path)
            # plt.close()

    service_list.sort(key=lambda x: x[1], reverse=True)

    # Output service_list.csv
    os.makedirs(output_file_path, exist_ok=True)
    service_list_file = os.path.join(output_file_path, "service_list.csv")

    service_list_df = pd.DataFrame(
        service_list, columns=["ServiceName", "AnomalyScore", "TimeRanges"]
    )
    service_list_df.to_csv(service_list_file, index=False)
    # print(f"Results saved to {service_list_file}")

    # 如果 service_list 不为空
    if len(service_list) > 0:
        return 1
        # return service_list_df['ServiceName']


# main
def metric_ad(file_path):
    normal_file_path = os.path.join(file_path, "normal/processed_metrics/")
    detect_file_path = os.path.join(file_path, "abnormal/processed_metrics/")
    output_file_path = os.path.join(file_path, "metric_ad_output")

    abnormal_service_name = PCA_detection(
        normal_file_path, detect_file_path, output_file_path
    )
    return abnormal_service_name


# Example usage:
# file_path = R'E:\OneDrive - CUHK-Shenzhen\RCA_Dataset\test_new_datasets\onlineboutique\cpu\checkoutservice-1011-1441'
# metric_ad(file_path)
