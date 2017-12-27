package cmd

import (
	"github.com/spf13/cobra"
)

const taskLogGroupFormat = "/fargate/task/%s"

var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "Run and manage one-time executions of Docker containers",
}

func init() {
	rootCmd.AddCommand(taskCmd)
}
