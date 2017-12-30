package cmd

import (
	"github.com/jpignata/fargate/console"
	ECS "github.com/jpignata/fargate/ecs"
	"github.com/spf13/cobra"
)

type TaskStopOperation struct {
	TaskGroupName string
	TaskIds       []string
}

var (
	flagTaskStopTasks []string
)

var taskStopCmd = &cobra.Command{
	Use:   "stop <task group name>",
	Short: "Stop tasks",
	Long: `Stop tasks

  Stops all tasks within a task group if run with only a task group name or stops
  individual tasks if one or more tasks are passed via the --task flag. Specify
  --task with a task ID parameter multiple times to stop multiple specific tasks.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		operation := &TaskStopOperation{
			TaskGroupName: args[0],
			TaskIds:       flagTaskStopTasks,
		}

		stopTasks(operation)
	},
}

func init() {
	taskCmd.AddCommand(taskStopCmd)

	taskStopCmd.Flags().StringSliceVarP(&flagTaskStopTasks, "task", "t", []string{}, "Stop specific task instances (can be specified multiple times)")
}

func stopTasks(operation *TaskStopOperation) {
	var taskCount int

	ecs := ECS.New(sess, clusterName)

	if len(operation.TaskIds) > 0 {
		taskCount = len(operation.TaskIds)

		ecs.StopTasks(operation.TaskIds)
	} else {
		var taskIds []string

		tasks := ecs.DescribeTasksForTaskGroup(operation.TaskGroupName)

		for _, task := range tasks {
			taskIds = append(taskIds, task.TaskId)
		}

		taskCount = len(taskIds)

		ecs.StopTasks(taskIds)
	}

	if taskCount == 1 {
		console.Info("Stopped %d task", taskCount)
	} else {
		console.Info("Stopped %d tasks", taskCount)
	}
}
