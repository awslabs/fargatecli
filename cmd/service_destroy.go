package cmd

import (
	"github.com/jpignata/fargate/console"
	ECS "github.com/jpignata/fargate/ecs"
	"github.com/spf13/cobra"
)

var serviceDestroyCmd = &cobra.Command{
	Use:   "destroy <service name>",
	Short: "Delete a service",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		destroyService(args[0])
	},
}

func init() {
	serviceCmd.AddCommand(serviceDestroyCmd)
}

func destroyService(serviceName string) {
	console.Info("[%s] Destroying service", serviceName)

	ecs := ECS.New(sess)
	ecs.DestroyService(serviceName)
}
