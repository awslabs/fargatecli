package elbv2

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/elbv2"
)

type ELBV2 struct {
	svc *elbv2.ELBV2
}

func New(sess *session.Session) ELBV2 {
	return ELBV2{
		svc: elbv2.New(sess),
	}
}
