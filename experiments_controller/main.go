package main

import (
	"log"
	"os"

	"github.com/CUHK-SE-Group/rcabench/router"

	"github.com/CUHK-SE-Group/rcabench/executor"

	"github.com/CUHK-SE-Group/rcabench/database"

	"github.com/CUHK-SE-Group/rcabench/config"

	_ "github.com/CUHK-SE-Group/rcabench/docs"

	"github.com/go-logr/stdr"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

	rootCmd.PersistentFlags().StringVarP(&port, "port", "p", "8080", "Port to run the server on")
	rootCmd.PersistentFlags().StringVarP(&conf, "conf", "c", "", "Path to configuration file")

	viper.BindPFlag("port", rootCmd.PersistentFlags().Lookup("port"))
	viper.BindPFlag("conf", rootCmd.PersistentFlags().Lookup("conf"))

	config.Init(viper.GetString("conf"))

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
			go executor.ConsumeTasks()
			port := viper.GetString("port") // 从 Viper 获取最终端口
			err := engine.Run(":" + port)
			if err != nil {
				panic(err)
			}
		},
	}

	rootCmd.AddCommand(producerCmd, consumerCmd, bothCmd)

	if err := rootCmd.Execute(); err != nil {
		logrus.Println(err)
		os.Exit(1)
	}
}
