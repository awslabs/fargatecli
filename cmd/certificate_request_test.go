package cmd

import (
	"fmt"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/jpignata/fargate/acm/mock/client"
	"github.com/jpignata/fargate/cmd/mock"
)

func TestCertificateRequestOperation(t *testing.T) {
	certificateArn := "arn:aws:acm:us-east-1:123456789012:certificate/12345678-1234-1234-1234-123456789012"
	domainName := "example.com"
	aliases := []string{"www.example.com"}

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockClient := client.NewMockClient(mockCtrl)
	mockOutput := &mock.Output{}

	operation := certificateRequestOperation{
		acm:        mockClient,
		aliases:    aliases,
		domainName: domainName,
		output:     mockOutput,
	}

	mockClient.EXPECT().RequestCertificate(domainName, aliases).Return(certificateArn, nil)

	operation.execute()

	if len(mockOutput.InfoMsgs) == 0 {
		t.Errorf("Expected info output from operation, got none")
	}
}

func TestCertificateRequestOperationError(t *testing.T) {
	domainName := "example.com"
	aliases := []string{}

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockClient := client.NewMockClient(mockCtrl)
	mockOutput := &mock.Output{}

	operation := certificateRequestOperation{
		acm:        mockClient,
		aliases:    aliases,
		domainName: domainName,
		output:     mockOutput,
	}

	mockClient.EXPECT().RequestCertificate(domainName, aliases).Return("", fmt.Errorf("oops, something went wrong"))

	operation.execute()

	if !mockOutput.Exited {
		t.Errorf("Expected premature exit; didn't")
	}

	if len(mockOutput.FatalMsgs) == 0 {
		t.Errorf("Expected error output from operation, got none")
	}
}

func TestCertificateRequestOperationInvalid(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockClient := client.NewMockClient(mockCtrl)
	mockOutput := &mock.Output{}

	operation := certificateRequestOperation{
		acm:        mockClient,
		domainName: "z", // Invalid
		output:     mockOutput,
	}

	operation.execute()

	if !mockOutput.Exited {
		t.Errorf("Expected premature exit; didn't")
	}
}

func TestCertificateRequestOperationValidateInvalidDomainName(t *testing.T) {
	operation := certificateRequestOperation{
		domainName: "z", // Invalid
	}

	errs := operation.validate()

	if len(errs) != 1 {
		t.Errorf("Invalid number of errors; want 1, got: %d", len(errs))
	}

	if strings.Index(errs[0].Error(), "The domain name requires at least 2 octets") == -1 {
		t.Errorf("Unexpected error; want: 'The domain name requires at leasr 2 octets', got: %s", errs[0].Error())
	}
}

func TestCertificateRequestOperationValidateInvalidAlias(t *testing.T) {
	operation := certificateRequestOperation{
		domainName: "example.com",
		aliases:    []string{"z"}, // Invalid
	}

	errs := operation.validate()

	if len(errs) != 1 {
		t.Errorf("Invalid number of errors; want 1, got: %d", len(errs))
	}

	if strings.Index(errs[0].Error(), "An alias requires at least 2 octets") == -1 {
		t.Errorf("Unexpected error; want: 'An alias requires at least 2 octets', got: %s", errs[0].Error())
	}
}
