from rca import start_rca
import os
from pathlib import Path
import inspect
import asyncio


def run_function(func, *args, **kwargs):
    # 检查函数是否是一个异步函数
    if inspect.iscoroutinefunction(func):
        # 如果是异步函数，用 asyncio.run 调用
        asyncio.run(func(*args, **kwargs))
    else:
        # 如果是普通函数，直接调用
        func(*args, **kwargs)


if __name__ == "__main__":
    workspace = os.environ["WORKSPACE"]
    if workspace == "":
        print("WARN: the WORKSPACE environ is not defined, using default '/app'")
        workspace = "/app"
    base_path = Path(workspace)

    normal_time_range = (
        [
            (
                int(os.environ["NORMAL_START"]) + 8 * 3600,
                int(os.environ["NORMAL_END"]) + 8 * 3600,
            )
        ]
        if os.environ.get("NORMAL_START") and os.environ.get("NORMAL_END")
        else []
    )

    abnormal_time_range = (
        [
            (
                int(os.environ["ABNORMAL_START"]) + 8 * 3600,
                int(os.environ["ABNORMAL_END"]) + 8 * 3600,
            )
        ]
        if os.environ.get("ABNORMAL_START") and os.environ.get("ABNORMAL_END")
        else []
    )

    run_function(
        start_rca,
        {
            "log_file": base_path / "input" / "logs.csv",
            "trace_file": base_path / "input" / "traces.csv",
            "trace_id_ts_file": base_path / "input" / "trace_id_ts.csv",
            "metric_file": base_path / "input" / "metrics.csv",
            "metric_sum_file": base_path / "input" / "metric_sum.csv",
            "metric_summary_file": base_path / "input" / "metrics_summary.csv",
            "metric_histogram_file": base_path / "input" / "metrics_histogram.csv",
            "event_file": base_path / "input" / "events.csv",
            "profiling_file": base_path / "input" / "profilings.csv",
            "normal_time_range": normal_time_range,
            "abnormal_time_range": abnormal_time_range,
        },
    )
