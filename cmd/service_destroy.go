package cmd

import (
	"github.com/jpignata/fargate/console"
	ECS "github.com/jpignata/fargate/ecs"
	"github.com/spf13/cobra"
)

type ServiceDestroyOperation struct {
	ServiceName string
}

var serviceDestroyCmd = &cobra.Command{
	Use:   "destroy <service name>",
	Short: "Delete a service",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		operation := &ServiceDestroyOperation{
			ServiceName: args[0],
		}

		destroyService(operation)
	},
}

func init() {
	serviceCmd.AddCommand(serviceDestroyCmd)
}

func destroyService(operation *ServiceDestroyOperation) {
	console.Info("[%s] Destroying service", operation.ServiceName)

	ecs := ECS.New(sess)
	ecs.DestroyService(operation.ServiceName)
}
