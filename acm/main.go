package acm

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/acm"
)

type ACM struct {
	svc *acm.ACM
}

func New(sess *session.Session) ACM {
	return ACM{
		svc: acm.New(sess),
	}
}
