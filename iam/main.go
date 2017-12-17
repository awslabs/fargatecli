package iam

import (
	"os"

	"github.com/aws/aws-sdk-go/aws/session"
	awsiam "github.com/aws/aws-sdk-go/service/iam"
	"github.com/fatih/color"
)

type IAM struct {
	svc *awsiam.IAM
}

func New() IAM {
	sess, err := session.NewSession()

	if err != nil {
		color.Red("Error creating IAM session: ", err)
		os.Exit(1)
	}

	return IAM{
		svc: awsiam.New(sess),
	}
}
