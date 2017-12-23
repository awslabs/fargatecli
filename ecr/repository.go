package ecr

import (
	"encoding/base64"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	awsecr "github.com/aws/aws-sdk-go/service/ecr"
	"github.com/jpignata/fargate/console"
)

func (ecr *ECR) CreateRepository(repositoryName string) string {
	console.Debug("Creating Amazon ECR repository")

	resp, err := ecr.svc.CreateRepository(
		&awsecr.CreateRepositoryInput{
			RepositoryName: aws.String(repositoryName),
		},
	)

	if err != nil {
		console.ErrorExit(err, "Couldn't create Amazon ECR repository")
	}

	console.Debug("Created Amazon ECR repository [%s]", *resp.Repository.RepositoryName)
	return aws.StringValue(resp.Repository.RepositoryUri)
}

func (ecr *ECR) IsRepositoryCreated(repositoryName string) bool {
	resp, err := ecr.svc.DescribeRepositories(
		&awsecr.DescribeRepositoriesInput{
			RepositoryNames: aws.StringSlice([]string{repositoryName}),
		},
	)

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			switch awsErr.Code() {
			case awsecr.ErrCodeRepositoryNotFoundException:
				return false
			default:
				console.ErrorExit(awsErr, "Could not create Cloudwatch Logs log group")
			}
		}

		console.ErrorExit(err, "Couldn't describe Amazon ECR repositories")
	}

	return len(resp.Repositories) == 1
}

func (ecr *ECR) GetRepositoryUri(repositoryName string) string {
	resp, err := ecr.svc.DescribeRepositories(
		&awsecr.DescribeRepositoriesInput{
			RepositoryNames: aws.StringSlice([]string{repositoryName}),
		},
	)

	if err != nil {
		console.ErrorExit(err, "Couldn't describe Amazon ECR repositories")
	}

	if len(resp.Repositories) != 1 {
		console.ErrorExit(err, "Couldn't find Amazon ECR repository: %s", repositoryName)
	}

	return aws.StringValue(resp.Repositories[0].RepositoryUri)
}

func (ecr *ECR) GetUsernameAndPassword() (username, password string) {
	resp, err := ecr.svc.GetAuthorizationToken(
		&awsecr.GetAuthorizationTokenInput{},
	)

	if err != nil {
		console.ErrorExit(err, "Couldn't get Amazon ECR authorization token")
	}

	token, _ := base64.StdEncoding.DecodeString(*resp.AuthorizationData[0].AuthorizationToken)
	s := strings.Split(string(token), ":")
	username = s[0]
	password = s[1]

	return
}
