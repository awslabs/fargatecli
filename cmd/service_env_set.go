package cmd

import (
	"github.com/jpignata/fargate/console"
	ECS "github.com/jpignata/fargate/ecs"
	"github.com/spf13/cobra"
)

type ServiceEnvSetOperation struct {
	ServiceName string
	EnvVars     []ECS.EnvVar
}

func (o *ServiceEnvSetOperation) Validate() {
	if len(o.EnvVars) == 0 {
		console.IssueExit("No environment variables specified")
	}
}

func (o *ServiceEnvSetOperation) SetEnvVars(inputEnvVars []string) {
	o.EnvVars = extractEnvVars(inputEnvVars)
}

var flagServiceEnvSetEnvVars []string

var serviceEnvSetCmd = &cobra.Command{
	Use: "set",
	Run: func(cmd *cobra.Command, args []string) {
		operation := &ServiceEnvSetOperation{
			ServiceName: args[0],
			EnvVars:     envVars,
		}

		operation.SetEnvVars(flagServiceEnvSetEnvVars)
		operation.Validate()
		serviceEnvSet(operation)
	},
}

func init() {
	serviceEnvSetCmd.Flags().StringSliceVarP(&flagServiceEnvSetEnvVars, "env", "e", []string{}, "Environment variables to set [e.g. KEY=value]")

	serviceEnvCmd.AddCommand(serviceEnvSetCmd)
}

func serviceEnvSet(operation *ServiceEnvSetOperation) {
	console.Info("Setting %s environment variables:", operation.ServiceName)

	for _, envVar := range operation.EnvVars {
		console.Info("- %s=%s", envVar.Key, envVar.Value)
	}

	ecs := ECS.New(sess)
	service := ecs.DescribeService(operation.ServiceName)
	taskDefinitionArn := ecs.AddEnvVarsToTaskDefinition(service.TaskDefinitionArn, operation.EnvVars)

	ecs.UpdateServiceTaskDefinition(operation.ServiceName, taskDefinitionArn)
}
