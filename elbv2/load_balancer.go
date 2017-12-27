package elbv2

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	awselbv2 "github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/jpignata/fargate/console"
)

type LoadBalancer struct {
	DNSName      string
	Name         string
	State        string
	StateReason  string
	Arn          string
	Type         string
	HostedZoneId string
}

type DescribeLoadBalancersInput struct {
	Names []string
	Arns  []string
}

type CreateLoadBalancerInput struct {
	Name      string
	SubnetIds []string
	Type      string
}

func (elbv2 *ELBV2) CreateLoadBalancer(input *CreateLoadBalancerInput) string {
	console.Debug("Creating ELB load balancer")

	resp, err := elbv2.svc.CreateLoadBalancer(
		&awselbv2.CreateLoadBalancerInput{
			Name:    aws.String(input.Name),
			Subnets: aws.StringSlice(input.SubnetIds),
			Type:    aws.String(input.Type),
		},
	)

	if err != nil || len(resp.LoadBalancers) != 1 {
		console.ErrorExit(err, "Could not create ELB load balancer")
	}

	return aws.StringValue(resp.LoadBalancers[0].LoadBalancerArn)
}

func (elbv2 *ELBV2) DescribeLoadBalancer(lbName string) LoadBalancer {
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

func (elbv2 *ELBV2) DescribeLoadBalancerByArn(lbArn string) LoadBalancer {
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

func (elbv2 *ELBV2) DeleteLoadBalancer(lbName string) {
	loadBalancer := elbv2.DescribeLoadBalancer(lbName)
	_, err := elbv2.svc.DeleteLoadBalancer(
		&awselbv2.DeleteLoadBalancerInput{
			LoadBalancerArn: aws.String(loadBalancer.Arn),
		},
	)

	if err != nil {
		console.ErrorExit(err, "Could not destroy ELB load balancer")
	}
}

func (elbv2 *ELBV2) DescribeLoadBalancers(i DescribeLoadBalancersInput) []LoadBalancer {
	var loadBalancers []LoadBalancer

	input := &awselbv2.DescribeLoadBalancersInput{}

	if len(i.Names) > 0 {
		input.SetNames(aws.StringSlice(i.Names))
	}

	if len(i.Arns) > 0 {
		input.SetLoadBalancerArns(aws.StringSlice(i.Arns))
	}

	err := elbv2.svc.DescribeLoadBalancersPages(
		input,
		func(resp *awselbv2.DescribeLoadBalancersOutput, lastPage bool) bool {
			for _, loadBalancer := range resp.LoadBalancers {
				loadBalancers = append(loadBalancers,
					LoadBalancer{
						Name:         aws.StringValue(loadBalancer.LoadBalancerName),
						DNSName:      aws.StringValue(loadBalancer.DNSName),
						State:        aws.StringValue(loadBalancer.State.Code),
						StateReason:  aws.StringValue(loadBalancer.State.Reason),
						Arn:          aws.StringValue(loadBalancer.LoadBalancerArn),
						Type:         aws.StringValue(loadBalancer.Type),
						HostedZoneId: aws.StringValue(loadBalancer.CanonicalHostedZoneId),
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
