package ec2

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	awsec2 "github.com/aws/aws-sdk-go/service/ec2"
)

const (
	defaultSecurityGroupName        = "fargate-default"
	defaultSecurityGroupDescription = "Default Fargate CLI SG"
)

func (ec2 SDKClient) GetDefaultSubnetIDs() ([]string, error) {
	var subnetIDs []string

	defaultFilter := &awsec2.Filter{
		Name:   aws.String("default-for-az"),
		Values: aws.StringSlice([]string{"true"}),
	}

	resp, err := ec2.client.DescribeSubnets(
		&awsec2.DescribeSubnetsInput{
			Filters: []*awsec2.Filter{defaultFilter},
		},
	)

	if err != nil {
		return subnetIDs, fmt.Errorf("could not retrieve default subnet IDs: %v", err)
	}

	for _, subnet := range resp.Subnets {
		subnetIDs = append(subnetIDs, aws.StringValue(subnet.SubnetId))
	}

	return subnetIDs, nil
}

func (ec2 SDKClient) GetDefaultSecurityGroupID() (string, error) {
	resp, err := ec2.client.DescribeSecurityGroups(
		&awsec2.DescribeSecurityGroupsInput{
			GroupNames: aws.StringSlice([]string{defaultSecurityGroupName}),
		},
	)

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			if aerr.Code() == "InvalidGroup.NotFound" {
				return "", nil
			}
		}

		return "", fmt.Errorf("could not retrieve default security group ID (%s): %v", defaultSecurityGroupName, err)
	}

	return aws.StringValue(resp.SecurityGroups[0].GroupId), nil
}

func (ec2 SDKClient) GetSubnetVPCID(subnetID string) (string, error) {
	resp, err := ec2.client.DescribeSubnets(
		&awsec2.DescribeSubnetsInput{
			SubnetIds: aws.StringSlice([]string{subnetID}),
		},
	)

	switch {
	case len(resp.Subnets) == 0:
		return "", fmt.Errorf("could not find VPC ID: subnet ID %s not found", subnetID)
	case err != nil:
		return "", fmt.Errorf("could not find VPC ID for subnet %s: %v", subnetID, err)
	default:
		return aws.StringValue(resp.Subnets[0].VpcId), nil
	}
}

func (ec2 SDKClient) CreateDefaultSecurityGroup() (string, error) {
	resp, err := ec2.client.CreateSecurityGroup(
		&awsec2.CreateSecurityGroupInput{
			GroupName:   aws.String(defaultSecurityGroupName),
			Description: aws.String(defaultSecurityGroupDescription),
		},
	)

	if err != nil {
		return "", fmt.Errorf("could not create default security group (%s): %v", defaultSecurityGroupName, err)
	}

	return aws.StringValue(resp.GroupId), nil
}

func (ec2 SDKClient) AuthorizeAllSecurityGroupIngress(groupId string) error {
	_, err := ec2.client.AuthorizeSecurityGroupIngress(
		&awsec2.AuthorizeSecurityGroupIngressInput{
			CidrIp:     aws.String("0.0.0.0/0"),
			GroupId:    aws.String(groupId),
			IpProtocol: aws.String("-1"),
		},
	)

	return err
}
