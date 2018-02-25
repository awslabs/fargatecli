package elbv2

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	awselbv2 "github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/golang/mock/gomock"
	"github.com/jpignata/fargate/elbv2/mock/sdk"
)

func TestCreateTargetGroup(t *testing.T) {
	name := "default"
	port := int64(80)
	protocol := "HTTP"
	vpcID := "vpc-1234567"
	targetGroupARN := "arn:aws:elasticloadbalancing:us-west-2:123456789012:targetgroup/my-targets/73e2d6bc24d8a067"

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockELBV2API := sdk.NewMockELBV2API(mockCtrl)
	elbv2 := SDKClient{client: mockELBV2API}

	i := &awselbv2.CreateTargetGroupInput{
		Name:       aws.String(name),
		Port:       aws.Int64(port),
		Protocol:   aws.String(protocol),
		TargetType: aws.String("ip"),
		VpcId:      aws.String(vpcID),
	}
	o := &awselbv2.CreateTargetGroupOutput{
		TargetGroups: []*awselbv2.TargetGroup{
			&awselbv2.TargetGroup{
				TargetGroupArn: aws.String(targetGroupARN),
			},
		},
	}

	mockELBV2API.EXPECT().CreateTargetGroup(i).Return(o, nil)

	arn, err := elbv2.CreateTargetGroup(
		CreateTargetGroupParameters{
			Name:     name,
			Port:     port,
			Protocol: protocol,
			VPCID:    vpcID,
		},
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if arn == "" {
		t.Errorf("expected ARN %s, got %s", targetGroupARN, arn)
	}
}

func TestCreateTargetGroupError(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockELBV2API := sdk.NewMockELBV2API(mockCtrl)
	elbv2 := SDKClient{client: mockELBV2API}

	mockELBV2API.EXPECT().CreateTargetGroup(gomock.Any()).Return(&awselbv2.CreateTargetGroupOutput{}, errors.New("boom"))

	arn, err := elbv2.CreateTargetGroup(
		CreateTargetGroupParameters{
			Name:     "default",
			Port:     int64(80),
			Protocol: "HTTP",
			VPCID:    "vpc-1234567",
		},
	)

	if err == nil {
		t.Fatalf("expected error, got none")
	}

	if arn != "" {
		t.Errorf("expected empty ARN, got %s", arn)
	}
}
