package acm

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/acm"
	"github.com/aws/aws-sdk-go/service/acm/acmiface"
)

type ACMClient interface {
	RequestCertificate(string, []string) (string, error)
}

type ACM struct {
	svc acmiface.ACMAPI
}

func New(sess *session.Session) ACM {
	return ACM{
		svc: acm.New(sess),
	}
}
