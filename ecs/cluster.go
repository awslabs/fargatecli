package ecs

import (
	"github.com/aws/aws-sdk-go/aws"
	awsecs "github.com/aws/aws-sdk-go/service/ecs"
	"github.com/jpignata/fargate/console"
)

const clusterName = "fargate"

func (ecs *ECS) CreateCluster() error {
	console.Debug("Creating ECS cluster")

	_, err := ecs.svc.CreateCluster(
		&awsecs.CreateClusterInput{
			ClusterName: aws.String(clusterName),
		},
	)

	if err != nil {
		return err
	}

	console.Debug("Created ECS cluster %s", clusterName)

	return nil
}
