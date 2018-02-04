package ecs

import (
	"github.com/aws/aws-sdk-go/aws"
	awsecs "github.com/aws/aws-sdk-go/service/ecs"
)

func (ecs *ECS) CreateCluster() (string, error) {
	input := &awsecs.CreateClusterInput{
		ClusterName: aws.String(ecs.ClusterName),
	}

	resp, err := ecs.svc.CreateCluster(input)

	return aws.StringValue(resp.Cluster.ClusterArn), err
}
