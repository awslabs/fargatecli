package cmd

import (
	"errors"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/jpignata/fargate/acm"
	"github.com/jpignata/fargate/acm/mock/client"
	"github.com/jpignata/fargate/cmd/mock"
)

func TestCertificateInfoOperation(t *testing.T) {
	domainName := "example.com"
	certificateArn := "arn:aws:acm:us-east-1:123456789012:certificate/12345678-1234-1234-1234-123456789012"
	inCertificate := acm.Certificate{
		Arn:        certificateArn,
		DomainName: domainName,
	}
	outCertificate := acm.Certificate{
		Arn:                     certificateArn,
		DomainName:              domainName,
		Type:                    "AMAZON_ISSUED",
		Status:                  "PENDING_VALIDATION",
		SubjectAlternativeNames: []string{"staging1.example.com", "staging2.example.com"},
		Validations: []acm.CertificateValidation{
			acm.CertificateValidation{
				Status:     "SUCCESS",
				DomainName: "staging.example.com",
				ResourceRecord: acm.CertificateResourceRecord{
					Name:  "_beeed67ae3f2d83f6cd3e19a8064947b.staging.example.com",
					Type:  "CNAME",
					Value: "_6ddc33cd42c3fe3d5eca4cb075013a0a.acm-validations.aws.",
				},
			},
		},
	}
	certificateList := acm.Certificates{inCertificate}

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockClient := client.NewMockClient(mockCtrl)
	mockOutput := &mock.Output{}

	mockClient.EXPECT().ListCertificates().Return(certificateList, nil)
	mockClient.EXPECT().InflateCertificate(inCertificate).Return(outCertificate, nil)

	certificateInfoOperation{
		acm:        mockClient,
		domainName: domainName,
		output:     mockOutput,
	}.execute()

	if len(mockOutput.FatalMsgs) > 0 {
		for _, fatal := range mockOutput.FatalMsgs {
			t.Errorf(fatal.Msg, fatal.Errors)
		}
	}

	if len(mockOutput.KeyValueMsgs) == 0 {
		t.Errorf("Expected key value output from operation, got none")
	}

	if mockOutput.KeyValueMsgs["Domain Name"] != domainName {
		t.Errorf("Expected Domain Name == %s, got %s", domainName, mockOutput.KeyValueMsgs["Domain Name"])
	}

	if mockOutput.KeyValueMsgs["Status"] != "Pending Validation" {
		t.Errorf("Expected Status == Pending Validation, got %s", mockOutput.KeyValueMsgs["Status"])
	}

	if mockOutput.KeyValueMsgs["Type"] != "Amazon Issued" {
		t.Errorf("Expected Type == Amazon Issued, got %s", mockOutput.KeyValueMsgs["Type"])
	}

	if mockOutput.KeyValueMsgs["Subject Alternative Names"] != "staging1.example.com, staging2.example.com" {
		t.Errorf(
			"Expected Subject Alternative Names == staging1.example.com, staging2.example.com, got %s",
			mockOutput.KeyValueMsgs["Subject Alternative Names"],
		)
	}

	if len(mockOutput.Tables) != 1 {
		t.Errorf("Expected 1 table, got %d", len(mockOutput.Tables))
	}

	if len(mockOutput.Tables[0].Rows) != 2 {
		t.Errorf("Expected 2 rows , got %d", len(mockOutput.Tables[0].Rows))
	}

	if mockOutput.Tables[0].Header != "Validations" {
		t.Errorf("Expected table with header Validations , got %s", mockOutput.Tables[0].Header)
	}

	if !reflect.DeepEqual(mockOutput.Tables[0].Rows[0], []string{"DOMAIN NAME", "STATUS", "RECORD"}) {
		t.Errorf("Expected table with validation column names , got %+v", mockOutput.Tables[0].Rows[0])
	}

	if mockOutput.Tables[0].Rows[1][0] != "staging.example.com" {
		t.Errorf("Expected Validation Domain Name == staging.example.com, got %s", mockOutput.Tables[0].Rows[1][0])
	}

	if mockOutput.Tables[0].Rows[1][1] != "Success" {
		t.Errorf("Expected Validation Status == Success, got %s", mockOutput.Tables[0].Rows[1][1])
	}

	if mockOutput.Tables[0].Rows[1][2] != "CNAME _beeed67ae3f2d83f6cd3e19a8064947b.staging.example.com -> _6ddc33cd42c3fe3d5eca4cb075013a0a.acm-validations.aws." {
		t.Errorf("Expected Validation Record == CNAME _beeed67ae3f2d83f6cd3e19a8064947b.staging.example.com -> _6ddc33cd42c3fe3d5eca4cb075013a0a.acm-validations.aws., got %s", mockOutput.Tables[0].Rows[1][2])
	}
}

func TestCertificateInfoOperationNotFound(t *testing.T) {
	certificateList := acm.Certificates{}

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockClient := client.NewMockClient(mockCtrl)
	mockOutput := &mock.Output{}

	mockClient.EXPECT().ListCertificates().Return(certificateList, nil)

	certificateInfoOperation{
		acm:        mockClient,
		domainName: "example.com",
		output:     mockOutput,
	}.execute()

	if len(mockOutput.FatalMsgs) > 0 {
		for _, fatal := range mockOutput.FatalMsgs {
			t.Errorf(fatal.Msg, fatal.Errors)
		}
	}

	if mockOutput.InfoMsgs[0] != "No certificate found for example.com" {
		t.Errorf("Expected info 'No certificate found for example.com', got: %+v", mockOutput.InfoMsgs)
	}
}

func TestCertificateInfoOperationListError(t *testing.T) {
	certificateList := acm.Certificates{}

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockClient := client.NewMockClient(mockCtrl)
	mockOutput := &mock.Output{}

	mockClient.EXPECT().ListCertificates().Return(certificateList, errors.New("boom"))

	certificateInfoOperation{
		acm:        mockClient,
		domainName: "example.com",
		output:     mockOutput,
	}.execute()

	if len(mockOutput.FatalMsgs) == 0 {
		t.Errorf("Expected error output, received none")
	}

	if mockOutput.FatalMsgs[0].Msg != "Could not find certificate for example.com" {
		t.Errorf("Expected info 'Could not find certificate for example.com', got: %s", mockOutput.FatalMsgs[0].Msg)
	}

	if !mockOutput.Exited {
		t.Errorf("Expected premature exit; didn't")
	}
}

func TestCertificateInfoOperationDescribeError(t *testing.T) {
	certificate := acm.Certificate{
		Arn:        "arn:aws:acm:us-east-1:123456789012:certificate/12345678-1234-1234-1234-123456789012",
		DomainName: "example.com",
	}
	certificateList := acm.Certificates{certificate}

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockClient := client.NewMockClient(mockCtrl)
	mockOutput := &mock.Output{}

	mockClient.EXPECT().ListCertificates().Return(certificateList, nil)
	mockClient.EXPECT().InflateCertificate(certificate).Return(certificate, errors.New("boom"))

	certificateInfoOperation{
		acm:        mockClient,
		domainName: "example.com",
		output:     mockOutput,
	}.execute()

	if len(mockOutput.FatalMsgs) == 0 {
		t.Errorf("Expected error output, received none")
	}

	if mockOutput.FatalMsgs[0].Msg != "Could not find certificate for example.com" {
		t.Errorf("Expected info 'Could not find certificate for example.com', got: %s", mockOutput.FatalMsgs[0].Msg)
	}

	if !mockOutput.Exited {
		t.Errorf("Expected premature exit; didn't")
	}
}
