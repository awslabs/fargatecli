package ecs

import (
	"os"

	"github.com/aws/aws-sdk-go/aws/session"
	awsecs "github.com/aws/aws-sdk-go/service/ecs"
	"github.com/fatih/color"
)

type ECS struct {
	svc *awsecs.ECS
}

func New() ECS {
	sess, err := session.NewSession()

	if err != nil {
		color.Red("Error creating ECS session: ", err)
		os.Exit(1)
	}

	return ECS{
		svc: awsecs.New(sess),
	}
}
