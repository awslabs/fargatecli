package ec2

import (
	"github.com/aws/aws-sdk-go/aws"
	awsec2 "github.com/aws/aws-sdk-go/service/ec2"
	"github.com/jpignata/fargate/console"
)

func (ec2 *EC2) GetDefaultVpcSubnetIds() []string {
	var subnetIds []string

	vpcFilter := &awsec2.Filter{
		Name:   aws.String("vpc-id"),
		Values: aws.StringSlice([]string{ec2.GetDefaultVpcId()}),
	}

	defaultFilter := &awsec2.Filter{
		Name:   aws.String("default-for-az"),
		Values: aws.StringSlice([]string{"true"}),
	}

	resp, err := ec2.svc.DescribeSubnets(
		&awsec2.DescribeSubnetsInput{
			Filters: []*awsec2.Filter{vpcFilter, defaultFilter},
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
