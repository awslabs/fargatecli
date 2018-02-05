package acm

//go:generate mockgen -package client -destination=mock/client/client.go github.com/jpignata/fargate/acm Client
//go:generate mockgen -package sdk -source ../vendor/github.com/aws/aws-sdk-go/service/acm/acmiface/interface.go -destination=mock/sdk/acmiface.go github.com/aws/aws-sdk-go/service/acm/acmiface ACMAPI

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/acm"
	"github.com/aws/aws-sdk-go/service/acm/acmiface"
)

// Client represents a method for accessing AWS Certificate Manager.
type Client interface {
	DeleteCertificate(string) error
	InflateCertificate(Certificate) (Certificate, error)
	ListCertificates() (Certificates, error)
	RequestCertificate(string, []string) (string, error)
	ImportCertificate([]byte, []byte, []byte) (string, error)
}

// SDKClient implements access to AWS Certificate Manager via the AWS SDK.
type SDKClient struct {
	client acmiface.ACMAPI
}

// New returns an SDKClient configured with the current session.
func New(sess *session.Session) SDKClient {
	return SDKClient{
		client: acm.New(sess),
	}
}
