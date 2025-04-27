from clickhouse_connect.driver.exceptions import OperationalError
import clickhouse_connect
import os
import pandas as pd
import subprocess
import json

namespace = os.getenv("NAMESPACE")

clickhouse_host = "10.10.10.58"
username = "default"
password = "password"
try:
    client = clickhouse_connect.get_client(
        host=clickhouse_host, username=username, password=password
    )
except OperationalError:
    print("ClickHouse is not up and reachable")
    exit(0)


def ping_host(host):
    """
    使用系统 ping 命令测试主机连通性
    """
    try:
        result = subprocess.run(
            ["ping", "-c", "1", "-W", "2", host],
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
        )
        if result.returncode == 0:
            print(f"{host} is reachable")
            return True
        else:
            print(f"{host} is unreachalbe")
            return False
    except Exception as e:
        print(f"Ping error: {e}")
        return False


def check_clickhouse_health():
    result = client.ping()
    if result:
        print("ClickHouse is up and reachable")
        return True
    else:
        print("ClickHouse is not up and reachable")
        return False


def generate_metric(start_time, end_time) -> pd.DataFrame:
    query = f"""
    SELECT
        TimeUnix,
        MetricName,
        MetricDescription,
        Value,
        ServiceName,
        MetricUnit,
        toJSONString(ResourceAttributes) AS ResourceAttributes,
        toJSONString(Attributes) AS Attributes
    FROM
        otel_metrics_gauge om
    WHERE
        om.ResourceAttributes['k8s.namespace.name'] = '{namespace}'
        AND om.TimeUnix BETWEEN '{start_time}' AND '{end_time}'
    """

    result = client.raw_query(query=query, fmt="CSVWithNames")

    return result


def generate_metric_sum(start_time, end_time) -> pd.DataFrame:
    query = f"""
    SELECT
        TimeUnix,
        MetricName,
        MetricDescription,
        Value,
        ServiceName,
        MetricUnit,
        toJSONString(ResourceAttributes) AS ResourceAttributes,
        toJSONString(Attributes) AS Attributes
    FROM
        otel_metrics_sum omg
    WHERE
        omg.ResourceAttributes['k8s.namespace.name'] = '{namespace}'
        AND omg.TimeUnix BETWEEN '{start_time}' AND '{end_time}'
    """

    result = client.raw_query(query=query, fmt="CSVWithNames")

    return result


def generate_metric_histogram(start_time, end_time) -> pd.DataFrame:
    query = f"""
    SELECT
        TimeUnix,
        MetricName,
        ServiceName,
        MetricUnit,
        toJSONString(ResourceAttributes) AS ResourceAttributes,
        toJSONString(Attributes) AS Attributes,
        Count,
        Sum,
        BucketCounts,
        ExplicitBounds,
        Min,
        Max,
        AggregationTemporality
    FROM
        otel_metrics_histogram omh
    WHERE
        omh.ResourceAttributes['service.namespace'] = '{namespace}'
        AND omh.TimeUnix BETWEEN '{start_time}' AND '{end_time}'
    """

    # 执行查询
    result = client.raw_query(query=query, fmt="CSVWithNames")

    return result


def generate_log(start_time, end_time) -> pd.DataFrame:
    query = f"""
    SELECT
        Timestamp,
        TimestampTime,
        TraceId,
        SpanId,
        SeverityText,
        SeverityNumber,
        ServiceName,
        Body,
        toJSONString(ResourceAttributes) AS ResourceAttributes,
        LogAttributes
    FROM
        otel_logs ol
    WHERE
        ol.ResourceAttributes['service.namespace'] = '{namespace}'
        AND ol.Timestamp BETWEEN '{start_time}' AND '{end_time}'
    """

    # 执行查询
    result = client.raw_query(query=query, fmt="CSVWithNames")

    return result


def generate_trace(start_time, end_time) -> pd.DataFrame:
    query = f"""
    SELECT Timestamp,
        TraceId,
        SpanId,
        ParentSpanId,
        TraceState,
        SpanName,
        SpanKind,
        ServiceName,
        toJSONString(ResourceAttributes) AS ResourceAttributes,
        SpanAttributes,
        Duration,
        StatusCode,
        StatusMessage
    FROM
        otel_traces ot
    WHERE
        ot.ResourceAttributes['service.namespace'] = '{namespace}'
        AND ot.Timestamp BETWEEN '{start_time}' AND '{end_time}'
    """
    # 执行查询
    result = client.raw_query(query=query, fmt="CSVWithNames")

    return result


def generate_trace_id_ts(start_time, end_time) -> pd.DataFrame:
    query = f"""
    SELECT TraceId,
        Start,
        End
    FROM
        otel_traces_trace_id_ts
    WHERE
        Start BETWEEN '{start_time}' AND '{end_time}'
        AND End BETWEEN '{start_time}' AND '{end_time}'
    """

    # 执行查询
    result = client.raw_query(query=query, fmt="CSVWithNames")

    return result


