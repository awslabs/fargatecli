package ecs

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
)

type ECS struct {
	svc  *ecs.ECS
	sess *session.Session
}

func New(sess *session.Session) ECS {
	return ECS{
		svc: ecs.New(sess),
	}
}
