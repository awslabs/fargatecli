package ecs

import (
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	awsecs "github.com/aws/aws-sdk-go/service/ecs"
	"github.com/jpignata/fargate/console"
)

const (
	networkInterfaceId = "networkInterfaceId"
	startedByFormat    = "fargate:%s"
)

type Task struct {
	DeploymentId  string
	TaskId        string
	Cpu           string
	CreatedAt     time.Time
	DesiredStatus string
	Image         string
	LastStatus    string
	Memory        string
	EniId         string
}

func (t *Task) RunningFor() time.Duration {
	return time.Now().Sub(t.CreatedAt).Truncate(time.Second)
}

type RunTaskInput struct {
	ClusterName       string
	Count             int64
	TaskDefinitionArn string
	TaskName          string
	SubnetIds         []string
}

func (ecs *ECS) DescribeTasksForService(serviceName string) []Task {
	var tasks []Task

	taskArns := ecs.ListTasksForService(serviceName)

	if len(taskArns) == 0 {
		return tasks
	}

	resp, err := ecs.svc.DescribeTasks(
		&awsecs.DescribeTasksInput{
			Cluster: aws.String(clusterName),
			Tasks:   aws.StringSlice(taskArns),
		},
	)

	if err != nil {
		console.ErrorExit(err, "Could not describe ECS tasks")
	}

	for _, t := range resp.Tasks {
		taskArn := aws.StringValue(t.TaskArn)
		contents := strings.Split(taskArn, "/")
		taskId := contents[len(contents)-1]

		task := Task{
			Cpu:           aws.StringValue(t.Cpu),
			CreatedAt:     aws.TimeValue(t.CreatedAt),
			DeploymentId:  ecs.getDeploymentId(aws.StringValue(t.TaskDefinitionArn)),
			DesiredStatus: aws.StringValue(t.DesiredStatus),
			LastStatus:    aws.StringValue(t.LastStatus),
			Memory:        aws.StringValue(t.Memory),
			TaskId:        taskId,
		}

		taskDefinition := ecs.DescribeTaskDefinition(aws.StringValue(t.TaskDefinitionArn))
		task.Image = aws.StringValue(taskDefinition.ContainerDefinitions[0].Image)

		if len(t.Attachments) == 1 {
			for _, detail := range t.Attachments[0].Details {
				if aws.StringValue(detail.Name) == networkInterfaceId {
					task.EniId = aws.StringValue(detail.Value)
					break
				}
			}
		}

		tasks = append(tasks, task)
	}

	return tasks
}

func (ecs *ECS) ListTasksForService(serviceName string) []string {
	resp, err := ecs.svc.ListTasks(
		&awsecs.ListTasksInput{
			Cluster:     aws.String(clusterName),
			ServiceName: aws.String(serviceName),
		},
	)

	if err != nil {
		console.ErrorExit(err, "Could not list ECS tasks")
	}

	return aws.StringValueSlice(resp.TaskArns)
}

func (ecs *ECS) RunTask(i *RunTaskInput) {
	_, err := ecs.svc.RunTask(
		&awsecs.RunTaskInput{
			Cluster:        aws.String(i.ClusterName),
			Count:          aws.Int64(i.Count),
			TaskDefinition: aws.String(i.TaskDefinitionArn),
			LaunchType:     aws.String(awsecs.CompatibilityFargate),
			StartedBy:      aws.String(fmt.Sprintf(startedByFormat, i.TaskName)),
			NetworkConfiguration: &awsecs.NetworkConfiguration{
				AwsvpcConfiguration: &awsecs.AwsVpcConfiguration{
					AssignPublicIp: aws.String(awsecs.AssignPublicIpEnabled),
					Subnets:        aws.StringSlice(i.SubnetIds),
				},
			},
		},
	)

	if err != nil {
		console.ErrorExit(err, "Could not run ECS task")
	}
}
