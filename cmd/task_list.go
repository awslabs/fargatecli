package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/jpignata/fargate/console"
	ECS "github.com/jpignata/fargate/ecs"
	"github.com/spf13/cobra"
)

var taskListCmd = &cobra.Command{
	Use:   "list",
	Short: "List running task groups",
	Run: func(cmd *cobra.Command, args []string) {
		listTaskGroups()
	},
}

func init() {
	taskCmd.AddCommand(taskListCmd)
}

func listTaskGroups() {
	ecs := ECS.New(sess)
	taskGroups := ecs.ListTaskGroups()

	if len(taskGroups) == 0 {
		console.InfoExit("No tasks running")
	}

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 1, '\t', 0)
	fmt.Fprintln(w, "NAME\tINSTANCES")

	for _, taskGroup := range taskGroups {
		fmt.Fprintf(w, "%s\t%d\n",
			taskGroup.TaskGroupName,
			taskGroup.Instances,
		)
	}

	w.Flush()
}
