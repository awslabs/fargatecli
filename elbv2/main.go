package elbv2

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/jpignata/fargate/console"
)

type ELBV2 struct {
	svc *elbv2.ELBV2
}

func New() ELBV2 {
	sess, err := session.NewSession()

	if err != nil {
		console.ErrorExit(err, "Could not create VPC session")
	}

	return ELBV2{
		svc: elbv2.New(sess),
	}
}
