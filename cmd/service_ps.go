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

var servicePsCmd = &cobra.Command{
	Use:   "ps <service name>",
	Short: "List running tasks in a service",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		psService(args[0])
	},
}

func init() {
	serviceCmd.AddCommand(servicePsCmd)
}

func psService(serviceName string) {
	var eniIds []string

	ecs := ECS.New(sess)
	ec2 := EC2.New(sess)
	tasks := ecs.DescribeTasksForService(serviceName)

	for _, task := range tasks {
		eniIds = append(eniIds, task.EniId)
	}

	if len(tasks) > 0 {
		enis := ec2.DescribeNetworkInterfaces(eniIds)

		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 0, 8, 1, '\t', 0)
		fmt.Fprintln(w, "IMAGE\tSTATUS\tDESIRED STATUS\tCREATED\tIP\tCPU\tMEMORY\t")

		for _, t := range tasks {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
				t.Image,
				util.Humanize(t.LastStatus),
				util.Humanize(t.DesiredStatus),
				t.CreatedAt,
				enis[t.EniId].PublicIpAddress,
				t.Cpu,
				t.Memory,
			)
		}

		w.Flush()
	} else {
		console.Info("No tasks found")
	}
}
