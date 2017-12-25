package cmd

import (
	"fmt"
	"strconv"

	"github.com/jpignata/fargate/console"
	ECS "github.com/jpignata/fargate/ecs"
	serviceutil "github.com/jpignata/fargate/service"
	"github.com/spf13/cobra"
)

var serviceUpdateCmd = &cobra.Command{
	Use:   "update <service name>",
	Short: "Updates all tasks within a service",
	Long:  "Updates all tasks within a service",
	Args:  cobra.ExactArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		if cpu == 0 && memory == 0 {
			console.ErrorExit(fmt.Errorf("--cpu and/or --memory must be supplied"), "Invalid command line arguments")
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		updateService(args[0])
	},
}

func init() {
	serviceCmd.AddCommand(serviceUpdateCmd)

	serviceUpdateCmd.Flags().Int16VarP(&cpu, "cpu", "c", 0, "Amount of cpu units to allocate for each task")
	serviceUpdateCmd.Flags().Int16VarP(&memory, "memory", "m", 0, "Amount of MiB to allocate for each task")
}

func updateService(serviceName string) {
	var (
		newCpu    int16
		newMemory int16
	)

	ecs := ECS.New(sess)
	service := ecs.DescribeService(serviceName)
	taskDefinition := ecs.DescribeTaskDefinition(service.TaskDefinitionArn)

	if cpu > 0 {
		newCpu = cpu
	} else {
		parsedCpu, err := strconv.ParseInt(*taskDefinition.Cpu, 10, 16)

		if err == nil {
			newCpu = int16(parsedCpu)
		} else {
			console.ErrorExit(err, "Invalid command line arguments")
		}
	}

	if memory > 0 {
		newMemory = memory
	} else {
		parsedMemory, err := strconv.ParseInt(*taskDefinition.Memory, 10, 16)

		if err == nil {
			newMemory = int16(parsedMemory)
		} else {
			console.ErrorExit(err, "Invalid command line arguments")
		}
	}

	err := serviceutil.ValidateCpuAndMemory(newCpu, newMemory)

	if err != nil {
		console.ErrorExit(err, "Invalid settings: %d CPU units / %d MiB", newCpu, newMemory)
	}

	console.Info("Updating service %s to %d CPU units / %d MiB", serviceName, newCpu, newMemory)

	newTaskDefinitionArn := ecs.UpdateTaskDefinitionCpuAndMemory(
		service.TaskDefinitionArn,
		strconv.FormatInt(int64(newCpu), 10),
		strconv.FormatInt(int64(newMemory), 10),
	)

	ecs.UpdateServiceTaskDefinition(serviceName, newTaskDefinitionArn)
}
