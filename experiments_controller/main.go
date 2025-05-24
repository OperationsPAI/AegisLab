// @title           RCABench API
// @version         1.0.1
// @description     RCABench - A comprehensive root cause analysis benchmarking platform for microservices
// @description     This API provides endpoints for managing datasets, algorithms, evaluations, and fault injections
// @description     for root cause analysis in distributed systems and microservices architectures.
// @contact.name    RCABench Team
// @contact.email   team@rcabench.com
// @license.name    MIT
// @license.url     https://opensource.org/licenses/MIT
// @host            localhost:8080
// @BasePath        /api/v1
// @schemes         http https
// @produce         json
// @consumes        json

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path"
	"runtime"

	cli "github.com/LGU-SE-Internal/chaos-experiment/client"
	chaos "github.com/LGU-SE-Internal/chaos-experiment/handler"
	"github.com/LGU-SE-Internal/rcabench/client"
	"github.com/LGU-SE-Internal/rcabench/client/k8s"
	"github.com/LGU-SE-Internal/rcabench/config"
	"github.com/LGU-SE-Internal/rcabench/database"
	_ "github.com/LGU-SE-Internal/rcabench/docs"
	"github.com/LGU-SE-Internal/rcabench/executor"
	"github.com/LGU-SE-Internal/rcabench/router"
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
		logrus.Fatalf("failed to bind flag: %v", err)
	}
	if err := viper.BindPFlag("conf", rootCmd.PersistentFlags().Lookup("conf")); err != nil {
		logrus.Fatalf("failed to bind flag: %v", err)
	}

	config.Init(viper.GetString("conf"))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	nsTargetMap, err := config.GetNsTargetMap()
	logrus.Infof("initalized nsTargetMap: %v", nsTargetMap)
	if err != nil {
		logrus.Fatal(err)
	}

	targetLabelKey := config.GetString("injection.target_label_key")
	if err := chaos.InitTargetConfig(nsTargetMap, targetLabelKey); err != nil {
		logrus.Fatal(err)
	}

	executor.InitConcurrencyLock(ctx)

	// Producer 子命令
	var producerCmd = &cobra.Command{
		Use:   "producer",
		Short: "Run as a producer",
		Run: func(cmd *cobra.Command, args []string) {
			logrus.Println("Running as producer")
			database.InitDB()
			client.InitTraceProvider()
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
			logrus.Println("Running as consumer")
			k8slogger.SetLogger(stdr.New(log.New(os.Stdout, "", log.LstdFlags)))
			database.InitDB()
			client.InitTraceProvider()
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
			client.InitTraceProvider()
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

	cli.NewK8sClient()
	rootCmd.AddCommand(producerCmd, consumerCmd, bothCmd)
	if err := rootCmd.Execute(); err != nil {
		logrus.Println(err.Error())
		os.Exit(1)
	}
}
