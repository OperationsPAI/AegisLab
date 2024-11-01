package main

import (
	"dagger/rcabench/executor"
	"dagger/rcabench/router"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "rcabench",
		Short: "RCA Bench is a benchmarking tool",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Please specify a mode: producer, consumer, or both")
		},
	}

	var producerCmd = &cobra.Command{
		Use:   "producer",
		Short: "Run as a producer",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Running as producer")
			engine := router.New()
			err := engine.Run(":8080")
			if err != nil {
				panic(err)
			}
		},
	}

	var consumerCmd = &cobra.Command{
		Use:   "consumer",
		Short: "Run as a consumer",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Running as consumer")
			executor.ConsumeTasks()
		},
	}

	var bothCmd = &cobra.Command{
		Use:   "both",
		Short: "Run as both producer and consumer",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Running as both producer and consumer")
			engine := router.New()
			go executor.ConsumeTasks()
			err := engine.Run(":8080")
			if err != nil {
				panic(err)
			}
		},
	}
	rootCmd.AddCommand(producerCmd, consumerCmd, bothCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
