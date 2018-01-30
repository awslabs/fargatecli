package acm

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/acm"
	"github.com/aws/aws-sdk-go/service/acm/acmiface"
)

type Client interface {
	RequestCertificate(string, []string) (string, error)
	DeleteCertificate(string) error
	GetCertificateArns(string) ([]string, error)
}

type SDKClient struct {
	client acmiface.ACMAPI
}

func New(sess *session.Session) SDKClient {
	return SDKClient{
		client: acm.New(sess),
	}
}
