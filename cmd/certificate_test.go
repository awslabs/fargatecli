package cmd

import (
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/jpignata/fargate/acm"
	"github.com/jpignata/fargate/acm/mock/client"
	"github.com/jpignata/fargate/cmd/mock"
)

func TestFindCertificate(t *testing.T) {
	certificate := acm.Certificate{
		DomainName: "www.example.com",
		ARN:        "arn:aws:acm:us-east-1:123456789012:certificate/12345678-1234-1234-1234-123456789012",
		Status:     "ISSUED",
		Type:       "AMAZON_ISSUED",
	}

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockClient := client.NewMockClient(mockCtrl)
	mockOutput := &mock.Output{}

	mockClient.EXPECT().ListCertificates().Return(acm.Certificates{certificate}, nil)
	mockClient.EXPECT().InflateCertificate(&certificate).Return(nil)

	operation := certificateOperation{
		acm:    mockClient,
		output: mockOutput,
	}
	foundCertificate, err := operation.findCertificate("www.example.com")

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if !reflect.DeepEqual(foundCertificate, certificate) {
		t.Errorf("Expected to find %+v, got: %v", certificate, foundCertificate)
	}
}

func TestFindCertificateNotFound(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockClient := client.NewMockClient(mockCtrl)
	mockOutput := &mock.Output{}

	mockClient.EXPECT().ListCertificates().Return(acm.Certificates{}, nil)

	operation := certificateOperation{
		acm:    mockClient,
		output: mockOutput,
	}
	foundCertificate, err := operation.findCertificate("www.example.com")

	if err != errCertificateNotFound {
		t.Errorf("Expected errCertificateNotFound, got %v", err)
	}

	if !reflect.DeepEqual(foundCertificate, acm.Certificate{}) {
		t.Errorf("Expected empty Certificate, got: %v", foundCertificate)
	}
}

func TestFindCertificateTooManyFound(t *testing.T) {
	certificates := acm.Certificates{
		acm.Certificate{DomainName: "www.example.com", ARN: "arn:1"},
		acm.Certificate{DomainName: "www.example.com", ARN: "arn:2"},
	}

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockClient := client.NewMockClient(mockCtrl)
	mockOutput := &mock.Output{}

	mockClient.EXPECT().ListCertificates().Return(certificates, nil)

	operation := certificateOperation{
		acm:    mockClient,
		output: mockOutput,
	}
	foundCertificate, err := operation.findCertificate("www.example.com")

	if err != errCertificateTooManyFound {
		t.Errorf("Expected errCertificateTooManyFound, got %v", err)
	}

	if !reflect.DeepEqual(foundCertificate, acm.Certificate{}) {
		t.Errorf("Expected empty Certificate, got: %v", foundCertificate)
	}
}
