package executor

import (
	"context"
	"dagger/rcabench/config"
	"dagger/rcabench/database"
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"dagger.io/dagger"
	"dagger.io/dagger/dag"
	"github.com/CUHK-SE-Group/chaos-experiment/client"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Rcabench struct{}

func (m *Rcabench) BuildBenchmarkDataImage(
	ctx context.Context,
	src *dagger.Directory,
	abnorStartTime, abnorEndTime, norStartTime, norEndTime time.Time,
) *dagger.Container {
	workspace := dag.Container().
		WithEnvVariable("CACHEBUSTER", time.Now().String()). // 强制重新编译
		WithDirectory("/app", src).
		WithWorkdir("/app").
		Directory("/app")

	logrus.Infof("timestamp: %v %v %v %v", abnorStartTime, abnorEndTime, norStartTime, norEndTime)

	hostSrv := dag.Host().Service([]dagger.PortForward{
		{Frontend: 8123, Backend: 8123},
	}, dagger.HostServiceOpts{Host: config.GetString("database.clickhouse_host")})

	return dag.Container().
		WithServiceBinding("clickhouse", hostSrv).
		WithEnvVariable("TIMEZONE", config.GetString("system.timezone")).
		Build(workspace, dagger.ContainerBuildOpts{
			BuildArgs: []dagger.BuildArg{
				{
					Name:  "ABNORMAL_START_VAR",
					Value: strconv.Itoa(int(abnorStartTime.Unix())),
				},
				{
					Name:  "ABNORMAL_END_VAR",
					Value: strconv.Itoa(int(abnorEndTime.Unix())),
				},
				{
					Name:  "NORMAL_START_VAR",
					Value: strconv.Itoa(int(norStartTime.Unix())),
				},
				{
					Name:  "NORMAL_END_VAR",
					Value: strconv.Itoa(int(norEndTime.Unix())),
				},
			},
		}).WithExec([]string{"python", "prepare_inputs.py"})
}

func (m *Rcabench) BuildAlgoBuilderImage(ctx context.Context, src *dagger.Directory) *dagger.Container {
	workspace := dag.Container().
		WithDirectory("/app", src).
		WithWorkdir("/app").
		Directory("/app")

	return dag.Container().
		Build(workspace, dagger.ContainerBuildOpts{
			Dockerfile: "builder.Dockerfile",
		})
}

func (m *Rcabench) BuildAlgoRunnerImage(
	ctx context.Context, bench_dir, src *dagger.Directory, start_script *dagger.File,
	abnorStartTime, abnorEndTime, norStartTime, norEndTime time.Time,
) *dagger.Container {
	data := m.BuildBenchmarkDataImage(ctx, bench_dir, abnorStartTime, abnorEndTime, norStartTime, norEndTime)
	builder := m.BuildAlgoBuilderImage(ctx, src)
	runner := builder.
		WithWorkdir("/app").
		WithDirectory("/app/input", data.Directory("/app/input")).
		WithFile("/app/rca.py", src.File("rca.py")).
		WithFile("/app/run_exp.py", start_script).
		WithEnvVariable("WORKSPACE", "/app").
		WithEnvVariable("TIMEZONE", config.GetString("system.timezone")).
		WithEnvVariable("ABNORMAL_START", strconv.Itoa(int(abnorStartTime.Unix()))).
		WithEnvVariable("ABNORMAL_END", strconv.Itoa(int(abnorEndTime.Unix()))).
		WithEnvVariable("NORMAL_START", strconv.Itoa(int(norStartTime.Unix()))).
		WithEnvVariable("NORMAL_END", strconv.Itoa(int(norEndTime.Unix())))
	return runner
}
func (m *Rcabench) Evaluate(
	ctx context.Context, bench_dir, src *dagger.Directory, start_script *dagger.File,
	abnorStartTime, abnorEndTime, norStartTime, norEndTime time.Time,
) *dagger.Container {
	return m.BuildAlgoRunnerImage(ctx, bench_dir, src, start_script, abnorStartTime, abnorEndTime, norStartTime, norEndTime).
		WithExposedPort(9000).
		WithExec([]string{"python", "run_exp.py"})
}

func executeAlgorithm(ctx context.Context, taskID string, payload map[string]interface{}) error {
	algPayload, err := parseAlgorithmExecutionPayload(payload)
	if err != nil {
		return err
	}

	var faultRecord database.FaultInjectionSchedule
	err = database.DB.Where("injection_name = ?", algPayload.DatasetName).First(&faultRecord).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("no matching fault injection record found for dataset: %s", algPayload.DatasetName)
		}
		return fmt.Errorf("failed to query database for dataset: %s, error: %v", algPayload.DatasetName, err)
	}

	var startTime, endTime time.Time
	if faultRecord.Status == database.DatasetSuccess {
		startTime = faultRecord.StartTime
		endTime = faultRecord.EndTime
	} else if faultRecord.Status == database.DatasetInitial {
		startTime, endTime, err = client.QueryCRDByName("ts", algPayload.DatasetName)
		if err != nil {
			return fmt.Errorf("failed to QueryCRDByName: %s, error: %v", algPayload.DatasetName, err)
		}
		if err := database.DB.Model(&faultRecord).Where("injection_name = ?", algPayload.DatasetName).
			Updates(map[string]interface{}{
				"start_time": startTime,
				"end_time":   endTime,
			}).Error; err != nil {
			return fmt.Errorf("failed to update start_time and end_time for dataset: %s, error: %v", algPayload.DatasetName, err)
		}
	}

	executionResult := database.ExecutionResult{
		Dataset: faultRecord.ID,
		TaskID:  taskID,
		Algo:    algPayload.Algorithm,
	}
	if err := database.DB.Create(&executionResult).Error; err != nil {
		return fmt.Errorf("failed to create execution result: %v", err)
	}

	updateTaskStatus(taskID, "Running", fmt.Sprintf("Running algorithm for task %s", taskID))

	pwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %v", err)
	}

	parentDir := filepath.Dir(pwd)

	benchPath := filepath.Join(parentDir, "benchmarks", algPayload.Benchmark)
	algoPath := filepath.Join(parentDir, "algorithms", algPayload.Algorithm)
	startScriptPath := filepath.Join(parentDir, "experiments", "run_exp.py")

	if _, err := os.Stat(benchPath); os.IsNotExist(err) {
		return fmt.Errorf("benchmark directory does not exist: %s", benchPath)
	}
	if _, err := os.Stat(algoPath); os.IsNotExist(err) {
		return fmt.Errorf("algorithm directory does not exist: %s", algoPath)
	}
	if _, err := os.Stat(startScriptPath); os.IsNotExist(err) {
		return fmt.Errorf("start script does not exist: %s", startScriptPath)
	}

	rc := &Rcabench{}
	con := rc.Evaluate(ctx, dag.Host().Directory(benchPath), dag.Host().Directory(algoPath), dag.Host().File(startScriptPath),
		startTime, endTime, startTime.Add(-20*time.Minute), startTime)

	if config.GetBool("debug") {
		_, err = con.Directory("/app/output").Export(ctx, "./output")
		if err != nil {
			return fmt.Errorf("failed to export result, details: %s", err.Error())
		}
		_, err = con.Directory("/app/input").Export(ctx, "./input")
		if err != nil {
			return fmt.Errorf("failed to export result, details: %s", err.Error())
		}
	}

	content, err := con.File("/app/output/result.csv").Contents(context.Background())
	if err != nil {
		updateTaskStatus(taskID, "Running", "There is no result.csv file in /app/output, please check whether it is nomal")
	} else {
		results, err := readCSVContent2Result(content, executionResult.ID)
		if err != nil {
			return fmt.Errorf("convert result.csv to database struct failed: %v", err)
		}
		if err := database.DB.Create(&results).Error; err != nil {
			return fmt.Errorf("save result.csv to database failed: %v", err)
		}
	}

	conclusion, err := con.File("/app/output/conclusion.csv").Contents(context.Background())
	if err != nil {
		updateTaskStatus(taskID, "Running", "There is no conclusion.csv file in /app/output, please check whether it is nomal")

	} else {
		results, err := readDetectorCSV(conclusion, executionResult.ID)
		if err != nil {
			return fmt.Errorf("convert result.csv to database struct failed: %v", err)
		}
		fmt.Println(results)
		if err := database.DB.Create(&results).Error; err != nil {
			return fmt.Errorf("save conclusion.csv to database failed: %v", err)
		}
	}
	return nil
}

