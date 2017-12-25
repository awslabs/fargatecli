package cmd

import (
	"github.com/jpignata/fargate/console"
	ECS "github.com/jpignata/fargate/ecs"
	"github.com/spf13/cobra"
)

var serviceRestartCmd = &cobra.Command{
	Use:   "restart <service name>",
	Short: "Restarts all tasks within a service",
	Long:  "Restarts all tasks within a service",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		restartService(args[0])
	},
}

func init() {
	serviceCmd.AddCommand(serviceRestartCmd)
}

func restartService(serviceName string) {
	console.Info("Restarting %s", serviceName)

	ecs := ECS.New(sess)
	service := ecs.DescribeService(serviceName)
	taskDefinitionArn := ecs.IncrementTaskDefinition(service.TaskDefinitionArn)

	ecs.UpdateServiceTaskDefinition(serviceName, taskDefinitionArn)
}
