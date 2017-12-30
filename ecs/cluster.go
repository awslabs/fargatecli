package ecs

import (
	"github.com/aws/aws-sdk-go/aws"
	awsecs "github.com/aws/aws-sdk-go/service/ecs"
	"github.com/jpignata/fargate/console"
)

func (ecs *ECS) CreateCluster() error {
	console.Debug("Creating ECS cluster")

	_, err := ecs.svc.CreateCluster(
		&awsecs.CreateClusterInput{
			ClusterName: aws.String(ecs.ClusterName),
		},
	)

	if err != nil {
		return err
	}

	console.Debug("Created ECS cluster %s", ecs.ClusterName)

	return nil
}
