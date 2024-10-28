
# Input Specification

```bash
dagger call evaluate --bench_dir ./benchmarks/clickhouse --src algorithms/dummyalgo/ --start-script experiments/run_exp.py export --path=./output
```

```bash
dagger call evaluate --bench_dir ./benchmarks/clickhouse \  # 指定 benchmark 的代码目录
--src algorithms/dummyalgo/ \                               # 指定算法代码目录
--start-script experiments/run_exp.py \                     # 指定启动脚本代码目录
export --path=./output                                      # 指定输出目录
```


`def start_rca(params: Dict):` 函数的输入参数示例

```json
{
    'log_file': 'log.csv',
    'trace_file': 'trace.csv',
    'trace_id_ts_file': 'trace.csv',
    'metric_file': 'metric.csv',
    'metric_sum_file': 'metric.csv',
    'metric_summary_file': 'metric.csv',
    'metric_histogram_file': 'metric.csv',
    'event_file': 'event.csv',
    'profiling_file': 'profile.csv',
    'normal_time_range': [(0, 10), (20, 30)],
    'abnormal_time_range': [(50, 60), (70, 80)],
    'output_file_path': '/app/output/result.csv'
}
```