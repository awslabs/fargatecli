package iam

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	awsiam "github.com/aws/aws-sdk-go/service/iam"
)

const ecsTaskExecutionRoleName = "ecsTaskExecutionRole"
const ecsTaskExecutionPolicyArn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
const ecsTaskExecutionRoleAssumeRolePolicyDocument = `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "ecs-tasks.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}`

func (iam *IAM) CreateEcsTaskExecutionRole() string {
	getRoleResp, err := iam.svc.GetRole(
		&awsiam.GetRoleInput{
			RoleName: aws.String(ecsTaskExecutionRoleName),
		},
	)

	if err == nil {
		return *getRoleResp.Role.Arn
	}

	createRoleResp, err := iam.svc.CreateRole(
		&awsiam.CreateRoleInput{
			AssumeRolePolicyDocument: aws.String(ecsTaskExecutionRoleAssumeRolePolicyDocument),
			RoleName:                 aws.String(ecsTaskExecutionRoleName),
		},
	)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	ecsTaskExecutionRoleArn := *createRoleResp.Role.Arn

	_, err = iam.svc.AttachRolePolicy(
		&awsiam.AttachRolePolicyInput{
			RoleName:  aws.String(ecsTaskExecutionRoleName),
			PolicyArn: aws.String(ecsTaskExecutionPolicyArn),
		},
	)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return ecsTaskExecutionRoleArn
}