// 解析算法执行任务的 Payload
func parseAlgorithmExecutionPayload(payload map[string]interface{}) (*AlgorithmExecutionPayload, error) {
	benchmark, ok := payload[EvalPayloadBench].(string)
	if !ok || benchmark == "" {
		return nil, fmt.Errorf("missing or invalid '%s' key in payload", EvalPayloadBench)
	}
	algorithm, ok := payload[EvalPayloadAlgo].(string)
	if !ok || algorithm == "" {
		return nil, fmt.Errorf("missing or invalid '%s' key in payload", EvalPayloadAlgo)
	}
	datasetName, ok := payload[EvalPayloadDataset].(string)
	if !ok || datasetName == "" {
		return nil, fmt.Errorf("missing or invalid '%s' key in payload", EvalPayloadDataset)
	}
	return &AlgorithmExecutionPayload{
		Benchmark:   benchmark,
		Algorithm:   algorithm,
		DatasetName: datasetName,
	}, nil
}

// 读取 CSV 内容并转换为结果
func readCSVContent2Result(csvContent string, executionID int) ([]database.GranularityResult, error) {
	reader := csv.NewReader(strings.NewReader(csvContent))

	// 读取表头
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %v", err)
	}

	expectedHeader := []string{"level", "result", "rank", "confidence"}
	if len(header) != len(expectedHeader) {
		return nil, fmt.Errorf("unexpected header length: got %d, expected %d", len(header), len(expectedHeader))
	}
	for i, field := range header {
		if field != expectedHeader[i] {
			return nil, fmt.Errorf("unexpected header field at column %d: got '%s', expected '%s'", i+1, field, expectedHeader[i])
		}
	}

	rows, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV rows: %v", err)
	}

	var results []database.GranularityResult
	for i, row := range rows {
		if len(row) != len(expectedHeader) {
			return nil, fmt.Errorf("row %d has incorrect number of columns: got %d, expected %d", i+1, len(row), len(expectedHeader))
		}

		level := row[0]
		result := row[1]
		rank, err := strconv.Atoi(row[2])
		if err != nil {
			return nil, fmt.Errorf("invalid rank value in row %d: %v", i+1, err)
		}
		confidence, err := strconv.ParseFloat(row[3], 64)
		if err != nil {
			return nil, fmt.Errorf("invalid confidence value in row %d: %v", i+1, err)
		}

		results = append(results, database.GranularityResult{
			ExecutionID: executionID,
			Level:       level,
			Result:      result,
			Rank:        rank,
			Confidence:  confidence,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		})
	}

	return results, nil
}

