package ec2

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	awsec2 "github.com/aws/aws-sdk-go/service/ec2"
)

const (
	defaultSecurityGroupName            = "fargate-default"
	defaultSecurityGroupDescription     = "Default Fargate CLI SG"
	defaultSecurityGroupIngressCIDR     = "0.0.0.0/0"
	defaultSecurityGroupIngressProtocol = "-1"
)

// GetDefaultSubnetIDs finds and returns the subnet IDs marked as default.
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

// GetDefaultSecurityGroupID returns the ID of the permissive security group created by default.
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

// GetSubnetVPCID returns the VPC ID for a given subnet ID.
func (ec2 SDKClient) GetSubnetVPCID(subnetID string) (string, error) {
	resp, err := ec2.client.DescribeSubnets(
		&awsec2.DescribeSubnetsInput{
			SubnetIds: aws.StringSlice([]string{subnetID}),
		},
	)

	switch {
	case err != nil:
		return "", fmt.Errorf("could not find VPC ID for subnet ID %s: %v", subnetID, err)
	case len(resp.Subnets) == 0:
		return "", fmt.Errorf("could not find VPC ID: subnet ID %s not found", subnetID)
	default:
		return aws.StringValue(resp.Subnets[0].VpcId), nil
	}
}

// CreateDefaultSecurityGroup creates a new security group for use as the default.
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

// AuthorizeAllSecurityGroupIngress configures a security group to allow all ingress traffic.
func (ec2 SDKClient) AuthorizeAllSecurityGroupIngress(groupID string) error {
	_, err := ec2.client.AuthorizeSecurityGroupIngress(
		&awsec2.AuthorizeSecurityGroupIngressInput{
			CidrIp:     aws.String(defaultSecurityGroupIngressCIDR),
			GroupId:    aws.String(groupID),
			IpProtocol: aws.String(defaultSecurityGroupIngressProtocol),
		},
	)

	return err
}
