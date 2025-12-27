//	@title			RCABench API
//	@version		0.0.0
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
	"time"

	"aegis/client"
	"aegis/client/k8s"
	"aegis/config"
	"aegis/consts"
	"aegis/database"
	"aegis/router"
	"aegis/service"
	"aegis/service/consumer"
	"aegis/utils"

	chaosCli "github.com/LGU-SE-Internal/chaos-experiment/client"
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

func initChaosExperiment() {
	k8sConfig := k8s.GetK8sRestConfig()
	if err := chaosCli.InitWithConfig(k8sConfig); err != nil {
		logrus.Fatalf("failed to initialize chaos experiment client: %v", err)
	}
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

	// Producer command - runs HTTP server for API endpoints
	var producerCmd = &cobra.Command{
		Use:   "producer",
		Short: "Run as a producer",
		Run: func(cmd *cobra.Command, args []string) {
			logrus.Println("Running as producer")
			database.InitDB()
			service.InitializeProducer()

			utils.InitValidator()
			client.InitTraceProvider()
			initChaosExperiment()

			engine := router.New()
			port := viper.GetString("port")
			if err := engine.Run(":" + port); err != nil {
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
			consts.InitialTime = utils.TimePtr(time.Now())
			consts.AppID = utils.GenerateULID(consts.InitialTime)

			k8slogger.SetLogger(stdr.New(log.New(os.Stdout, "", log.LstdFlags)))
			initChaosExperiment()
			go k8s.GetK8sController().Initialize(ctx, cancel, consumer.NewHandler())

			database.InitDB()
			service.InitializeConsumer()
			service.InitConcurrencyLock(ctx)

			client.InitTraceProvider()

			go consumer.StartScheduler(ctx)
			consumer.ConsumeTasks(ctx)
		},
	}

	// Both subcommand
	var bothCmd = &cobra.Command{
		Use:   "both",
		Short: "Run as both producer and consumer",
		Run: func(cmd *cobra.Command, args []string) {
			logrus.Println("Running as both producer and consumer")
			consts.InitialTime = utils.TimePtr(time.Now())
			consts.AppID = utils.GenerateULID(consts.InitialTime)

			k8slogger.SetLogger(stdr.New(log.New(os.Stdout, "", log.LstdFlags)))
			initChaosExperiment()
			go k8s.GetK8sController().Initialize(ctx, cancel, consumer.NewHandler())

			database.InitDB()
			service.InitializeProducer()
			service.InitializeConsumer()
			service.InitConcurrencyLock(ctx)

			utils.InitValidator()
			client.InitTraceProvider()

			go consumer.StartScheduler(ctx)
			go consumer.ConsumeTasks(ctx)

			engine := router.New()
			port := viper.GetString("port")
			if err := engine.Run(":" + port); err != nil {
				panic(err)
			}
		},
	}

	rootCmd.AddCommand(producerCmd, consumerCmd, bothCmd)
	if err := rootCmd.Execute(); err != nil {
		logrus.Println(err.Error())
		os.Exit(1)
	}
}
