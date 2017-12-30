package cmd

import (
	"fmt"
	ECS "github.com/jpignata/fargate/ecs"
	"github.com/spf13/cobra"
)

type ServiceEnvListOperation struct {
	ServiceName string
}

var serviceEnvListCmd = &cobra.Command{
	Use:   "list <service-name>",
	Short: "Show environment variables",
	Run: func(cmd *cobra.Command, args []string) {
		operation := &ServiceEnvListOperation{
			ServiceName: args[0],
		}

		serviceEnvList(operation)
	},
}

func init() {
	serviceEnvCmd.AddCommand(serviceEnvListCmd)
}

func serviceEnvList(operation *ServiceEnvListOperation) {
	ecs := ECS.New(sess, clusterName)
	service := ecs.DescribeService(operation.ServiceName)
	envVars := ecs.GetEnvVarsFromTaskDefinition(service.TaskDefinitionArn)

	for _, envVar := range envVars {
		fmt.Printf("%s=%s\n", envVar.Key, envVar.Value)
	}
}
