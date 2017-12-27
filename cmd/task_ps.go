package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/jpignata/fargate/console"
	EC2 "github.com/jpignata/fargate/ec2"
	ECS "github.com/jpignata/fargate/ecs"
	"github.com/jpignata/fargate/util"
	"github.com/spf13/cobra"
)

type TaskProcessListOperation struct {
	TaskName string
}

var taskPsCmd = &cobra.Command{
	Use:   "ps <task name>",
	Short: "List running instances for a task",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		operation := &TaskProcessListOperation{
			TaskName: args[0],
		}

		getTaskProcessList(operation)
	},
}

func init() {
	taskCmd.AddCommand(taskPsCmd)
}

func getTaskProcessList(operation *TaskProcessListOperation) {
	var eniIds []string

	ecs := ECS.New(sess)
	ec2 := EC2.New(sess)
	tasks := ecs.DescribeTasksForTask(operation.TaskName)

	for _, task := range tasks {
		if task.EniId != "" {
			eniIds = append(eniIds, task.EniId)
		}
	}

	if len(tasks) == 0 {
		console.InfoExit("No tasks found")
	}

	enis := ec2.DescribeNetworkInterfaces(eniIds)

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 1, '\t', 0)
	fmt.Fprintln(w, "ID\tIMAGE\tSTATUS\tRUNNING\tIP\tCPU\tMEMORY\t")

	for _, t := range tasks {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			t.TaskId,
			t.Image,
			util.Humanize(t.LastStatus),
			t.RunningFor(),
			enis[t.EniId].PublicIpAddress,
			t.Cpu,
			t.Memory,
		)
	}

	w.Flush()
}
