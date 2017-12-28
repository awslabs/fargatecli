package cmd

import (
	"github.com/spf13/cobra"
)

const taskLogGroupFormat = "/fargate/task/%s"

var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "Run and manage one-time executions of containers",
	Long: `Run and manage one-time executions of containers

Tasks are one-time execution of your containers on AWS Fargate. Tasks run with
the specified configuration until either manually stopped or interrupted for
any reason.`,
}

func init() {
	rootCmd.AddCommand(taskCmd)
}
