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
	Use:   "destroy <service-name>",
	Short: "Destroy a service",
	Long: `Destroy service

Deletes a service. In order to destroy a service, it must first be scaled to 0
running.`,
	Args: cobra.ExactArgs(1),
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
	ecs := ECS.New(sess)
	ecs.DestroyService(operation.ServiceName)
	console.Info("Destroyed service %s", operation.ServiceName)
}
