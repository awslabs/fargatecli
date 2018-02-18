package cmd

import (
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
		vpcOperation: vpcOperation{
			ec2:              mockEC2Client,
			securityGroupIDs: securityGroupIDs,
			subnetIDs:        subnetIDs,
			vpcID:            vpcID,
		},
		elbv2:  mockELBV2Client,
		lbType: lbType,
		lbName: lbName,
		output: mockOutput,
		ports:  []Port{Port{80, "HTTP"}},
	}

	operation.execute()

	if len(mockOutput.InfoMsgs) != 1 {
		t.Fatalf("expected 1 info msg, got: %d", len(mockOutput.InfoMsgs))
	}

	if expected, got := "Created load balancer lb", mockOutput.InfoMsgs[0]; expected != got {
		t.Errorf("expected: %s, got: %s", expected, got)
	}
}

func TestLBCreateOperationLBError(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockELBV2Client := elbv2client.NewMockClient(mockCtrl)
	mockACMClient := acmclient.NewMockClient(mockCtrl)
	mockEC2Client := ec2client.NewMockClient(mockCtrl)
	mockOutput := &mock.Output{}

	mockELBV2Client.EXPECT().CreateLoadBalancer(gomock.Any()).Return("", errors.New("boom"))

	operation := lbCreateOperation{
		certificateOperation: certificateOperation{
			acm: mockACMClient,
		},
		vpcOperation: vpcOperation{
			ec2: mockEC2Client,
		},
		elbv2:  mockELBV2Client,
		lbType: "application",
		lbName: "web",
		output: mockOutput,
		ports:  []Port{Port{80, "HTTP"}},
	}

	operation.execute()

	if len(mockOutput.FatalMsgs) != 1 {
		t.Fatalf("expected 1 fatal msg, got: %d", len(mockOutput.FatalMsgs))
	}

	if expected, got := "Could not create load balancer", mockOutput.FatalMsgs[0].Msg; expected != got {
		t.Errorf("expected: %s, got: %s", expected, got)
	}
}

func TestLBCreateOperationTargetGroupError(t *testing.T) {
	lbName := "lb"
	lbType := "application"
	lbARN := "arn:aws:elasticloadbalancing:us-east-1:123456789012:loadbalancer/app/lb/50dc6c495c0c9188"
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

	mockELBV2Client.EXPECT().CreateLoadBalancer(createLoadBalancerInput).Return(lbARN, nil)
	mockELBV2Client.EXPECT().CreateTargetGroup(createTargetGroupInput).Return("", errors.New("boom"))

	operation := lbCreateOperation{
		certificateOperation: certificateOperation{
			acm: mockACMClient,
		},
		vpcOperation: vpcOperation{
			ec2:              mockEC2Client,
			securityGroupIDs: securityGroupIDs,
			subnetIDs:        subnetIDs,
			vpcID:            vpcID,
		},
		elbv2:  mockELBV2Client,
		lbType: lbType,
		lbName: lbName,
		output: mockOutput,
		ports:  []Port{Port{80, "HTTP"}},
	}

	operation.execute()

	if len(mockOutput.FatalMsgs) != 1 {
		t.Fatalf("expected 1 fatal msg, got: %d", len(mockOutput.FatalMsgs))
	}

	if expected, got := "Could not create default target group", mockOutput.FatalMsgs[0].Msg; expected != got {
		t.Errorf("expected: %s, got: %s", expected, got)
	}
}

func TestLBCreateOperationListenerError(t *testing.T) {
	lbName := "lb"
	lbType := "application"
	lbARN := "arn:aws:elasticloadbalancing:us-east-1:123456789012:loadbalancer/app/lb/50dc6c495c0c9188"
	tgARN := "arn:aws:elasticloadbalancing:us-east-1:123456789012:targetgroup/my-targets/73e2d6bc24d8a067"
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
	mockELBV2Client.EXPECT().CreateListener(createListenerInput).Return("", errors.New("boom"))

	operation := lbCreateOperation{
		certificateOperation: certificateOperation{
			acm: mockACMClient,
		},
		vpcOperation: vpcOperation{
			ec2:              mockEC2Client,
			securityGroupIDs: securityGroupIDs,
			subnetIDs:        subnetIDs,
			vpcID:            vpcID,
		},
		elbv2:  mockELBV2Client,
		lbType: lbType,
		lbName: lbName,
		output: mockOutput,
		ports:  []Port{Port{80, "HTTP"}},
	}

	operation.execute()

	if len(mockOutput.FatalMsgs) != 1 {
		t.Fatalf("expected 1 fatal msg, got: %d", len(mockOutput.FatalMsgs))
	}

	if expected, got := "Could not create listener", mockOutput.FatalMsgs[0].Msg; expected != got {
		t.Errorf("expected: %s, got: %s", expected, got)
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
			t.Fatalf("expected no errors, got: %v", errs)
		}

		if !reflect.DeepEqual(operation.ports, test.outputPorts) {
			t.Errorf("expected ports %v, got: %v", test.outputPorts, operation.ports)
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
		t.Errorf("expected error %v, got: %v", expected, errs[0])
	}
}

