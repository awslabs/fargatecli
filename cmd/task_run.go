package cmd

import (
	CWL "github.com/jpignata/fargate/cloudwatchlogs"
	"github.com/jpignata/fargate/console"
	"github.com/jpignata/fargate/docker"
	EC2 "github.com/jpignata/fargate/ec2"
	ECR "github.com/jpignata/fargate/ecr"
	ECS "github.com/jpignata/fargate/ecs"
	"github.com/jpignata/fargate/git"
	IAM "github.com/jpignata/fargate/iam"
	"github.com/spf13/cobra"
)

type TaskRunOperation struct {
	Num      int64
	Cpu      string
	EnvVars  []ECS.EnvVar
	Image    string
	Memory   string
	TaskName string
}

func (o *TaskRunOperation) Validate() {
	err := validateCpuAndMemory(o.Cpu, o.Memory)

	if err != nil {
		console.ErrorExit(err, "Invalid settings: %s CPU units / %s MiB", o.Cpu, o.Memory)
	}

	if o.Num < 1 {
		console.ErrorExit(err, "Invalid number of tasks: %d, num must be > 1", o.Num)
	}
}

func (o *TaskRunOperation) SetEnvVars(inputEnvVars []string) {
	o.EnvVars = extractEnvVars(inputEnvVars)
}

var (
	flagTaskRunNum     int64
	flagTaskRunCpu     string
	flagTaskRunEnvVars []string
	flagTaskRunImage   string
	flagTaskRunMemory  string
)

var taskRunCmd = &cobra.Command{
	Use:   "run <task name>",
	Short: "Run new tasks",
	Long: `Run new tasks

You must specify a task group name in order to interact with the task(s) in
subsequent commands to view logs, stop and inspect tasks. Task group names do
not have to be unique -- multiple configurations of task instances can be
started with the same task group.

Multiple instances of a task can be run by specifying a number in the --count
flag. If no number is specified, a single task instance will be run.

CPU and memory settings can be optionally specified as CPU units and mebibytes
respectively using the --cpu and --memory flags. Every 1024 CPU units is
equivilent to a single vCPU. AWS Fargate only supports certain combinations of
CPU and memory configurations:

| CPU (CPU Units) | Memory (MiB)                          |
| --------------- | ------------------------------------- |
| 256             | 512, 1024, or 2048                    |
| 512             | 1024 through 4096 in 1GiB increments  |
| 1024            | 2048 through 8192 in 1GiB increments  |
| 2048            | 4096 through 16384 in 1GiB increments |
| 4096            | 8192 through 30720 in 1GiB increments |

If not specified, fargate will launch minimally sized tasks at 0.25 vCPU (256
CPU units) and 0.5GB (512 MiB) of memory.

The Docker container image to use in the task can be optionally specified via
the --image flag. If not specified, fargate will build a new Docker container
image from the current working directory and push it to Amazon ECR in a
repository named for the task group. If the current working directory is a git
repository, the container image will be tagged with the short ref of the HEAD
commit. If not, a timestamp in the format of YYYYMMDDHHMMSS will be used.

Environment variables can be specified via the --env flag. Specify --env with a
key=value parameter multiple times to add multiple variables.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		operation := &TaskRunOperation{
			Num:      flagTaskRunNum,
			Cpu:      flagTaskRunCpu,
			Image:    flagTaskRunImage,
			Memory:   flagTaskRunMemory,
			TaskName: args[0],
		}

		operation.SetEnvVars(flagTaskRunEnvVars)
		operation.Validate()

		runTask(operation)
	},
}

func init() {
	taskRunCmd.Flags().Int64VarP(&flagTaskRunNum, "num", "n", 1, "Number of task instances to run")
	taskRunCmd.Flags().StringSliceVarP(&flagTaskRunEnvVars, "env", "e", []string{}, "Environment variables to set [e.g. KEY=value] (can be specified multiple times)")
	taskRunCmd.Flags().StringVarP(&flagTaskRunCpu, "cpu", "c", "256", "Amount of cpu units to allocate for each task")
	taskRunCmd.Flags().StringVarP(&flagTaskRunImage, "image", "i", "", "Docker image to run; if omitted Fargate will build an image from the Dockerfile in the current directory")
	taskRunCmd.Flags().StringVarP(&flagTaskRunMemory, "memory", "m", "512", "Amount of MiB to allocate for each task")

	taskCmd.AddCommand(taskRunCmd)
}

func runTask(operation *TaskRunOperation) {
	var repositoryUri string

	cwl := CWL.New(sess)
	ec2 := EC2.New(sess)
	ecr := ECR.New(sess)
	ecs := ECS.New(sess)
	iam := IAM.New(sess)

	if ecr.IsRepositoryCreated(operation.TaskName) {
		repositoryUri = ecr.GetRepositoryUri(operation.TaskName)
	} else {
		repositoryUri = ecr.CreateRepository(operation.TaskName)
	}

	repository := docker.Repository{Uri: repositoryUri}
	subnetIds := ec2.GetDefaultVpcSubnetIds()
	securityGroupId := ec2.GetDefaultSecurityGroupId()
	ecsTaskExecutionRoleArn := iam.CreateEcsTaskExecutionRole()
	logGroupName := cwl.CreateLogGroup(taskLogGroupFormat, operation.TaskName)

	if operation.Image == "" {
		var tag string

		username, password := ecr.GetUsernameAndPassword()

		if git.IsCwdGitRepo() {
			tag = git.GetShortSha()
		} else {
			tag = docker.GenerateTag()
		}

		repository.Login(username, password)
		repository.Build(tag)
		repository.Push(tag)

		operation.Image = repository.UriFor(tag)
	}

	taskDefinitionArn := ecs.CreateTaskDefinition(
		&ECS.CreateTaskDefinitionInput{
			Cpu:              operation.Cpu,
			EnvVars:          operation.EnvVars,
			ExecutionRoleArn: ecsTaskExecutionRoleArn,
			Image:            operation.Image,
			Memory:           operation.Memory,
			Name:             operation.TaskName,
			LogGroupName:     logGroupName,
			LogRegion:        region,
			Type:             "task",
		},
	)

	ecs.RunTask(
		&ECS.RunTaskInput{
			ClusterName:       clusterName,
			Count:             operation.Num,
			TaskName:          operation.TaskName,
			TaskDefinitionArn: taskDefinitionArn,
			SubnetIds:         subnetIds,
			SecurityGroupId:   securityGroupId,
		},
	)

	console.Info("Running task %s", operation.TaskName)
}
