package cmd

import (
	"strings"

	"github.com/jpignata/fargate/console"
	ECS "github.com/jpignata/fargate/ecs"
	"github.com/jpignata/fargate/util"
	"github.com/spf13/cobra"
)

var keys []string

var serviceEnvUnsetCmd = &cobra.Command{
	Use: "unset",
	PreRun: func(cmd *cobra.Command, args []string) {
		extractEnvVars()
	},
	Run: func(cmd *cobra.Command, args []string) {
		serviceEnvUnset(args[0])
	},
}

func init() {
	serviceEnvUnsetCmd.Flags().StringSliceVarP(&keys, "key", "k", []string{}, "Environment variable keys to unset [e.g. KEY, NGINX_PORT]")

	serviceEnvCmd.AddCommand(serviceEnvUnsetCmd)
}

func serviceEnvUnset(serviceName string) {
	if len(keys) == 0 {
		console.ErrorExit(nil, "No keys specified")
	}

	upperKeys := util.Map(keys, strings.ToUpper)

	console.Info("Unsetting %s environment variables:", serviceName)

	for _, key := range upperKeys {
		console.Info("- %s", key)
	}

	ecs := ECS.New(sess)
	service := ecs.DescribeService(serviceName)
	taskDefinitionArn := ecs.RemoveEnvVarsFromTaskDefinition(service.TaskDefinitionArn, upperKeys)

	ecs.UpdateServiceTaskDefinition(serviceName, taskDefinitionArn)
}
