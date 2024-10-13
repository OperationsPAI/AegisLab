import asyncio

import pandas as pd
from aiomultiprocess import Pool


async def process_window(window_slice: pd.DataFrame, current_start):
    """Process a time window of trace data."""
    span_groups = window_slice.groupby(["ServiceName", "SpanName"])
    mean_duration = span_groups["Duration"].mean()
    std_duration = span_groups["Duration"].std()
    parent_service_name = span_groups["ParentServiceName"].first()
    trace_id = span_groups["TraceId"].first()

    window_result = pd.DataFrame(
        {
            "MeanDuration": mean_duration,
            "StdDuration": std_duration,
            "ParentServiceName": parent_service_name,
            "TraceId": trace_id,
            "StartTime": current_start,
        }
    )

    window_result["ServiceName"] = window_result.index.get_level_values("ServiceName")
    window_result["SpanName"] = window_result.index.get_level_values("SpanName")

    return window_result


async def calculate_abnormal(trace_df: pd.DataFrame, pool: Pool):
    """Process abnormal trace data and calculate mean and standard deviation of service spans within a rolling window."""
    trace_df["Timestamp"] = pd.to_datetime(trace_df["Timestamp"])
    trace_df = trace_df.set_index("Timestamp").sort_values(by="Timestamp")
    rolling_data = pd.DataFrame()

    start_time = trace_df.index.min()
    end_time = trace_df.index.max()

    window_size = pd.Timedelta(minutes=2)
    step_size = pd.Timedelta(minutes=1)

    args = []
    for current_start in pd.date_range(start=start_time, end=end_time, freq=step_size):
        current_end = current_start + window_size
        if current_end > end_time:
            break

        window_slice = trace_df.loc[current_start:current_end]
        args.append([window_slice, current_start])

    results = await pool.starmap(process_window, args)

    rolling_data = pd.concat(results)

    return rolling_data[
        ["ServiceName", "SpanName", "MeanDuration", "StdDuration", "ParentServiceName", "TraceId", "StartTime"]
    ]


def calculate_normal(trace_df: pd.DataFrame):
    """Process normal trace data and calculate mean and standard deviation of service spans."""
    span_groups = trace_df.groupby(["ServiceName", "SpanName"])
    mean_duration = span_groups["Duration"].mean()
    std_duration = span_groups["Duration"].std()
    parent_service_name = span_groups["ParentServiceName"].first()
    trace_id = span_groups["TraceId"].first()
    data = []
    for (service, span_name), mean in mean_duration.items():
        data.append(
            {
                "ServiceName": service,
                "SpanName": span_name,
                "MeanDuration": mean,
                "StdDuration": std_duration[(service, span_name)],
                "ParentServiceName": parent_service_name[(service, span_name)],
                "TraceId": trace_id[(service, span_name)],
            }
        )
    return pd.DataFrame(data)
