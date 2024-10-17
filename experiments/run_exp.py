# This file will be moved into image workspace, and start the evaluation by running the script

from rca import start_rca
import os
from pathlib import Path


if __name__ == "__main__":
    workspace = os.environ["WORKSPACE"]
    if workspace == "":
        print("WARN: the WORKSPACE environ is not defined, using default '/app'")
        workspace = "/app"
    base_path = Path(workspace)

    normal_time_range = (
        [(int(os.environ["NORMAL_START"]), int(os.environ["NORMAL_END"]))]
        if os.environ.get("NORMAL_START") and os.environ.get("NORMAL_END")
        else []
    )

    abnormal_time_range = (
        [(int(os.environ["ABNORMAL_START"]), int(os.environ["ABNORMAL_END"]))]
        if os.environ.get("ABNORMAL_START") and os.environ.get("ABNORMAL_END")
        else []
    )

    start_rca(
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
        }
    )
