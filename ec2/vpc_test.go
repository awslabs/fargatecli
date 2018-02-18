package ec2

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	awsec2 "github.com/aws/aws-sdk-go/service/ec2"
	"github.com/golang/mock/gomock"
	"github.com/jpignata/fargate/ec2/mock/sdk"
)

func TestGetDefaultSubnetIDs(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	subnetID := "subnet-abcdef"
	subnet := &awsec2.Subnet{
		SubnetId: aws.String(subnetID),
	}
	filter := &awsec2.Filter{
		Name:   aws.String("default-for-az"),
		Values: aws.StringSlice([]string{"true"}),
	}
	input := &awsec2.DescribeSubnetsInput{
		Filters: []*awsec2.Filter{filter},
	}
	output := &awsec2.DescribeSubnetsOutput{
		Subnets: []*awsec2.Subnet{subnet},
	}

	mockEC2Client := sdk.NewMockEC2API(mockCtrl)
	ec2 := SDKClient{client: mockEC2Client}

	mockEC2Client.EXPECT().DescribeSubnets(input).Return(output, nil)

	out, err := ec2.GetDefaultSubnetIDs()

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if out[0] != subnetID {
		t.Errorf("expected %s, got %s", subnetID, out[0])
	}
}

func TestGetDefaultSubnetIDsError(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockEC2Client := sdk.NewMockEC2API(mockCtrl)
	ec2 := SDKClient{client: mockEC2Client}

	mockEC2Client.EXPECT().DescribeSubnets(gomock.Any()).Return(&awsec2.DescribeSubnetsOutput{}, errors.New("boom"))

	out, err := ec2.GetDefaultSubnetIDs()

	if len(out) > 0 {
		t.Errorf("expected no results, got %v", out)
	}

	if err == nil {
		t.Errorf("expected error, got none")
	}
}

func TestGetDefaultSecurityGroupID(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	securityGroupID := "sg-abcdef"
	securityGroup := &awsec2.SecurityGroup{
		GroupId: aws.String(securityGroupID),
	}
	input := &awsec2.DescribeSecurityGroupsInput{
		GroupNames: aws.StringSlice([]string{"fargate-default"}),
	}
	output := &awsec2.DescribeSecurityGroupsOutput{
		SecurityGroups: []*awsec2.SecurityGroup{securityGroup},
	}

	mockEC2Client := sdk.NewMockEC2API(mockCtrl)
	ec2 := SDKClient{client: mockEC2Client}

	mockEC2Client.EXPECT().DescribeSecurityGroups(input).Return(output, nil)

	out, err := ec2.GetDefaultSecurityGroupID()

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if out != securityGroupID {
		t.Errorf("expected %s, got %s", securityGroupID, out)
	}
}

func TestGetDefaultSecurityGroupIDError(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockEC2Client := sdk.NewMockEC2API(mockCtrl)
	ec2 := SDKClient{client: mockEC2Client}

	mockEC2Client.EXPECT().DescribeSecurityGroups(gomock.Any()).Return(&awsec2.DescribeSecurityGroupsOutput{}, errors.New("boom"))

	out, err := ec2.GetDefaultSecurityGroupID()

	if out != "" {
		t.Errorf("expected no result, got %v", out)
	}

	if err == nil {
		t.Errorf("expected error, got none")
	}
}

func TestGetDefaultSecurityGroupIDGroupNotFound(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockEC2Client := sdk.NewMockEC2API(mockCtrl)
	ec2 := SDKClient{client: mockEC2Client}
	awserr := awserr.New("InvalidGroup.NotFound", "Group not found", errors.New("boom"))

	mockEC2Client.EXPECT().DescribeSecurityGroups(gomock.Any()).Return(&awsec2.DescribeSecurityGroupsOutput{}, awserr)

	out, err := ec2.GetDefaultSecurityGroupID()

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if out != "" {
		t.Errorf("expected no result, got %v", out)
	}
}

