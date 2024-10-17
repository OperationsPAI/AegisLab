import asyncio
from pathlib import Path

import pandas as pd


async def detect(base_dir, data):
    """Detect anomalies in service spans based on duration mean using k-σ rule."""
    base_dir = Path(base_dir)
    output_dir = base_dir / "trace_ad_output"
    case_name = "_".join(
        base_dir.parts[-2:]
    )  # Extract the case name from the base directory

    system_anomalous, anomalies = detect_anomalies(*data, k=3)
    if system_anomalous:
        await save_anomalies_to_csv(anomalies, output_dir / "service_list.csv")
    else:
        print(f"No trace-related anomalies detected in case {case_name}.")
    return system_anomalous, anomalies


def detect_anomalies(normal_df, abnormal_df, k=3):
    """Detect anomalies in service spans based on duration mean using k-σ rule."""

    anomalies = {}

    for start_time, group in abnormal_df.groupby("StartTime"):
        start_time = pd.to_datetime(start_time)
        for _, row in group.iterrows():
            service_name = row["ServiceName"]
            span_name = row["SpanName"]
            mean_duration = row["MeanDuration"]

            normal_stats = normal_df[
                (normal_df["ServiceName"] == service_name)
                & (normal_df["SpanName"] == span_name)
            ]

            if not normal_stats.empty:
                normal_mean = normal_stats["MeanDuration"].iloc[0]
                normal_std = normal_stats["StdDuration"].iloc[0]

                if mean_duration > normal_mean + k * normal_std:
                    if service_name not in anomalies:
                        anomalies[service_name] = {
                            "TimeRanges": (
                                start_time,
                                start_time + pd.Timedelta(minutes=2),
                            ),
                        }
                    else:
                        anomalies[service_name]["TimeRanges"] = (
                            min(anomalies[service_name]["TimeRanges"][0], start_time),
                            max(
                                anomalies[service_name]["TimeRanges"][1],
                                start_time + pd.Timedelta(minutes=2),
                            ),
                        )

    system_anomalous = len(anomalies) > 0
    return system_anomalous, anomalies


async def save_anomalies_to_csv(anomalies, output_file):
    """Save the anomalies dictionary to a CSV file."""
    data = []
    for service, details in anomalies.items():
        time_range = details["TimeRanges"]
        start_time = time_range[0].strftime("%Y-%m-%d %H:%M:%S")
        end_time = time_range[1].strftime("%Y-%m-%d %H:%M:%S")
        data.append(
            {"ServiceName": service, "StartTime": start_time, "EndTime": end_time}
        )

    df = pd.DataFrame(data)
    await asyncio.to_thread(df.to_csv, output_file, index=False, mode="w")

    # print(f"Anomalies saved to {output_file}")
