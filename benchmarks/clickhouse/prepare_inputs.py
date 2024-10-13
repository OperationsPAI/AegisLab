from clickhouse_driver import Client
import pandas as pd
import time
import os

namespace = "ts-dev"


def generate_metric(start_time, end_time) -> pd.DataFrame:
    # 连接到 ClickHouse 客户端
    client = Client("10.26.1.146", user="default", password="nn", database="default")

    # 定义查询语句
    query = """
SELECT 
    TimeUnix,
    MetricName, 
    Value, 
    ServiceName,
    MetricUnit,
    ResourceAttributes,
    Attributes
FROM 
    otel_metrics_gauge
WHERE 
    (ResourceAttributes['k8s.namespace.name'] = %(namespace)s
     OR NOT mapContains(ResourceAttributes, 'k8s.namespace.name'))
    AND TimeUnix BETWEEN %(start_time)s AND %(end_time)s
    """

    # 设置查询参数
    params = {"namespace": namespace, "start_time": start_time, "end_time": end_time}

    # 执行查询
    result = client.execute(query, params)

    # 定义 DataFrame 的列名
    selected_columns = [
        "TimeUnix",
        "MetricName",
        "Value",
        "ServiceName",
        "MetricUnit",
        "ResourceAttributes",
        "Attributes",
    ]

    # 将查询结果转换为 pandas DataFrame
    df = pd.DataFrame(result, columns=selected_columns)

    return df


def generate_metric_sum(start_time, end_time) -> pd.DataFrame:
    # 连接到 ClickHouse 客户端
    client = Client("10.26.1.146", user="default", password="nn", database="default")

    # 定义查询语句
    query = f"""
    SELECT 
        TimeUnix,
        MetricName, 
        Value, 
        ServiceName,
        MetricUnit,
        ResourceAttributes,
        Attributes
    FROM 
        otel_metrics_sum
    WHERE 
        (ResourceAttributes['k8s.namespace.name'] = %(namespace)s
        OR NOT mapContains(ResourceAttributes, 'k8s.namespace.name'))
        AND TimeUnix BETWEEN %(start_time)s AND %(end_time)s
    """

    # 设置查询参数
    params = {"namespace": namespace, "start_time": start_time, "end_time": end_time}

    # 执行查询
    result = client.execute(query, params)

    # 定义 DataFrame 的列名
    selected_columns = [
        "TimeUnix",
        "MetricName",
        "Value",
        "ServiceName",
        "MetricUnit",
        "ResourceAttributes",
        "Attributes",
    ]

    # 将查询结果转换为 pandas DataFrame
    df = pd.DataFrame(result, columns=selected_columns)

    return df


def generate_metric_histogram(start_time, end_time) -> pd.DataFrame:
    # 连接到 ClickHouse 客户端
    client = Client("10.26.1.146", user="default", password="nn", database="default")

    # 定义查询语句
    query = f"""
    SELECT 
        TimeUnix,
        MetricName, 
        ServiceName,
        MetricUnit,
        ResourceAttributes,
        Attributes,
        Count,
        Sum,
        BucketCounts,
        ExplicitBounds,
        Min,
        Max,
        AggregationTemporality
    FROM 
        otel_metrics_histogram
    WHERE 
        (ResourceAttributes['k8s.namespace.name'] = %(namespace)s
        OR NOT mapContains(ResourceAttributes, 'k8s.namespace.name'))
        AND TimeUnix BETWEEN %(start_time)s AND %(end_time)s
    """

    # 设置查询参数
    params = {"namespace": namespace, "start_time": start_time, "end_time": end_time}

    # 执行查询
    result = client.execute(query, params)

    # 定义 DataFrame 的列名
    selected_columns = [
        "TimeUnix",
        "MetricName",
        "ServiceName",
        "MetricUnit",
        "ResourceAttributes",
        "Attributes",
        "Count",
        "Sum",
        "BucketCounts",
        "ExplicitBounds",
        "Min",
        "Max",
        "AggregationTemporality",
    ]

    # 将查询结果转换为 pandas DataFrame
    df = pd.DataFrame(result, columns=selected_columns)

    return df