func readDetectorCSV(csvContent string, executionID int) ([]database.Detector, error) {
	reader := csv.NewReader(strings.NewReader(csvContent))

	// 读取表头
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %v", err)
	}

	expectedHeader := []string{"SpanName", "Issues", "AvgDuration", "SuccRate", "P90", "P95", "P99"}
	if len(header) != len(expectedHeader) {
		return nil, fmt.Errorf("unexpected header length: got %d, expected %d", len(header), len(expectedHeader))
	}
	for i, field := range header {
		if field != expectedHeader[i] {
			return nil, fmt.Errorf("unexpected header field at column %d: got '%s', expected '%s'", i+1, field, expectedHeader[i])
		}
	}

	// 读取所有行
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV rows: %v", err)
	}

	var results []database.Detector
	for i, row := range rows {
		if len(row) != len(expectedHeader) {
			return nil, fmt.Errorf("row %d has incorrect number of columns: got %d, expected %d", i+1, len(row), len(expectedHeader))
		}

		spanName := row[0]
		issues := row[1]

		// 处理空值
		var avgDuration, succRate, p90, p95, p99 *float64

		// 如果字段非空，转换为 float64，否则设置为 nil
		if row[2] != "" {
			val, err := strconv.ParseFloat(row[2], 64)
			if err != nil {
				return nil, fmt.Errorf("invalid AvgDuration value in row %d: %v", i+1, err)
			}
			avgDuration = &val
		}
		if row[3] != "" {
			val, err := strconv.ParseFloat(row[3], 64)
			if err != nil {
				return nil, fmt.Errorf("invalid SuccRate value in row %d: %v", i+1, err)
			}
			succRate = &val
		}
		if row[4] != "" {
			val, err := strconv.ParseFloat(row[4], 64)
			if err != nil {
				return nil, fmt.Errorf("invalid P90 value in row %d: %v", i+1, err)
			}
			p90 = &val
		}
		if row[5] != "" {
			val, err := strconv.ParseFloat(row[5], 64)
			if err != nil {
				return nil, fmt.Errorf("invalid P95 value in row %d: %v", i+1, err)
			}
			p95 = &val
		}
		if row[6] != "" {
			val, err := strconv.ParseFloat(row[6], 64)
			if err != nil {
				return nil, fmt.Errorf("invalid P99 value in row %d: %v", i+1, err)
			}
			p99 = &val
		}

		// 将数据添加到结果
		results = append(results, database.Detector{
			ExecutionID: executionID,
			SpanName:    spanName,
			Issues:      issues,
			AvgDuration: avgDuration,
			SuccRate:    succRate,
			P90:         p90,
			P95:         p95,
			P99:         p99,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		})
	}

	return results, nil
}
