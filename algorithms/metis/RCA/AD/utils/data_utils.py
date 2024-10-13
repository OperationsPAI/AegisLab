from itertools import starmap

import pandas as pd


async def calculate_rates(data, normal=True):
    """wrapper function to calculate error and warn rates."""
    return pd.concat(list(starmap(calculate_normal if normal else calculate_abnormal, data))).round(3)


def process_window(window_slice: pd.DataFrame, pod, current_start):
    """Process a time window of log data."""

    proportion = window_slice["log_level"].value_counts(normalize=True)

    window_result = pd.DataFrame(
        {
            "ServiceName": [pod],
            "ErrorRate": [proportion.get("ERROR", 0)],
            "WarnRate": [proportion.get("WARN", 0)],
            "StartTime": [current_start],
        },
        index=[0],  # let pandas know it's a single row
    )
    return window_result


def calculate_abnormal(log_df: pd.DataFrame, pod):
    """Process abnormal log data and calculate error and warn rates within a rolling window."""
    log_df.set_index("Timestamp", inplace=True)

    start_time = log_df.index.min()
    end_time = log_df.index.max()

    window_size = pd.Timedelta(minutes=2)
    step_size = pd.Timedelta(minutes=1)

    args = []
    for current_start in pd.date_range(start=start_time, end=end_time, freq=step_size):

        current_end = current_start + window_size

        if current_end > end_time:
            break
        window_slice = log_df.loc[current_start:current_end]

        args.append((window_slice, pod, current_start))
    try:
        results = pd.concat(list(starmap(process_window, args)))
    except ValueError:
        results = pd.DataFrame()

    return results


def compute_statistics(df: pd.DataFrame):
    """Compute mean and standard deviation of error and warn rates."""
    return pd.DataFrame(
        {
            "ServiceName": [df["ServiceName"].iloc[0]],
            "ErrorRateMean": [df["ErrorRate"].mean()],
            "ErrorRateStd": [df["ErrorRate"].std()],
            "WarnRateMean": [df["WarnRate"].mean()],
            "WarnRateStd": [df["WarnRate"].std()],
        },
        index=[0],
    )


def calculate_normal(log_df: pd.DataFrame, pod):
    """Calculate error and warn rates for normal log data."""

    log_df.set_index("Timestamp", inplace=True)

    start_time = log_df.index.min()
    end_time = log_df.index.max()

    window_size = pd.Timedelta(seconds=10)
    args = []

    for current_start in pd.date_range(start=start_time, end=end_time, freq=window_size):
        current_end = current_start + window_size

        if current_end > end_time:
            break

        window_slice = log_df.loc[current_start:current_end]
        args.append((window_slice, pod, current_start))

    results = pd.concat(list(starmap(process_window, args)))

    return compute_statistics(results)
