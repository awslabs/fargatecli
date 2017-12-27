package ecs

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	awsecs "github.com/aws/aws-sdk-go/service/ecs"
	"github.com/jpignata/fargate/console"
)

const (
	networkInterfaceId        = "networkInterfaceId"
	startedByFormat           = "fargate:%s"
	taskGroupStartedByPattern = "fargate:(.*)"
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
	StartedBy     string
}

func (t *Task) RunningFor() time.Duration {
	return time.Now().Sub(t.CreatedAt).Truncate(time.Second)
}

type TaskGroup struct {
	TaskGroupName string
	Instances     int64
}

type RunTaskInput struct {
	ClusterName       string
	Count             int64
	TaskDefinitionArn string
	TaskName          string
	SubnetIds         []string
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

func (ecs *ECS) DescribeTasksForService(serviceName string) []Task {
	return ecs.listTasks(
		&awsecs.ListTasksInput{
			Cluster:     aws.String(clusterName),
			LaunchType:  aws.String(awsecs.CompatibilityFargate),
			ServiceName: aws.String(serviceName),
		},
	)
}

func (ecs *ECS) DescribeTasksForTaskGroup(taskGroupName string) []Task {
	return ecs.listTasks(
		&awsecs.ListTasksInput{
			StartedBy: aws.String(fmt.Sprintf(startedByFormat, taskGroupName)),
			Cluster:   aws.String(clusterName),
		},
	)
}

func (ecs *ECS) ListTaskGroups() []*TaskGroup {
	var taskGroups []*TaskGroup

	taskGroupStartedByRegexp := regexp.MustCompile(taskGroupStartedByPattern)

	input := &awsecs.ListTasksInput{
		Cluster: aws.String(clusterName),
	}

OUTER:
	for _, task := range ecs.listTasks(input) {
		matches := taskGroupStartedByRegexp.FindStringSubmatch(task.StartedBy)

		if len(matches) == 2 {
			taskGroupName := matches[1]

			for _, taskGroup := range taskGroups {
				if taskGroup.TaskGroupName == taskGroupName {
					taskGroup.Instances++
					continue OUTER
				}
			}

			taskGroups = append(
				taskGroups,
				&TaskGroup{
					TaskGroupName: taskGroupName,
					Instances:     1,
				},
			)
		}
	}

	return taskGroups
}

func (ecs *ECS) StopTasks(taskIds []string) {
	for _, taskId := range taskIds {
		ecs.StopTask(taskId)
	}
}

func (ecs *ECS) StopTask(taskId string) {
	_, err := ecs.svc.StopTask(
		&awsecs.StopTaskInput{
			Cluster: aws.String(clusterName),
			Task:    aws.String(taskId),
		},
	)

	if err != nil {
		console.ErrorExit(err, "Could not stop ECS task")
	}
}

func (ecs *ECS) listTasks(input *awsecs.ListTasksInput) []Task {
	var tasks []Task
	var taskArnBatches [][]string

	err := ecs.svc.ListTasksPages(
		input,
		func(resp *awsecs.ListTasksOutput, lastPage bool) bool {
			if len(resp.TaskArns) > 0 {
				taskArnBatches = append(taskArnBatches, aws.StringValueSlice(resp.TaskArns))
			}

			return true
		},
	)

	if err != nil {
		console.ErrorExit(err, "Could not list ECS tasks")
	}

	if len(taskArnBatches) > 0 {
		for _, taskArnBatch := range taskArnBatches {
			for _, task := range ecs.describeTasks(taskArnBatch) {
				tasks = append(tasks, task)
			}
		}
	}

	return tasks
}

func (ecs *ECS) describeTasks(taskArns []string) []Task {
	var tasks []Task

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
			StartedBy:     aws.StringValue(t.StartedBy),
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
