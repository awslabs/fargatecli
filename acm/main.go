package acm

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/acm"
	"github.com/jpignata/fargate/console"
)

type ACM struct {
	svc *acm.ACM
}

func New() ACM {
	sess, err := session.NewSession()

	if err != nil {
		console.ErrorExit(err, "Could not create ACM session")
	}

	return ACM{
		svc: acm.New(sess),
	}
}
