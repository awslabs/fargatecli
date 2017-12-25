package elbv2

import (
	"github.com/aws/aws-sdk-go/aws"
	awselbv2 "github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/jpignata/fargate/console"
)

type CreateTargetGroupInput struct {
	Name     string
	Protocol string
	Port     int64
	VpcId    string
}

type TargetGroup struct {
	Name            string
	Arn             string
	LoadBalancerArn string
}

func (elbv2 *ELBV2) CreateTargetGroup(input *CreateTargetGroupInput) string {
	console.Debug("Creating ELB target group")

	resp, err := elbv2.svc.CreateTargetGroup(
		&awselbv2.CreateTargetGroupInput{
			Name:       aws.String(input.Name),
			Port:       aws.Int64(input.Port),
			Protocol:   aws.String(input.Protocol),
			TargetType: aws.String(awselbv2.TargetTypeEnumIp),
			VpcId:      aws.String(input.VpcId),
		},
	)

	if err != nil || len(resp.TargetGroups) != 1 {
		console.ErrorExit(err, "Could not create ELB target group")
	}

	targetGroupArn := aws.StringValue(resp.TargetGroups[0].TargetGroupArn)

	console.Debug("Created ELB target group [%s]", input.Name)

	return targetGroupArn
}

func (elbv2 *ELBV2) DeleteTargetGroup(targetGroupName string) {
	console.Debug("Deleting ELB target group")

	targetGroup := elbv2.describeTargetGroupByName(targetGroupName)

	elbv2.svc.DeleteTargetGroup(
		&awselbv2.DeleteTargetGroupInput{
			TargetGroupArn: targetGroup.TargetGroupArn,
		},
	)
}

func (elbv2 *ELBV2) GetTargetGroupLoadBalancerArn(targetGroupName string) string {
	targetGroup := elbv2.describeTargetGroupByName(targetGroupName)

	if len(targetGroup.LoadBalancerArns) == 0 {
		return aws.StringValue(targetGroup.LoadBalancerArns[0])
	} else {
		return ""
	}
}

func (elbv2 *ELBV2) DescribeTargetGroups(targetGroupArns []string) []TargetGroup {
	var targetGroups []TargetGroup

	resp, err := elbv2.svc.DescribeTargetGroups(
		&awselbv2.DescribeTargetGroupsInput{
			TargetGroupArns: aws.StringSlice(targetGroupArns),
		},
	)

	if err != nil {
		console.ErrorExit(err, "Could not describe ELB target groups")
	}

	for _, targetGroup := range resp.TargetGroups {
		tg := TargetGroup{
			Name: aws.StringValue(targetGroup.TargetGroupName),
			Arn:  aws.StringValue(targetGroup.TargetGroupArn),
		}

		if len(targetGroup.LoadBalancerArns) > 0 {
			tg.LoadBalancerArn = aws.StringValue(targetGroup.LoadBalancerArns[0])
		}

		targetGroups = append(targetGroups, tg)
	}

	return targetGroups
}

func (elbv2 *ELBV2) describeTargetGroupByName(targetGroupName string) *awselbv2.TargetGroup {
	resp, err := elbv2.svc.DescribeTargetGroups(
		&awselbv2.DescribeTargetGroupsInput{
			Names: aws.StringSlice([]string{targetGroupName}),
		},
	)

	if err != nil {
		console.ErrorExit(err, "Could not describe ELB target groups")
	}

	if len(resp.TargetGroups) != 1 {
		console.IssueExit("Could not describe ELB target groups")
	}

	return resp.TargetGroups[0]
}

func (elbv2 *ELBV2) describeTargetGroupByArn(targetGroupArn string) *awselbv2.TargetGroup {
	resp, err := elbv2.svc.DescribeTargetGroups(
		&awselbv2.DescribeTargetGroupsInput{
			TargetGroupArns: aws.StringSlice([]string{targetGroupArn}),
		},
	)

	if err != nil {
		console.ErrorExit(err, "Could not describe ELB target groups")
	}

	if len(resp.TargetGroups) != 1 {
		console.IssueExit("Could not describe ELB target groups")
	}

	return resp.TargetGroups[0]
}
