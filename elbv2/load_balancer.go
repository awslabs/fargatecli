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
	Arn              string
	Type             string
	HostedZoneId     string
	SecurityGroupIds []string
	SubnetIds        []string
	VpcId            string
}

type LoadBalancers []LoadBalancer

type DescribeLoadBalancersInput struct {
	Names []string
	Arns  []string
}

type CreateLoadBalancerInput struct {
	Name             string
	SubnetIds        []string
	Type             string
	SecurityGroupIds []string
}

func (elbv2 SDKClient) DescribeLoadBalancersByName(lbNames []string) (LoadBalancers, error) {
	var loadBalancers []LoadBalancer

	input := &awselbv2.DescribeLoadBalancersInput{Names: aws.StringSlice(lbNames)}
	handler := func(resp *awselbv2.DescribeLoadBalancersOutput, lastPage bool) bool {
		for _, loadBalancer := range resp.LoadBalancers {
			var subnetIds []string

			for _, availabilityZone := range loadBalancer.AvailabilityZones {
				subnetIds = append(subnetIds, aws.StringValue(availabilityZone.SubnetId))
			}

			loadBalancers = append(loadBalancers,
				LoadBalancer{
					Arn:              aws.StringValue(loadBalancer.LoadBalancerArn),
					DNSName:          aws.StringValue(loadBalancer.DNSName),
					HostedZoneId:     aws.StringValue(loadBalancer.CanonicalHostedZoneId),
					VpcId:            aws.StringValue(loadBalancer.VpcId),
					Name:             aws.StringValue(loadBalancer.LoadBalancerName),
					SecurityGroupIds: aws.StringValueSlice(loadBalancer.SecurityGroups),
					State:            aws.StringValue(loadBalancer.State.Code),
					StateReason:      aws.StringValue(loadBalancer.State.Reason),
					SubnetIds:        subnetIds,
					Type:             aws.StringValue(loadBalancer.Type),
				},
			)
		}

		return true
	}

	err := elbv2.client.DescribeLoadBalancersPages(input, handler)

	return loadBalancers, err
}

func (elbv2 SDKClient) CreateLoadBalancer(i *CreateLoadBalancerInput) string {
	console.Debug("Creating ELB load balancer")
	input := &awselbv2.CreateLoadBalancerInput{
		Name:    aws.String(i.Name),
		Subnets: aws.StringSlice(i.SubnetIds),
		Type:    aws.String(i.Type),
	}

	if i.Type == awselbv2.LoadBalancerTypeEnumApplication {
		input.SetSecurityGroups(aws.StringSlice(i.SecurityGroupIds))
	}

	resp, err := elbv2.client.CreateLoadBalancer(input)

	if err != nil || len(resp.LoadBalancers) != 1 {
		console.ErrorExit(err, "Could not create ELB load balancer")
	}

	return aws.StringValue(resp.LoadBalancers[0].LoadBalancerArn)
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

func (elbv2 SDKClient) DescribeLoadBalancerByArn(lbArn string) LoadBalancer {
	loadBalancers := elbv2.DescribeLoadBalancers(
		DescribeLoadBalancersInput{
			Arns: []string{lbArn},
		},
	)

	if len(loadBalancers) == 0 {
		console.ErrorExit(fmt.Errorf("%s not found", lbArn), "Could not find ELB load balancer")
	}

	return loadBalancers[0]
}

func (elbv2 SDKClient) DeleteLoadBalancer(lbName string) {
	loadBalancer := elbv2.DescribeLoadBalancer(lbName)
	_, err := elbv2.client.DeleteLoadBalancer(
		&awselbv2.DeleteLoadBalancerInput{
			LoadBalancerArn: aws.String(loadBalancer.Arn),
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

	if len(i.Arns) > 0 {
		input.SetLoadBalancerArns(aws.StringSlice(i.Arns))
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
						Arn:              aws.StringValue(loadBalancer.LoadBalancerArn),
						DNSName:          aws.StringValue(loadBalancer.DNSName),
						HostedZoneId:     aws.StringValue(loadBalancer.CanonicalHostedZoneId),
						VpcId:            aws.StringValue(loadBalancer.VpcId),
						Name:             aws.StringValue(loadBalancer.LoadBalancerName),
						SecurityGroupIds: aws.StringValueSlice(loadBalancer.SecurityGroups),
						State:            aws.StringValue(loadBalancer.State.Code),
						StateReason:      aws.StringValue(loadBalancer.State.Reason),
						SubnetIds:        subnetIds,
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
