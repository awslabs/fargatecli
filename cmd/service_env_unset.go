package cmd

import (
	"strings"

	"github.com/jpignata/fargate/console"
	ECS "github.com/jpignata/fargate/ecs"
	"github.com/jpignata/fargate/util"
	"github.com/spf13/cobra"
)

type ServiceEnvUnsetOperation struct {
	ServiceName string
	Keys        []string
}

func (o *ServiceEnvUnsetOperation) Validate() {
	if len(o.Keys) == 0 {
		console.IssueExit("No keys specified")
	}
}

func (o *ServiceEnvUnsetOperation) SetKeys(keys []string) {
	o.Keys = util.Map(keys, strings.ToUpper)
}

var serviceEnvUnsetCmd = &cobra.Command{
	Use: "unset",
	Run: func(cmd *cobra.Command, args []string) {
		operation := &ServiceEnvUnsetOperation{
			ServiceName: args[0],
		}

		operation.SetKeys(flagServiceEnvUnsetKeys)
		operation.Validate()
		serviceEnvUnset(operation)
	},
}

var flagServiceEnvUnsetKeys []string

func init() {
	serviceEnvUnsetCmd.Flags().StringSliceVarP(&flagServiceEnvUnsetKeys, "key", "k", []string{}, "Environment variable keys to unset [e.g. KEY, NGINX_PORT]")

	serviceEnvCmd.AddCommand(serviceEnvUnsetCmd)
}

func serviceEnvUnset(operation *ServiceEnvUnsetOperation) {
	console.Info("Unsetting %s environment variables:", operation.ServiceName)

	for _, key := range operation.Keys {
		console.Info("- %s", key)
	}

	ecs := ECS.New(sess)
	service := ecs.DescribeService(operation.ServiceName)
	taskDefinitionArn := ecs.RemoveEnvVarsFromTaskDefinition(service.TaskDefinitionArn, operation.Keys)

	ecs.UpdateServiceTaskDefinition(operation.ServiceName, taskDefinitionArn)
}
