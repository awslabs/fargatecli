package cloudwatchlogs

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	awscwl "github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/jpignata/fargate/console"
)

type GetLogsInput struct {
	Filter         string
	LogGroupName   string
	LogStreamNames []string
	EndTime        time.Time
	StartTime      time.Time
}

type LogLine struct {
	EventId       string
	LogStreamName string
	Message       string
	Timestamp     time.Time
}

func (cwl *CloudWatchLogs) CreateLogGroup(logGroupName string, a ...interface{}) string {
	formattedLogGroupName := fmt.Sprintf(logGroupName, a...)
	_, err := cwl.svc.CreateLogGroup(
		&awscwl.CreateLogGroupInput{
			LogGroupName: aws.String(formattedLogGroupName),
		},
	)

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			switch awsErr.Code() {
			case awscwl.ErrCodeResourceAlreadyExistsException:
				return formattedLogGroupName
			default:
				console.ErrorExit(awsErr, "Could not create Cloudwatch Logs log group")
			}
		}
	}

	return formattedLogGroupName
}

func (cwl *CloudWatchLogs) GetLogs(i *GetLogsInput) []LogLine {
	var logLines []LogLine

	input := &awscwl.FilterLogEventsInput{
		LogGroupName: aws.String(i.LogGroupName),
		Interleaved:  aws.Bool(true),
	}

	if !i.StartTime.IsZero() {
		input.SetStartTime(i.StartTime.UTC().UnixNano() / int64(time.Millisecond))
	}

	if !i.EndTime.IsZero() {
		input.SetEndTime(i.EndTime.UTC().UnixNano() / int64(time.Millisecond))
	}

	if i.Filter != "" {
		input.SetFilterPattern(i.Filter)
	}

	if len(i.LogStreamNames) > 0 {
		input.SetLogStreamNames(aws.StringSlice(i.LogStreamNames))
	}

	err := cwl.svc.FilterLogEventsPages(
		input,
		func(resp *awscwl.FilterLogEventsOutput, lastPage bool) bool {
			for _, event := range resp.Events {
				logLines = append(logLines,
					LogLine{
						EventId:       aws.StringValue(event.EventId),
						Message:       aws.StringValue(event.Message),
						LogStreamName: aws.StringValue(event.LogStreamName),
						Timestamp:     time.Unix(0, aws.Int64Value(event.Timestamp)*int64(time.Millisecond)),
					},
				)

			}

			return true
		},
	)

	if err != nil {
		console.ErrorExit(err, "Could not get logs")
	}

	return logLines
}
