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
	Use:   "logs <service-name>",
	Short: "Show logs from tasks in a service",
	Long: `Show logs from tasks in a service

Return either a specific segment of service logs or tail logs in real-time
using the --follow option. Logs are prefixed by their log stream name which is
in the format of "fargate/\<service-name>/\<task-id>."

Follow will continue to run and return logs until interrupted by Control-C. If
--follow is passed --end cannot be specified.

Logs can be returned for specific tasks within a service by passing a task ID
via the --task flag. Pass --task with a task ID multiple times in order to
retrieve logs from multiple specific tasks.

A specific window of logs can be requested by passing --start and --end options
with a time expression. The time expression can be either a duration or a
timestamp:

  - Duration (e.g. -1h [one hour ago], -1h10m30s [one hour, ten minutes, and
    thirty seconds ago], 2h [two hours from now])
  - Timestamp with optional timezone in the format of YYYY-MM-DD HH:MM:SS [TZ];
    timezone will default to UTC if omitted (e.g. 2017-12-22 15:10:03 EST)

You can filter logs for specific term by passing a filter expression via the
--filter flag. Pass a single term to search for that term, pass multiple terms
to search for log messages that include all terms.`,
	Args: cobra.ExactArgs(1),
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
