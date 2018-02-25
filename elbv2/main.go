// Package elbv2 is a client for Elastic Load Balancing (v2).
package elbv2

//go:generate mockgen -package client -destination=mock/client/client.go github.com/jpignata/fargate/elbv2 Client
//go:generate mockgen -package sdk -source ../vendor/github.com/aws/aws-sdk-go/service/elbv2/elbv2iface/interface.go -destination=mock/sdk/elbv2iface.go github.com/aws/aws-sdk-go/service/elbv2/elbv2iface ELBV2API

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/aws/aws-sdk-go/service/elbv2/elbv2iface"
)

// Client represents a method for accessing Elastic Load Balancing (v2).
type Client interface {
	CreateListener(CreateListenerParameters) (string, error)
	DescribeListeners(string) (Listeners, error)

	DescribeLoadBalancers() (LoadBalancers, error)
	DescribeLoadBalancersByName([]string) (LoadBalancers, error)
	CreateLoadBalancer(CreateLoadBalancerParameters) (string, error)

	CreateTargetGroup(CreateTargetGroupParameters) (string, error)
}

// SDKClient implements access to Elastic Load Balancing (v2) via the AWS SDK.
type SDKClient struct {
	client elbv2iface.ELBV2API
}

// New returns an SDKClient configured with the given session.
func New(sess *session.Session) SDKClient {
	return SDKClient{
		client: elbv2.New(sess),
	}
}
