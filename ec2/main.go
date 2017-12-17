package ec2

import (
	"os"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/fatih/color"
)

type EC2 struct {
	svc *ec2.EC2
}

func New() EC2 {
	sess, err := session.NewSession()

	if err != nil {
		color.Red("Error creating VPC session: ", err)
		os.Exit(1)
	}

	return EC2{
		svc: ec2.New(sess),
	}
}