def generate_data_nezha(start_time, end_time) -> pd.DataFrame:
    query = """
    SELECT 
        TimeUnix,
        MetricName, 
        Value, 
        ResourceAttributes['k8s.pod.name'] as PodName
    FROM 
        otel_metrics_gauge
    WHERE 
        ResourceAttributes['k8s.namespace.name'] = %(namespace)s
        AND TimeUnix BETWEEN %(start_time)s AND %(end_time)s
    """

    # 设置查询参数
    params = {"namespace": namespace, "start_time": start_time, "end_time": end_time}

    # 执行查询
    result = client.execute(query, params)

    # 定义 DataFrame 的列名
    selected_columns = ["TimeStamp", "MetricName", "Value", "PodName"]

    # 将查询结果转换为 pandas DataFrame
    df = pd.DataFrame(result, columns=selected_columns)

    df_pivot = df.pivot_table(
        index=["TimeStamp", "PodName"], columns="MetricName", values="Value"
    ).reset_index()

    return df_pivot


def save_to_csv(result: bytes, filename: str):
    with open(filename, "w") as f:
        f.write(result.decode("utf-8"))
    print(f"Data has been successfully saved to {filename}")


def convert_to_clickhouse_time(unix_timestamp):
    """将 UNIX 时间戳转换为 ClickHouse 支持的时间格式"""
    return (
        pd.to_datetime(unix_timestamp, utc=True, unit="s")
        .astimezone(os.environ["TIMEZONE"])
        .strftime("%Y-%m-%d %H:%M:%S")
    )


if __name__ == "__main__":
    if not check_clickhouse_health():
        exit(0)

    if os.environ.get("NORMAL_START") and os.environ.get("NORMAL_END"):
        normal_time_range = [
            int(os.environ["NORMAL_START"]),
            int(os.environ["NORMAL_END"]),
        ]
    else:
        print("env NORMAL_START and NORMAL_END not found")
        exit(0)

    if os.environ.get("ABNORMAL_START") and os.environ.get("ABNORMAL_END"):
        abnormal_time_range = [
            int(os.environ["ABNORMAL_START"]),
            int(os.environ["ABNORMAL_END"]),
        ]
    else:
        print("env ABNORMAL_START and ABNORMAL_END not found")
        exit(0)

    print("Normal Time Range (Unix):", normal_time_range)
    print("Abnormal Time Range (Unix):", abnormal_time_range)

    output_path = os.environ.get("OUTPUT_PATH")

    normal_start_time = normal_time_range[0]
    normal_end_time = normal_time_range[1]

    abnormal_start_time = abnormal_time_range[0]
    abnormal_end_time = abnormal_time_range[1]

    # 转换为 ClickHouse 格式
    normal_start_time_clickhouse = convert_to_clickhouse_time(normal_start_time)
    normal_end_time_clickhouse = convert_to_clickhouse_time(normal_end_time)
    abnormal_start_time_clickhouse = convert_to_clickhouse_time(abnormal_start_time)
    abnormal_end_time_clickhouse = convert_to_clickhouse_time(abnormal_end_time)

    print(
        "Normal Time Range (ClickHouse):",
        [normal_start_time_clickhouse, normal_end_time_clickhouse],
    )
    print(
        "Abnormal Time Range (ClickHouse):",
        [abnormal_start_time_clickhouse, abnormal_end_time_clickhouse],
    )

    if normal_end_time != abnormal_start_time:
        print(
            "The time range may not suitable for discontinuous time range, please check it."
        )

    os.makedirs(output_path, exist_ok=True)

    normal_data = [
        (generate_metric, "normal_metrics.csv"),
        (generate_metric_sum, "normal_metric_sum.csv"),
        (generate_metric_histogram, "normal_metrics_histogram.csv"),
        (generate_log, "normal_logs.csv"),
        (generate_trace, "normal_traces.csv"),
        (generate_trace_id_ts, "normal_trace_id_ts.csv"),
    ]

    abnormal_data = [
        (generate_metric, "abnormal_metrics.csv"),
        (generate_metric_sum, "abnormal_metric_sum.csv"),
        (generate_metric_histogram, "abnormal_metrics_histogram.csv"),
        (generate_log, "abnormal_logs.csv"),
        (generate_trace, "abnormal_traces.csv"),
        (generate_trace_id_ts, "abnormal_trace_id_ts.csv"),
    ]

    for func, filename in normal_data:
        print(f"executing {func.__name__}")
        save_to_csv(
            func(normal_start_time_clickhouse, normal_end_time_clickhouse),
            f"{output_path}/{filename}",
        )

    for func, filename in abnormal_data:
        print(f"executing {func.__name__}")
        save_to_csv(
            func(abnormal_start_time_clickhouse, abnormal_end_time_clickhouse),
            f"{output_path}/{filename}",
        )

    with open(f"{output_path}/time_ranges.json", "w") as f:
        val = {
            "normal_start": str(normal_start_time),
            "normal_end": str(normal_end_time),
            "abnormal_start": str(abnormal_start_time),
            "abnormal_end": str(abnormal_end_time),
        }
        json.dump(val, f)

    files = list(os.listdir(output_path))

    # compress to data.tar.gz
    subprocess.run(["tar", "-czf", "data.tar.gz", *files], check=True, cwd=output_path)

    # remove files
    for file in files:
        os.remove(f"{output_path}/{file}")
