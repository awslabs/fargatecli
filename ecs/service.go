package ecs

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	awsecs "github.com/aws/aws-sdk-go/service/ecs"
	"github.com/jpignata/fargate/console"
)

type CreateServiceInput struct {
	Cluster           string
	DesiredCount      int64
	Name              string
	Port              int64
	SubnetIds         []string
	TargetGroupArn    string
	TaskDefinitionArn string
	SecurityGroupId   string
}

type Service struct {
	Cluster           string
	Cpu               string
	Deployments       []Deployment
	DesiredCount      int64
	EnvVars           []EnvVar
	Events            []Event
	Image             string
	Memory            string
	Name              string
	PendingCount      int64
	RunningCount      int64
	TargetGroupArn    string
	TaskDefinitionArn string
}

type Event struct {
	CreatedAt time.Time
	Message   string
}

type Deployment struct {
	CreatedAt    time.Time
	DesiredCount int64
	Id           string
	Image        string
	PendingCount int64
	RunningCount int64
	Status       string
}

func (s *Service) AddEvent(e Event) {
	s.Events = append(s.Events, e)
}

func (s *Service) AddDeployment(d Deployment) {
	s.Deployments = append(s.Deployments, d)
}

func (ecs *ECS) CreateService(input *CreateServiceInput) {
	console.Debug("Creating ECS service")

	createServiceInput := &awsecs.CreateServiceInput{
		Cluster:        aws.String(input.Cluster),
		DesiredCount:   aws.Int64(input.DesiredCount),
		ServiceName:    aws.String(input.Name),
		TaskDefinition: aws.String(input.TaskDefinitionArn),
		LaunchType:     aws.String(awsecs.CompatibilityFargate),
		NetworkConfiguration: &awsecs.NetworkConfiguration{
			AwsvpcConfiguration: &awsecs.AwsVpcConfiguration{
				AssignPublicIp: aws.String(awsecs.AssignPublicIpEnabled),
				Subnets:        aws.StringSlice(input.SubnetIds),
				SecurityGroups: aws.StringSlice([]string{input.SecurityGroupId}),
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

func (ecs *ECS) DescribeService(serviceName string) Service {
	services := ecs.DescribeServices([]string{serviceName})

	if len(services) == 0 {
		console.ErrorExit(fmt.Errorf("Could not find %s", serviceName), "Could not describe ECS service")
	}

	return services[0]
}

func (ecs *ECS) GetDesiredCount(serviceName string) int64 {
	service := ecs.DescribeService(serviceName)
	return service.DesiredCount
}

func (ecs *ECS) SetDesiredCount(serviceName string, desiredCount int64) {
	_, err := ecs.svc.UpdateService(
		&awsecs.UpdateServiceInput{
			Cluster:      aws.String(ecs.ClusterName),
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
			Cluster: aws.String(ecs.ClusterName),
			Service: aws.String(serviceName),
		},
	)

	if err != nil {
		console.ErrorExit(err, "Could not destroy ECS service")
	}
}

func (ecs *ECS) ListServices() []Service {
	var services []Service
	var serviceArnBatches [][]string

	err := ecs.svc.ListServicesPages(
		&awsecs.ListServicesInput{
			Cluster:    aws.String(ecs.ClusterName),
			LaunchType: aws.String(awsecs.CompatibilityFargate),
		},

		func(resp *awsecs.ListServicesOutput, lastPage bool) bool {
			if len(resp.ServiceArns) > 0 {
				serviceArnBatches = append(serviceArnBatches, aws.StringValueSlice(resp.ServiceArns))
			}

			return true
		},
	)

	if err != nil {
		console.ErrorExit(err, "Could not list ECS services")
	}

	if len(serviceArnBatches) > 0 {
		for _, serviceArnBatch := range serviceArnBatches {
			for _, service := range ecs.DescribeServices(serviceArnBatch) {
				services = append(services, service)
			}
		}
	}

	return services
}

func (ecs *ECS) DescribeServices(serviceArns []string) []Service {
	var services []Service

	resp, err := ecs.svc.DescribeServices(
		&awsecs.DescribeServicesInput{
			Cluster:  aws.String(ecs.ClusterName),
			Services: aws.StringSlice(serviceArns),
		},
	)

	if err != nil {
		console.ErrorExit(err, "Could not describe ECS services")
	}

	for _, service := range resp.Services {
		s := Service{
			Name:              aws.StringValue(service.ServiceName),
			DesiredCount:      aws.Int64Value(service.DesiredCount),
			PendingCount:      aws.Int64Value(service.PendingCount),
			RunningCount:      aws.Int64Value(service.RunningCount),
			TaskDefinitionArn: aws.StringValue(service.TaskDefinition),
		}

		taskDefinition := ecs.DescribeTaskDefinition(aws.StringValue(service.TaskDefinition))

		s.Cpu = aws.StringValue(taskDefinition.Cpu)
		s.Memory = aws.StringValue(taskDefinition.Memory)

		if len(service.LoadBalancers) > 0 {
			s.TargetGroupArn = aws.StringValue(service.LoadBalancers[0].TargetGroupArn)
		}

		if len(taskDefinition.ContainerDefinitions) > 0 {
			s.Image = aws.StringValue(taskDefinition.ContainerDefinitions[0].Image)

			for _, env := range taskDefinition.ContainerDefinitions[0].Environment {
				s.EnvVars = append(
					s.EnvVars,
					EnvVar{
						Key:   aws.StringValue(env.Name),
						Value: aws.StringValue(env.Value),
					},
				)
			}
		}

		for _, event := range service.Events {
			s.AddEvent(
				Event{
					CreatedAt: aws.TimeValue(event.CreatedAt),
					Message:   aws.StringValue(event.Message),
				},
			)
		}

		for _, d := range service.Deployments {
			deployment := Deployment{
				Status:       aws.StringValue(d.Status),
				DesiredCount: aws.Int64Value(d.DesiredCount),
				PendingCount: aws.Int64Value(d.PendingCount),
				RunningCount: aws.Int64Value(d.RunningCount),
				CreatedAt:    aws.TimeValue(d.CreatedAt),
				Id:           ecs.getDeploymentId(aws.StringValue(d.TaskDefinition)),
			}

			deploymentTaskDefinition := ecs.DescribeTaskDefinition(aws.StringValue(d.TaskDefinition))
			deployment.Image = aws.StringValue(deploymentTaskDefinition.ContainerDefinitions[0].Image)

			s.AddDeployment(deployment)
		}

		services = append(services, s)
	}

	return services
}

func (ecs *ECS) UpdateServiceTaskDefinition(serviceName, taskDefinitionArn string) {
	_, err := ecs.svc.UpdateService(
		&awsecs.UpdateServiceInput{
			Cluster:        aws.String(ecs.ClusterName),
			Service:        aws.String(serviceName),
			TaskDefinition: aws.String(taskDefinitionArn),
		},
	)

	if err != nil {
		console.ErrorExit(err, "Could not update ECS service task definition")
	}
}
