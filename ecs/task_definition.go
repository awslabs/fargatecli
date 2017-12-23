package ecs

import (
	"github.com/aws/aws-sdk-go/aws"
	awsecs "github.com/aws/aws-sdk-go/service/ecs"
	"github.com/jpignata/fargate/console"
)

var taskDefinitionCache = make(map[string]*awsecs.TaskDefinition)

type CreateTaskDefinitionInput struct {
	Cpu              string
	EnvVars          map[string]string
	ExecutionRoleArn string
	Image            string
	Memory           string
	Name             string
	Port             int64
}

func (ecs *ECS) CreateTaskDefinition(input *CreateTaskDefinitionInput) string {
	console.Debug("Creating ECS task definition")

	const essential = true
	const launchType = "FARGATE"
	const networkMode = "awsvpc"

	compatbilities := []string{launchType}

	containerDefinition := &awsecs.ContainerDefinition{
		Name:        aws.String(input.Name),
		Essential:   aws.Bool(essential),
		Image:       aws.String(input.Image),
		Environment: input.Environment(),
	}

	if input.Port != 0 {
		containerDefinition.SetPortMappings(
			[]*awsecs.PortMapping{
				&awsecs.PortMapping{
					ContainerPort: aws.Int64(int64(input.Port)),
				},
			},
		)
	}

	resp, err := ecs.svc.RegisterTaskDefinition(
		&awsecs.RegisterTaskDefinitionInput{
			Family:                  aws.String(input.Name),
			RequiresCompatibilities: aws.StringSlice(compatbilities),
			ContainerDefinitions:    []*awsecs.ContainerDefinition{containerDefinition},
			NetworkMode:             aws.String(networkMode),
			Memory:                  aws.String(input.Memory),
			Cpu:                     aws.String(input.Cpu),
			ExecutionRoleArn:        aws.String(input.ExecutionRoleArn),
		},
	)

	if err != nil {
		console.ErrorExit(err, "Couldn't register ECS task definition")
	}

	td := resp.TaskDefinition

	console.Debug("Created ECS task definition [%s:%d]", aws.StringValue(td.Family), aws.Int64Value(td.Revision))

	return aws.StringValue(td.TaskDefinitionArn)
}

func (input *CreateTaskDefinitionInput) Environment() []*awsecs.KeyValuePair {
	var environment []*awsecs.KeyValuePair

	for name, value := range input.EnvVars {
		environment = append(environment,
			&awsecs.KeyValuePair{
				Name:  aws.String(name),
				Value: aws.String(value),
			},
		)
	}

	return environment
}

func (ecs *ECS) DescribeTaskDefinition(taskDefinitionArn string) *awsecs.TaskDefinition {
	if taskDefinitionCache[taskDefinitionArn] != nil {
		return taskDefinitionCache[taskDefinitionArn]
	}

	resp, err := ecs.svc.DescribeTaskDefinition(
		&awsecs.DescribeTaskDefinitionInput{
			TaskDefinition: aws.String(taskDefinitionArn),
		},
	)

	if err != nil {
		console.ErrorExit(err, "Could not describe ECS task definition")
	}

	taskDefinitionCache[taskDefinitionArn] = resp.TaskDefinition

	return taskDefinitionCache[taskDefinitionArn]
}

func (ecs *ECS) UpdateTaskDefinitionImage(taskDefinitionArn, image string) string {
	taskDefinition := ecs.DescribeTaskDefinition(taskDefinitionArn)
	taskDefinition.ContainerDefinitions[0].Image = aws.String(image)

	resp, err := ecs.svc.RegisterTaskDefinition(
		&awsecs.RegisterTaskDefinitionInput{
			ContainerDefinitions:    taskDefinition.ContainerDefinitions,
			Cpu:                     taskDefinition.Cpu,
			ExecutionRoleArn:        taskDefinition.ExecutionRoleArn,
			Family:                  taskDefinition.Family,
			Memory:                  taskDefinition.Memory,
			NetworkMode:             taskDefinition.NetworkMode,
			RequiresCompatibilities: taskDefinition.RequiresCompatibilities,
		},
	)

	if err != nil {
		console.ErrorExit(err, "Could not register ECS task definition")
	}

	return aws.StringValue(resp.TaskDefinition.TaskDefinitionArn)
}
