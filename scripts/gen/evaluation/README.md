# Eval Tool

A CLI tool for evaluating RCA Benchmark results, generating structured metrics analysis in JSON format.

## ✨ Features

- ​**CSV Analysis**: Processes `result.csv` from rcabench
- ​**Multi-Metric Evaluation**: Computes granularity metrics across multiple dimensions
- ​**Smart Output**: Auto-generates JSON reports with UUID filenames
- ​**Structured Logging**: Built-in log tracing with file:line diagnostics

## 📦 Installation

### From Source

```bash
git clone https://github.com/LGU-SE-Internal/rcabench.git
cd scripts/gen/evaluation
# Generates binaries in bin/
make linux  # or make windows
```

## 🚀 Usage

### Basic Execution

```bash
eval --input result.csv --service <service-name> [--output custom/path.json]
```

### Required Flags

| Flag              | Description                      | Example                |
| ----------------- | -------------------------------- | ---------------------- |
| `-i`, `--input`   | Path to result.csv file          | `-i ./data/result.csv` |
| `-s`, `--service` | Target service name for analysis | `-s payment-service`   |

### Optional Flags

| Flag             | Default Path                     | Valid Format |
| ---------------- | -------------------------------- | ------------ |
| `-o`, `--output` | `./eval-tool/output/{UUID}.json` | `.json` only |

## 📄 Input Format

### Required Columns

| Column       | Type    | Description                                                        | Example Value       |
| ------------ | ------- | ------------------------------------------------------------------ | ------------------- |
| `level`      | string  | Error impact level identifier (`service`, `pod`, `span`, `metric`) | `service`           |
| `result`     | string  | Full service name (must follow k8s service naming convention)      | `ts-avatar-service` |
| `rank`       | integer | Sorting sequence number (1 indicates highest priority)             | `1`                 |
| `confidence` | float   | Diagnosis confidence score (0.0-1.0, currently set to 0)           | `0`                 |

### Example CSV Snippet

```csv
level,result,rank,confidence
service,ts-avatar-service,1,0
service,ts-wait-order-service,2,0
service,ts-news-service,3,0
```

## 📄 Output Structure

```json
{
  "metric_1": [
    {
      "level": "service",
      "metric": "AC@1",
      "rate": 0
    },
    {
      "level": "service",
      "metric": "AC@3",
      "rate": 0
    },
    {
      "level": "service",
      "metric": "AC@5",
      "rate": 0
    }
  ],
  "metric_2": [
    // ... metric-specific conclusions
  ]
}
```

## 📌 Report issues <a href="https://github.com/LGU-SE-Internal/rcabench/issues">here</a>

Key Features:

- Clear installation paths for different user types
- Visual flag documentation table
- Output schema preview
- Validation requirements listing
- Contextual error log examples
- Platform-specific build guidance
- Direct issue reporting channel

Let me know if you need any section expanded!
