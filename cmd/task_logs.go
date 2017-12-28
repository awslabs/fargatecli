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
	Short: "Show logs from tasks",
	Long: `Show logs from tasks

Return either a specific segment of task logs or tail logs in real-time using
the --follow option. Logs are prefixed by their log stream name which is in the
format of "fargate/<task-group-name>/<task-id>." If --follow is passed --end
cannot be specified.

Logs can be returned for specific tasks within a task group by passing a task ID
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
to search for log messages that include all terms. See
http://docs.aws.amazon.com/AmazonCloudWatch/latest/logs/FilterAndPatternSyntax.html#matching-terms-events
for more details.`,
	Args: cobra.ExactArgs(1),
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
