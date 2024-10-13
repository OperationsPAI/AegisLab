import asyncio
import os
from datetime import datetime, timedelta
from pathlib import Path

import aiofiles
import clickhouse_connect
import toml
from clickhouse_connect.driver.asyncclient import AsyncClient

# Read credentials from environment variables
CLICKHOUSE_USER = os.getenv("CLICKHOUSE_USER", "default")
CLICKHOUSE_PASSWORD = os.getenv("CLICKHOUSE_PASSWORD", "nn")


def generate_query(query_type, start_time, end_time, namespace):
    """Generate the query for logs, metrics, and traces."""
    if query_type == "log":
        return Rf"""
            SELECT 
                Timestamp, 
                ResourceAttributes['k8s.namespace.uid'] AS k8s_namespace_uid,
                ResourceAttributes['k8s.namespace.name'] AS k8s_namespace_name,
                ResourceAttributes['k8s.pod.uid'] AS k8s_pod_uid,
                ResourceAttributes['k8s.{"container" if namespace == "ts" else "deployment"}.name'] AS k8s_container_name,
                Body 
            FROM 
                otel_logs 
            WHERE 
                Timestamp BETWEEN '{start_time}' AND '{end_time}'
                AND ResourceAttributes['k8s.namespace.name'] = '{namespace}'
        """
    elif query_type == "metric":
        return Rf"""
            (
                SELECT
                    ResourceAttributes['k8s.namespace.name'] AS k8s_namespace_name,
                    ResourceAttributes['k8s.pod.uid'] AS k8s_pod_uid,
                    ResourceAttributes['k8s.pod.name'] AS k8s_pod_name,
                    ResourceAttributes['k8s.{"container" if namespace == "ts" else "deployment"}.name'] AS k8s_container_name,
                    MetricName, 
                    MetricDescription, 
                    TimeUnix, 
                    Value,
                    MetricUnit,
                    Attributes['direction'] AS direction
                FROM otel_metrics_gauge
                WHERE TimeUnix BETWEEN '{start_time}' AND '{end_time}'
                AND ResourceAttributes['k8s.namespace.name'] = '{namespace}'
            )
            UNION ALL
            (
                SELECT
                    ResourceAttributes['k8s.namespace.name'] AS k8s_namespace_name,
                    ResourceAttributes['k8s.pod.uid'] AS k8s_pod_uid,
                    ResourceAttributes['k8s.pod.name'] AS k8s_pod_name,
                    ResourceAttributes['k8s.{"container" if namespace == "ts" else "deployment"}.name'] AS k8s_container_name,
                    MetricName, 
                    MetricDescription,
                    TimeUnix, 
                    Value,
                    MetricUnit,
                    Attributes['direction'] AS direction
                FROM otel_metrics_sum oms
                WHERE MetricName IN ('k8s.pod.network.io', 'k8s.pod.network.errors')
                AND TimeUnix BETWEEN '{start_time}' AND '{end_time}'
                AND ResourceAttributes['k8s.namespace.name'] = '{namespace}'
            )
        """
    elif query_type == "trace":
        return (
            Rf"""
            WITH 
                trace_ids AS (
                    SELECT DISTINCT traceID
                    FROM apo.jaeger_spans_local
                    WHERE `timestamp` BETWEEN '{start_time}' AND '{end_time}'
                ),

                parent AS (
                    SELECT
                        lower(Hex(reinterpretAsString(base64Decode(JSONExtractString(model, 'span_id'))))) AS SpanId,
                        JSONExtractString(t1.model, 'process', 'service_name') AS ServiceName,
                        traceID
                    FROM
                        apo.jaeger_spans_local
                    WHERE
                        traceID IN (SELECT traceID FROM trace_ids)
                )

            SELECT
                t1.`timestamp` AS Timestamp,
                t1.traceID AS TraceId,
                lower(Hex(reinterpretAsString(base64Decode(JSONExtractString(t1.model, 'span_id'))))) AS SpanId,
                JSONExtractString(t1.model, 'operation_name') AS SpanName,
                JSONExtractString(t1.model, 'process', 'service_name') AS ServiceName,
                JSONExtractUInt(t1.model, 'duration') AS Duration,
                COALESCE(parent.SpanId, 'Unknown') AS ParentSpanId,
                COALESCE(parent.SServiceName, 'Unknown') AS ParentServiceName
            FROM
                apo.jaeger_spans_local t1
            LEFT JOIN
                parent
            ON
                parent.SpanId = lower(Hex(reinterpretAsString(base64Decode(JSONExtractString(t1.model, 'references', 1, 'span_id')))))
            WHERE
                t1.`timestamp` BETWEEN '{start_time}' AND '{end_time}'
        """
            if namespace == "ts"
            else Rf"""
            WITH
                trace_ids AS (
                    SELECT DISTINCT TraceId
                    FROM otel_traces
                    WHERE `Timestamp` BETWEEN '{start_time}' AND '{end_time}'
                ),
                parent AS (
                    SELECT
                        SpanId,
                        ServiceName
                    FROM
                        otel_traces
                    WHERE
                        TraceId IN (SELECT TraceId FROM trace_ids)
                )

            SELECT
                ot1.`Timestamp`,
                ot1.TraceId,
                ot1.SpanId,
                ot1.ParentSpanId,
                ot1.SpanName,
                ot1.ServiceName,
                ot1.Duration,
                parent.ServiceName AS ParentServiceName
            FROM
                otel_traces ot1
            LEFT JOIN
                parent
            ON
                ot1.ParentSpanId = parent.SpanId
            WHERE
                ot1.`Timestamp` BETWEEN '{start_time}' AND '{end_time}'
        """
        )
    else:
        raise ValueError("Invalid query type")


