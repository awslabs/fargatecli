package elbv2

import (
	"github.com/aws/aws-sdk-go/aws"
	awselbv2 "github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/jpignata/fargate/console"
)

type TargetGroup struct {
	Name            string
	Arn             string
	LoadBalancerARN string
}

type CreateTargetGroupParameters struct {
	Name     string
	Port     int64
	Protocol string
	VPCID    string
}

func (elbv2 SDKClient) CreateTargetGroup(i CreateTargetGroupParameters) (string, error) {
	resp, err := elbv2.client.CreateTargetGroup(
		&awselbv2.CreateTargetGroupInput{
			Name:       aws.String(i.Name),
			Port:       aws.Int64(i.Port),
			Protocol:   aws.String(i.Protocol),
			TargetType: aws.String(awselbv2.TargetTypeEnumIp),
			VpcId:      aws.String(i.VPCID),
		},
	)

	if err != nil {
		return "", err
	}

	return aws.StringValue(resp.TargetGroups[0].TargetGroupArn), nil
}

func (elbv2 SDKClient) DeleteTargetGroup(targetGroupName string) {
	console.Debug("Deleting ELB target group")

	targetGroup := elbv2.describeTargetGroupByName(targetGroupName)

	elbv2.client.DeleteTargetGroup(
		&awselbv2.DeleteTargetGroupInput{
			TargetGroupArn: targetGroup.TargetGroupArn,
		},
	)
}

func (elbv2 SDKClient) DeleteTargetGroupByArn(targetGroupARN string) {
	_, err := elbv2.client.DeleteTargetGroup(
		&awselbv2.DeleteTargetGroupInput{
			TargetGroupArn: aws.String(targetGroupARN),
		},
	)

	if err != nil {
		console.ErrorExit(err, "Could not delete ELB target group")
	}
}

func (elbv2 SDKClient) GetTargetGroupArn(targetGroupName string) string {
	resp, _ := elbv2.client.DescribeTargetGroups(
		&awselbv2.DescribeTargetGroupsInput{
			Names: aws.StringSlice([]string{targetGroupName}),
		},
	)

	if len(resp.TargetGroups) == 1 {
		return aws.StringValue(resp.TargetGroups[0].TargetGroupArn)
	}

	return ""
}

func (elbv2 SDKClient) GetTargetGroupLoadBalancerArn(targetGroupARN string) string {
	targetGroup := elbv2.describeTargetGroupByArn(targetGroupARN)

	if len(targetGroup.LoadBalancerArns) > 0 {
		return aws.StringValue(targetGroup.LoadBalancerArns[0])
	} else {
		return ""
	}
}

func (elbv2 SDKClient) DescribeTargetGroups(targetGroupARNs []string) []TargetGroup {
	var targetGroups []TargetGroup

	resp, err := elbv2.client.DescribeTargetGroups(
		&awselbv2.DescribeTargetGroupsInput{
			TargetGroupArns: aws.StringSlice(targetGroupARNs),
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
			tg.LoadBalancerARN = aws.StringValue(targetGroup.LoadBalancerArns[0])
		}

		targetGroups = append(targetGroups, tg)
	}

	return targetGroups
}

func (elbv2 SDKClient) describeTargetGroupByName(targetGroupName string) *awselbv2.TargetGroup {
	resp, err := elbv2.client.DescribeTargetGroups(
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

func (elbv2 SDKClient) describeTargetGroupByArn(targetGroupARN string) *awselbv2.TargetGroup {
	resp, err := elbv2.client.DescribeTargetGroups(
		&awselbv2.DescribeTargetGroupsInput{
			TargetGroupArns: aws.StringSlice([]string{targetGroupARN}),
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
