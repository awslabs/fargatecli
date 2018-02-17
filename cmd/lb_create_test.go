package cmd

import (
	//"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/jpignata/fargate/acm"
	acmclient "github.com/jpignata/fargate/acm/mock/client"
	"github.com/jpignata/fargate/cmd/mock"
	ec2client "github.com/jpignata/fargate/ec2/mock/client"
	"github.com/jpignata/fargate/elbv2"
	elbv2client "github.com/jpignata/fargate/elbv2/mock/client"
)

var (
	certificates = acm.Certificates{
		acm.Certificate{
			DomainName: "example.com",
			ARN:        "arn:aws:acm:us-east-1:123456789012:certificate/12345678-1234-1234-1234-123456789012",
		},
	}
)

func TestLBCreateOperation(t *testing.T) {
	lbName := "lb"
	lbType := "application"
	lbARN := "arn:aws:elasticloadbalancing:us-east-1:123456789012:loadbalancer/app/lb/50dc6c495c0c9188"
	tgARN := "arn:aws:elasticloadbalancing:us-east-1:123456789012:targetgroup/my-targets/73e2d6bc24d8a067"
	listenerARN := "arn:aws:elasticloadbalancing:us-east-1:123456789012:listener/app/my-load-balancer/50dc6c495c0c9188/f2f7dc8efc522ab2"
	subnetIDs := []string{"subnet-1234567", "subnet-abcdef8"}
	securityGroupIDs := []string{"sg-1234567"}
	vpcID := "vpc-1234567"

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockELBV2Client := elbv2client.NewMockClient(mockCtrl)
	mockACMClient := acmclient.NewMockClient(mockCtrl)
	mockEC2Client := ec2client.NewMockClient(mockCtrl)
	mockOutput := &mock.Output{}

	createLoadBalancerInput := elbv2.CreateLoadBalancerInput{
		Name:             lbName,
		SecurityGroupIDs: securityGroupIDs,
		SubnetIDs:        subnetIDs,
		Type:             lbType,
	}
	createTargetGroupInput := elbv2.CreateTargetGroupInput{
		Name:     "lb-default",
		Port:     80,
		Protocol: "HTTP",
		VPCID:    vpcID,
	}
	createListenerInput := elbv2.CreateListenerInput{
		DefaultTargetGroupARN: tgARN,
		LoadBalancerARN:       lbARN,
		Port:                  80,
		Protocol:              "HTTP",
	}

	mockELBV2Client.EXPECT().CreateLoadBalancer(createLoadBalancerInput).Return(lbARN, nil)
	mockELBV2Client.EXPECT().CreateTargetGroup(createTargetGroupInput).Return(tgARN, nil)
	mockELBV2Client.EXPECT().CreateListener(createListenerInput).Return(listenerARN, nil)

	operation := lbCreateOperation{
		certificateOperation: certificateOperation{
			acm: mockACMClient,
		},
		elbv2:            mockELBV2Client,
		ec2:              mockEC2Client,
		lbType:           lbType,
		loadBalancerName: lbName,
		output:           mockOutput,
		ports:            []Port{Port{80, "HTTP"}},
		securityGroupIDs: securityGroupIDs,
		subnetIDs:        subnetIDs,
		vpcID:            vpcID,
	}

	operation.execute()

	if len(mockOutput.InfoMsgs) != 1 {
		t.Fatalf("expected 1 info msg, got: %d", len(mockOutput.InfoMsgs))
	}

	if expected, got := "Created load balancer lb", mockOutput.InfoMsgs[0]; expected != got {
		t.Errorf("expected %s, got %s", expected, got)
	}
}
