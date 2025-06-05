from rcabench.openapi.api_client import ApiClient, Configuration
from rcabench.openapi import TraceApi, InjectionApi, TaskApi
from rcabench.model.trace import StreamEvent
from rcabench.rcabench import RCABenchSDK
import typer
import json
from rich.console import Console
from rich.json import JSON
from typing import Any

app = typer.Typer(pretty_exceptions_show_locals=False)
console = Console()


def print_json_highlighted(data: Any) -> None:
    if isinstance(data, str):
        json_str = data
    else:
        json_str = json.dumps(data, indent=2, ensure_ascii=False, default=str)

    json_obj = JSON(json_str)
    console.print(json_obj)


@app.command()
def ns_stat(
    host: str = "http://10.10.10.220:32080",
) -> None:
    """get namespace lock status"""
    configuration: Configuration = Configuration(host=host)

    with ApiClient(configuration=configuration) as client:
        api: InjectionApi = InjectionApi(client)
        resp = api.api_v1_injections_ns_status_get()
        print_json_highlighted(resp.data)


@app.command()
def task_queue(
    host: str = "http://10.10.10.220:32080",
) -> None:
    """get task queue status"""
    configuration: Configuration = Configuration(host=host)

    with ApiClient(configuration=configuration) as client:
        api: TaskApi = TaskApi(client)
        resp = api.api_v1_tasks_queue_get()
        data = resp.data
        assert data is not None, "No data returned from task queue"
        if hasattr(data, "model_dump"):
            print_json_highlighted(data.model_dump())
        else:
            print_json_highlighted(data)


@app.command()
def trace_stat(
    host: str = "http://10.10.10.220:32080",
    lookback: str = "1h",
) -> None:
    """get trace analysis information"""
    configuration: Configuration = Configuration(host=host)

    with ApiClient(configuration=configuration) as client:
        api: TraceApi = TraceApi(client)
        resp = api.api_v1_traces_analyze_get(lookback=lookback)
        data = resp.data
        assert data is not None, "No data returned from task queue"
        print_json_highlighted(data)


@app.command()
def trace_stream(
    trace_id: str,
    host: str = "http://10.10.10.220:32080",
    timeout: float = 600000,
) -> None:
    """get one trace stream event by trace_id"""
    sdk = RCABenchSDK(host)
    for e in sdk.trace.stream_trace_events(trace_id, timeout=timeout):
        assert isinstance(e, StreamEvent)
        assert isinstance(e.payload, str)
        e.payload = json.loads(e.payload)
        print_json_highlighted(e.model_dump())


if __name__ == "__main__":
    app()
