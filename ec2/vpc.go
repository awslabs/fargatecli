package ec2

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	awsec2 "github.com/aws/aws-sdk-go/service/ec2"
	"github.com/fatih/color"
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
		fmt.Println("Could not get default vpc subnets")
		os.Exit(1)
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
		color.Red("Could not find a default Vpc")
		os.Exit(1)
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
		color.Red("Could not EC2.DescribeVpcs")
		fmt.Println(err)
		os.Exit(1)
	}

	return resp.Vpcs
}
