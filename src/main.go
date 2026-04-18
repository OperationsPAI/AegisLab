//	@title			RCABench API
//	@version		1.0.0
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
	"os"

	"aegis/app"
	gatewayapp "aegis/app/gateway"
	iamapp "aegis/app/iam"
	orchestratorapp "aegis/app/orchestrator"
	resourceapp "aegis/app/resource"
	runtimeapp "aegis/app/runtime"
	systemapp "aegis/app/system"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/fx"
)

func newModeCommand(use, short string, run func()) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short,
		Run: func(cmd *cobra.Command, args []string) {
			run()
		},
	}
}

func main() {
	var port string
	var conf string
	var rootCmd = &cobra.Command{
		Use:   "rcabench",
		Short: "RCA Bench is a benchmarking tool",
		Run: func(cmd *cobra.Command, args []string) {
			logrus.Println("Please specify a mode: producer, consumer, both, api-gateway, iam-service, orchestrator-service, resource-service, runtime-worker-service, or system-service")
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

	producerCmd := newModeCommand("producer", "Run as a producer", func() {
		fx.New(app.ProducerOptions(viper.GetString("conf"), viper.GetString("port"))).Run()
	})
	consumerCmd := newModeCommand("consumer", "Run as a consumer", func() {
		fx.New(app.ConsumerOptions(viper.GetString("conf"))).Run()
	})
	bothCmd := newModeCommand("both", "Run as both producer and consumer", func() {
		fx.New(app.BothOptions(viper.GetString("conf"), viper.GetString("port"))).Run()
	})
	apiGatewayCmd := newModeCommand("api-gateway", "Run as the API gateway", func() {
		fx.New(gatewayapp.Options(viper.GetString("conf"), viper.GetString("port"))).Run()
	})
	iamServiceCmd := newModeCommand("iam-service", "Run as the IAM service", func() {
		fx.New(iamapp.Options(viper.GetString("conf"))).Run()
	})
	orchestratorServiceCmd := newModeCommand("orchestrator-service", "Run as the orchestrator service", func() {
		fx.New(orchestratorapp.Options(viper.GetString("conf"))).Run()
	})
	resourceServiceCmd := newModeCommand("resource-service", "Run as the resource service", func() {
		fx.New(resourceapp.Options(viper.GetString("conf"))).Run()
	})
	runtimeWorkerServiceCmd := newModeCommand("runtime-worker-service", "Run as the runtime worker service", func() {
		fx.New(runtimeapp.Options(viper.GetString("conf"))).Run()
	})
	systemServiceCmd := newModeCommand("system-service", "Run as the system service", func() {
		fx.New(systemapp.Options(viper.GetString("conf"))).Run()
	})

	rootCmd.AddCommand(
		producerCmd,
		consumerCmd,
		bothCmd,
		apiGatewayCmd,
		iamServiceCmd,
		orchestratorServiceCmd,
		resourceServiceCmd,
		runtimeWorkerServiceCmd,
		systemServiceCmd,
	)
	if err := rootCmd.Execute(); err != nil {
		logrus.Println(err.Error())
		os.Exit(1)
	}
}
