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
	Ecs         ECS.ECS
	Service     ECS.Service
}

func (o *ServiceUpdateOperation) Validate() {
	if o.Cpu == "" && o.Memory == "" {
		console.ErrorExit(fmt.Errorf("--cpu and/or --memory must be supplied"), "Invalid command line arguments")
	}

	o.Service = o.Ecs.DescribeService(o.ServiceName)
	cpu, memory := o.Ecs.GetCpuAndMemoryFromTaskDefinition(o.Service.TaskDefinitionArn)

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
	Use:   "update <service name>",
	Short: "Update cpu and/or memory settings",
	Long:  "Update cpu and/or memory settings",
	Args:  cobra.ExactArgs(1),
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
	console.Info("Updating service %s to %d CPU units / %d MiB", operation.ServiceName, operation.Cpu, operation.Memory)

	newTaskDefinitionArn := operation.Ecs.UpdateTaskDefinitionCpuAndMemory(
		operation.Service.TaskDefinitionArn,
		operation.Cpu,
		operation.Memory,
	)

	operation.Ecs.UpdateServiceTaskDefinition(operation.ServiceName, newTaskDefinitionArn)
}
