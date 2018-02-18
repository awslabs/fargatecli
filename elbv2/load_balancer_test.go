package elbv2

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	awselbv2 "github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/jpignata/fargate/elbv2/mock/sdk"
)

func TestDescribeLoadBalancersByName(t *testing.T) {
	resp := &awselbv2.DescribeLoadBalancersOutput{
		LoadBalancers: []*awselbv2.LoadBalancer{
			&awselbv2.LoadBalancer{
				LoadBalancerArn:       aws.String("arn:aws:elasticloadbalancing:us-west-2:123456789012:loadbalancer/app/my-load-balancer/50dc6c495c0c9188"),
				DNSName:               aws.String("my-load-balancer-424835706.us-west-2.elb.amazonaws.com"),
				CanonicalHostedZoneId: aws.String("Z2P70J7EXAMPLE"),
				VpcId:            aws.String("vpc-3ac0fb5f"),
				LoadBalancerName: aws.String("web"),
				SecurityGroups:   []*string{aws.String("sg-5943793c")},
				AvailabilityZones: []*awselbv2.AvailabilityZone{
					&awselbv2.AvailabilityZone{
						SubnetId: aws.String("subnet-8360a9e7"),
					},
				},
				Type: aws.String("application"),
				State: &awselbv2.LoadBalancerState{
					Code:   aws.String("active"),
					Reason: aws.String(""),
				},
			},
		},
	}

	mockClient := sdk.MockDescribeLoadBalancersClient{Resp: resp}
	elbv2 := SDKClient{client: mockClient}
	loadBalancers, err := elbv2.DescribeLoadBalancersByName([]string{"web"})

	if err != nil {
		t.Errorf("Expected no error, got %+v", err)
	}

	if len(loadBalancers) != 1 {
		t.Errorf("Expected 1 load balancer, got %d", len(loadBalancers))
	}
}

func TestDescribeLoadBalancersByNameError(t *testing.T) {
	mockClient := sdk.MockDescribeLoadBalancersClient{Error: errors.New("boom")}
	elbv2 := SDKClient{client: mockClient}
	_, err := elbv2.DescribeLoadBalancersByName([]string{"web"})

	if err == nil {
		t.Error("Expected error, got none")
	}
}
