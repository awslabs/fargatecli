package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	flagServiceLogsFilter    string
	flagServiceLogsEndTime   string
	flagServiceLogsStartTime string
	flagServiceLogsFollow    bool
	flagServiceLogsTasks     []string
)

var serviceLogsCmd = &cobra.Command{
	Use:   "logs <service name>",
	Short: "View logs from a service",
	Args:  cobra.ExactArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
	},
	Run: func(cmd *cobra.Command, args []string) {
		operation := &GetLogsOperation{
			LogGroupName: fmt.Sprintf(serviceLogGroupFormat, args[0]),
			Filter:       flagServiceLogsFilter,
			Follow:       flagServiceLogsFollow,
			Namespace:    args[0],
		}

		operation.AddTasks(flagServiceLogsTasks)
		operation.AddStartTime(flagServiceLogsStartTime)
		operation.AddEndTime(flagServiceLogsEndTime)

		GetLogs(operation)
	},
}

func init() {
	serviceCmd.AddCommand(serviceLogsCmd)

	serviceLogsCmd.Flags().BoolVarP(&flagServiceLogsFollow, "follow", "f", false, "Poll logs and continuously print new events")
	serviceLogsCmd.Flags().StringVar(&flagServiceLogsFilter, "filter", "", "Filter pattern to apply")
	serviceLogsCmd.Flags().StringVar(&flagServiceLogsStartTime, "start", "", "Earliest time to return logs (e.g. -1h, 2018-01-01 09:36:00 EST")
	serviceLogsCmd.Flags().StringVar(&flagServiceLogsEndTime, "end", "", "Latest time to return logs (e.g. 3y, 2021-01-20 12:00:00 EST")
	serviceLogsCmd.Flags().StringSliceVarP(&flagServiceLogsTasks, "task", "t", []string{}, "Show logs from specific task (can be specified multiple times)")
}
