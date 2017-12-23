package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/jpignata/fargate/console"
	ECS "github.com/jpignata/fargate/ecs"
	"github.com/spf13/cobra"
)

var serviceListCmd = &cobra.Command{
	Use:   "list",
	Short: "List services",
	Run: func(cmd *cobra.Command, args []string) {
		listServices()
	},
}

func init() {
	serviceCmd.AddCommand(serviceListCmd)
}

func listServices() {
	ecs := ECS.New()
	services := ecs.ListServices()

	if len(services) > 0 {
		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 0, 8, 1, '\t', 0)
		fmt.Fprintln(w, "NAME\tIMAGE\tCPU\tMEMORY\tDESIRED\tRUNNING\tPENDING\t")

		for _, service := range services {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\t%d\t%d\t\n",
				service.Name,
				service.Image,
				service.Cpu,
				service.Memory,
				service.DesiredCount,
				service.RunningCount,
				service.PendingCount,
			)
		}

		w.Flush()
	} else {
		console.Info("No services found")
	}
}
