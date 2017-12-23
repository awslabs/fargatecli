package route53

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
)

type Route53 struct {
	svc *route53.Route53
}

func New(sess *session.Session) Route53 {
	return Route53{
		svc: route53.New(sess),
	}
}
