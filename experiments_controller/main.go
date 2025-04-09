package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path"
	"runtime"

	"github.com/CUHK-SE-Group/chaos-experiment/client"
	"github.com/CUHK-SE-Group/rcabench/client/k8s"
	"github.com/CUHK-SE-Group/rcabench/config"
	"github.com/CUHK-SE-Group/rcabench/database"
	_ "github.com/CUHK-SE-Group/rcabench/docs"
	"github.com/CUHK-SE-Group/rcabench/executor"
	"github.com/CUHK-SE-Group/rcabench/router"
	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/go-logr/stdr"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	k8slogger "sigs.k8s.io/controller-runtime/pkg/log"
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
	logrus.Info("Logger initialized")
}

func main() {
	var port string
	var conf string
	var rootCmd = &cobra.Command{
		Use:   "rcabench",
		Short: "RCA Bench is a benchmarking tool",
		Run: func(cmd *cobra.Command, args []string) {
			logrus.Println("Please specify a mode: producer, consumer, or both")
		},
	}

	rootCmd.PersistentFlags().StringVarP(&port, "port", "p", "8080", "Port to run the server on")
	rootCmd.PersistentFlags().StringVarP(&conf, "conf", "c", "/etc/rcabench/config.prod.toml", "Path to configuration file")

	if err := viper.BindPFlag("port", rootCmd.PersistentFlags().Lookup("port")); err != nil {
		logrus.Fatalf("Failed to bind flag: %v", err)
	}
	if err := viper.BindPFlag("conf", rootCmd.PersistentFlags().Lookup("conf")); err != nil {
		logrus.Fatalf("Failed to bind flag: %v", err)
	}

	config.Init(viper.GetString("conf"))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Producer 子命令
	var producerCmd = &cobra.Command{
		Use:   "producer",
		Short: "Run as a producer",
		Run: func(cmd *cobra.Command, args []string) {
			logrus.Println("Running as producer")
			database.InitDB()
			engine := router.New()
			port := viper.GetString("port") // 从 Viper 获取最终端口
			err := engine.Run(":" + port)
			if err != nil {
				panic(err)
			}
		},
	}

	// Consumer 子命令
	var consumerCmd = &cobra.Command{
		Use:   "consumer",
		Short: "Run as a consumer",
		Run: func(cmd *cobra.Command, args []string) {
			database.InitDB()
			k8slogger.SetLogger(stdr.New(log.New(os.Stdout, "", log.LstdFlags)))
			logrus.Println("Running as consumer")
			go k8s.Init(ctx, executor.Exec)
			go executor.StartScheduler(ctx)
			executor.ConsumeTasks()
		},
	}

	// Both 子命令
	var bothCmd = &cobra.Command{
		Use:   "both",
		Short: "Run as both producer and consumer",
		Run: func(cmd *cobra.Command, args []string) {
			logrus.Println("Running as both producer and consumer")
			k8slogger.SetLogger(stdr.New(log.New(os.Stdout, "", log.LstdFlags)))
			engine := router.New()
			database.InitDB()
			go k8s.Init(ctx, executor.Exec)
			go executor.ConsumeTasks()
			go executor.StartScheduler(ctx)
			port := viper.GetString("port") // 从 Viper 获取最终端口
			err := engine.Run(":" + port)
			if err != nil {
				panic(err)
			}
		},
	}
	client.NewK8sClient()
	rootCmd.AddCommand(producerCmd, consumerCmd, bothCmd)
	if err := rootCmd.Execute(); err != nil {
		logrus.Println(err.Error())
		os.Exit(1)
	}
}
