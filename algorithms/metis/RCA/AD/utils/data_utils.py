from itertools import starmap

import pandas as pd

from ..abstract import Calculator


class TraceDurationCalculator(Calculator):
    @staticmethod
    async def calculate(data, normal):
        """wrapper function to calculate mean and standard deviation of service spans."""
        return (
            TraceDurationCalculator._calculate_normal(data)
            if normal
            else TraceDurationCalculator._calculate_abnormal(data)
        )

    @staticmethod
    def _process_window(window_slice: pd.DataFrame, current_start):
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

    @staticmethod
    def _calculate_abnormal(trace_df: pd.DataFrame):
        """Process abnormal trace data and calculate mean and standard deviation of service spans within a rolling window."""
        trace_df["Timestamp"] = pd.to_datetime(trace_df["Timestamp"])
        trace_df = trace_df.set_index("Timestamp").sort_values(by="Timestamp")

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

        results = pd.concat(list(starmap(TraceDurationCalculator._process_window, args)))

        return results[
            ["ServiceName", "SpanName", "MeanDuration", "StdDuration", "ParentServiceName", "TraceId", "StartTime"]
        ]

    @staticmethod
    def _calculate_normal(trace_df: pd.DataFrame):
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


class LogLevelCalculator(Calculator):
    @staticmethod
    async def calculate(data, normal=True):
        """wrapper function to calculate error and warn rates."""
        calculator = LogLevelCalculator._calculate_normal if normal else LogLevelCalculator._calculate_abnormal
        return pd.concat(list(starmap(calculator, data))).round(3)

    @staticmethod
    def _process_window(window_slice: pd.DataFrame, pod, current_start):
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

    @staticmethod
    def _calculate_abnormal(log_df: pd.DataFrame, pod):
        """Process abnormal log data and calculate error and warn rates within a rolling window."""
        log_df["Timestamp"] = pd.to_datetime(log_df["Timestamp"])
        log_df.set_index("Timestamp", inplace=True)

        start_time = log_df.index.min()
        end_time = log_df.index.max()

        window_size = pd.Timedelta(minutes=2)
        step_size = pd.Timedelta(minutes=1)

        args = []
        for current_start in pd.date_range(start=start_time, end=end_time, freq=step_size):

            current_end = current_start + window_size
            window_slice = log_df.loc[current_start:current_end]

            args.append((window_slice, pod, current_start))
        try:
            results = pd.concat(list(starmap(LogLevelCalculator._process_window, args)))
        except ValueError:
            results = pd.DataFrame()

        return results

    @staticmethod
    def _compute_statistics(df: pd.DataFrame):
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

    @staticmethod
    def _calculate_normal(log_df: pd.DataFrame, pod):
        """Calculate error and warn rates for normal log data."""
        log_df["Timestamp"] = pd.to_datetime(log_df["Timestamp"])
        log_df.set_index("Timestamp", inplace=True)

        start_time = log_df.index.min()
        end_time = log_df.index.max()

        window_size = pd.Timedelta(seconds=30)
        args = []

        for current_start in pd.date_range(start=start_time, end=end_time, freq=window_size):
            current_end = current_start + window_size

            window_slice = log_df.loc[current_start:current_end]
            args.append((window_slice, pod, current_start))

        try:
            results = pd.concat(list(starmap(LogLevelCalculator._process_window, args)))
        except ValueError:
            results = pd.DataFrame()

        return LogLevelCalculator._compute_statistics(results)
