import asyncio
from pathlib import Path
from typing import List

import pandas as pd
from aiomultiprocess import Pool
from pandas import DataFrame

from ..abstract import Parser

INCLUDE_METRIC_NAME = [
    "k8s.pod.cpu.usage",
    "k8s.pod.cpu_limit_utilization",
    "k8s.pod.memory.usage",
    "k8s.pod.memory_limit_utilization",
    "k8s.pod.network.errors",
    "k8s.pod.network.io",
]
NETWORK_COLUMNS = ['k8s.pod.network.errors', 'receive_bytes', 'transmit_bytes']
# filesystem_columns = ['k8s.pod.filesystem.capacity', 'k8s.pod.filesystem.usage']


EXCLUDED_PODS = [
    "redis-cart",
    "loadgenerator",
    "mysql",
    "otel-demo-opensearch",
    "kafka",
    "otelcol",
    "grafana",
    "flagd",
    "jaeger",
    "prometheus",
]


class MetricParser(Parser):
    def __init__(self, base_dir: Path):
        super().__init__(base_dir)
        self.file_name = "metrics.csv"
        self.normal_group_metric_path = base_dir / "normal" / self.file_name
        self.abnormal_group_metric_path = base_dir / "abnormal" / self.file_name

    async def parse(self, pool: Pool):
        tasks = [self._process_metric_data(pool, normal=normal) for normal in [True, False]]
        df = await asyncio.gather(*tasks)
        return df

    async def _process_metric_data(self, pool: Pool, normal: bool):
        file_path = self.normal_group_metric_path if normal else self.abnormal_group_metric_path
        df = pd.read_csv(file_path, engine='pyarrow').dropna(subset=["k8s_pod_name"])
        output_dir = self.base_dir / ("normal" if normal else "abnormal") / "processed_metrics"
        output_dir.mkdir(parents=True, exist_ok=True)
        results = await asyncio.gather(
            *[
                process_pod(pod, group, output_dir, pool, 10000)
                for pod, group in df.groupby("k8s_pod_name")
                if pod and not any(substring in pod for substring in EXCLUDED_PODS)
            ]
        )
        return results


async def batch_process(batch: pd.DataFrame, pod):
    batch = batch[batch['MetricName'].isin(INCLUDE_METRIC_NAME)]
    # Replace metric names based on direction
    batch.loc[batch['MetricName'] == 'k8s.pod.network.io', 'MetricName'] = batch.loc[
        batch['MetricName'] == 'k8s.pod.network.io', 'direction'
    ].map({'transmit': 'transmit_bytes', 'receive': 'receive_bytes'})

    # Pivot the DataFrame to have timestamps as index and metrics as columns
    pivoted = batch.pivot_table(index='TimeUnix', columns='MetricName', values='Value', aggfunc='first')

    # Fill NaN values with empty strings
    df = pivoted.fillna('')

    # 处理网络指标：差分计算
    if all(col in df.columns for col in NETWORK_COLUMNS):
        if len(df) > 1:
            df[NETWORK_COLUMNS] = df[NETWORK_COLUMNS].diff()
            df[NETWORK_COLUMNS] = df[NETWORK_COLUMNS].fillna(0)  # 填充NaN值
            # print(f'Network metrics processed for: {filename}')
        else:
            print(f'Skipped network metrics (not enough rows): {pod}')
    else:
        print(f'Skipped network metrics (missing columns): {pod}')
    return df


async def process_pod(pod, group: pd.DataFrame, output_dir: Path, pool: Pool, batch_size=10000):
    if batch_size >= len(group) * 0.5:
        final_df = await pool.apply(batch_process, args=(group, pod))
    else:
        processed_batches = []
        batches = [group.iloc[start : start + batch_size].copy() for start in range(0, len(group), batch_size)]
        processed_batches = await pool.starmap(batch_process, zip(batches, [pod] * len(batches)))
        final_df = pd.concat(processed_batches, ignore_index=True)
    # Sort by timestamp
    final_df = final_df.sort_values(by='TimeUnix')

    # 保存处理后的数据
    output_path = output_dir / f"{pod}.csv"
    await asyncio.to_thread(final_df.to_csv, output_path, index=True)
    return final_df, pod
