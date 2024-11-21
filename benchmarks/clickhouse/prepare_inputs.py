import os
import time

import clickhouse_connect
import pandas as pd
from clickhouse_connect.driver.client import Client

namespace = "ts"
clickhouse_host = "10.10.10.58"


def generate_metric(start_time, end_time) -> pd.DataFrame:
    # 连接到 ClickHouse 客户端
    client = clickhouse_connect.get_client(
        host=clickhouse_host, username="default", password="password"
    )
    # 定义查询语句
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
    # 连接到 ClickHouse 客户端
    client = clickhouse_connect.get_client(
        host=clickhouse_host, username="default", password="password"
    )

    # 定义查询语句
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
    # 连接到 ClickHouse 客户端
    client = clickhouse_connect.get_client(
        host=clickhouse_host, username="default", password="password"
    )

    # 定义查询语句
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
    # 连接到 ClickHouse 客户端
    client = clickhouse_connect.get_client(
        host=clickhouse_host, username="default", password="password"
    )

    # 定义查询语句
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
    # 连接到 ClickHouse 客户端
    client = clickhouse_connect.get_client(
        host=clickhouse_host, username="default", password="password"
    )

    # 定义查询语句
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
        host=clickhouse_host, username="default", password="password"
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
    # 连接到 ClickHouse 客户端
    client = Client(
        clickhouse_host, user="default", password="password", database="default"
    )

    # 定义查询语句
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


if __name__ == "__main__":
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

    print(normal_time_range)
    print(abnormal_time_range)
    normal_start_time = normal_time_range[0]
    normal_end_time = normal_time_range[1]

    abnormal_start_time = abnormal_time_range[0]
    abnormal_end_time = abnormal_time_range[1]

    if normal_end_time != abnormal_time_range:
        print(
            "the time range may not suitable for discontinuous time range, please check it."
        )
    os.mkdir("input")

    save_to_csv(
        generate_metric(normal_start_time, abnormal_end_time), "input/metrics.csv"
    )
    save_to_csv(
        generate_metric_sum(normal_start_time, abnormal_end_time), "input/metric.csv"
    )
    save_to_csv(
        generate_metric_histogram(normal_start_time, abnormal_end_time),
        "input/metrics_histogram.csv",
    )
    save_to_csv(generate_log(normal_start_time, abnormal_end_time), "input/logs.csv")
    save_to_csv(
        generate_trace(normal_start_time, abnormal_end_time), "input/traces.csv"
    )
    save_to_csv(
        generate_trace_id_ts(normal_start_time, abnormal_end_time),
        "input/trace_id_ts.csv",
    )
