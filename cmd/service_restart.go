package cmd

import (
	"github.com/jpignata/fargate/console"
	ECS "github.com/jpignata/fargate/ecs"
	"github.com/spf13/cobra"
)

type ServiceRestartOperation struct {
	ServiceName string
}

var serviceRestartCmd = &cobra.Command{
	Use:   "restart <service name>",
	Short: "Restart all tasks within a service",
	Long: `Restart all tasks within a service

	Begins a deployment and restarts all tasks within a service. This is useful
	if you have some external data source cached that you need to refresh.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		operation := &ServiceRestartOperation{
			ServiceName: args[0],
		}

		restartService(operation)
	},
}

func init() {
	serviceCmd.AddCommand(serviceRestartCmd)
}

func restartService(operation *ServiceRestartOperation) {
	console.Info("Restarting %s", operation.ServiceName)

	ecs := ECS.New(sess)
	service := ecs.DescribeService(operation.ServiceName)
	taskDefinitionArn := ecs.IncrementTaskDefinition(service.TaskDefinitionArn)

	ecs.UpdateServiceTaskDefinition(operation.ServiceName, taskDefinitionArn)
}
