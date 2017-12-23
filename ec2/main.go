package ec2

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type EC2 struct {
	svc *ec2.EC2
}

func New(sess *session.Session) EC2 {
	return EC2{
		svc: ec2.New(sess),
	}
}
