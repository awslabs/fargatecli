package cmd

import (
	"github.com/jpignata/fargate/console"
	ECS "github.com/jpignata/fargate/ecs"
	"github.com/spf13/cobra"
)

var serviceEnvSetCmd = &cobra.Command{
	Use: "set",
	PreRun: func(cmd *cobra.Command, args []string) {
		extractEnvVars()
	},
	Run: func(cmd *cobra.Command, args []string) {
		serviceEnvSet(args[0])
	},
}

func init() {
	serviceEnvSetCmd.Flags().StringSliceVarP(&envVarsRaw, "env", "e", []string{}, "Environment variables to set [e.g. KEY=value]")

	serviceEnvCmd.AddCommand(serviceEnvSetCmd)
}

func serviceEnvSet(serviceName string) {
	if len(envVars) == 0 {
		console.ErrorExit(nil, "No environment variables specified")
	}

	console.Info("Setting %s environment variables:", serviceName)

	for _, envVar := range envVars {
		console.Info("- %s=%s", envVar.Key, envVar.Value)
	}

	ecs := ECS.New(sess)
	service := ecs.DescribeService(serviceName)
	taskDefinitionArn := ecs.AddEnvVarsToTaskDefinition(service.TaskDefinitionArn, envVars)

	ecs.UpdateServiceTaskDefinition(serviceName, taskDefinitionArn)
}
