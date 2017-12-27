package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	flagTaskLogsFilter    string
	flagTaskLogsEndTime   string
	flagTaskLogsStartTime string
	flagTaskLogsFollow    bool
	flagTaskLogsTasks     []string
)

var taskLogsCmd = &cobra.Command{
	Use:   "logs <task name>",
	Short: "View logs from a task group",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		operation := &GetLogsOperation{
			LogGroupName: fmt.Sprintf(taskLogGroupFormat, args[0]),
			Filter:       flagTaskLogsFilter,
			Follow:       flagTaskLogsFollow,
			Namespace:    args[0],
		}

		operation.AddTasks(flagTaskLogsTasks)
		operation.AddStartTime(flagTaskLogsStartTime)
		operation.AddEndTime(flagTaskLogsEndTime)

		GetLogs(operation)
	},
}

func init() {
	taskCmd.AddCommand(taskLogsCmd)

	taskLogsCmd.Flags().BoolVarP(&flagTaskLogsFollow, "follow", "f", false, "Poll logs and continuously print new events")
	taskLogsCmd.Flags().StringVar(&flagTaskLogsFilter, "filter", "", "Filter pattern to apply")
	taskLogsCmd.Flags().StringVar(&flagTaskLogsStartTime, "start", "", "Earliest time to return logs (e.g. -1h, 2018-01-01 09:36:00 EST")
	taskLogsCmd.Flags().StringVar(&flagTaskLogsEndTime, "end", "", "Latest time to return logs (e.g. 3y, 2021-01-20 12:00:00 EST")
	taskLogsCmd.Flags().StringSliceVarP(&flagTaskLogsTasks, "task", "t", []string{}, "Show logs from specific task (can be specified multiple times)")
}