func TestSetPortsCommingled(t *testing.T) {
	o := lbCreateOperation{}
	errs := o.setPorts([]string{"HTTP:80", "TCP:3386"})

	if len(errs) != 1 {
		t.Fatalf("expected error, got none")
	}

	if expected := errors.New("load balancers do not support commingled TCP and HTTP/HTTPS ports"); errs[0].Error() != expected.Error() {
		t.Errorf("expected error %v, got: %v", expected, errs[0])
	}
}

func TestSetPortsCantInflate(t *testing.T) {
	o := lbCreateOperation{}
	errs := o.setPorts([]string{"bargle"})

	if len(errs) != 1 {
		t.Fatalf("expected error, got none")
	}

	if expected := errors.New("could not parse port number from bargle"); errs[0].Error() != expected.Error() {
		t.Errorf("expected error %v, got: %v", expected, errs[0])
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
			t.Fatalf("expected no error, got: %v", err)
		}

		if o.lbType != test.lbType {
			t.Errorf("expected: %s, got: %s", test.lbType, o.lbType)
		}
	}
}

func TestInferTypeNoPorts(t *testing.T) {
	o := lbCreateOperation{}
	err := o.inferType()

	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	if o.lbType != "" {
		t.Errorf("expected type to not be inferred, got: %s", o.lbType)
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
		t.Errorf("expected error %s, got: %v", expected, err)
	}
}

func TestSetCertificateARNs(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	domainName := "example.com"
	certificateARN := "arn:aws:acm:us-east-1:123456789012:certificate/12345678-1234-1234-1234-123456789012"
	certificate := acm.Certificate{
		ARN:        certificateARN,
		DomainName: domainName,
		Status:     "ISSUED",
	}
	certificateList := acm.Certificates{certificate}
	mockClient := acmclient.NewMockClient(mockCtrl)

	mockClient.EXPECT().ListCertificates().Return(certificateList, nil)
	mockClient.EXPECT().InflateCertificate(certificate).Return(certificate, nil)

	o := lbCreateOperation{
		certificateOperation: certificateOperation{
			acm: mockClient,
		},
	}

	errs := o.setCertificateARNs([]string{domainName})

	if len(errs) > 0 {
		t.Fatalf("expected no errors, got: %v", errs)
	}

	if len(o.certificateARNs) != 1 {
		t.Fatalf("expected 1 certificate ARN, got: %d", len(o.certificateARNs))
	}

	if o.certificateARNs[0] != certificateARN {
		t.Errorf("expected certificate ARN %s, got: %s", certificateARN, o.certificateARNs[0])
	}
}

func TestSetCertificateARNsNotIssued(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	domainName := "example.com"
	certificateARN := "arn:aws:acm:us-east-1:123456789012:certificate/12345678-1234-1234-1234-123456789012"
	certificate := acm.Certificate{
		ARN:        certificateARN,
		DomainName: domainName,
		Status:     "FAILED",
	}
	certificateList := acm.Certificates{certificate}
	mockClient := acmclient.NewMockClient(mockCtrl)

	mockClient.EXPECT().ListCertificates().Return(certificateList, nil)
	mockClient.EXPECT().InflateCertificate(certificate).Return(certificate, nil)

	o := lbCreateOperation{
		certificateOperation: certificateOperation{
			acm: mockClient,
		},
	}

	errs := o.setCertificateARNs([]string{domainName})

	if len(errs) == 0 {
		t.Fatalf("expected 1 errors, got none")
	}

	if expected := "certificate example.com is in state failed"; errs[0].Error() != expected {
		t.Fatalf("expected error %s, got: %v", expected, errs[0])
	}

	if len(o.certificateARNs) > 0 {
		t.Fatalf("expected no certificate ARNs, got: %v", o.certificateARNs)
	}
}

