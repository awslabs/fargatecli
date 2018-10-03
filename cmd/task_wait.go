package cmd

import (
	"github.com/jpignata/fargate/console"
	ECS "github.com/jpignata/fargate/ecs"
	"github.com/spf13/cobra"
)

var taskWaitCmd = &cobra.Command{
	Use:   "wait",
	Short: "Wait for running task groups to complete",
	Run: func(cmd *cobra.Command, args []string) {
		waitTaskGroups()
	},
}

func init() {
	taskCmd.AddCommand(taskWaitCmd)
}

func waitTaskGroups() {
	ecs := ECS.New(sess, clusterName)
	taskreturn := ecs.WaitTaskGroups()

	if taskreturn == true {
		console.InfoExit("All tasks have stopped")
	}

}
