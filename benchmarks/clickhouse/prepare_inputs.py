from clickhouse_driver import Client
import pandas as pd
import time

def generate_data() -> pd.DataFrame:
    # 连接到 ClickHouse 客户端
    client = Client('10.26.1.146', user='default', password='nn', database='default')
    
    # 获取当前时间和开始时间（过去 10 分钟）
    end_time = int(time.time())
    start_time = end_time - 600

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
    params = {
        'start_time': start_time,
        'end_time': end_time
    }

    # 执行查询
    result = client.execute(query, params)

    # 定义 DataFrame 的列名
    selected_columns = ['TimeStamp', 'MetricName', 'Value', "PodName"]

    # 将查询结果转换为 pandas DataFrame
    df = pd.DataFrame(result, columns=selected_columns)

    return df

def save_to_csv(df: pd.DataFrame, filename: str):
    """
    将 DataFrame 保存为 CSV 文件。

    :param df: 要保存的 pandas DataFrame
    :param filename: 保存的文件名（包括路径）
    """
    try:
        df.to_csv(filename, index=False, encoding='utf-8-sig')
        print(f"数据已成功保存到 {filename}")
    except Exception as e:
        print(f"保存 CSV 文件时出错: {e}")

if __name__ == "__main__":
    data_df = generate_data()
    
    csv_filename = 'input.csv'
    
    save_to_csv(data_df, csv_filename)