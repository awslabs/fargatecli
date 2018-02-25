package elbv2

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	awselbv2 "github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/golang/mock/gomock"
	"github.com/jpignata/fargate/elbv2/mock/sdk"
)

var (
	lbARN        = "arn:aws:elasticloadbalancing:us-west-2:123456789012:loadbalancer/app/my-load-balancer/50dc6c495c0c9188"
	dnsName      = "my-load-balancer-424835706.us-west-2.elb.amazonaws.com"
	hostedZoneID = "Z2P70J7EXAMPLE"
	vpcID        = "vpc-3ac0fb5f"
	lbName       = "web"
	subnet       = "subnet-8360a9e7"
	lbType       = "application"
	status       = "active"

	resp = &awselbv2.DescribeLoadBalancersOutput{
		LoadBalancers: []*awselbv2.LoadBalancer{
			&awselbv2.LoadBalancer{
				LoadBalancerArn:       aws.String(lbARN),
				DNSName:               aws.String(dnsName),
				CanonicalHostedZoneId: aws.String(hostedZoneID),
				VpcId:            aws.String(vpcID),
				LoadBalancerName: aws.String(lbName),
				SecurityGroups:   []*string{aws.String("sg-5943793c")},
				AvailabilityZones: []*awselbv2.AvailabilityZone{
					&awselbv2.AvailabilityZone{
						SubnetId: aws.String(subnet),
					},
				},
				Type: aws.String(lbType),
				State: &awselbv2.LoadBalancerState{
					Code: aws.String(status),
				},
			},
		},
	}
)

func TestDescribeLoadBalancers(t *testing.T) {
	mockClient := sdk.MockDescribeLoadBalancersClient{Resp: resp}
	elbv2 := SDKClient{client: mockClient}
	loadBalancers, err := elbv2.DescribeLoadBalancers()

	if err != nil {
		t.Errorf("expected no error, got %+v", err)
	}

	if len(loadBalancers) != 1 {
		t.Errorf("expected 1 load balancer, got %d", len(loadBalancers))
	}

	if loadBalancers[0].ARN != lbARN {
		t.Errorf("expected ARN %s, got %s", lbARN, loadBalancers[0].ARN)
	}

	if loadBalancers[0].DNSName != dnsName {
		t.Errorf("expected DNSName %s, got %s", dnsName, loadBalancers[0].DNSName)
	}

	if loadBalancers[0].HostedZoneID != hostedZoneID {
		t.Errorf("expected HostedZoneID %s, got %s", hostedZoneID, loadBalancers[0].HostedZoneID)
	}

	if loadBalancers[0].VPCID != vpcID {
		t.Errorf("expected VPCID %s, got %s", vpcID, loadBalancers[0].VPCID)
	}

	if loadBalancers[0].Name != lbName {
		t.Errorf("expected Name %s, got %s", lbName, loadBalancers[0].Name)
	}

	if loadBalancers[0].SubnetIDs[0] != subnet {
		t.Errorf("expected subnet %s, got %s", subnet, loadBalancers[0].SubnetIDs[0])
	}

	if loadBalancers[0].Type != lbType {
		t.Errorf("expected type  %s, got %s", lbType, loadBalancers[0].Type)
	}

	if loadBalancers[0].Status != status {
		t.Errorf("expected status %s, got %s", status, loadBalancers[0].Status)
	}
}

func TestDescribeLoadBalancersByName(t *testing.T) {
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

func TestDescribeLoadBalancersByARN(t *testing.T) {
	mockClient := sdk.MockDescribeLoadBalancersClient{Resp: resp}
	elbv2 := SDKClient{client: mockClient}
	loadBalancers, err := elbv2.DescribeLoadBalancersByARN([]string{lbARN})

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

func TestCreateLoadBalancer(t *testing.T) {
	lbARN := "arn:aws:elasticloadbalancing:us-west-2:123456789012:loadbalancer/app/my-load-balancer/50dc6c495c0c9188"
	name := "cool-load-balancer"
	subnetIDs := []string{"subnet-1234567"}
	securityGroupIDs := []string{"sg-1234567"}
	lbType := "application"

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockELBV2API := sdk.NewMockELBV2API(mockCtrl)
	elbv2 := SDKClient{client: mockELBV2API}
	i := &awselbv2.CreateLoadBalancerInput{
		Name:           aws.String(name),
		Subnets:        aws.StringSlice(subnetIDs),
		SecurityGroups: aws.StringSlice(securityGroupIDs),
		Type:           aws.String(lbType),
	}
	o := &awselbv2.CreateLoadBalancerOutput{
		LoadBalancers: []*awselbv2.LoadBalancer{
			&awselbv2.LoadBalancer{
				LoadBalancerArn: aws.String(lbARN),
			},
		},
	}
	params := CreateLoadBalancerParameters{
		Name:             name,
		SubnetIDs:        subnetIDs,
		Type:             lbType,
		SecurityGroupIDs: securityGroupIDs,
	}

	mockELBV2API.EXPECT().CreateLoadBalancer(i).Return(o, nil)

	arn, err := elbv2.CreateLoadBalancer(params)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if arn != lbARN {
		t.Errorf("expected ARN %s, got %s", lbARN, arn)
	}
}

func TestCreateLoadBalancerWithError(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockELBV2API := sdk.NewMockELBV2API(mockCtrl)
	elbv2 := SDKClient{client: mockELBV2API}

	mockELBV2API.EXPECT().CreateLoadBalancer(gomock.Any()).Return(nil, errors.New("boom"))

	_, err := elbv2.CreateLoadBalancer(CreateLoadBalancerParameters{})

	if err == nil {
		t.Fatalf("expected error, got none")
	}
}
