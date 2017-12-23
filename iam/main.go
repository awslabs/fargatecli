package iam

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
)

type IAM struct {
	svc *iam.IAM
}

func New(sess *session.Session) IAM {
	return IAM{
		svc: iam.New(sess),
	}
}
