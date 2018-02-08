package ec2

//go:generate mockgen -package client -destination=mock/client/client.go github.com/jpignata/fargate/ec2 Client
//go:generate mockgen -package sdk -source ../vendor/github.com/aws/aws-sdk-go/service/ec2/ec2iface/interface.go -destination=mock/sdk/ec2iface.go github.com/aws/aws-sdk-go/service/ec2/ec2iface EC2API

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
)

type Client interface {
	GetDefaultSubnetIDs() ([]string, error)
	GetSubnetVPCID(string) (string, error)
	GetDefaultSecurityGroupID() (string, error)
	CreateDefaultSecurityGroup() (string, error)
	AuthorizeAllSecurityGroupIngress(string) error
}

type SDKClient struct {
	client ec2iface.EC2API
}

func New(sess *session.Session) SDKClient {
	return SDKClient{
		client: ec2.New(sess),
	}
}
