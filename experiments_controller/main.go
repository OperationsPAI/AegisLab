package main

import (
	"dagger/rcabench/config"
	"dagger/rcabench/database"
	"dagger/rcabench/executor"
	"dagger/rcabench/router"
	"log"
	"os"

	"github.com/go-logr/stdr"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	k8slogger "sigs.k8s.io/controller-runtime/pkg/log"
)

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

	var producerCmd = &cobra.Command{
		Use:   "producer",
		Short: "Run as a producer",
		Run: func(cmd *cobra.Command, args []string) {
			logrus.Println("Running as producer")
			database.InitDB()
			engine := router.New()
			err := engine.Run(":" + port)
			if err != nil {
				panic(err)
			}
		},
	}

	var consumerCmd = &cobra.Command{
		Use:   "consumer",
		Short: "Run as a consumer",
		Run: func(cmd *cobra.Command, args []string) {
			database.InitDB()
			k8slogger.SetLogger(stdr.New(log.New(os.Stdout, "", log.LstdFlags)))
			logrus.Println("Running as consumer")
			executor.ConsumeTasks()
		},
	}

	var bothCmd = &cobra.Command{
		Use:   "both",
		Short: "Run as both producer and consumer",
		Run: func(cmd *cobra.Command, args []string) {
			logrus.Println("Running as both producer and consumer")
			k8slogger.SetLogger(stdr.New(log.New(os.Stdout, "", log.LstdFlags)))
			engine := router.New()
			database.InitDB()
			go executor.ConsumeTasks()
			err := engine.Run(":" + port)
			if err != nil {
				panic(err)
			}
		},
	}
	rootCmd.PersistentFlags().StringVarP(&port, "port", "p", "8080", "Port to run the server on")
	rootCmd.PersistentFlags().StringVarP(&conf, "conf", "c", "./config.toml", "database path")
	config.Init(conf)

	rootCmd.AddCommand(producerCmd, consumerCmd, bothCmd)

	if err := rootCmd.Execute(); err != nil {
		logrus.Println(err)
		os.Exit(1)
	}
}
