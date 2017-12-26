package cmd

import (
	"github.com/jpignata/fargate/console"
	"github.com/jpignata/fargate/docker"
	ECR "github.com/jpignata/fargate/ecr"
	ECS "github.com/jpignata/fargate/ecs"
	"github.com/jpignata/fargate/git"
	"github.com/spf13/cobra"
)

type ServiceDeployOperation struct {
	ServiceName string
	Image       string
}

var flagServiceDeployImage string

var serviceDeployCmd = &cobra.Command{
	Use:   "deploy <service name>",
	Short: "Deploy a service from an image or local Dockerfile",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		operation := &ServiceDeployOperation{
			ServiceName: args[0],
			Image:       flagServiceDeployImage,
		}

		deployService(operation)
	},
}

func init() {
	serviceDeployCmd.Flags().StringVarP(&flagServiceDeployImage, "image", "i", "", "Docker image to run in the service; if omitted Fargate will build an image from the Dockerfile in the current directory")

	serviceCmd.AddCommand(serviceDeployCmd)
}

func deployService(operation *ServiceDeployOperation) {
	console.Info("Deploying %s", operation.ServiceName)

	ecs := ECS.New(sess)
	service := ecs.DescribeService(operation.ServiceName)

	if operation.Image == "" {
		var tag string

		ecr := ECR.New(sess)
		repositoryUri := ecr.GetRepositoryUri(operation.ServiceName)
		repository := docker.Repository{Uri: repositoryUri}
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

	taskDefinitionArn := ecs.UpdateTaskDefinitionImage(service.TaskDefinitionArn, operation.Image)
	ecs.UpdateServiceTaskDefinition(operation.ServiceName, taskDefinitionArn)
}
