package ecs

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
)

type ECS struct {
	svc         *ecs.ECS
	ClusterName string
}

func New(sess *session.Session, clusterName string) ECS {
	return ECS{
		ClusterName: clusterName,
		svc:         ecs.New(sess),
	}
}