func TestSetCertificateARNsNotFound(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockClient := acmclient.NewMockClient(mockCtrl)

	mockClient.EXPECT().ListCertificates().Return(acm.Certificates{}, nil)

	o := lbCreateOperation{
		certificateOperation: certificateOperation{
			acm: mockClient,
		},
	}

	errs := o.setCertificateARNs([]string{"example.com"})

	if len(errs) == 0 {
		t.Fatalf("expected 1 errors, got none")
	}

	if expected := "no certificate found for example.com"; errs[0].Error() != expected {
		t.Fatalf("expected error %s, got: %v", expected, errs[0])
	}

	if len(o.certificateARNs) > 0 {
		t.Fatalf("expected no certificate ARNs, got: %v", o.certificateARNs)
	}
}

func TestSetCertificateARNsTooManyFound(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	certificate := acm.Certificate{DomainName: "example.com"}
	certificateList := acm.Certificates{certificate, certificate}
	mockClient := acmclient.NewMockClient(mockCtrl)

	mockClient.EXPECT().ListCertificates().Return(certificateList, nil)

	o := lbCreateOperation{
		certificateOperation: certificateOperation{
			acm: mockClient,
		},
	}

	errs := o.setCertificateARNs([]string{"example.com"})

	if len(errs) == 0 {
		t.Fatalf("expected 1 errors, got none")
	}

	if expected := "multiple certificates found for example.com"; errs[0].Error() != expected {
		t.Fatalf("expected error %s, got: %v", expected, errs[0])
	}

	if len(o.certificateARNs) > 0 {
		t.Fatalf("expected no certificate ARNs, got: %v", o.certificateARNs)
	}
}

func TestSetCertificateARNsError(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockClient := acmclient.NewMockClient(mockCtrl)

	mockClient.EXPECT().ListCertificates().Return(acm.Certificates{}, errors.New("boom"))

	o := lbCreateOperation{
		certificateOperation: certificateOperation{
			acm: mockClient,
		},
	}

	errs := o.setCertificateARNs([]string{"example.com"})

	if len(errs) == 0 {
		t.Fatalf("expected 1 errors, got none")
	}

	if expected := "could not find certificate ARN: boom"; errs[0].Error() != expected {
		t.Fatalf("expected error %s, got: %v", expected, errs[0])
	}

	if len(o.certificateARNs) > 0 {
		t.Fatalf("expected no certificate ARNs, got: %v", o.certificateARNs)
	}
}

func TestValidate(t *testing.T) {
	o := lbCreateOperation{
		lbName: "web",
		lbType: "application",
		vpcOperation: vpcOperation{
			subnetIDs: []string{"subnet-abcdef", "subnet-1234567"},
		},
	}

	errs := o.validate()

	if len(errs) > 0 {
		t.Errorf("expected no errors, got: %v", errs)
	}
}

func TestValidateNoName(t *testing.T) {
	o := lbCreateOperation{
		lbType: "application",
		vpcOperation: vpcOperation{
			subnetIDs: []string{"subnet-abcdef", "subnet-1234567"},
		},
	}

	errs := o.validate()

	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got: %v", errs)
	}

	if expected := "--name is required"; errs[0].Error() != expected {
		t.Errorf("expected: %s, got: %v", expected, errs)
	}
}

func TestValidateApplicationLBNoSubnets(t *testing.T) {
	o := lbCreateOperation{
		lbName: "web",
		lbType: "application",
	}

	errs := o.validate()

	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got: %v", errs)
	}

	if expected := "HTTP/HTTPS load balancers require two subnet IDs from unique Availability Zones"; errs[0].Error() != expected {
		t.Errorf("expected: %s, got: %v", expected, errs)
	}
}

func TestValidateNetworkLBWithSGs(t *testing.T) {
	o := lbCreateOperation{
		lbName: "web",
		lbType: "network",
		vpcOperation: vpcOperation{
			securityGroupIDs: []string{"sg-abcdef"},
		},
	}

	errs := o.validate()

	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got: %v", errs)
	}

	if expected := "security groups can only be specified for HTTP/HTTPS load balancers"; errs[0].Error() != expected {
		t.Errorf("expected: %s, got: %v", expected, errs)
	}
}

