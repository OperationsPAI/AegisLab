package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"runtime"

	"aegis/config"
	"aegis/consts"
	"aegis/database"
	"aegis/dto"
	"aegis/service/common"

	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

func init() {
	logrus.SetReportCaller(true)
	logrus.SetFormatter(&nested.Formatter{
		CustomCallerFormatter: func(f *runtime.Frame) string {
			filename := path.Base(f.File)
			return fmt.Sprintf(" (%s:%d)", filename, f.Line)
		},
		FieldsOrder:     []string{"component", "category"},
		HideKeys:        true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
	logrus.SetLevel(logrus.InfoLevel)
	logrus.Info("Task submitter initialized")
}

func main() {
	var conf string
	var rootCmd = &cobra.Command{
		Use:   "task-submitter",
		Short: "Submit tasks to RCABench task queue",
		Long:  "A command-line tool to submit various types of tasks to the RCABench task queue system",
	}

	rootCmd.PersistentFlags().StringVarP(&conf, "conf", "c", "/etc/rcabench/config.prod.toml", "Path to configuration file")
	if err := viper.BindPFlag("conf", rootCmd.PersistentFlags().Lookup("conf")); err != nil {
		logrus.Fatalf("failed to bind flag: %v", err)
	}

	// Initialize configuration and database for all commands
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		config.Init(viper.GetString("conf"))
		database.InitDB()
		// Redis client is initialized lazily via GetRedisClient()
	}

	// Build Container Task Command
	var buildContainerCmd = &cobra.Command{
		Use:   "build-container",
		Short: "Submit a build container task",
		Long:  "Submit a task to build a container image from source code",
		Run:   submitBuildContainerTask,
	}
	buildContainerCmd.Flags().String("image-ref", "", "Container image reference (required)")
	buildContainerCmd.Flags().String("source-path", "", "Path to source code (required)")
	buildContainerCmd.Flags().String("dockerfile", "Dockerfile", "Path to Dockerfile")
	buildContainerCmd.Flags().StringSlice("build-args", []string{}, "Build arguments (key=value)")
	buildContainerCmd.Flags().Bool("immediate", true, "Execute immediately")
	buildContainerCmd.Flags().Int("project-id", 0, "Project ID (required)")
	buildContainerCmd.Flags().Int("user-id", 0, "User ID (required)")
	buildContainerCmd.MarkFlagRequired("image-ref")
	buildContainerCmd.MarkFlagRequired("source-path")
	buildContainerCmd.MarkFlagRequired("project-id")
	buildContainerCmd.MarkFlagRequired("user-id")

	// Restart Pedestal Task Command
	var restartPedestalCmd = &cobra.Command{
		Use:   "restart-pedestal",
		Short: "Submit a restart pedestal task",
		Long:  "Submit a task to restart a pedestal (benchmark environment)",
		Run:   submitRestartPedestalTask,
	}
	restartPedestalCmd.Flags().Int("pedestal-id", 0, "Pedestal ID (required)")
	restartPedestalCmd.Flags().String("namespace", "", "Kubernetes namespace (required)")
	restartPedestalCmd.Flags().Bool("immediate", true, "Execute immediately")
	restartPedestalCmd.Flags().Int("user-id", 0, "User ID")
	restartPedestalCmd.MarkFlagRequired("pedestal-id")
	restartPedestalCmd.MarkFlagRequired("namespace")

	// Fault Injection Task Command
	var faultInjectionCmd = &cobra.Command{
		Use:   "fault-injection",
		Short: "Submit a fault injection task",
		Long:  "Submit a task to inject a fault into the system",
		Run:   submitFaultInjectionTask,
	}
	faultInjectionCmd.Flags().String("node-json", "", "Chaos node configuration as JSON (required)")
	faultInjectionCmd.Flags().Int("benchmark-id", 0, "Benchmark ID (required)")
	faultInjectionCmd.Flags().String("namespace", "", "Kubernetes namespace (required)")
	faultInjectionCmd.Flags().Int("pedestal-id", 0, "Pedestal ID (required)")
	faultInjectionCmd.Flags().Int("pre-duration", 60, "Pre-injection duration in seconds")
	faultInjectionCmd.Flags().Bool("immediate", true, "Execute immediately")
	faultInjectionCmd.Flags().Int("user-id", 0, "User ID")
	faultInjectionCmd.MarkFlagRequired("node-json")
	faultInjectionCmd.MarkFlagRequired("benchmark-id")
	faultInjectionCmd.MarkFlagRequired("namespace")
	faultInjectionCmd.MarkFlagRequired("pedestal-id")

	// Build Datapack Task Command
	var buildDatapackCmd = &cobra.Command{
		Use:   "build-datapack",
		Short: "Submit a build datapack task",
		Long:  "Submit a task to build a datapack from collected data",
		Run:   submitBuildDatapackTask,
	}
	buildDatapackCmd.Flags().Int("injection-id", 0, "Injection ID (required)")
	buildDatapackCmd.Flags().String("namespace", "", "Kubernetes namespace (required)")
	buildDatapackCmd.Flags().Bool("immediate", true, "Execute immediately")
	buildDatapackCmd.Flags().Int("project-id", 0, "Project ID (required)")
	buildDatapackCmd.Flags().Int("user-id", 0, "User ID (required)")
	buildDatapackCmd.MarkFlagRequired("injection-id")
	buildDatapackCmd.MarkFlagRequired("namespace")
	buildDatapackCmd.MarkFlagRequired("project-id")
	buildDatapackCmd.MarkFlagRequired("user-id")

	// Run Algorithm Task Command
	var runAlgorithmCmd = &cobra.Command{
		Use:   "run-algorithm",
		Short: "Submit an algorithm execution task",
		Long:  "Submit a task to run an RCA algorithm on a datapack",
		Run:   submitRunAlgorithmTask,
	}
	runAlgorithmCmd.Flags().Int("execution-id", 0, "Execution ID (required)")
	runAlgorithmCmd.Flags().String("algorithm-image", "", "Algorithm container image (required)")
	runAlgorithmCmd.Flags().Int("injection-id", 0, "Injection ID (required)")
	runAlgorithmCmd.Flags().String("namespace", "", "Kubernetes namespace (required)")
	runAlgorithmCmd.Flags().Bool("immediate", true, "Execute immediately")
	runAlgorithmCmd.Flags().Int("user-id", 0, "User ID")
	runAlgorithmCmd.MarkFlagRequired("execution-id")
	runAlgorithmCmd.MarkFlagRequired("algorithm-image")
	runAlgorithmCmd.MarkFlagRequired("injection-id")
	runAlgorithmCmd.MarkFlagRequired("namespace")

	// Collect Result Task Command
	var collectResultCmd = &cobra.Command{
		Use:   "collect-result",
		Short: "Submit a result collection task",
		Long:  "Submit a task to collect and process algorithm execution results",
		Run:   submitCollectResultTask,
	}
	collectResultCmd.Flags().Int("execution-id", 0, "Execution ID (required)")
	collectResultCmd.Flags().String("namespace", "", "Kubernetes namespace (required)")
	collectResultCmd.Flags().Bool("immediate", true, "Execute immediately")
	collectResultCmd.Flags().Int("user-id", 0, "User ID")
	collectResultCmd.MarkFlagRequired("execution-id")
	collectResultCmd.MarkFlagRequired("namespace")

	rootCmd.AddCommand(
		buildContainerCmd,
		restartPedestalCmd,
		faultInjectionCmd,
		buildDatapackCmd,
		runAlgorithmCmd,
		collectResultCmd,
	)

	if err := rootCmd.Execute(); err != nil {
		logrus.Println(err.Error())
		os.Exit(1)
	}
}

func submitBuildContainerTask(cmd *cobra.Command, args []string) {
	ctx := context.Background()

	imageRef, _ := cmd.Flags().GetString("image-ref")
	sourcePath, _ := cmd.Flags().GetString("source-path")
	dockerfile, _ := cmd.Flags().GetString("dockerfile")
	buildArgsSlice, _ := cmd.Flags().GetStringSlice("build-args")
	immediate, _ := cmd.Flags().GetBool("immediate")
	userID, _ := cmd.Flags().GetInt("user-id")

	buildArgs := make(map[string]string)
	for _, arg := range buildArgsSlice {
		key, value, found := parseKeyValue(arg)
		if found {
			buildArgs[key] = value
		}
	}

	payload := map[string]any{
		"imageRef":   imageRef,
		"sourcePath": sourcePath,
		"buildOptions": map[string]any{
			"dockerfile": dockerfile,
			"buildArgs":  buildArgs,
		},
	}

	task := &dto.UnifiedTask{
		TaskID:    uuid.NewString(),
		Type:      consts.TaskTypeBuildContainer,
		Immediate: immediate,
		RetryPolicy: dto.RetryPolicy{
			MaxAttempts: 3,
			BackoffSec:  10,
		},
		Payload: payload,
		TraceID: uuid.NewString(),
		GroupID: uuid.NewString(),
		UserID:  userID,
		State:   consts.TaskPending,
	}

	setGroupCarrier(ctx, task)

	if err := common.SubmitTask(ctx, task); err != nil {
		logrus.Fatalf("Failed to submit build container task: %v", err)
	}

	logrus.Infof("Successfully submitted build container task: %s", task.TaskID)
	fmt.Printf("Task ID: %s\nTrace ID: %s\nGroup ID: %s\n", task.TaskID, task.TraceID, task.GroupID)
}

func submitRestartPedestalTask(cmd *cobra.Command, args []string) {
	ctx := context.Background()

	pedestalID, _ := cmd.Flags().GetInt("pedestal-id")
	namespace, _ := cmd.Flags().GetString("namespace")
	immediate, _ := cmd.Flags().GetBool("immediate")
	userID, _ := cmd.Flags().GetInt("user-id")

	payload := map[string]any{
		"pedestalID": pedestalID,
		"namespace":  namespace,
	}

	task := &dto.UnifiedTask{
		TaskID:    uuid.NewString(),
		Type:      consts.TaskTypeRestartPedestal,
		Immediate: immediate,
		RetryPolicy: dto.RetryPolicy{
			MaxAttempts: 1,
			BackoffSec:  10,
		},
		Payload: payload,
		TraceID: uuid.NewString(),
		GroupID: uuid.NewString(),
		UserID:  userID,
		State:   consts.TaskPending,
	}

	setGroupCarrier(ctx, task)

	if err := common.SubmitTask(ctx, task); err != nil {
		logrus.Fatalf("Failed to submit restart pedestal task: %v", err)
	}

	logrus.Infof("Successfully submitted restart pedestal task: %s", task.TaskID)
	fmt.Printf("Task ID: %s\nTrace ID: %s\nGroup ID: %s\n", task.TaskID, task.TraceID, task.GroupID)
}

func submitFaultInjectionTask(cmd *cobra.Command, args []string) {
	ctx := context.Background()

	nodeJSON, _ := cmd.Flags().GetString("node-json")
	benchmarkID, _ := cmd.Flags().GetInt("benchmark-id")
	namespace, _ := cmd.Flags().GetString("namespace")
	pedestalID, _ := cmd.Flags().GetInt("pedestal-id")
	preDuration, _ := cmd.Flags().GetInt("pre-duration")
	immediate, _ := cmd.Flags().GetBool("immediate")
	userID, _ := cmd.Flags().GetInt("user-id")

	var node map[string]any
	if err := json.Unmarshal([]byte(nodeJSON), &node); err != nil {
		logrus.Fatalf("Failed to parse node JSON: %v", err)
	}

	payload := map[string]any{
		"benchmark": map[string]any{
			"id": benchmarkID,
		},
		"preDuration": preDuration,
		"node":        node,
		"namespace":   namespace,
		"pedestalID":  pedestalID,
		"labels":      []dto.LabelItem{},
	}

	task := &dto.UnifiedTask{
		TaskID:    uuid.NewString(),
		Type:      consts.TaskTypeFaultInjection,
		Immediate: immediate,
		RetryPolicy: dto.RetryPolicy{
			MaxAttempts: 3,
			BackoffSec:  10,
		},
		Payload: payload,
		TraceID: uuid.NewString(),
		GroupID: uuid.NewString(),
		UserID:  userID,
		State:   consts.TaskPending,
	}

	setGroupCarrier(ctx, task)

	if err := common.SubmitTask(ctx, task); err != nil {
		logrus.Fatalf("Failed to submit fault injection task: %v", err)
	}

	logrus.Infof("Successfully submitted fault injection task: %s", task.TaskID)
	fmt.Printf("Task ID: %s\nTrace ID: %s\nGroup ID: %s\n", task.TaskID, task.TraceID, task.GroupID)
}

func submitBuildDatapackTask(cmd *cobra.Command, args []string) {
	ctx := context.Background()

	injectionID, _ := cmd.Flags().GetInt("injection-id")
	namespace, _ := cmd.Flags().GetString("namespace")
	immediate, _ := cmd.Flags().GetBool("immediate")
	projectID, _ := cmd.Flags().GetInt("project-id")
	userID, _ := cmd.Flags().GetInt("user-id")

	// Query injection from database to get benchmark and datapack info
	injection, err := getInjectionInfo(injectionID)
	if err != nil {
		logrus.Fatalf("Failed to get injection info: %v", err)
	}

	payload := map[string]any{
		consts.BuildBenchmark:        injection.Benchmark,
		consts.BuildDatapack:         injection.Datapack,
		consts.BuildDatasetVersionID: nil,
		consts.BuildLabels:           []dto.LabelItem{},
		consts.InjectNamespace:       namespace,
	}

	task := &dto.UnifiedTask{
		TaskID:    uuid.NewString(),
		Type:      consts.TaskTypeBuildDatapack,
		Immediate: immediate,
		RetryPolicy: dto.RetryPolicy{
			MaxAttempts: 3,
			BackoffSec:  10,
		},
		Payload:   payload,
		TraceID:   uuid.NewString(),
		GroupID:   uuid.NewString(),
		ProjectID: projectID,
		UserID:    userID,
		State:     consts.TaskPending,
	}

	setGroupCarrier(ctx, task)

	if err := common.SubmitTask(ctx, task); err != nil {
		logrus.Fatalf("Failed to submit build datapack task: %v", err)
	}

	logrus.Infof("Successfully submitted build datapack task: %s", task.TaskID)
	fmt.Printf("Task ID: %s\nTrace ID: %s\nGroup ID: %s\n", task.TaskID, task.TraceID, task.GroupID)
}

func submitRunAlgorithmTask(cmd *cobra.Command, args []string) {
	ctx := context.Background()

	executionID, _ := cmd.Flags().GetInt("execution-id")
	algorithmImage, _ := cmd.Flags().GetString("algorithm-image")
	injectionID, _ := cmd.Flags().GetInt("injection-id")
	namespace, _ := cmd.Flags().GetString("namespace")
	immediate, _ := cmd.Flags().GetBool("immediate")
	userID, _ := cmd.Flags().GetInt("user-id")

	payload := map[string]any{
		"executionID": executionID,
		"algorithm": map[string]any{
			"image": algorithmImage,
		},
		"injectionID": injectionID,
		"namespace":   namespace,
	}

	task := &dto.UnifiedTask{
		TaskID:    uuid.NewString(),
		Type:      consts.TaskTypeRunAlgorithm,
		Immediate: immediate,
		RetryPolicy: dto.RetryPolicy{
			MaxAttempts: 3,
			BackoffSec:  10,
		},
		Payload: payload,
		TraceID: uuid.NewString(),
		GroupID: uuid.NewString(),
		UserID:  userID,
		State:   consts.TaskPending,
	}

	setGroupCarrier(ctx, task)

	if err := common.SubmitTask(ctx, task); err != nil {
		logrus.Fatalf("Failed to submit run algorithm task: %v", err)
	}

	logrus.Infof("Successfully submitted run algorithm task: %s", task.TaskID)
	fmt.Printf("Task ID: %s\nTrace ID: %s\nGroup ID: %s\n", task.TaskID, task.TraceID, task.GroupID)
}

func submitCollectResultTask(cmd *cobra.Command, args []string) {
	ctx := context.Background()

	executionID, _ := cmd.Flags().GetInt("execution-id")
	namespace, _ := cmd.Flags().GetString("namespace")
	immediate, _ := cmd.Flags().GetBool("immediate")
	userID, _ := cmd.Flags().GetInt("user-id")

	payload := map[string]any{
		"executionID": executionID,
		"namespace":   namespace,
	}

	task := &dto.UnifiedTask{
		TaskID:    uuid.NewString(),
		Type:      consts.TaskTypeCollectResult,
		Immediate: immediate,
		RetryPolicy: dto.RetryPolicy{
			MaxAttempts: 3,
			BackoffSec:  10,
		},
		Payload: payload,
		TraceID: uuid.NewString(),
		GroupID: uuid.NewString(),
		UserID:  userID,
		State:   consts.TaskPending,
	}

	setGroupCarrier(ctx, task)

	if err := common.SubmitTask(ctx, task); err != nil {
		logrus.Fatalf("Failed to submit collect result task: %v", err)
	}

	logrus.Infof("Successfully submitted collect result task: %s", task.TaskID)
	fmt.Printf("Task ID: %s\nTrace ID: %s\nGroup ID: %s\n", task.TaskID, task.TraceID, task.GroupID)
}

// Helper functions

type injectionInfo struct {
	Benchmark dto.ContainerVersionItem
	Datapack  dto.InjectionItem
}

func getInjectionInfo(injectionID int) (*injectionInfo, error) {
	var injection database.FaultInjection
	if err := database.DB.
		Preload("Task").
		Preload("Benchmark").
		Preload("Benchmark.Container").
		Where("id = ?", injectionID).
		First(&injection).Error; err != nil {
		return nil, fmt.Errorf("failed to find injection with id %d: %w", injectionID, err)
	}

	if injection.Benchmark == nil {
		return nil, fmt.Errorf("injection %d has no associated benchmark", injectionID)
	}

	benchmarkItem := dto.NewContainerVersionItem(injection.Benchmark)
	datapackItem := dto.NewInjectionItem(&injection)

	return &injectionInfo{
		Benchmark: benchmarkItem,
		Datapack:  datapackItem,
	}, nil
}

func setGroupCarrier(ctx context.Context, task *dto.UnifiedTask) {
	task.GroupCarrier = make(propagation.MapCarrier)
	otel.GetTextMapPropagator().Inject(ctx, task.GroupCarrier)
}

func parseKeyValue(s string) (string, string, bool) {
	for i := 0; i < len(s); i++ {
		if s[i] == '=' {
			return s[:i], s[i+1:], true
		}
	}
	return "", "", false
}
