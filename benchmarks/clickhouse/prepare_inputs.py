import os
import clickhouse_connect
import pandas as pd
import subprocess
from datetime import datetime


namespace = "ts"
clickhouse_host = "clickhouse"
username = "default"
password = "password"
client = clickhouse_connect.get_client(
    host=clickhouse_host, username=username, password=password
)


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
            print(f"{host} 可达")
            return True
        else:
            print(f"{host} 不可达")
            return False
    except Exception as e:
        print(f"Ping 异常: {e}")
        return False


def check_clickhouse_health():
    try:
        result = client.ping()
        if result:
            print("ClickHouse 服务正常")
            return True
        else:
            print("clickhouse 服务异常")
    except Exception as e:
        print(f"ClickHouse 服务异常: {e}")
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
        omh.ResourceAttributes['k8s.namespace.name'] = '{namespace}'
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
    # 连接到 ClickHouse 客户端
    client = clickhouse_connect.get_client(
        host=clickhouse_host, username=username, password=password
    )

    # 定义查询语句
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
        ResourceAttributes['k8s.namespace.name'] = 'ts'
        AND TimeUnix BETWEEN %(start_time)s AND %(end_time)s
    """

    # 设置查询参数
    params = {"start_time": start_time, "end_time": end_time}

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
    print(f"数据已成功保存到 {filename}")


def convert_to_clickhouse_time(unix_timestamp):
    """将 UNIX 时间戳转换为 ClickHouse 支持的时间格式"""
    return (
        pd.to_datetime(unix_timestamp, utc=True, unit="s")
        .astimezone("Asia/Shanghai")
        .strftime("%Y-%m-%d %H:%M:%S")
    )


if __name__ == "__main__":
    check_clickhouse_health()

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

    os.makedirs("input", exist_ok=True)

    print("executing generate_metric")
    save_to_csv(
        generate_metric(normal_start_time_clickhouse, abnormal_end_time_clickhouse),
        "input/metrics.csv",
    )
    print("executing generate_metric_sum")
    save_to_csv(
        generate_metric_sum(normal_start_time_clickhouse, abnormal_end_time_clickhouse),
        "input/metric_sum.csv",
    )
    print("executing generate_metric_histogram")
    save_to_csv(
        generate_metric_histogram(
            normal_start_time_clickhouse, abnormal_end_time_clickhouse
        ),
        "input/metrics_histogram.csv",
    )
    print("executing generate_log")
    save_to_csv(
        generate_log(normal_start_time_clickhouse, abnormal_end_time_clickhouse),
        "input/logs.csv",
    )
    print("executing generate_trace")
    save_to_csv(
        generate_trace(normal_start_time_clickhouse, abnormal_end_time_clickhouse),
        "input/traces.csv",
    )
    print("executing generate_trace_id_ts")
    save_to_csv(
        generate_trace_id_ts(
            normal_start_time_clickhouse, abnormal_end_time_clickhouse
        ),
        "input/trace_id_ts.csv",
    )