func TestNewLBCreateOperation(t *testing.T) {
	domainName := "example.com"
	certificateARN := "arn:aws:acm:us-east-1:123456789012:certificate/12345678-1234-1234-1234-123456789012"
	certificate := acm.Certificate{
		ARN:        certificateARN,
		DomainName: domainName,
		Status:     "ISSUED",
	}
	mockOutput := &mock.Output{}
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockEC2 := ec2client.NewMockClient(mockCtrl)
	mockACM := acmclient.NewMockClient(mockCtrl)
	mockELBV2 := elbv2client.NewMockClient(mockCtrl)

	mockEC2.EXPECT().GetSubnetVPCID("subnet-1234567").Return("vpc-1234567", nil)
	mockACM.EXPECT().ListCertificates().Return(acm.Certificates{certificate}, nil)
	mockACM.EXPECT().InflateCertificate(certificate).Return(certificate, nil)

	o, errs := newLBCreateOperation(
		"web",
		[]string{"example.com"},
		[]string{"80", "443"},
		[]string{"sg-abcdef"},
		[]string{"subnet-1234567", "subnet-abcdef"},
		mockOutput,
		mockACM,
		mockEC2,
		mockELBV2,
	)

	if len(errs) > 0 {
		t.Fatalf("expected no error, got: %v", errs)
	}

	if o.acm != mockACM {
		t.Errorf("acm client not set")
	}

	if o.ec2 != mockEC2 {
		t.Errorf("ec2 client not set")
	}

	if o.elbv2 != mockELBV2 {
		t.Errorf("elbv2 client not set")
	}

	if o.output != mockOutput {
		t.Errorf("output not set")
	}

	if o.lbName != "web" {
		t.Errorf("expected lbName == web, got: %s", o.lbName)
	}

	if len(o.ports) != 2 {
		t.Fatalf("expected 2 ports, got: %d", len(o.ports))
	}

	if o.ports[0].Number != 80 || o.ports[0].Protocol != "HTTP" {
		t.Errorf("expected port HTTP:80, got: %v", o.ports)
	}

	if o.ports[1].Number != 443 || o.ports[1].Protocol != "HTTPS" {
		t.Errorf("expected port HTTPS:443, got: %v", o.ports)
	}

	if o.lbType != "application" {
		t.Errorf("expected lbType == application, got: %s", o.lbType)
	}

	if o.certificateARNs[0] != certificateARN {
		t.Errorf("expected certificate ARN %s, got: %s", certificateARN, o.certificateARNs)
	}
}

func TestNewLBCreateOperationNoName(t *testing.T) {
	mockOutput := &mock.Output{}
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	ec2 := ec2client.NewMockClient(mockCtrl)
	ec2.EXPECT().GetSubnetVPCID(gomock.Any()).Return("vpc-1234567", nil)

	_, err := newLBCreateOperation(
		"",
		[]string{},
		[]string{"80"},
		[]string{"sg-abcdef"},
		[]string{"subnet-abcdef", "subnet-1234567"},
		mockOutput,
		acmclient.NewMockClient(mockCtrl),
		ec2,
		elbv2client.NewMockClient(mockCtrl),
	)

	if err == nil {
		t.Fatalf("expected errors, got none")
	}

	if expected := "--name is required"; err[0].Error() != expected {
		t.Errorf("expected: %s, got: %v", expected, err)
	}
}

func TestNewLBCreateOperationNoPort(t *testing.T) {
	mockOutput := &mock.Output{}
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockEC2 := ec2client.NewMockClient(mockCtrl)

	mockEC2.EXPECT().GetDefaultSubnetIDs().Return([]string{"subnet-1234567"}, nil)
	mockEC2.EXPECT().GetSubnetVPCID("subnet-1234567").Return("vpc-1234567", nil)

	_, err := newLBCreateOperation(
		"web",
		[]string{},
		[]string{},
		[]string{},
		[]string{},
		mockOutput,
		acmclient.NewMockClient(mockCtrl),
		mockEC2,
		elbv2client.NewMockClient(mockCtrl),
	)

	if err == nil {
		t.Fatalf("expected errors, got none")
	}

	if expected := "at least one --port must be specified"; err[0].Error() != expected {
		t.Errorf("expected: %s, got: %v", expected, err)
	}
}

