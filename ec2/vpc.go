package ec2

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	awsec2 "github.com/aws/aws-sdk-go/service/ec2"
	"github.com/jpignata/fargate/console"
)

const (
	defaultSecurityGroupName        = "fargate-default"
	defaultSecurityGroupDescription = "Default Fargate CLI SG"
)

func (ec2 *EC2) GetDefaultVpcSubnetIds() []string {
	var subnetIds []string

	defaultFilter := &awsec2.Filter{
		Name:   aws.String("default-for-az"),
		Values: aws.StringSlice([]string{"true"}),
	}

	resp, err := ec2.svc.DescribeSubnets(
		&awsec2.DescribeSubnetsInput{
			Filters: []*awsec2.Filter{defaultFilter},
		},
	)

	if err != nil {
		console.IssueExit("Could not find default VPC subnets")
	}

	for _, subnet := range resp.Subnets {
		subnetIds = append(subnetIds, *subnet.SubnetId)
	}

	return subnetIds
}

func (ec2 *EC2) GetDefaultVpcId() string {
	return *ec2.getDefaultVpc().VpcId
}

func (ec2 *EC2) GetDefaultSecurityGroupId() string {
	resp, err := ec2.svc.DescribeSecurityGroups(
		&awsec2.DescribeSecurityGroupsInput{
			GroupNames: aws.StringSlice([]string{defaultSecurityGroupName}),
		},
	)

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case "InvalidGroup.NotFound":
				return ec2.createDefaultSecurityGroup()
			default:
				console.ErrorExit(err, "Could not find EC2 security group")
			}
		}
	}

	return aws.StringValue(resp.SecurityGroups[0].GroupId)
}

func (ec2 *EC2) GetSubnetVpcId(subnetId string) string {
	subnets := ec2.describeSubnets([]string{subnetId})

	if len(subnets) != 1 {
		console.IssueExit("Subnet ID %s not found", subnetId)
	}

	return aws.StringValue(subnets[0].VpcId)
}

func (ec2 *EC2) describeSubnets(subnetIds []string) []*awsec2.Subnet {
	resp, err := ec2.svc.DescribeSubnets(
		&awsec2.DescribeSubnetsInput{
			SubnetIds: aws.StringSlice(subnetIds),
		},
	)

	if err != nil {
		console.ErrorExit(err, "Could not describe EC2 security groups")
	}

	return resp.Subnets
}

func (ec2 *EC2) getDefaultVpc() *awsec2.Vpc {
	filter := &awsec2.Filter{
		Name:   aws.String("isDefault"),
		Values: aws.StringSlice([]string{"true"}),
	}
	vpcs := ec2.describeVpcs([]*awsec2.Filter{filter})

	if len(vpcs) != 1 {
		console.IssueExit("Could not find a default VPC")
	}

	return vpcs[0]
}

func (ec2 *EC2) describeVpcs(filters []*awsec2.Filter) []*awsec2.Vpc {
	resp, err := ec2.svc.DescribeVpcs(
		&awsec2.DescribeVpcsInput{
			Filters: filters,
		},
	)

	if err != nil {
		console.ErrorExit(err, "Could not describe VPCs")
	}

	return resp.Vpcs
}

func (ec2 *EC2) createDefaultSecurityGroup() string {
	resp, err := ec2.svc.CreateSecurityGroup(
		&awsec2.CreateSecurityGroupInput{
			GroupName:   aws.String(defaultSecurityGroupName),
			Description: aws.String(defaultSecurityGroupDescription),
		},
	)

	if err != nil {
		console.ErrorExit(err, "Could not create default EC2 security group")
	}

	groupId := resp.GroupId

	ec2.authorizeAllSecurityGroupIngress(groupId)

	return aws.StringValue(groupId)
}

func (ec2 *EC2) authorizeAllSecurityGroupIngress(groupId *string) {
	_, err := ec2.svc.AuthorizeSecurityGroupIngress(
		&awsec2.AuthorizeSecurityGroupIngressInput{
			CidrIp:     aws.String("0.0.0.0/0"),
			GroupId:    groupId,
			IpProtocol: aws.String("-1"),
		},
	)

	if err != nil {
		console.ErrorExit(err, "Could not create default EC2 security group")
	}
}
