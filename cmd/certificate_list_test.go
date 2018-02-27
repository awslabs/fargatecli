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

func TestCertificateListOperation(t *testing.T) {
	certificateList := acm.Certificates{
		acm.Certificate{
			ARN:                     "arn:aws:acm:us-east-1:123456789012:certificate/12345678-1234-1234-1234-123456789012",
			DomainName:              "example.com",
			Type:                    "AMAZON_ISSUED",
			Status:                  "PENDING_VALIDATION",
			SubjectAlternativeNames: []string{"staging1.example.com", "staging2.example.com"},
		},
	}

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockClient := client.NewMockClient(mockCtrl)
	mockOutput := &mock.Output{}

	mockClient.EXPECT().ListCertificates().Return(certificateList, nil)
	mockClient.EXPECT().InflateCertificate(&certificateList[0]).Return(nil)

	certificateListOperation{
		acm:    mockClient,
		output: mockOutput,
	}.execute()

	if len(mockOutput.Tables) != 1 {
		t.Errorf("Expected table, got none")
	}

	if len(mockOutput.Tables[0].Rows) != 2 {
		t.Errorf("Expected table with 2 rows, got %d", len(mockOutput.Tables[0].Rows))
	}

	if !reflect.DeepEqual(
		mockOutput.Tables[0].Rows[0],
		[]string{"CERTIFICATE", "TYPE", "STATUS", "SUBJECT ALTERNATIVE NAMES"},
	) {
		t.Errorf("Expected column headers, found %+v", mockOutput.Tables[0].Rows[0])
	}

	if mockOutput.Tables[0].Rows[1][0] != "example.com" {
		t.Errorf("Expected Domain Name == example.com, found %s", mockOutput.Tables[0].Rows[1][0])
	}

	if mockOutput.Tables[0].Rows[1][1] != "Amazon Issued" {
		t.Errorf("Expected Type == Amazon Issued, found %s", mockOutput.Tables[0].Rows[1][1])
	}

	if mockOutput.Tables[0].Rows[1][2] != "Pending Validation" {
		t.Errorf("Expected Status == Pending Validation, found %s", mockOutput.Tables[0].Rows[1][2])
	}

	if mockOutput.Tables[0].Rows[1][3] != "staging1.example.com, staging2.example.com" {
		t.Errorf("Expected Subject Alternative Names == staging1.example.com, staging2.example.com, found %s", mockOutput.Tables[0].Rows[1][3])
	}
}

func TestCertificateListOperationOrdered(t *testing.T) {
	certificateList := acm.Certificates{
		acm.Certificate{
			ARN:        "arn:aws:acm:us-east-1:123456789012:certificate/12345678-1234-1234-1234-123456789012-f",
			DomainName: "f.com",
		},
		acm.Certificate{
			ARN:        "arn:aws:acm:us-east-1:123456789012:certificate/12345678-1234-1234-1234-123456789012-c",
			DomainName: "c.com",
		},
		acm.Certificate{
			ARN:        "arn:aws:acm:us-east-1:123456789012:certificate/12345678-1234-1234-1234-123456789012-a",
			DomainName: "a.com",
		},
		acm.Certificate{
			ARN:        "arn:aws:acm:us-east-1:123456789012:certificate/12345678-1234-1234-1234-123456789012-d",
			DomainName: "d.com",
		},
		acm.Certificate{
			ARN:        "arn:aws:acm:us-east-1:123456789012:certificate/12345678-1234-1234-1234-123456789012-e",
			DomainName: "e.com",
		},
		acm.Certificate{
			ARN:        "arn:aws:acm:us-east-1:123456789012:certificate/12345678-1234-1234-1234-123456789012-b",
			DomainName: "b.com",
		},
	}

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockClient := client.NewMockClient(mockCtrl)
	mockOutput := &mock.Output{}

	mockClient.EXPECT().ListCertificates().Return(certificateList, nil)
	mockClient.EXPECT().InflateCertificate(gomock.Any()).Return(nil).Times(len(certificateList))

	certificateListOperation{
		acm:    mockClient,
		output: mockOutput,
	}.execute()

	for i, r := range mockOutput.Tables[0].Rows {
		// Skip header and first row
		if i <= 1 {
			continue
		}

		// Compare domain to the domain of the previous row, check it is lexicographically subsequent
		if mockOutput.Tables[0].Rows[i-1][0] > r[0] {
			t.Errorf("Expected alphabetical order, got %+v", mockOutput.Tables[0].Rows[1:len(mockOutput.Tables[0].Rows)+1])
			break
		}
	}
}

func TestCertificateListOperationNotFound(t *testing.T) {
	certificateList := acm.Certificates{}

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockClient := client.NewMockClient(mockCtrl)
	mockOutput := &mock.Output{}

	mockClient.EXPECT().ListCertificates().Return(certificateList, nil)

	certificateListOperation{
		acm:    mockClient,
		output: mockOutput,
	}.execute()

	if len(mockOutput.FatalMsgs) > 0 {
		for _, fatal := range mockOutput.FatalMsgs {
			t.Errorf(fatal.Msg, fatal.Errors)
		}
	}

	if mockOutput.InfoMsgs[0] != "No certificates found" {
		t.Errorf("Expected info 'No certificate found', got: %+v", mockOutput.InfoMsgs)
	}
}

func TestCertificateListOperationListError(t *testing.T) {
	certificateList := acm.Certificates{}

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockClient := client.NewMockClient(mockCtrl)
	mockOutput := &mock.Output{}

	mockClient.EXPECT().ListCertificates().Return(certificateList, errors.New("boom"))

	certificateListOperation{
		acm:    mockClient,
		output: mockOutput,
	}.execute()

	if len(mockOutput.FatalMsgs) == 0 {
		t.Errorf("Expected error output, received none")
	}

	if mockOutput.FatalMsgs[0].Msg != "Could not list certificates" {
		t.Errorf("Expected info 'Could not list certificates', got: %s", mockOutput.FatalMsgs[0].Msg)
	}

	if !mockOutput.Exited {
		t.Errorf("Expected premature exit; didn't")
	}
}

func TestCertificateListOperationDescribeError(t *testing.T) {
	certificate := acm.Certificate{
		ARN:        "arn:aws:acm:us-east-1:123456789012:certificate/12345678-1234-1234-1234-123456789012",
		DomainName: "example.com",
	}
	certificateList := acm.Certificates{certificate}

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockClient := client.NewMockClient(mockCtrl)
	mockOutput := &mock.Output{}

	mockClient.EXPECT().ListCertificates().Return(certificateList, nil)
	mockClient.EXPECT().InflateCertificate(&certificate).Return(errors.New("boom"))

	certificateListOperation{
		acm:    mockClient,
		output: mockOutput,
	}.execute()

	if len(mockOutput.FatalMsgs) == 0 {
		t.Errorf("Expected error output, received none")
	}

	if mockOutput.FatalMsgs[0].Msg != "Could not list certificates" {
		t.Errorf("Expected info 'Could not list certificates', got: %s", mockOutput.FatalMsgs[0].Msg)
	}

	if !mockOutput.Exited {
		t.Errorf("Expected premature exit; didn't")
	}
}
