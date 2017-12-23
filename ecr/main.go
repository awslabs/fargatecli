package ecr

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
)

type ECR struct {
	svc *ecr.ECR
}

func New(sess *session.Session) ECR {
	return ECR{
		svc: ecr.New(sess),
	}
}
