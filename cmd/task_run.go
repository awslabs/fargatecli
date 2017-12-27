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
	Short: "Run a task",
	Long: `Run a task

A task is a one-time execution of a Docker container. Tasks will run until they
are manually stopped or interrupted for any reason.

If you specify an image using the --image flag, the task will run a container
using that Docker container image. If you do not specify the --image flag,
fargate will build and push a Docker container image based upon the Dockerfile
in the current working directory.
`,
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
	taskRunCmd.Flags().StringSliceVarP(&flagTaskRunEnvVars, "env", "e", []string{}, "Environment variables to set [e.g. KEY=value]")
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
		},
	)

	console.Info("Running task %s", operation.TaskName)
}
