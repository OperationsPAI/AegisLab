from rca import start_rca
import os
from pathlib import Path
import inspect
import asyncio
import tempfile
import subprocess


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
    input_path = os.environ["INPUT_PATH"]
    if input_path == "":
        print("WARN: the INPUT_PATH environ is not defined")
        exit(1)

    if os.path.exists(f"{input_path}/data.tar.gz"):
        tempdir = tempfile.TemporaryDirectory()
        subprocess.run(
            ["tar", "-xzf", f"{input_path}/data.tar.gz", "-C", tempdir.name],
            check=True,
        )
        input_path = tempdir.name

    key = "OUTPUT_PATH"
    output_path_value = os.getenv(key)
    if not output_path_value:
        print("WARN: the OUTPUT_PATH environ is not defined")
        output_path = "/app/output"
        os.environ[key] = output_path

    base_path = Path(workspace)
    input_path = Path(input_path)

    run_function(
        start_rca,
        {
            "normal_log_file": input_path / "normal_logs.csv",
            "normal_trace_file": input_path / "normal_traces.csv",
            "normal_trace_id_ts_file": input_path / "normal_trace_id_ts.csv",
            "normal_metric_file": input_path / "normal_metrics.csv",
            "normal_metric_sum_file": input_path / "normal_metric_sum.csv",
            "normal_metric_summary_file": input_path / "normal_metrics_summary.csv",
            "normal_metric_histogram_file": input_path / "normal_metrics_histogram.csv",
            "normal_event_file": input_path / "normal_events.csv",
            "normal_profiling_file": input_path / "normal_profilings.csv",
            "abnormal_log_file": input_path / "abnormal_logs.csv",
            "abnormal_trace_file": input_path / "abnormal_traces.csv",
            "abnormal_trace_id_ts_file": input_path / "abnormal_trace_id_ts.csv",
            "abnormal_metric_file": input_path / "abnormal_metrics.csv",
            "abnormal_metric_sum_file": input_path / "abnormal_metric_sum.csv",
            "abnormal_metric_summary_file": input_path / "abnormal_metrics_summary.csv",
            "abnormal_metric_histogram_file": input_path
            / "abnormal_metrics_histogram.csv",
            "abnormal_event_file": input_path / "abnormal_events.csv",
            "abnormal_profiling_file": input_path / "abnormal_profilings.csv",
        },
    )
