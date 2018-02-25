package elbv2

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	awselbv2 "github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/golang/mock/gomock"
	"github.com/jpignata/fargate/elbv2/mock/sdk"
)

func TestListenerString(t *testing.T) {
	listener := Listener{
		Port:     80,
		Protocol: "HTTP",
	}

	if expected, got := "HTTP:80", listener.String(); got != expected {
		t.Errorf("expected %s, got %s", expected, got)
	}
}

func TestListenersString(t *testing.T) {
	listeners := Listeners{
		Listener{
			Port:     80,
			Protocol: "HTTP",
		},
		Listener{
			Port:     443,
			Protocol: "HTTPS",
		},
	}

	if expected, got := "HTTP:80, HTTPS:443", listeners.String(); got != expected {
		t.Errorf("expected %s, got %s", expected, got)
	}
}

func TestCreateListenerParametersSetCertificateARNs(t *testing.T) {
	certificateARNs := []string{"arn:aws:acm:us-east-1:123456789012:certificate/12345678-1234-1234-1234-123456789012"}
	params := CreateListenerParameters{}

	params.SetCertificateARNs(certificateARNs)

	if !reflect.DeepEqual(params.CertificateARNs, certificateARNs) {
		t.Errorf("expected %v, got %v", certificateARNs, params.CertificateARNs)
	}
}

func TestDescribeListeners(t *testing.T) {
	listenerARN := "arn:aws:elasticloadbalancing:us-west-2:123456789012:listener/app/my-load-balancer/50dc6c495c0c9188/f2f7dc8efc522ab2"
	certificateARN := "arn:aws:acm:us-east-1:123456789012:certificate/12345678-1234-1234-1234-123456789012"

	resp := &awselbv2.DescribeListenersOutput{
		Listeners: []*awselbv2.Listener{
			&awselbv2.Listener{
				ListenerArn: aws.String(listenerARN),
				Port:        aws.Int64(80),
				Protocol:    aws.String("HTTP"),
				Certificates: []*awselbv2.Certificate{
					&awselbv2.Certificate{
						CertificateArn: aws.String(certificateARN),
					},
				},
			},
		},
	}

	mockClient := sdk.MockDescribeListenersClient{Resp: resp}
	elbv2 := SDKClient{client: mockClient}
	listeners, err := elbv2.DescribeListeners("lbARN")

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if len(listeners) != 1 {
		t.Errorf("expected 1 listener, got %d", len(listeners))
	}

	if listeners[0].ARN != listenerARN {
		t.Errorf("expected ARN %s, got %s", listenerARN, listeners[0].ARN)
	}

	if expected := int64(80); expected != listeners[0].Port {
		t.Errorf("expected Port %d, got %d", expected, listeners[0].Port)
	}

	if expected := "HTTP"; expected != listeners[0].Protocol {
		t.Errorf("expected Port %s, got %s", expected, listeners[0].Protocol)
	}

	if listeners[0].CertificateARNs[0] != certificateARN {
		t.Errorf("expected certificate ARN %s, got %s", certificateARN, listeners[0].CertificateARNs[0])
	}
}

func TestCreateListeners(t *testing.T) {
	lbARN := "arn:aws:elasticloadbalancing:us-west-2:123456789012:loadbalancer/app/my-load-balancer/50dc6c495c0c9188"
	port := int64(443)
	protocol := "HTTPS"
	defaultTargetGroupARN := "arn:aws:elasticloadbalancing:us-west-2:123456789012:targetgroup/my-targets/73e2d6bc24d8a067"
	listenerARN := "arn:aws:elasticloadbalancing:us-west-2:123456789012:listener/app/my-load-balancer/50dc6c495c0c9188/f2f7dc8efc522ab2"
	certificateARNs := []string{"arn:aws:acm:us-east-1:123456789012:certificate/12345678-1234-1234-1234-123456789012"}

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockELBV2API := sdk.NewMockELBV2API(mockCtrl)
	elbv2 := SDKClient{client: mockELBV2API}
	i := &awselbv2.CreateListenerInput{
		Port:            aws.Int64(port),
		Protocol:        aws.String(protocol),
		LoadBalancerArn: aws.String(lbARN),
		Certificates: []*awselbv2.Certificate{
			&awselbv2.Certificate{
				CertificateArn: aws.String(certificateARNs[0]),
			},
		},
		DefaultActions: []*awselbv2.Action{
			&awselbv2.Action{
				TargetGroupArn: aws.String(defaultTargetGroupARN),
				Type:           aws.String("forward"),
			},
		},
	}
	o := &awselbv2.CreateListenerOutput{
		Listeners: []*awselbv2.Listener{
			&awselbv2.Listener{
				ListenerArn: aws.String(listenerARN),
			},
		},
	}
	params := CreateListenerParameters{
		CertificateARNs:       certificateARNs,
		Port:                  port,
		Protocol:              protocol,
		LoadBalancerARN:       lbARN,
		DefaultTargetGroupARN: defaultTargetGroupARN,
	}

	mockELBV2API.EXPECT().CreateListener(i).Return(o, nil)

	arn, err := elbv2.CreateListener(params)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if arn != listenerARN {
		t.Errorf("expected ARN %s, got %s", lbARN, arn)
	}
}
