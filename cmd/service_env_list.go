package cmd

import (
	"fmt"
	ECS "github.com/jpignata/fargate/ecs"
	"github.com/spf13/cobra"
)

var serviceEnvListCmd = &cobra.Command{
	Use: "list",
	PreRun: func(cmd *cobra.Command, args []string) {
		extractEnvVars()
	},
	Run: func(cmd *cobra.Command, args []string) {
		serviceEnvList(args[0])
	},
}

func init() {
	serviceEnvCmd.AddCommand(serviceEnvListCmd)
}

func serviceEnvList(serviceName string) {
	ecs := ECS.New(sess)
	service := ecs.DescribeService(serviceName)
	envVars := ecs.GetEnvVarsFromTaskDefinition(service.TaskDefinitionArn)

	for _, envVar := range envVars {
		fmt.Printf("%s=%s\n", envVar.Key, envVar.Value)
	}
}
