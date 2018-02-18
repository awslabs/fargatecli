package cmd

import (
	"errors"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/jpignata/fargate/cmd/mock"
	"github.com/jpignata/fargate/elbv2"
	elbv2client "github.com/jpignata/fargate/elbv2/mock/client"
)

func TestLBListOperation(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockClient := elbv2client.NewMockClient(mockCtrl)
	mockOutput := &mock.Output{}

	loadBalancer1 := elbv2.LoadBalancer{
		ARN:     "arn:aws:elasticloadbalancing:us-east-1:123456789012:loadbalancer/app/lb/50dc6c495c0c9188",
		DNSName: "test-12345678.us-east-1.elb.amazonaws.com",
		Name:    "test",
		Type:    "application",
		State:   "active",
	}
	loadBalancer2 := elbv2.LoadBalancer{
		ARN:     "arn:aws:elasticloadbalancing:us-east-1:123456789012:loadbalancer/app/lb/93fa3d386bec918a",
		DNSName: "test-abcdef.us-east-1.elb.amazonaws.com",
		Name:    "test2",
		Type:    "application",
		State:   "active",
	}
	listener1 := elbv2.Listener{
		ARN:      "arn:aws:elasticloadbalancing:us-east-1:123456789012:listener/app/my-load-balancer/50dc6c495c0c9188/f2f7dc8efc522ab2",
		Port:     80,
		Protocol: "HTTP",
	}
	listener2 := elbv2.Listener{
		ARN:      "arn:aws:elasticloadbalancing:us-east-1:123456789012:listener/app/my-load-balancer/50dc6c495c0c9188/f2f7dc8efc522ab2",
		Port:     8080,
		Protocol: "HTTP",
	}
	loadBalancers := elbv2.LoadBalancers{loadBalancer1, loadBalancer2}
	listeners1 := elbv2.Listeners{listener1}
	listeners2 := elbv2.Listeners{listener2}

	mockClient.EXPECT().DescribeLoadBalancers().Return(loadBalancers, nil)
	mockClient.EXPECT().DescribeListeners(loadBalancer1.ARN).Return(listeners1, nil)
	mockClient.EXPECT().DescribeListeners(loadBalancer2.ARN).Return(listeners2, nil)

	lbListOperation{
		elbv2:  mockClient,
		output: mockOutput,
	}.execute()

	if len(mockOutput.Tables) == 0 {
		t.Fatalf("expected table, got none")
	}

	if len(mockOutput.Tables[0].Rows) != 3 {
		t.Errorf("expected table with 3 rows, got %d", len(mockOutput.Tables[0].Rows))
	}

	if expected, got := []string{"NAME", "TYPE", "STATUS", "DNS NAME", "PORTS"}, mockOutput.Tables[0].Rows[0]; !reflect.DeepEqual(expected, got) {
		t.Errorf("expected column headers: %v, got: %v", expected, got)
	}

	row1 := mockOutput.Tables[0].Rows[1]

	if row1[0] != loadBalancer1.Name {
		t.Errorf("expected name: %s, got: %s", loadBalancer1.Name, row1[0])
	}

	if expected := Titleize(loadBalancer1.Type); row1[1] != expected {
		t.Errorf("expected type: %s, got: %s", expected, row1[1])
	}

	if expected := Titleize(loadBalancer1.State); row1[2] != expected {
		t.Errorf("expected status: %s, got: %s", expected, row1[2])
	}

	if row1[3] != loadBalancer1.DNSName {
		t.Errorf("expected DNS name: %s, got: %s", loadBalancer1.DNSName, row1[3])
	}

	if expected := "HTTP:80"; row1[4] != expected {
		t.Errorf("expected ports: %s, got: %s", expected, row1[4])
	}

	row2 := mockOutput.Tables[0].Rows[2]

	if row2[0] != loadBalancer2.Name {
		t.Errorf("expected name: %s, got: %s", loadBalancer2.Name, row2[0])
	}

	if expected := Titleize(loadBalancer2.Type); row2[1] != expected {
		t.Errorf("expected type: %s, got: %s", expected, row2[1])
	}

	if expected := Titleize(loadBalancer2.State); row2[2] != expected {
		t.Errorf("expected status: %s, got: %s", expected, row2[2])
	}

	if row2[3] != loadBalancer2.DNSName {
		t.Errorf("expected DNS name: %s, got: %s", loadBalancer1.DNSName, row2[3])
	}

	if expected := "HTTP:8080"; row2[4] != expected {
		t.Errorf("expected ports: %s, got: %s", expected, row2[4])
	}
}

func TestLBListOperationLBDescribeError(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockClient := elbv2client.NewMockClient(mockCtrl)
	mockOutput := &mock.Output{}

	mockClient.EXPECT().DescribeLoadBalancers().Return(elbv2.LoadBalancers{}, errors.New("boom"))

	lbListOperation{
		elbv2:  mockClient,
		output: mockOutput,
	}.execute()

	if len(mockOutput.FatalMsgs) == 0 {
		t.Fatalf("expected fatal output, got none")
	}

	if expected, got := "Could not list load balancers", mockOutput.FatalMsgs[0].Msg; got != expected {
		t.Errorf("expected fatal output: %s, got: %s", expected, got)
	}
}

func TestLBListOperationListenerDescribeError(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockClient := elbv2client.NewMockClient(mockCtrl)
	mockOutput := &mock.Output{}

	loadBalancers := elbv2.LoadBalancers{
		elbv2.LoadBalancer{ARN: "lbARN"},
	}

	mockClient.EXPECT().DescribeLoadBalancers().Return(loadBalancers, nil)
	mockClient.EXPECT().DescribeListeners("lbARN").Return(elbv2.Listeners{}, errors.New("boom"))

	lbListOperation{
		elbv2:  mockClient,
		output: mockOutput,
	}.execute()

	if len(mockOutput.FatalMsgs) == 0 {
		t.Fatalf("expected fatal output, got none")
	}

	if expected, got := "Could not list load balancers", mockOutput.FatalMsgs[0].Msg; got != expected {
		t.Errorf("expected fatal output: %s, got: %s", expected, got)
	}
}

func TestLBListOperationNoneFound(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockClient := elbv2client.NewMockClient(mockCtrl)
	mockOutput := &mock.Output{}

	mockClient.EXPECT().DescribeLoadBalancers().Return(elbv2.LoadBalancers{}, nil)

	lbListOperation{
		elbv2:  mockClient,
		output: mockOutput,
	}.execute()

	if len(mockOutput.InfoMsgs) == 0 {
		t.Fatalf("expected info output, got none")
	}

	if expected, got := "No load balancers found", mockOutput.InfoMsgs[0]; got != expected {
		t.Errorf("expected info output: %s, got: %s", expected, got)
	}
}
