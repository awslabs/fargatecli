package cmd

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/jpignata/fargate/acm"
	"github.com/jpignata/fargate/acm/mock/client"
	"github.com/jpignata/fargate/cmd/mock"
)

func TestCertificateDestroyOperation(t *testing.T) {
	certificateARN := "arn:aws:acm:us-east-1:123456789012:certificate/12345678-1234-1234-1234-123456789012"
	domainName := "example.com"
	certificate := acm.Certificate{
		ARN:        certificateARN,
		DomainName: domainName,
	}

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockClient := client.NewMockClient(mockCtrl)
	mockOutput := &mock.Output{}

	mockClient.EXPECT().ListCertificates().Return(acm.Certificates{certificate}, nil)
	mockClient.EXPECT().InflateCertificate(certificate).Return(certificate, nil)
	mockClient.EXPECT().DeleteCertificate(certificateARN).Return(nil)

	certificateDestroyOperation{
		certificateOperation: certificateOperation{
			acm: mockClient,
		},
		domainName: domainName,
		output:     mockOutput,
	}.execute()

	if len(mockOutput.FatalMsgs) > 0 {
		for _, fatal := range mockOutput.FatalMsgs {
			t.Errorf(fatal.Msg, fatal.Errors)
		}
	}

	if len(mockOutput.InfoMsgs) == 0 {
		t.Errorf("Expected info output from operation, got none")
	}
}

func TestCertificateDestroyOperationCertNotFound(t *testing.T) {
	domainName := "example.com"

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockClient := client.NewMockClient(mockCtrl)
	mockOutput := &mock.Output{}

	mockClient.EXPECT().ListCertificates().Return(acm.Certificates{}, nil)

	certificateDestroyOperation{
		certificateOperation: certificateOperation{
			acm: mockClient,
		},
		domainName: domainName,
		output:     mockOutput,
	}.execute()

	if len(mockOutput.FatalMsgs) == 0 {
		t.Errorf("Expected fatal output from operation, got none")
	}

	if !mockOutput.Exited {
		t.Errorf("Expected premature exit; didn't")
	}
}

func TestCertificateDestroyOperationMoreThanOneCertFound(t *testing.T) {
	certificateARN1 := "arn:aws:acm:us-east-1:123456789012:certificate/12345678-1234-1234-1234-123456789012"
	certificateARN2 := "arn:aws:acm:us-east-1:123456789012:certificate/abcdef01-2345-6789-0abc-def012345678"
	domainName := "example.com"
	certificate1 := acm.Certificate{
		ARN:        certificateARN1,
		DomainName: domainName,
	}
	certificate2 := acm.Certificate{
		ARN:        certificateARN2,
		DomainName: domainName,
	}

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockClient := client.NewMockClient(mockCtrl)
	mockOutput := &mock.Output{}

	mockClient.EXPECT().ListCertificates().Return(acm.Certificates{certificate1, certificate2}, nil)

	certificateDestroyOperation{
		certificateOperation: certificateOperation{
			acm: mockClient,
		},
		domainName: domainName,
		output:     mockOutput,
	}.execute()

	if len(mockOutput.FatalMsgs) == 0 {
		t.Errorf("Expected fatal output from operation, got none")
	}

	if !mockOutput.Exited {
		t.Errorf("Expected premature exit; didn't")
	}
}

func TestCertificateDestroyOperationListError(t *testing.T) {
	domainName := "example.com"

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockClient := client.NewMockClient(mockCtrl)
	mockOutput := &mock.Output{}

	mockClient.EXPECT().ListCertificates().Return(acm.Certificates{}, errors.New("something went boom"))

	certificateDestroyOperation{
		certificateOperation: certificateOperation{
			acm: mockClient,
		},
		domainName: domainName,
		output:     mockOutput,
	}.execute()

	if len(mockOutput.FatalMsgs) == 0 {
		t.Errorf("Expected fatal output from operation, got none")
	}

	if !mockOutput.Exited {
		t.Errorf("Expected premature exit; didn't")
	}
}

func TestCertificateDestroyOperationDeleteError(t *testing.T) {
	certificateARN := "arn:aws:acm:us-east-1:123456789012:certificate/12345678-1234-1234-1234-123456789012"
	domainName := "example.com"
	certificate := acm.Certificate{
		ARN:        certificateARN,
		DomainName: domainName,
	}

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockClient := client.NewMockClient(mockCtrl)
	mockOutput := &mock.Output{}

	mockClient.EXPECT().ListCertificates().Return(acm.Certificates{certificate}, nil)
	mockClient.EXPECT().InflateCertificate(certificate).Return(certificate, nil)
	mockClient.EXPECT().DeleteCertificate(certificateARN).Return(errors.New(":-("))

	certificateDestroyOperation{
		certificateOperation: certificateOperation{
			acm: mockClient,
		},
		domainName: domainName,
		output:     mockOutput,
	}.execute()

	if len(mockOutput.FatalMsgs) == 0 {
		t.Errorf("Expected fatal output from operation, got none")
	}

	if !mockOutput.Exited {
		t.Errorf("Expected premature exit; didn't")
	}
}
