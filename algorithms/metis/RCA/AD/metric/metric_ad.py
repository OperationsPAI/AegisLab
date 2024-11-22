import csv
import os
from pathlib import Path
from typing import Sequence, Tuple, Union

import numpy as np
import pandas as pd
from aiomultiprocess import Pool
from sklearn.decomposition import PCA
from sklearn.preprocessing import MinMaxScaler, StandardScaler

from ..abstract.detector import Detector


class MetricDetector(Detector):
    def __init__(self, base_dir: Path):
        super().__init__(base_dir)
        self.normal_file_dir = base_dir / "normal/processed_metrics/"
        self.detect_file_dir = base_dir / "abnormal/processed_metrics/"
        self.output_dir = self.base_dir / "metric_ad_output"
        self.system_anomalous = False
        self.service_list = []
        self.output_dir.mkdir(parents=True, exist_ok=True)

    async def detect(
        self, data: Sequence[Sequence[Tuple[pd.DataFrame, str]]], pool: Pool
    ):
        """wrapper function to detect anomalies."""
        await self._detect_anomalies(data, pool)
        if self.system_anomalous:
            service_list_file = os.path.join(self.output_dir, "service_list.csv")
            with open(service_list_file, mode="w", newline="") as file:
                writer = csv.writer(file)
                writer.writerow(["ServiceName", "AnomalyScore", "TimeRanges"])
                writer.writerows(self.service_list)
        # print(f"Results saved to {service_list_file}")
        return self.system_anomalous, self.service_list

    async def _detect_anomalies(
        self, data: Sequence[Sequence[Tuple[pd.DataFrame, str]]], pool: Pool
    ):
        normal_groups, abnormal_groups = (
            {name: df for df, name in group} for group in data
        )
        all_services = normal_groups.keys() & abnormal_groups.keys()
        self.service_list = await pool.starmap(
            self._process_func,
            [
                (normal_groups[service], abnormal_groups[service], service)
                for service in all_services
            ],
        )
        self.service_list = [x for x in self.service_list if x]
        self.service_list.sort(key=lambda x: x[1], reverse=True)

        if len(self.service_list) > 0:
            self.system_anomalous = True

    async def _process_func(self, normal_df, abnormal_df, service_name):
        normal_timestamps, normal_features = MetricDetector.get_time_and_features(
            normal_df
        )
        detect_timestamps, detect_features = MetricDetector.get_time_and_features(
            abnormal_df
        )
        # Initialize PCA anomaly detector
        pca_detector = PCAError(pca_dim="auto", svd_solver="full")
        pca_detector.fit(normal_features, verbose=False)

        # Detect anomalies over the entire period
        window_anomalies, complete_anomaly_scores = pca_detector.detect(detect_features)

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
            anomaly_score = all_anomaly_scores.max()

            # print(service_name,":",anomaly_score)

            # If AnomalyScore is greater than or equal to 0.001, add to service list
            if anomaly_score >= 0.001:
                return [
                    service_name,
                    round(anomaly_score, 3),
                    f"{earliest_anomaly_str}-{latest_anomaly_str}",
                ]
        else:
            # print(f"No anomalies detected in {file_name}.")
            return []

    @staticmethod
    def get_time_and_features(df: pd.DataFrame):
        df.index = pd.to_datetime(df.index)
        # 如果 'receive_bytes' 列不全为 0，则进行替换 0 为 NaN 和插值
        # if not (df['receive_bytes'] == 0).all():
        #     df['receive_bytes'] = df['receive_bytes'].replace(0, np.nan)
        #     df['receive_bytes'] = df['receive_bytes'].interpolate(method='linear')
        # 如果 'transmit_bytes' 列不全为 0，则进行替换 0 为 NaN 和插值
        # if not (df['transmit_bytes'] == 0).all():
        #     df['transmit_bytes'] = df['transmit_bytes'].replace(0, np.nan)
        #     df['transmit_bytes'] = df['transmit_bytes'].interpolate(method='linear')
        df.replace(np.nan, 0, inplace=True)
        timestamps = df.index
        features = df.values
        return timestamps, features


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
