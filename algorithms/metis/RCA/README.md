- [Usage](#usage)
  - [evaluation.py](#evaluationpy)
  - [collect\_data.py](#collect_datapy)
    - [安装依赖](#安装依赖)
    - [环境变量](#环境变量)
    - [输入](#输入)
    - [输出](#输出)
      - [文件夹结构](#文件夹结构)
      - [CSV 文件](#csv-文件)
  - [log\_ad](#log_ad)
  - [trace\_ad](#trace_ad)


## Usage

### 环境安装

使用以下命令来创建一个Metis的python环境并安装所有指定的包：

```bash
python3.9 -m pip install -r requirements.txt
```


### evaluation.py

代码主函数，请首先在base_dirs中添加需要检测的故障case路径（请替换为你的数据路径），比如：

```
base_dirs = [
        R"E:\OneDrive - CUHK-Shenzhen\RCA_Dataset\test_new_datasets\onlineboutique\cpu\checkoutservice-1011-1441",
        R"E:\OneDrive - CUHK-Shenzhen\RCA_Dataset\test_new_datasets\onlineboutique\memory\checkoutservice-1011-1500",
        R"E:\OneDrive - CUHK-Shenzhen\RCA_Dataset\test_new_datasets\onlineboutique\pod_failure\paymentservice-1011-1525",
    ]
```

使用以下命令触发根因分析：
```bash
python evaluation.py
```

`service ranking`和`events`输出在故障数据路径下的`rcl_output`文件夹中。


### collect_data.py

脚本使用异步 I/O 操作从 ClickHouse 收集数据。它会根据输入的时间戳、`namespace`、`chaos_type` 和 `service_name`，从 ClickHouse 中提取 log、metric 和 trace 数据，并将数据保存到 CSV 文件中。

#### 安装依赖

在运行脚本之前，请确保你已安装以下Python库：

- `clickhouse_connect`
- `aiofiles`
- `asyncio`

你可以使用以下命令安装这些库：

```bash
pip install clickhouse-connect aiofiles
```

#### 环境变量

脚本从环境变量中读取ClickHouse的连接凭据。请设置以下环境变量：

- `CLICKHOUSE_USER`：ClickHouse用户名（默认值为`default`）
- `CLICKHOUSE_PASSWORD`：ClickHouse密码（默认值为`nn`）


#### 输入

可以通过两种方式输入：

1. **从 TOML 文件加载：**  
   脚本会自动检查名为 `chaos_config.toml` 的文件。如果文件存在，将从中加载异常事件，包括时间戳、`namespace`、`chaos_type` 和 `service_name`。

   示例 `chaos_config.toml` 文件：
   
   ```toml
   [[chaos_events]]
   timestamp = "2024-09-21 11:29:39"
   namespace = "ts"
   chaos_type = "HTTP_Abort"
   service = "travel2" 
   [[chaos_events]]
   timestamp = "2024-09-26 21:25:49"
   namespace = "onlineboutique"
   chaos_type = "cpu"
   service = "adservice"
   ```
   需要安装 `toml` 库，可以使用以下命令安装：

   ```bash
   pip install toml
   ```

2. **交互式输入：**  
   如果未找到 TOML 文件，脚本将提示用户手动输入。输入格式为时间戳（`YYYY-MM-DD HH:MM:SS`）、`chaos_type` 和 `service_name`，示例如下：

   ```
   Enter the timestamp for anomaly injection (YYYY-MM-DD HH:MM:SS, or press Enter to stop): 2024-09-21 11:29:39
   Enter namespace: ts
   Enter the chaos type: HTTP_Abort
   Enter the service name: travel2
   Enter the timestamp for anomaly injection (YYYY-MM-DD HH:MM:SS, or press Enter to stop): 2024-09-26 21:25:49
   Enter namespace: onlineboutique
   Enter the chaos type: cpu
   Enter the service name: adservice
   Enter the timestamp for anomaly injection (YYYY-MM-DD HH:MM:SS, or press Enter to stop):
   No valid timestamp provided. Stopping input.
   ```

输入完成后，脚本将持续要求输入新的 `chaos_type` 和 `service_name` 以及对应的时间戳，直到用户输入空行或无效的时间戳为止。


#### 输出

##### 文件夹结构

根据输入的时间戳、`namespace`、`chaos_type` 和 `service_name`，脚本会创建相应的文件夹。输出文件夹的结构如下：

```
{namespace}/{chaos_type}/{service_name}/normal/
{namespace}/{chaos_type}/{service_name}/abnormal/
```

- `normal` 文件夹包含异常注入时间前 14 分钟到前 4 分钟的数据。
- `abnormal` 文件夹包含异常注入时间前 4 分钟到后 6 分钟的数据。


##### CSV 文件

在每个子文件夹内，脚本会创建三个CSV文件，分别用于保存日志、指标和追踪数据。这些文件的命名规则为 `{data_type}s.csv`，其中 `{data_type}` 为 `log`、`metric` 或 `trace`。文件内容如下：

- **`logs.csv`**:
  - `Timestamp`
  - `k8s_namespace_uid`
  - `k8s_namespace_name`
  - `k8s_pod_uid`
  - `k8s_container_name`
  - `Body`
  
  **`logs.csv` 示例**:
  ```csv
  "Timestamp","k8s_namespace_uid","k8s_namespace_name","k8s_pod_uid","k8s_container_name","Body"
  "2024-09-13 15:42:03.000077493","","ts","f86d3639-9af3-4c72-be9b-3a9a359351fa","ts-order-service","15:42:02.999 INFO  o.s.OrderServiceImpl#178 TraceID:  SpanID:  [queryOrders][Step 2][Check All Requirement End]"
  ```
- **`metrics.csv`**:
  - `k8s_namespace_name`
  - `k8s_pod_uid`
  - `k8s_pod_name`
  - `k8s_pod_uid`
  - `k8s_container_name`
  - `MetricName`
  - `MetricDescription`
  - `TimeUnix`
  - `Value`
  
  **`metrics.csv` 示例**:
  ```csv
  "k8s_namespace_name","k8s_pod_uid","k8s_pod_name","k8s_container_name","MetricName","MetricDescription","TimeUnix","Value"
  "onlineboutique","b1e516b8-437e-4cac-a2eb-023dd8894d9e","emailservice-579754779-x2cwp","","k8s.pod.memory.rss","Pod memory rss","2024-09-26 21:22:44.572471143",54677504
  ```

- **`traces.csv`**:
  - `Timestamp`
  - `TraceId`
  - `SpanId`
  - `SpanName`
  - `ServiceName`
  - `Duration`
  - `ParentSpanId`
  - `ParentServiceName`
  
  **`traces.csv` 示例**:
  ```csv
  Timestamp,TraceId,SpanId,SpanName,ServiceName,Duration,ParentSpanId,ParentServiceName
  2024-09-19 16:42:18,c1b96c9325202e2f47bc20319e3b5c18,cf388dbcce905860,GET /api/v1/verifycode/verify/{verifyCode},ts-verification-code-service,1294232,9aa6e3f427fe86ad,ts-auth-service
  ```

### log_ad

在 main 中设置好 `base` 和 `file_name`，输出会在 `base_dir` 下的 `log_ad_output` 文件夹

有异步 IO 和 multiprocess， **必须配置 redis，不然会死锁**

`error_weight`, `warn_weight` 和 `k` 在 `detect_anomalies` 配置，默认是 0.7, 0.3, 3

配置好之后 `python main.py` 运行

### trace_ad

在 main 中设置好 `base` 和 `file_name`，输出会在 `base_dir` 下的 `trace_ad_output` 文件夹

`k` 在 `detect_anomalies` 配置，默认是 3



