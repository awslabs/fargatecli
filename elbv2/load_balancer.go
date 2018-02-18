package elbv2

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	awselbv2 "github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/jpignata/fargate/console"
)

type LoadBalancer struct {
	ARN              string
	DNSName          string
	HostedZoneID     string
	Listeners        Listeners
	Name             string
	SecurityGroupIDs []string
	State            string
	StateReason      string
	SubnetIDs        []string
	Type             string
	VPCID            string
}

type LoadBalancers []LoadBalancer

type CreateLoadBalancerInput struct {
	Name             string
	SecurityGroupIDs []string
	SubnetIDs        []string
	Type             string
}

func (elbv2 SDKClient) DescribeLoadBalancers() (LoadBalancers, error) {
	return elbv2.describeLoadBalancers(&awselbv2.DescribeLoadBalancersInput{})
}

func (elbv2 SDKClient) DescribeLoadBalancersByName(lbNames []string) (LoadBalancers, error) {
	return elbv2.describeLoadBalancers(
		&awselbv2.DescribeLoadBalancersInput{Names: aws.StringSlice(lbNames)},
	)
}

func (elbv2 SDKClient) DescribeLoadBalancersByARN(lbARNs []string) (LoadBalancers, error) {
	return elbv2.describeLoadBalancers(
		&awselbv2.DescribeLoadBalancersInput{LoadBalancerArns: aws.StringSlice(lbARNs)},
	)
}

func (elbv2 SDKClient) CreateLoadBalancer(i CreateLoadBalancerInput) (string, error) {
	sdki := &awselbv2.CreateLoadBalancerInput{
		Name:    aws.String(i.Name),
		Subnets: aws.StringSlice(i.SubnetIDs),
		Type:    aws.String(i.Type),
	}

	if i.Type == awselbv2.LoadBalancerTypeEnumApplication {
		sdki.SetSecurityGroups(aws.StringSlice(i.SecurityGroupIDs))
	}

	resp, err := elbv2.client.CreateLoadBalancer(sdki)

	if err != nil {
		return "", err
	}

	return aws.StringValue(resp.LoadBalancers[0].LoadBalancerArn), nil
}

func (elbv2 SDKClient) DescribeLoadBalancer(lbName string) LoadBalancer {
	loadBalancers, _ := elbv2.DescribeLoadBalancersByName([]string{lbName})

	if len(loadBalancers) == 0 {
		console.ErrorExit(fmt.Errorf("%s not found", lbName), "Could not find ELB load balancer")
	}

	return loadBalancers[0]
}

func (elbv2 SDKClient) DescribeLoadBalancerByARN(lbARN string) LoadBalancer {
	loadBalancers, _ := elbv2.DescribeLoadBalancersByARN([]string{lbARN})

	if len(loadBalancers) == 0 {
		console.ErrorExit(fmt.Errorf("%s not found", lbARN), "Could not find ELB load balancer")
	}

	return loadBalancers[0]
}

func (elbv2 SDKClient) DeleteLoadBalancer(lbName string) {
	loadBalancer := elbv2.DescribeLoadBalancer(lbName)
	_, err := elbv2.client.DeleteLoadBalancer(
		&awselbv2.DeleteLoadBalancerInput{
			LoadBalancerArn: aws.String(loadBalancer.ARN),
		},
	)

	if err != nil {
		console.ErrorExit(err, "Could not destroy ELB load balancer")
	}
}

func (elbv2 SDKClient) describeLoadBalancers(i *awselbv2.DescribeLoadBalancersInput) (LoadBalancers, error) {
	var loadBalancers []LoadBalancer

	handler := func(resp *awselbv2.DescribeLoadBalancersOutput, lastPage bool) bool {
		for _, loadBalancer := range resp.LoadBalancers {
			var subnetIDs []string

			for _, availabilityZone := range loadBalancer.AvailabilityZones {
				subnetIDs = append(subnetIDs, aws.StringValue(availabilityZone.SubnetId))
			}

			loadBalancers = append(loadBalancers,
				LoadBalancer{
					ARN:              aws.StringValue(loadBalancer.LoadBalancerArn),
					DNSName:          aws.StringValue(loadBalancer.DNSName),
					HostedZoneID:     aws.StringValue(loadBalancer.CanonicalHostedZoneId),
					VPCID:            aws.StringValue(loadBalancer.VpcId),
					Name:             aws.StringValue(loadBalancer.LoadBalancerName),
					SecurityGroupIDs: aws.StringValueSlice(loadBalancer.SecurityGroups),
					State:            aws.StringValue(loadBalancer.State.Code),
					StateReason:      aws.StringValue(loadBalancer.State.Reason),
					SubnetIDs:        subnetIDs,
					Type:             aws.StringValue(loadBalancer.Type),
				},
			)
		}

		return true
	}

	err := elbv2.client.DescribeLoadBalancersPages(i, handler)

	return loadBalancers, err
}