async def fetch_data(client: AsyncClient, query, filepath, semaphore):
    """Fetch data from ClickHouse and save it to a file."""
    async with semaphore:
        async with aiofiles.open(filepath, "w") as f:
            for _ in range(3):
                try:
                    result = await client.raw_query(query=query, fmt="CSVWithNames")
                    print(f"Data written to {filepath}")
                    break
                except Exception as e:
                    print(f"Error fetching data: {e}. Retrying...")
            else:
                print(f"Failed to fetch data for {filepath}, query: {query}")
                return
            await f.write(result.decode("utf-8"))


async def collect_and_save_data(client, folder, start_time, end_time, data_type, namespace, semaphore):
    """Collect and save data in batches."""
    filepath = Path(folder) / f"{data_type}s.csv"
    query = generate_query(data_type, start_time, end_time, namespace)
    await fetch_data(client, query, filepath, semaphore)


def create_folders(namespace: str, chaos_type: str, service: str):
    """Create normal and abnormal folders for storing data."""
    normal_folder = Path(namespace) / chaos_type / service / "normal"
    abnormal_folder = Path(namespace) / chaos_type / service / "abnormal"
    normal_folder.mkdir(parents=True, exist_ok=True)
    abnormal_folder.mkdir(parents=True, exist_ok=True)
    return normal_folder, abnormal_folder


async def process_case(timestamp, namespace, chaos_type, service, client, semaphore):
    """Process a single chaos event."""
    # Parse the input time
    timestamp = datetime.strptime(timestamp.strip(), "%Y-%m-%d %H:%M:%S")

    # Calculate time windows
    # abnormal_start = timestamp - timedelta(minutes=4)
    # abnormal_end = timestamp + timedelta(minutes=6)
    # normal_start = abnormal_start - timedelta(minutes=10)
    # normal_end = abnormal_start

    abnormal_start = timestamp
    abnormal_end = timestamp + timedelta(minutes=10)
    normal_start = abnormal_start - timedelta(minutes=5)
    normal_end = abnormal_start
    
    normal_folder, abnormal_folder = create_folders(namespace, chaos_type, service)
    tasks = [
        collect_and_save_data(client, folder, start_time, end_time, data_type, namespace, semaphore)
        for folder, start_time, end_time in [
            (normal_folder, normal_start, normal_end),
            (abnormal_folder, abnormal_start, abnormal_end),
        ]
        for data_type in ["log", "metric", "trace"]
    ]
    await asyncio.gather(*tasks)


def load_from_toml(config_path):
    """Load chaos events from a TOML file."""
    chaos_config = toml.load(config_path)
    args = []

    for event in chaos_config.get("chaos_events", []):
        input_timestamp = event.get("timestamp")
        input_namespace = event.get("namespace")
        input_chaos_type = event.get("chaos_type")
        input_service = event.get("service")

        try:
            # Validate timestamp format
            datetime.strptime(input_timestamp, "%Y-%m-%d %H:%M:%S")
        except ValueError:
            print(f"Invalid timestamp format for {input_timestamp}. Skipping this event.")
            continue

        args.append([input_timestamp, input_namespace, input_chaos_type, input_service])

    return args


def interactive_input():
    """Collect chaos events interactively from user input."""
    args = []

    while True:
        input_timestamp = input(
            "Enter the timestamp for anomaly injection (YYYY-MM-DD HH:MM:SS, or press Enter to stop): "
        ).strip()

        # Stop the loop if no valid timestamp is entered
        if not input_timestamp:
            print("No valid timestamp provided. Stopping input.")
            break

        try:
            # Try parsing the timestamp to check validity
            datetime.strptime(input_timestamp, "%Y-%m-%d %H:%M:%S")
        except ValueError:
            print("Invalid timestamp format. Please try again.")
            continue

        input_namespace = input("Enter namespace: ").strip()
        input_chaos_type = input("Enter the chaos type: ").strip()
        input_service = input("Enter the service name: ").strip()

        args.append([input_timestamp, input_namespace, input_chaos_type, input_service])

    return args


async def main():
    config_path = "chaos_config.toml"
    client = await clickhouse_connect.create_async_client(
        host="10.26.1.146", username=CLICKHOUSE_USER, password=CLICKHOUSE_PASSWORD
    )
    semaphore = asyncio.Semaphore(2)
    # Check if the TOML file exists
    if os.path.exists(config_path):
        print(f"Loading chaos events from {config_path}...")
        args = load_from_toml(config_path)
    else:
        print("No TOML file found. Switching to interactive input.")
        args = interactive_input()

    print("Chaos events:", args)
    await asyncio.gather(*(process_case(*arg, client, semaphore) for arg in args))


if __name__ == "__main__":
    asyncio.run(main())
