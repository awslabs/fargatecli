package cmd

import (
	"fmt"

	"github.com/jpignata/fargate/console"
	ECS "github.com/jpignata/fargate/ecs"
	"github.com/spf13/cobra"
)

type ServiceUpdateOperation struct {
	ServiceName string
	Cpu         string
	Memory      string
	Service     ECS.Service
}

func (o *ServiceUpdateOperation) Validate() {
	ecs := ECS.New(sess, clusterName)

	if o.Cpu == "" && o.Memory == "" {
		console.ErrorExit(fmt.Errorf("--cpu and/or --memory must be supplied"), "Invalid command line arguments")
	}

	o.Service = ecs.DescribeService(o.ServiceName)
	cpu, memory := ecs.GetCpuAndMemoryFromTaskDefinition(o.Service.TaskDefinitionArn)

	if o.Cpu == "" {
		o.Cpu = cpu
	}

	if o.Memory == "" {
		o.Memory = memory
	}

	err := validateCpuAndMemory(o.Cpu, o.Memory)

	if err != nil {
		console.ErrorExit(err, "Invalid settings: %d CPU units / %d MiB", o.Cpu, o.Memory)
	}
}

var (
	flagServiceUpdateCpu    string
	flagServiceUpdateMemory string
)

var serviceUpdateCmd = &cobra.Command{
	Use:   "update <service-name> --cpu <cpu-units> | --memory <MiB>",
	Short: "Update service configuration",
	Long: `Update service configuration

CPU and memory settings are specified as CPU units and mebibytes respectively
using the --cpu and --memory flags. Every 1024 CPU units is equivilent to a
single vCPU. AWS Fargate only supports certain combinations of CPU and memory
configurations:

| CPU (CPU Units) | Memory (MiB)                          |
| --------------- | ------------------------------------- |
| 256             | 512, 1024, or 2048                    |
| 512             | 1024 through 4096 in 1GiB increments  |
| 1024            | 2048 through 8192 in 1GiB increments  |
| 2048            | 4096 through 16384 in 1GiB increments |
| 4096            | 8192 through 30720 in 1GiB increments |

At least one of --cpu or --memory must be specified.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		operation := &ServiceUpdateOperation{
			ServiceName: args[0],
			Cpu:         flagServiceUpdateCpu,
			Memory:      flagServiceUpdateMemory,
		}

		updateService(operation)
	},
}

func init() {
	serviceCmd.AddCommand(serviceUpdateCmd)

	serviceUpdateCmd.Flags().StringVarP(&flagServiceUpdateCpu, "cpu", "c", "", "Amount of cpu units to allocate for each task")
	serviceUpdateCmd.Flags().StringVarP(&flagServiceUpdateMemory, "memory", "m", "", "Amount of MiB to allocate for each task")
}

func updateService(operation *ServiceUpdateOperation) {
	ecs := ECS.New(sess, clusterName)

	newTaskDefinitionArn := ecs.UpdateTaskDefinitionCpuAndMemory(
		operation.Service.TaskDefinitionArn,
		operation.Cpu,
		operation.Memory,
	)

	ecs.UpdateServiceTaskDefinition(operation.ServiceName, newTaskDefinitionArn)
	console.Info("Updated service %s to %d CPU units / %d MiB", operation.ServiceName, operation.Cpu, operation.Memory)
}
