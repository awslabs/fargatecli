package cmd

import (
	"github.com/jpignata/fargate/console"
	"github.com/jpignata/fargate/docker"
	ECR "github.com/jpignata/fargate/ecr"
	ECS "github.com/jpignata/fargate/ecs"
	"github.com/jpignata/fargate/git"
	"github.com/spf13/cobra"
)

var serviceDeployCmd = &cobra.Command{
	Use:  "deploy <service name>",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		deployService(args[0])
	},
}

func init() {
	serviceDeployCmd.Flags().StringVarP(&image, "image", "i", "", "Docker image to run in the service; if omitted Fargate will build an image from the Dockerfile in the current directory")

	serviceCmd.AddCommand(serviceDeployCmd)
}

func deployService(serviceName string) {
	console.Info("Deploying %s", serviceName)

	ecs := ECS.New(sess)
	service := ecs.DescribeService(serviceName)

	if image == "" {
		var tag string

		ecr := ECR.New(sess)
		repositoryUri := ecr.GetRepositoryUri(serviceName)
		repository := docker.Repository{
			Uri: repositoryUri,
		}

		username, password := ecr.GetUsernameAndPassword()

		if git.IsCwdGitRepo() {
			tag = git.GetShortSha()
		} else {
			tag = docker.GenerateTag()
		}

		repository.Login(username, password)
		repository.Build(tag)
		repository.Push(tag)

		image = repository.UriFor(tag)
	}

	taskDefinitionArn := ecs.UpdateTaskDefinitionImage(service.TaskDefinitionArn, image)
	ecs.UpdateServiceTaskDefinition(serviceName, taskDefinitionArn)
}
