package ecs

import (
	"github.com/aws/aws-sdk-go/aws"
	awsecs "github.com/aws/aws-sdk-go/service/ecs"
)

func (ecs *ECS) CreateCluster() (string, error) {
	input := &awsecs.CreateClusterInput{
		ClusterName: aws.String(ecs.ClusterName),
	}

	if output, err := ecs.svc.CreateCluster(input); err == nil {
		return aws.StringValue(output.Cluster.ClusterArn), nil
	} else {
		return "", err
	}
}
