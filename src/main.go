//	@title			RCABench API
//	@version		1.1.44
//	@description	RCABench - A comprehensive root cause analysis benchmarking platform for microservices

//	@contact.name	RCABench Team
//	@contact.email	team@rcabench.com

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

//	@host	http://localhost:8082

//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Type "Bearer" followed by a space and JWT token.

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path"
	"runtime"

	"aegis/client"
	"aegis/client/k8s"
	"aegis/config"
	"aegis/database"
	"aegis/router"
	"aegis/service"
	"aegis/service/consumer"
	"aegis/utils"

	cli "github.com/LGU-SE-Internal/chaos-experiment/client"
	chaos "github.com/LGU-SE-Internal/chaos-experiment/handler"
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
	logrus.SetLevel(logrus.InfoLevel)
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

	nsTargetMap, err := utils.GetNsCountMap()
	logrus.Infof("initalized nsTargetMap: %v", nsTargetMap)
	if err != nil {
		logrus.Fatal(err)
	}

	targetLabelKey := config.GetString("injection.target_label_key")
	if err := chaos.InitTargetConfig(nsTargetMap, targetLabelKey); err != nil {
		logrus.Fatal(err)
	}

	// Producer command - runs HTTP server for API endpoints
	var producerCmd = &cobra.Command{
		Use:   "producer",
		Short: "Run as a producer",
		Run: func(cmd *cobra.Command, args []string) {
			logrus.Println("Running as producer")
			database.InitDB()
			service.InitializeData(ctx)

			utils.InitValidator()
			client.InitTraceProvider()

			engine := router.New()
			port := viper.GetString("port")
			err := engine.Run(":" + port)
			if err != nil {
				panic(err)
			}
		},
	}

	// Consumer command - runs background workers and Kubernetes controllers
	var consumerCmd = &cobra.Command{
		Use:   "consumer",
		Short: "Run as a consumer",
		Run: func(cmd *cobra.Command, args []string) {
			logrus.Println("Running as consumer")
			k8slogger.SetLogger(stdr.New(log.New(os.Stdout, "", log.LstdFlags)))
			database.InitDB()
			service.InitConcurrencyLock(ctx)

			client.InitTraceProvider()
			go k8s.Init(ctx, consumer.NewHandler())
			go consumer.StartScheduler(ctx)
			consumer.ConsumeTasks()
		},
	}

	// Both subcommand
	var bothCmd = &cobra.Command{
		Use:   "both",
		Short: "Run as both producer and consumer",
		Run: func(cmd *cobra.Command, args []string) {
			logrus.Println("Running as both producer and consumer")
			k8slogger.SetLogger(stdr.New(log.New(os.Stdout, "", log.LstdFlags)))
			database.InitDB()
			service.InitializeData(ctx)
			service.InitConcurrencyLock(ctx)

			utils.InitValidator()
			client.InitTraceProvider()
			go k8s.Init(ctx, consumer.NewHandler())
			go consumer.StartScheduler(ctx)
			go consumer.ConsumeTasks()

			engine := router.New()
			port := viper.GetString("port")
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
