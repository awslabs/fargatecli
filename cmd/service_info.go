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

var serviceInfoCmd = &cobra.Command{
	Use:   "info <service name>",
	Short: "Display information about a service",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		infoService(args[0])
	},
}

func init() {
	serviceCmd.AddCommand(serviceInfoCmd)
}

func infoService(serviceName string) {
	var eniIds []string

	ecs := ECS.New()
	ec2 := EC2.New()
	service := ecs.DescribeService(serviceName)
	tasks := ecs.DescribeTasksForService(serviceName)

	for _, task := range tasks {
		eniIds = append(eniIds, task.EniId)
	}

	console.KeyValue("Service Name", "%s\n", serviceName)
	console.KeyValue("Status", "\n")
	console.KeyValue("  Desired", "%d\n", service.DesiredCount)
	console.KeyValue("  Running", "%d\n", service.RunningCount)
	console.KeyValue("  Pending", "%d\n", service.PendingCount)
	console.KeyValue("Image", "%s\n", service.Image)
	console.KeyValue("Cpu", "%s\n", service.Cpu)
	console.KeyValue("Memory", "%s\n", service.Memory)

	if len(tasks) > 0 {
		console.Header("== Tasks ==")

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
	}

	if len(service.Deployments) > 0 {
		console.Header("== Deployments ==")

		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 0, 8, 1, '\t', 0)
		fmt.Fprintln(w, "IMAGE\tSTATUS\tCREATED\tDESIRED\tRUNNING\tPENDING")

		for _, d := range service.Deployments {
			fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%d\t%d\n",
				d.Image,
				util.Humanize(d.Status),
				d.CreatedAt,
				d.DesiredCount,
				d.RunningCount,
				d.PendingCount,
			)
		}

		w.Flush()
	}

	if len(service.Events) > 0 {
		console.Header("== Events ==")

		for i, event := range service.Events {
			fmt.Printf("[%s] %s\n", event.CreatedAt, event.Message)

			if i == 10 {
				break
			}
		}
	}
}
