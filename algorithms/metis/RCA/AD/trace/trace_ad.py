import asyncio
from pathlib import Path
from typing import Sequence

import pandas as pd

from ..abstract.detector import Detector


class TraceDetector(Detector):
    def __init__(self, base_dir: Path, data: Sequence[pd.DataFrame]):
        super().__init__(base_dir)
        self.normal_df = data[0]
        self.abnormal_df = data[1]
        self.output_dir = self.base_dir / "trace_ad_output"
        self.output_file = self.output_dir / "service_list.csv"
        self.system_anomalous = False
        self.anomalies = {}
        self.output_dir.mkdir(parents=True, exist_ok=True)

    async def detect(self):
        """Detect anomalies in service spans based on duration mean using k-σ rule."""

        self._detect_anomalies(k=3)
        if self.system_anomalous:
            await self._save_anomalies_to_csv()
        else:
            print(f"No trace-related anomalies detected in case {'_'.join(self.base_dir.parts[-2:])}.")
        return self.system_anomalous, self.anomalies

    def _detect_anomalies(self, k=3):
        """Detect anomalies in service spans based on duration mean using k-σ rule."""

        for start_time, group in self.abnormal_df.groupby("StartTime"):
            start_time = pd.to_datetime(start_time)
            for _, row in group.iterrows():
                service_name = row["ServiceName"]
                span_name = row["SpanName"]
                mean_duration = row["MeanDuration"]

                normal_stats = self.normal_df[
                    (self.normal_df["ServiceName"] == service_name) & (self.normal_df["SpanName"] == span_name)
                ]

                if not normal_stats.empty:
                    normal_mean = normal_stats["MeanDuration"].iloc[0]
                    normal_std = normal_stats["StdDuration"].iloc[0]

                    if mean_duration > normal_mean + k * normal_std:
                        if service_name not in self.anomalies:
                            self.anomalies[service_name] = {
                                "TimeRanges": (start_time, start_time + pd.Timedelta(minutes=2)),
                            }
                        else:
                            self.anomalies[service_name]["TimeRanges"] = (
                                min(self.anomalies[service_name]["TimeRanges"][0], start_time),
                                max(
                                    self.anomalies[service_name]["TimeRanges"][1], start_time + pd.Timedelta(minutes=2)
                                ),
                            )

        self.system_anomalous = len(self.anomalies) > 0

    async def _save_anomalies_to_csv(self):
        """Save the anomalies dictionary to a CSV file."""
        data = []
        for service, details in self.anomalies.items():
            time_range = details["TimeRanges"]
            start_time = time_range[0].strftime("%Y-%m-%d %H:%M:%S")
            end_time = time_range[1].strftime("%Y-%m-%d %H:%M:%S")
            data.append({"ServiceName": service, "StartTime": start_time, "EndTime": end_time})

        df = pd.DataFrame(data)
        await asyncio.to_thread(df.to_csv, self.output_file, index=False, mode="w")

        # print(f"Anomalies saved to {self.output_file}")
