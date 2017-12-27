package cmd

import (
	"fmt"

	"github.com/jpignata/fargate/console"
	EC2 "github.com/jpignata/fargate/ec2"
	ECS "github.com/jpignata/fargate/ecs"
	"github.com/jpignata/fargate/util"
	"github.com/spf13/cobra"
)

type TaskInfoOperation struct {
	TaskGroupName string
	TaskIds       []string
}

var flagTaskInfoTasks []string

var taskInfoCmd = &cobra.Command{
	Use:   "info <task group name>",
	Short: "Display configuration information about tasks instances",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		operation := &TaskInfoOperation{
			TaskGroupName: args[0],
			TaskIds:       flagTaskInfoTasks,
		}

		getTaskInfo(operation)
	},
}

func init() {
	taskCmd.AddCommand(taskInfoCmd)

	taskInfoCmd.Flags().StringSliceVarP(&flagTaskInfoTasks, "task", "t", []string{}, "Get info for specific task instances (can be specified multiple times)")
}

func getTaskInfo(operation *TaskInfoOperation) {
	var tasks []ECS.Task
	var eniIds []string

	ecs := ECS.New(sess)
	ec2 := EC2.New(sess)

	if len(operation.TaskIds) > 0 {
		tasks = ecs.DescribeTasks(operation.TaskIds)
	} else {
		tasks = ecs.DescribeTasksForTaskGroup(operation.TaskGroupName)
	}

	if len(tasks) == 0 {
		console.InfoExit("No tasks found")
	}

	for _, task := range tasks {
		if task.EniId != "" {
			eniIds = append(eniIds, task.EniId)
		}
	}

	enis := ec2.DescribeNetworkInterfaces(eniIds)

	console.KeyValue("Task Group Name", "%s\n", operation.TaskGroupName)
	console.KeyValue("Task Instances", "%d\n", len(tasks))

	for _, task := range tasks {
		console.KeyValue("  "+task.TaskId, "\n")
		console.KeyValue("    Image", "%s\n", task.Image)
		console.KeyValue("    Status", "%s\n", util.Humanize(task.LastStatus))
		console.KeyValue("    Started At", "%s\n", task.CreatedAt)
		console.KeyValue("    IP", "%s\n", enis[task.EniId].PublicIpAddress)
		console.KeyValue("    CPU", "%s\n", task.Cpu)
		console.KeyValue("    Memory", "%s\n", task.Memory)

		if len(task.EnvVars) > 0 {
			console.KeyValue("    Environment Variables", "\n")

			for _, envVar := range task.EnvVars {
				fmt.Printf("      %s=%s\n", envVar.Key, envVar.Value)
			}
		}
	}
}
