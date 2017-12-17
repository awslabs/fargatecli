package ecs

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	awsecs "github.com/aws/aws-sdk-go/service/ecs"
	"github.com/jpignata/fargate/console"
)

type CreateServiceInput struct {
	Cluster           string
	Port              int64
	Name              string
	SubnetIds         []string
	TargetGroupArn    string
	TaskDefinitionArn string
}

func (ecs *ECS) CreateService(input *CreateServiceInput) {
	console.Debug("Creating ECS service")

	const desiredCount = 1
	const launchType = "FARGATE"
	const assignPublicIp = "ENABLED"

	createServiceInput := &awsecs.CreateServiceInput{
		Cluster:        aws.String(input.Cluster),
		DesiredCount:   aws.Int64(desiredCount),
		ServiceName:    aws.String(input.Name),
		TaskDefinition: aws.String(input.TaskDefinitionArn),
		LaunchType:     aws.String(launchType),
		NetworkConfiguration: &awsecs.NetworkConfiguration{
			AwsvpcConfiguration: &awsecs.AwsVpcConfiguration{
				AssignPublicIp: aws.String(assignPublicIp),
				Subnets:        aws.StringSlice(input.SubnetIds),
			},
		},
	}

	if input.TargetGroupArn != "" && input.Port > 0 {
		createServiceInput.SetLoadBalancers(
			[]*awsecs.LoadBalancer{
				&awsecs.LoadBalancer{
					TargetGroupArn: aws.String(input.TargetGroupArn),
					ContainerPort:  aws.Int64(input.Port),
					ContainerName:  aws.String(input.Name),
				},
			},
		)
	}

	_, err := ecs.svc.CreateService(createServiceInput)

	if err != nil {
		console.ErrorExit(err, "Couldn't create ECS service")
	}

	console.Debug("Created ECS service [%s]", input.Name)

	return
}

func (ecs *ECS) GetDesiredCount(serviceName string) int64 {
	resp, err := ecs.svc.DescribeServices(
		&awsecs.DescribeServicesInput{
			Services: aws.StringSlice([]string{serviceName}),
			Cluster:  aws.String(clusterName),
		},
	)

	if err != nil {
		console.ErrorExit(err, "Could not get desired count")
	}

	if len(resp.Services) == 0 {
		console.ErrorExit(fmt.Errorf("Could not find %s", serviceName), "Could not get desired count")
	}

	return aws.Int64Value(resp.Services[0].DesiredCount)
}

func (ecs *ECS) SetDesiredCount(serviceName string, desiredCount int64) {
	_, err := ecs.svc.UpdateService(
		&awsecs.UpdateServiceInput{
			Cluster:      aws.String(clusterName),
			Service:      aws.String(serviceName),
			DesiredCount: aws.Int64(desiredCount),
		},
	)

	if err != nil {
		console.ErrorExit(err, "Could not scale ECS service")
	}
}

func (ecs *ECS) DestroyService(serviceName string) {
	_, err := ecs.svc.DeleteService(
		&awsecs.DeleteServiceInput{
			Cluster: aws.String(clusterName),
			Service: aws.String(serviceName),
		},
	)

	if err != nil {
		console.ErrorExit(err, "Could not destroy ECS service")
	}
}