func TestGetSubnetVPCID(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	subnetID := "subnet-abcdef"
	vpcID := "vpc-abcdef"
	subnet := &awsec2.Subnet{
		VpcId: aws.String(vpcID),
	}
	input := &awsec2.DescribeSubnetsInput{
		SubnetIds: aws.StringSlice([]string{subnetID}),
	}
	output := &awsec2.DescribeSubnetsOutput{
		Subnets: []*awsec2.Subnet{subnet},
	}

	mockEC2Client := sdk.NewMockEC2API(mockCtrl)
	ec2 := SDKClient{client: mockEC2Client}

	mockEC2Client.EXPECT().DescribeSubnets(input).Return(output, nil)

	out, err := ec2.GetSubnetVPCID(subnetID)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if out != vpcID {
		t.Errorf("expected %s, got %s", vpcID, out)
	}
}

func TestGetSubnetVPCIDError(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockEC2Client := sdk.NewMockEC2API(mockCtrl)
	ec2 := SDKClient{client: mockEC2Client}

	mockEC2Client.EXPECT().DescribeSubnets(gomock.Any()).Return(&awsec2.DescribeSubnetsOutput{}, errors.New("boom"))

	out, err := ec2.GetSubnetVPCID("subnet-abcdef")

	if err == nil {
		t.Errorf("expected error, got none")
	}

	if expected := errors.New("could not find VPC ID for subnet ID subnet-abcdef: boom"); err.Error() != expected.Error() {
		t.Errorf("expected error %v, got %v", expected, err)
	}

	if out != "" {
		t.Errorf("expected no result, got %s", out)
	}
}

func TestGetSubnetVPCIDSubnetNotFound(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	subnetID := "subnet-abcdef"
	mockEC2Client := sdk.NewMockEC2API(mockCtrl)
	ec2 := SDKClient{client: mockEC2Client}

	mockEC2Client.EXPECT().DescribeSubnets(gomock.Any()).Return(&awsec2.DescribeSubnetsOutput{}, nil)

	out, err := ec2.GetSubnetVPCID(subnetID)

	if err == nil {
		t.Errorf("expected error, got none")
	}

	if expected := errors.New("could not find VPC ID: subnet ID subnet-abcdef not found"); err.Error() != expected.Error() {
		t.Errorf("expected error %v, got %v", expected, err)
	}

	if out != "" {
		t.Errorf("expected no result, got %s", out)
	}
}

func TestCreateDefaultSecurityGroup(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	securityGroupID := "sg-abcdef"
	input := &awsec2.CreateSecurityGroupInput{
		GroupName:   aws.String("fargate-default"),
		Description: aws.String("Default Fargate CLI SG"),
	}
	output := &awsec2.CreateSecurityGroupOutput{
		GroupId: aws.String(securityGroupID),
	}

	mockEC2Client := sdk.NewMockEC2API(mockCtrl)
	ec2 := SDKClient{client: mockEC2Client}

	mockEC2Client.EXPECT().CreateSecurityGroup(input).Return(output, nil)

	out, err := ec2.CreateDefaultSecurityGroup()

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if out != securityGroupID {
		t.Errorf("expected %s, got %s", securityGroupID, out)
	}
}

func TestCreateDefaultSecurityGroupError(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockEC2Client := sdk.NewMockEC2API(mockCtrl)
	ec2 := SDKClient{client: mockEC2Client}

	mockEC2Client.EXPECT().CreateSecurityGroup(gomock.Any()).Return(&awsec2.CreateSecurityGroupOutput{}, errors.New("boom"))

	out, err := ec2.CreateDefaultSecurityGroup()

	if err == nil {
		t.Errorf("expected error, got none")
	}

	if expected := errors.New("could not create default security group (fargate-default): boom"); err.Error() != expected.Error() {
		t.Errorf("expected error %v, got %v", expected, err)
	}

	if out != "" {
		t.Errorf("expected no result, got %s", out)
	}
}

func TestAuthorizeAllSecurityGroupIngress(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	securityGroupID := "sg-abcdef"
	input := &awsec2.AuthorizeSecurityGroupIngressInput{
		CidrIp:     aws.String("0.0.0.0/0"),
		GroupId:    aws.String("sg-abcdef"),
		IpProtocol: aws.String("-1"),
	}

	mockEC2Client := sdk.NewMockEC2API(mockCtrl)
	ec2 := SDKClient{client: mockEC2Client}

	mockEC2Client.EXPECT().AuthorizeSecurityGroupIngress(input).Return(&awsec2.AuthorizeSecurityGroupIngressOutput{}, nil)

	err := ec2.AuthorizeAllSecurityGroupIngress(securityGroupID)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}
