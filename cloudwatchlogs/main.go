package cloudwatchlogs

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
)

type CloudWatchLogs struct {
	svc *cloudwatchlogs.CloudWatchLogs
}

func New(sess *session.Session) CloudWatchLogs {
	return CloudWatchLogs{
		svc: cloudwatchlogs.New(sess),
	}
}