func TestNewLBCreateOperationDescribeSubnetsError(t *testing.T) {
	mockOutput := &mock.Output{}
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	ec2 := ec2client.NewMockClient(mockCtrl)
	ec2.EXPECT().GetSubnetVPCID(gomock.Any()).Return("", errors.New("boom"))

	_, err := newLBCreateOperation(
		"web",
		[]string{},
		[]string{"80"},
		[]string{"sg-abcdef"},
		[]string{"subnet-abcdef", "subnet-1234567"},
		mockOutput,
		acmclient.NewMockClient(mockCtrl),
		ec2,
		elbv2client.NewMockClient(mockCtrl),
	)

	if err == nil {
		t.Fatalf("expected errors, got none")
	}

	if expected := "boom"; err[0].Error() != expected {
		t.Errorf("expected: %s, got: %v", expected, err)
	}
}

func TestNewLBCreateOperationInvalidProtocol(t *testing.T) {
	mockOutput := &mock.Output{}
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	ec2 := ec2client.NewMockClient(mockCtrl)
	ec2.EXPECT().GetSubnetVPCID(gomock.Any()).Return("vpc-1234567", nil)

	_, err := newLBCreateOperation(
		"web",
		[]string{},
		[]string{"SMTP:25"},
		[]string{"sg-abcdef"},
		[]string{"subnet-abcdef", "subnet-1234567"},
		mockOutput,
		acmclient.NewMockClient(mockCtrl),
		ec2,
		elbv2client.NewMockClient(mockCtrl),
	)

	if err == nil {
		t.Fatalf("expected errors, got none")
	}

	if expected := "invalid protocol SMTP (specify TCP, HTTP, or HTTPS)"; err[0].Error() != expected {
		t.Errorf("expected: %s, got: %v", expected, err)
	}
}

func TestNewLBCreateOperationUseDefaultSG(t *testing.T) {
	mockOutput := &mock.Output{}
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	ec2 := ec2client.NewMockClient(mockCtrl)
	ec2.EXPECT().GetDefaultSecurityGroupID().Return("sg-1234567", nil)
	ec2.EXPECT().GetSubnetVPCID(gomock.Any()).Return("vpc-1234567", nil)

	o, err := newLBCreateOperation(
		"web",
		[]string{},
		[]string{"80"},
		[]string{},
		[]string{"subnet-abcdef", "subnet-1234567"},
		mockOutput,
		acmclient.NewMockClient(mockCtrl),
		ec2,
		elbv2client.NewMockClient(mockCtrl),
	)

	if err != nil {
		t.Fatalf("expected no errors, got: %v", err)
	}

	if o.securityGroupIDs[0] != "sg-1234567" {
		t.Errorf("expected SG sg-1234567, got: %v", o.securityGroupIDs)
	}
}

func TestNewLBCreateOperationDefaultSGError(t *testing.T) {
	mockOutput := &mock.Output{}
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	ec2 := ec2client.NewMockClient(mockCtrl)
	ec2.EXPECT().GetDefaultSecurityGroupID().Return("", errors.New("boom"))
	ec2.EXPECT().GetSubnetVPCID(gomock.Any()).Return("vpc-1234567", nil)

	_, errs := newLBCreateOperation(
		"web",
		[]string{},
		[]string{"80"},
		[]string{},
		[]string{"subnet-abcdef", "subnet-1234567"},
		mockOutput,
		acmclient.NewMockClient(mockCtrl),
		ec2,
		elbv2client.NewMockClient(mockCtrl),
	)

	if len(errs) == 0 {
		t.Fatalf("expected error, got none")
	}

	if expected := "boom"; errs[0].Error() != expected {
		t.Errorf("expected error %s, got: %v", expected, errs)
	}
}

func TestNewLBCreateOperationCertificateError(t *testing.T) {
	mockOutput := &mock.Output{}
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockEC2 := ec2client.NewMockClient(mockCtrl)
	mockACM := acmclient.NewMockClient(mockCtrl)
	mockELBV2 := elbv2client.NewMockClient(mockCtrl)

	mockEC2.EXPECT().GetSubnetVPCID("subnet-1234567").Return("vpc-1234567", nil)
	mockACM.EXPECT().ListCertificates().Return(acm.Certificates{}, errors.New("boom"))

	_, errs := newLBCreateOperation(
		"web",
		[]string{"example.com"},
		[]string{"80", "443"},
		[]string{"sg-abcdef"},
		[]string{"subnet-1234567", "subnet-abcdef"},
		mockOutput,
		mockACM,
		mockEC2,
		mockELBV2,
	)

	if len(errs) != 1 {
		t.Fatalf("expected error, got none")
	}

	if expected := "could not find certificate ARN: boom"; errs[0].Error() != expected {
		t.Errorf("expected error %s, got: %v", expected, errs)
	}
}
