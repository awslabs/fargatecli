package cmd

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/jpignata/fargate/cmd/mock"
	"github.com/jpignata/fargate/elbv2"
	elbv2client "github.com/jpignata/fargate/elbv2/mock/client"
	"github.com/jpignata/fargate/route53"
	route53client "github.com/jpignata/fargate/route53/mock/client"
)

var (
	lb = elbv2.LoadBalancer{
		ARN:              "arn:aws:elasticloadbalancing:us-west-2:123456789012:loadbalancer/app/my-load-balancer/50dc6c495c0c9188",
		DNSName:          "my-load-balancer-424835706.us-west-2.elb.amazonaws.com",
		HostedZoneId:     "Z2P70J7EXAMPLE",
		VPCID:            "vpc-3ac0fb5f",
		Name:             "web",
		SecurityGroupIDs: []string{"sg-5943793c"},
		State:            "active",
		StateReason:      "",
		SubnetIDs:        []string{"subnet-8360a9e7"},
		Type:             "application",
	}
	hostedZone = route53.HostedZone{
		Name: "example.com.",
		ID:   "Z111111QQQQQQQ",
	}
)

func TestLbAliasOperation(t *testing.T) {
	domainName := "example.com."
	lbName := "web"
	dnsName := "my-load-balancer-424835706.us-west-2.elb.amazonaws.com"
	hostedZoneID := "Z2P70J7EXAMPLE"

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockELBV2Client := elbv2client.NewMockClient(mockCtrl)
	mockRoute53Client := route53client.NewMockClient(mockCtrl)
	mockOutput := &mock.Output{}

	createAliasInput := route53.CreateAliasInput{
		HostedZoneID:       hostedZone.ID,
		RecordType:         "A",
		Name:               domainName,
		Target:             dnsName,
		TargetHostedZoneID: hostedZoneID,
	}

	operation := lbAliasOperation{
		lbOperation: lbOperation{
			elbv2: mockELBV2Client,
		},
		aliasDomain: domainName,
		lbName:      lbName,
		output:      mockOutput,
		route53:     mockRoute53Client,
	}

	mockELBV2Client.EXPECT().DescribeLoadBalancersByName([]string{"web"}).Return(elbv2.LoadBalancers{lb}, nil)
	mockRoute53Client.EXPECT().ListHostedZones().Return(route53.HostedZones{hostedZone}, nil)
	mockRoute53Client.EXPECT().CreateAlias(createAliasInput).Return("ID", nil)

	operation.execute()

	if len(mockOutput.InfoMsgs) == 0 {
		t.Errorf("Expected info output from operation, got none")
	}
}

func TestLbAliasOperationFindLbError(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockELBV2Client := elbv2client.NewMockClient(mockCtrl)
	mockRoute53Client := route53client.NewMockClient(mockCtrl)
	mockOutput := &mock.Output{}

	operation := lbAliasOperation{
		lbOperation: lbOperation{
			elbv2: mockELBV2Client,
		},
		aliasDomain: "example.com",
		lbName:      "web",
		output:      mockOutput,
		route53:     mockRoute53Client,
	}

	mockELBV2Client.EXPECT().DescribeLoadBalancersByName([]string{"web"}).Return(elbv2.LoadBalancers{}, errors.New("boom"))

	operation.execute()

	if len(mockOutput.FatalMsgs) == 0 {
		t.Error("Expected fatal output, got none")
	} else if mockOutput.FatalMsgs[0].Msg != "Could not alias load balancer" {
		t.Errorf("Expected fatal output == 'Could not alias load balancer', got: %s", mockOutput.FatalMsgs[0])
	}
}

func TestLbAliasOperationListHostedZonesError(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockELBV2Client := elbv2client.NewMockClient(mockCtrl)
	mockRoute53Client := route53client.NewMockClient(mockCtrl)
	mockOutput := &mock.Output{}

	operation := lbAliasOperation{
		lbOperation: lbOperation{
			elbv2: mockELBV2Client,
		},
		aliasDomain: "example.com",
		lbName:      "web",
		output:      mockOutput,
		route53:     mockRoute53Client,
	}

	mockELBV2Client.EXPECT().DescribeLoadBalancersByName([]string{"web"}).Return(elbv2.LoadBalancers{elbv2.LoadBalancer{}}, nil)
	mockRoute53Client.EXPECT().ListHostedZones().Return(route53.HostedZones{}, errors.New("boom"))

	operation.execute()

	if len(mockOutput.FatalMsgs) == 0 {
		t.Error("Expected fatal output, got none")
	} else if mockOutput.FatalMsgs[0].Msg != "Could not alias load balancer" {
		t.Errorf("Expected fatal output == 'Could not alias load balancer', got: %s", mockOutput.FatalMsgs[0])
	}
}

func TestLbAliasOperationAliasError(t *testing.T) {
	domainName := "example.com."
	lbName := "web"
	dnsName := "my-load-balancer-424835706.us-west-2.elb.amazonaws.com"
	hostedZoneID := "Z2P70J7EXAMPLE"

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockELBV2Client := elbv2client.NewMockClient(mockCtrl)
	mockRoute53Client := route53client.NewMockClient(mockCtrl)
	mockOutput := &mock.Output{}

	createAliasInput := route53.CreateAliasInput{
		HostedZoneID:       hostedZone.ID,
		RecordType:         "A",
		Name:               domainName,
		Target:             dnsName,
		TargetHostedZoneID: hostedZoneID,
	}

	operation := lbAliasOperation{
		lbOperation: lbOperation{
			elbv2: mockELBV2Client,
		},
		aliasDomain: domainName,
		lbName:      lbName,
		output:      mockOutput,
		route53:     mockRoute53Client,
	}

	mockELBV2Client.EXPECT().DescribeLoadBalancersByName([]string{"web"}).Return(elbv2.LoadBalancers{lb}, nil)
	mockRoute53Client.EXPECT().ListHostedZones().Return(route53.HostedZones{hostedZone}, nil)
	mockRoute53Client.EXPECT().CreateAlias(createAliasInput).Return("", errors.New("boom"))

	operation.execute()

	if len(mockOutput.FatalMsgs) == 0 {
		t.Error("Expected fatal output, got none")
	} else if mockOutput.FatalMsgs[0].Msg != "Could not alias load balancer" {
		t.Errorf("Expected fatal output == 'Could not alias load balancer', got: %s", mockOutput.FatalMsgs[0])
	}
}

func TestLbAliasOperationHostedZoneNotFound(t *testing.T) {
	domainName := "example.com."
	lbName := "web"

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockELBV2Client := elbv2client.NewMockClient(mockCtrl)
	mockRoute53Client := route53client.NewMockClient(mockCtrl)
	mockOutput := &mock.Output{}

	operation := lbAliasOperation{
		lbOperation: lbOperation{
			elbv2: mockELBV2Client,
		},
		aliasDomain: domainName,
		lbName:      lbName,
		output:      mockOutput,
		route53:     mockRoute53Client,
	}

	mockELBV2Client.EXPECT().DescribeLoadBalancersByName([]string{"web"}).Return(elbv2.LoadBalancers{lb}, nil)
	mockRoute53Client.EXPECT().ListHostedZones().Return(route53.HostedZones{}, nil)

	operation.execute()

	if len(mockOutput.WarnMsgs) == 0 {
		t.Error("Expected warn output, got none")
	} else if mockOutput.WarnMsgs[0] != "Could not find hosted zone for example.com." {
		t.Errorf("Expected warn output == 'Could not find hosted zone for example.com.', got: %s", mockOutput.WarnMsgs[0])
	}
}
