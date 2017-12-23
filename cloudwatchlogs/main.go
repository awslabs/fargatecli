package cloudwatchlogs

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/jpignata/fargate/console"
)

type CloudWatchLogs struct {
	svc *cloudwatchlogs.CloudWatchLogs
}

func New() CloudWatchLogs {
	sess, err := session.NewSession()

	if err != nil {
		console.ErrorExit(err, "Error creating CloudWatch Logs session")
	}

	return CloudWatchLogs{
		svc: cloudwatchlogs.New(sess),
	}
}
