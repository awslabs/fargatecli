package cmd

import (
	//"errors"
	"errors"
	"reflect"
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

func TestSetPorts(t *testing.T) {
	tests := []struct {
		inputPorts  []string
		outputPorts []Port
	}{
		{[]string{"80"}, []Port{Port{80, "HTTP"}}},
		{[]string{"http:80"}, []Port{Port{80, "HTTP"}}},
		{[]string{"hTTp:80"}, []Port{Port{80, "HTTP"}}},
		{[]string{"HTTP:80"}, []Port{Port{80, "HTTP"}}},
		{[]string{"443"}, []Port{Port{443, "HTTPS"}}},
		{[]string{"https:443"}, []Port{Port{443, "HTTPS"}}},
		{[]string{"hTTpS:443"}, []Port{Port{443, "HTTPS"}}},
		{[]string{"HTTPS:443"}, []Port{Port{443, "HTTPS"}}},
		{[]string{"8080"}, []Port{Port{8080, "TCP"}}},
		{[]string{"tcp:8080"}, []Port{Port{8080, "TCP"}}},
		{[]string{"HTTP:8080"}, []Port{Port{8080, "HTTP"}}},
		{[]string{"80", "443"}, []Port{Port{80, "HTTP"}, Port{443, "HTTPS"}}},
		{[]string{"tcp:3386", "TCP:5000"}, []Port{Port{3386, "TCP"}, Port{5000, "TCP"}}},
	}

	for _, test := range tests {
		operation := lbCreateOperation{}
		errs := operation.setPorts(test.inputPorts)

		if len(errs) > 0 {
			t.Fatalf("expected no errors, got %v", errs)
		}

		if !reflect.DeepEqual(operation.ports, test.outputPorts) {
			t.Errorf("expected ports %v, got %v", test.outputPorts, operation.ports)
		}
	}
}

func TestSetPortsMissing(t *testing.T) {
	o := lbCreateOperation{}
	errs := o.setPorts([]string{})

	if len(errs) != 1 {
		t.Fatalf("expected error, got none")
	}

	if expected := errors.New("at least one --port must be specified"); errs[0].Error() != expected.Error() {
		t.Errorf("expected error %v, got %v", expected, errs[0])
	}
}

func TestSetPortsCommingled(t *testing.T) {
	o := lbCreateOperation{}
	errs := o.setPorts([]string{"HTTP:80", "TCP:3386"})

	if len(errs) != 1 {
		t.Fatalf("expected error, got none")
	}

	if expected := errors.New("load balancers do not support commingled TCP and HTTP/HTTPS ports"); errs[0].Error() != expected.Error() {
		t.Errorf("expected error %v, got %v", expected, errs[0])
	}
}

func TestSetPortsCantInflate(t *testing.T) {
	o := lbCreateOperation{}
	errs := o.setPorts([]string{"bargle"})

	if len(errs) != 1 {
		t.Fatalf("expected error, got none")
	}

	if expected := errors.New("could not parse port number from bargle"); errs[0].Error() != expected.Error() {
		t.Errorf("expected error %v, got %v", expected, errs[0])
	}
}

func TestInferType(t *testing.T) {
	tests := []struct {
		inputPorts []string
		lbType     string
	}{
		{[]string{"80"}, "application"},
		{[]string{"443"}, "application"},
		{[]string{"80", "443"}, "application"},
		{[]string{"8080"}, "network"},
		{[]string{"1"}, "network"},
		{[]string{"5000", "2112"}, "network"},
	}

	for _, test := range tests {
		o := lbCreateOperation{}

		o.setPorts(test.inputPorts)
		err := o.inferType()

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if o.lbType != test.lbType {
			t.Errorf("expected %s, got %s", test.lbType, o.lbType)
		}
	}
}

func TestInferTypeNoPorts(t *testing.T) {
	o := lbCreateOperation{}
	err := o.inferType()

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if o.lbType != "" {
		t.Errorf("expected type to not be inferred, got %s", o.lbType)
	}
}

func TestInferTypeInvalidProtocol(t *testing.T) {
	o := lbCreateOperation{
		ports: []Port{Port{80, "INTERWEB"}},
	}
	err := o.inferType()

	if err == nil {
		t.Fatalf("expected error, got none")
	}

	if expected := "could not infer type from port settings"; err.Error() != expected {
		t.Errorf("expected error %s, got %v", expected, err)
	}
}
