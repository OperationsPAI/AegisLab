# Run the server 

```bash
docker compose up
```

# Develop the server

```bash
cd experiments_controller
docker compose up redis dagger-engine
_EXPERIMENTAL_DAGGER_RUNNER_HOST=tcp://localhost:5678 go run main.go both # --port 8082 
```

# Input Specification

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