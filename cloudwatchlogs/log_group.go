package cloudwatchlogs

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	awscwl "github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/jpignata/fargate/console"
)

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
