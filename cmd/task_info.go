package cmd

import (
	"fmt"
	"strings"

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
	Short: "Inspect tasks",
	Long: `Inspect tasks

Shows extended information for each running task within a task group or for
specific tasks specified with the --task flag. Information includes environment
variables which could differ between tasks in a task group. To inspect multiple
specific tasks within a task group specific --task with a task ID multiple
times.`,
	Args: cobra.ExactArgs(1),
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

	ecs := ECS.New(sess, clusterName)
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
		eni := enis[task.EniId]

		console.KeyValue("  "+task.TaskId, "\n")
		console.KeyValue("    Image", "%s\n", task.Image)
		console.KeyValue("    Status", "%s\n", util.Humanize(task.LastStatus))
		console.KeyValue("    Started At", "%s\n", task.CreatedAt)
		console.KeyValue("    IP", "%s\n", eni.PublicIpAddress)
		console.KeyValue("    CPU", "%s\n", task.Cpu)
		console.KeyValue("    Memory", "%s\n", task.Memory)
		console.KeyValue("    Subnet", "%s\n", task.SubnetId)
		console.KeyValue("    Security Groups", "%s\n", strings.Join(eni.SecurityGroupIds, ", "))

		if len(task.EnvVars) > 0 {
			console.KeyValue("    Environment Variables", "\n")

			for _, envVar := range task.EnvVars {
				fmt.Printf("      %s=%s\n", envVar.Key, envVar.Value)
			}
		}
	}
}
