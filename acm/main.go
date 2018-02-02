package acm

//go:generate mockgen -package client -destination=mock/client/client.go github.com/jpignata/fargate/acm Client
//go:generate mockgen -package sdk -destination=mock/sdk/acmiface.go github.com/aws/aws-sdk-go/service/acm/acmiface ACMAPI

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/acm"
	"github.com/aws/aws-sdk-go/service/acm/acmiface"
)

type Client interface {
	DeleteCertificate(string) error
	InflateCertificate(Certificate) (Certificate, error)
	ListCertificates() (Certificates, error)
	RequestCertificate(string, []string) (string, error)
}

type SDKClient struct {
	client acmiface.ACMAPI
}

func New(sess *session.Session) SDKClient {
	return SDKClient{
		client: acm.New(sess),
	}
}
