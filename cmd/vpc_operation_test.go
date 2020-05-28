package cmd

import (
	"errors"
	"testing"

	"github.com/awslabs/fargatecli/cmd/mock"
	ec2client "github.com/awslabs/fargatecli/ec2/mock/client"
	"github.com/golang/mock/gomock"
)

func TestSetSubnetIDs(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockEC2Client := ec2client.NewMockClient(mockCtrl)
	mockOutput := &mock.Output{}

	mockEC2Client.EXPECT().GetSubnetVPCID("subnet-1234567").Return("vpc-1234567", nil)

	operation := vpcOperation{
		ec2:    mockEC2Client,
		output: mockOutput,
	}

	err := operation.setSubnetIDs([]string{"subnet-1234567"})

	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	if len(operation.subnetIDs) != 1 {
		t.Fatalf("expected 1 subnet ID, got: %d", len(operation.subnetIDs))
	}

	if expected := "subnet-1234567"; operation.subnetIDs[0] != expected {
		t.Errorf("expected: %s, got: %s", expected, operation.subnetIDs[0])
	}

	if expected := "vpc-1234567"; operation.vpcID != expected {
		t.Errorf("expected: %s, got: %s", expected, operation.vpcID)
	}
}

func TestSetSubnetIDsError(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockEC2Client := ec2client.NewMockClient(mockCtrl)
	mockOutput := &mock.Output{}

	mockEC2Client.EXPECT().GetSubnetVPCID("subnet-1234567").Return("", errors.New("boom"))

	operation := vpcOperation{
		ec2:    mockEC2Client,
		output: mockOutput,
	}

	err := operation.setSubnetIDs([]string{"subnet-1234567"})

	if err == nil {
		t.Errorf("expected error, got none")
	}

	if expected := "boom"; err.Error() != expected {
		t.Errorf("expected: %s, got: %v", expected, err)
	}
}

func TestSetDefaultSecurityGroupID(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockEC2Client := ec2client.NewMockClient(mockCtrl)
	mockOutput := &mock.Output{}

	mockEC2Client.EXPECT().SetDefaultSecurityGroupID().Return("sg-1234567", nil) //SGCreate fallback is tested in vpc_test.go

	operation := vpcOperation{
		ec2:    mockEC2Client,
		output: mockOutput,
	}

	err := operation.setDefaultSecurityGroupID()

	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	if len(operation.securityGroupIDs) != 1 {
		t.Fatalf("expected 1 security group ID, got: %d", len(operation.securityGroupIDs))
	}

	if expected := "sg-1234567"; operation.securityGroupIDs[0] != expected {
		t.Errorf("expected: %s, got: %s", expected, operation.securityGroupIDs[0])
	}
}

func TestSetDefaultSecurityGroupIDLookupError(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockEC2Client := ec2client.NewMockClient(mockCtrl)
	mockOutput := &mock.Output{}

	mockEC2Client.EXPECT().SetDefaultSecurityGroupID().Return("", errors.New("boom"))

	operation := vpcOperation{
		ec2:    mockEC2Client,
		output: mockOutput,
	}

	err := operation.setDefaultSecurityGroupID()

	if err == nil {
		t.Errorf("expected error, got none")
	}

	if expected := "boom"; err.Error() != expected {
		t.Errorf("expected: %s, got: %v", expected, err)
	}
}

func TestSetDefaultSecurityGroupIDWithCreateError(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockEC2Client := ec2client.NewMockClient(mockCtrl)
	mockOutput := &mock.Output{}

	mockEC2Client.EXPECT().SetDefaultSecurityGroupID().Return("", errors.New("boom"))

	operation := vpcOperation{
		ec2:    mockEC2Client,
		output: mockOutput,
	}

	err := operation.setDefaultSecurityGroupID()

	if err == nil {
		t.Errorf("expected error, got none")
	}

	if expected := "boom"; err.Error() != expected {
		t.Errorf("expected: %s, got: %v", expected, err)
	}
}

func TestSetDefaultSecurityGroupIDWithAuthorizeError(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockEC2Client := ec2client.NewMockClient(mockCtrl)
	mockOutput := &mock.Output{}

	mockEC2Client.EXPECT().SetDefaultSecurityGroupID().Return("sg-1234567", errors.New("boom"))

	operation := vpcOperation{
		ec2:    mockEC2Client,
		output: mockOutput,
	}

	err := operation.setDefaultSecurityGroupID()

	if err == nil {
		t.Errorf("expected error, got none")
	}

	if expected := "boom"; err.Error() != expected {
		t.Errorf("expected: %s, got: %v", expected, err)
	}
}

func TestSetDefaultSubnetIDs(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockEC2Client := ec2client.NewMockClient(mockCtrl)
	mockOutput := &mock.Output{}

	mockEC2Client.EXPECT().GetDefaultSubnetIDs().Return([]string{"subnet-1234567", "subnet-abcdef"}, nil)
	mockEC2Client.EXPECT().GetSubnetVPCID("subnet-1234567").Return("vpc-1234567", nil)

	operation := vpcOperation{
		ec2:    mockEC2Client,
		output: mockOutput,
	}

	err := operation.setDefaultSubnetIDs()

	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	if len(operation.subnetIDs) != 2 {
		t.Fatalf("expected 2 subnet IDs, got: %d", len(operation.subnetIDs))
	}

	if expected := "subnet-1234567"; operation.subnetIDs[0] != expected {
		t.Errorf("expected: %s, got: %s", expected, operation.subnetIDs[0])
	}

	if expected := "vpc-1234567"; operation.vpcID != expected {
		t.Errorf("expected: %s, got: %s", expected, operation.vpcID)
	}
}

func TestSetDefaultSubnetIDsLookupError(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockEC2Client := ec2client.NewMockClient(mockCtrl)
	mockOutput := &mock.Output{}

	mockEC2Client.EXPECT().GetDefaultSubnetIDs().Return([]string{}, errors.New("boom"))

	operation := vpcOperation{
		ec2:    mockEC2Client,
		output: mockOutput,
	}

	err := operation.setDefaultSubnetIDs()

	if err == nil {
		t.Errorf("expected error, got none")
	}

	if expected := "boom"; err.Error() != expected {
		t.Errorf("expected: %s, got: %v", expected, err)
	}
}

func TestSetDefaultSubnetIDsVPCError(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockEC2Client := ec2client.NewMockClient(mockCtrl)
	mockOutput := &mock.Output{}

	mockEC2Client.EXPECT().GetDefaultSubnetIDs().Return([]string{"subnet-1234567", "subnet-abcdef"}, nil)
	mockEC2Client.EXPECT().GetSubnetVPCID("subnet-1234567").Return("", errors.New("boom"))

	operation := vpcOperation{
		ec2:    mockEC2Client,
		output: mockOutput,
	}

	err := operation.setDefaultSubnetIDs()

	if err == nil {
		t.Errorf("expected error, got none")
	}

	if expected := "boom"; err.Error() != expected {
		t.Errorf("expected: %s, got: %v", expected, err)
	}
}