def generate_log(start_time, end_time) -> pd.DataFrame:
    # 连接到 ClickHouse 客户端
    client = Client("10.26.1.146", user="default", password="nn", database="default")

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
        ResourceAttributes,
        LogAttributes
    FROM 
        otel_logs
    WHERE 
        (ResourceAttributes['k8s.namespace.name'] = %(namespace)s
        OR NOT mapContains(ResourceAttributes, 'k8s.namespace.name'))
        AND Timestamp BETWEEN %(start_time)s AND %(end_time)s
    """

    # 设置查询参数
    params = {"namespace": namespace, "start_time": start_time, "end_time": end_time}

    # 执行查询
    result = client.execute(query, params)

    # 定义 DataFrame 的列名
    selected_columns = [
        "Timestamp",
        "TimestampTime",
        "TraceId",
        "SpanId",
        "SeverityText",
        "SeverityNumber",
        "ServiceName",
        "Body",
        "ResourceAttributes",
        "LogAttributes",
    ]

    # 将查询结果转换为 pandas DataFrame
    df = pd.DataFrame(result, columns=selected_columns)

    return df


def generate_trace(start_time, end_time) -> pd.DataFrame:
    # 连接到 ClickHouse 客户端
    client = Client("10.26.1.146", user="default", password="nn", database="default")

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
        ResourceAttributes,
        SpanAttributes,
        Duration,
        StatusCode,
        StatusMessage
    FROM 
        otel_traces
    WHERE 
        (ResourceAttributes['service.namespace'] = %(namespace)s)
        AND Timestamp BETWEEN %(start_time)s AND %(end_time)s
    """

    # 设置查询参数
    params = {"namespace": namespace, "start_time": start_time, "end_time": end_time}

    # 执行查询
    result = client.execute(query, params)

    # 定义 DataFrame 的列名
    selected_columns = [
        "Timestamp",
        "TraceId",
        "SpanId",
        "ParentSpanId",
        "TraceState",
        "SpanName",
        "SpanKind",
        "ServiceName",
        "ResourceAttributes",
        "SpanAttributes",
        "Duration",
        "StatusCode",
        "StatusMessage",
    ]

    # 将查询结果转换为 pandas DataFrame
    df = pd.DataFrame(result, columns=selected_columns)

    return df


def generate_trace_id_ts(start_time, end_time) -> pd.DataFrame:
    # 连接到 ClickHouse 客户端
    client = Client("10.26.1.146", user="default", password="nn", database="default")

    # 定义查询语句
    query = f"""
    SELECT TraceId,
        Start, 
        End
    FROM 
        otel_traces_trace_id_ts
    WHERE 
        Start BETWEEN %(start_time)s AND %(end_time)s
        AND End BETWEEN %(start_time)s AND %(end_time)s
    """

    # 设置查询参数
    params = {"namespace": namespace, "start_time": start_time, "end_time": end_time}

    # 执行查询
    result = client.execute(query, params)

    # 定义 DataFrame 的列名
    selected_columns = ["TraceId", "Start", "End"]

    # 将查询结果转换为 pandas DataFrame
    df = pd.DataFrame(result, columns=selected_columns)

    return df


def generate_data_nezha(start_time, end_time) -> pd.DataFrame:
    # 连接到 ClickHouse 客户端
    client = Client("10.26.1.146", user="default", password="nn", database="default")

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


def save_to_csv(df: pd.DataFrame, filename: str):
    """
    将 DataFrame 保存为 CSV 文件。

    :param df: 要保存的 pandas DataFrame
    :param filename: 保存的文件名（包括路径）
    """
    try:
        df.to_csv(filename, index=False, encoding="utf-8-sig")
        print(f"数据已成功保存到 {filename}")
    except Exception as e:
        print(f"保存 CSV 文件时出错: {e}")


if __name__ == "__main__":
    # 获取当前时间和开始时间（过去 10 分钟）
    end_time = int(time.time())
    start_time = end_time - 600

    os.mkdir("input")

    save_to_csv(generate_metric(start_time, end_time), "input/metrics.csv")
    save_to_csv(generate_metric_sum(start_time, end_time), "input/metric_sum.csv")
    save_to_csv(
        generate_metric_histogram(start_time, end_time), "input/metrics_histogram.csv"
    )
    save_to_csv(generate_log(start_time, end_time), "input/logs.csv")
    save_to_csv(generate_trace(start_time, end_time), "input/traces.csv")
    save_to_csv(generate_trace_id_ts(start_time, end_time), "input/trace_id_ts.csv")
