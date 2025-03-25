package main

import (
	"eval-tool/pkg/executor"
	"eval-tool/pkg/util"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"

	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	inputFile  string
	outputPath string
	service    string
)

var rootCmd = &cobra.Command{
	Use:   "eval",
	Short: "eval is a tool to evaluate the result.csv produced by rcabench",
	RunE: func(cmd *cobra.Command, args []string) error {
		return processData(inputFile, outputPath, service)
	},
}

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

	rootCmd.PersistentFlags().StringVarP(&inputFile, "input", "i", "", "result.csv file path (required)")
	rootCmd.PersistentFlags().StringVarP(&outputPath, "output", "o", "", "custom output path (optional)")
	rootCmd.PersistentFlags().StringVarP(&service, "service", "s", "", "injection service name (required)")

	rootCmd.MarkPersistentFlagRequired("input")
	rootCmd.MarkPersistentFlagRequired("service")
}

func processData(inputFile, outputPath, service string) error {
	recordsPtr, err := util.ReadCSVFile(inputFile)
	if err != nil {
		logrus.Error(err)
		return err
	}

	records := *recordsPtr
	granularityResults, err := executor.GetGranularityResults(records)
	if err != nil {
		logrus.Error(err)
		return err
	}

	results := make(map[string][]*executor.Conclusion, len(executor.GetMetrics()))

	for metric, evalFunc := range executor.GetMetrics() {
		conclusions, err := evalFunc(granularityResults, service)
		if err != nil {
			logrus.Errorf("calculate metric %s failed: %v", metric, err)
		}

		results[metric] = conclusions
	}

	workspace, err := os.Getwd()
	if err != nil {
		logrus.Errorf("failed to get the current director: %v", err)
		return err
	}

	if outputPath == "" {
		outputPath = filepath.Join(workspace, "eval-tool", "output", fmt.Sprintf("%s.json", uuid.New().String()))
	}
	if !util.IsJSONByExt(outputPath) {
		errorMsg := "output file must be a JSON file"
		logrus.Errorf(errorMsg)
		return fmt.Errorf(errorMsg)
	}

	util.OutputToJSON(results, outputPath)

	return nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		logrus.Error("failed to run")
		os.Exit(1)
	}
}
