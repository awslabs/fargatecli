package route53

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/jpignata/fargate/console"
)

type Route53 struct {
	svc *route53.Route53
}

func New() Route53 {
	sess, err := session.NewSession()

	if err != nil {
		console.ErrorExit(err, "Could not create Route53 session")
	}

	return Route53{
		svc: route53.New(sess),
	}
}
