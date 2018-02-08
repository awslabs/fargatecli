package elbv2

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	awselbv2 "github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/jpignata/fargate/console"
)

type LoadBalancer struct {
	DNSName          string
	Name             string
	State            string
	StateReason      string
	ARN              string
	Type             string
	HostedZoneId     string
	SecurityGroupIDs []string
	SubnetIDs        []string
	VPCID            string
}

type LoadBalancers []LoadBalancer

type DescribeLoadBalancersInput struct {
	Names []string
	ARNs  []string
}

type CreateLoadBalancerInput struct {
	Name             string
	SubnetIDs        []string
	SecurityGroupIDs []string
	Type             string
}

func (elbv2 SDKClient) DescribeLoadBalancersByName(lbNames []string) (LoadBalancers, error) {
	var loadBalancers []LoadBalancer

	input := &awselbv2.DescribeLoadBalancersInput{Names: aws.StringSlice(lbNames)}
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
					HostedZoneId:     aws.StringValue(loadBalancer.CanonicalHostedZoneId),
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

	err := elbv2.client.DescribeLoadBalancersPages(input, handler)

	return loadBalancers, err
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
	loadBalancers := elbv2.DescribeLoadBalancers(
		DescribeLoadBalancersInput{
			Names: []string{lbName},
		},
	)

	if len(loadBalancers) == 0 {
		console.ErrorExit(fmt.Errorf("%s not found", lbName), "Could not find ELB load balancer")
	}

	return loadBalancers[0]
}

func (elbv2 SDKClient) DescribeLoadBalancerByArn(lbARN string) LoadBalancer {
	loadBalancers := elbv2.DescribeLoadBalancers(
		DescribeLoadBalancersInput{
			ARNs: []string{lbARN},
		},
	)

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

func (elbv2 SDKClient) DescribeLoadBalancers(i DescribeLoadBalancersInput) []LoadBalancer {
	var loadBalancers []LoadBalancer

	input := &awselbv2.DescribeLoadBalancersInput{}

	if len(i.Names) > 0 {
		input.SetNames(aws.StringSlice(i.Names))
	}

	if len(i.ARNs) > 0 {
		input.SetLoadBalancerArns(aws.StringSlice(i.ARNs))
	}

	err := elbv2.client.DescribeLoadBalancersPages(
		input,
		func(resp *awselbv2.DescribeLoadBalancersOutput, lastPage bool) bool {
			for _, loadBalancer := range resp.LoadBalancers {
				var subnetIds []string

				for _, availabilityZone := range loadBalancer.AvailabilityZones {
					subnetIds = append(subnetIds, aws.StringValue(availabilityZone.SubnetId))
				}

				loadBalancers = append(loadBalancers,
					LoadBalancer{
						ARN:              aws.StringValue(loadBalancer.LoadBalancerArn),
						DNSName:          aws.StringValue(loadBalancer.DNSName),
						HostedZoneId:     aws.StringValue(loadBalancer.CanonicalHostedZoneId),
						VPCID:            aws.StringValue(loadBalancer.VpcId),
						Name:             aws.StringValue(loadBalancer.LoadBalancerName),
						SecurityGroupIDs: aws.StringValueSlice(loadBalancer.SecurityGroups),
						State:            aws.StringValue(loadBalancer.State.Code),
						StateReason:      aws.StringValue(loadBalancer.State.Reason),
						SubnetIDs:        subnetIds,
						Type:             aws.StringValue(loadBalancer.Type),
					},
				)
			}

			return true
		},
	)

	if err != nil {
		console.ErrorExit(err, "Could not describe ELB load balancers")
	}

	return loadBalancers
}
