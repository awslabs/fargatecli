package ecr

import (
	"os"

	"github.com/aws/aws-sdk-go/aws/session"
	awsecr "github.com/aws/aws-sdk-go/service/ecr"
	"github.com/fatih/color"
)

type ECR struct {
	svc *awsecr.ECR
}

func New() ECR {
	sess, err := session.NewSession()

	if err != nil {
		color.Red("Error creating ECR session: ", err)
		os.Exit(1)
	}

	return ECR{
		svc: awsecr.New(sess),
	}
}
