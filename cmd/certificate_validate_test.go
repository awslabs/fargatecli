package cmd

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/jpignata/fargate/acm"
	acmclient "github.com/jpignata/fargate/acm/mock/client"
	"github.com/jpignata/fargate/cmd/mock"
	"github.com/jpignata/fargate/route53"
	route53client "github.com/jpignata/fargate/route53/mock/client"
)

func TestCertificateValidateOperation(t *testing.T) {
	resourceRecordType := "CNAME"
	resourceRecordName := "_beeed67ae3f2d83f6cd3e19a8064947b.staging.example.com"
	resourceRecordValue := "_6ddc33cd42c3fe3d5eca4cb075013a0a.acm-validations.aws."

	hostedZones := route53.HostedZones{
		route53.HostedZone{
			Name: "example.com.",
			ID:   "Z2FDTNDATAQYW2",
		},
	}
	certificates := acm.Certificates{
		acm.Certificate{
			ARN:        "arn:aws:acm:us-east-1:123456789012:certificate/12345678-1234-1234-1234-123456789012",
			DomainName: "example.com",
		},
	}
	certificate := acm.Certificate{
		ARN:        "arn:aws:acm:us-east-1:123456789012:certificate/12345678-1234-1234-1234-123456789012",
		DomainName: "example.com",
		Status:     "PENDING_VALIDATION",
		Type:       "AMAZON_ISSUED",
		Validations: []acm.CertificateValidation{
			acm.CertificateValidation{
				Status:     "PENDING_VALIDATION",
				DomainName: "example.com",
				ResourceRecord: acm.CertificateResourceRecord{
					Name:  resourceRecordName,
					Type:  resourceRecordType,
					Value: resourceRecordValue,
				},
			},
		},
	}

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockRoute53Client := route53client.NewMockClient(mockCtrl)
	mockACMClient := acmclient.NewMockClient(mockCtrl)
	mockOutput := &mock.Output{}

	createResourceRecordInput := route53.CreateResourceRecordInput{
		HostedZoneID: hostedZones[0].ID,
		RecordType:   resourceRecordType,
		Name:         resourceRecordName,
		Value:        resourceRecordValue,
	}

	mockACMClient.EXPECT().ListCertificates().Return(certificates, nil)
	mockACMClient.EXPECT().InflateCertificate(certificates[0]).Return(certificate, nil)
	mockRoute53Client.EXPECT().ListHostedZones().Return(hostedZones, nil)
	mockRoute53Client.EXPECT().CreateResourceRecord(createResourceRecordInput).Return("/change/1", nil)

	certificateValidateOperation{
		certificateOperation: certificateOperation{
			acm: mockACMClient,
		},
		domainName: "example.com",
		output:     mockOutput,
		route53:    mockRoute53Client,
	}.execute()

	if len(mockOutput.InfoMsgs) == 0 {
		t.Error("Expected info output, got none")
	} else if mockOutput.InfoMsgs[0] != "[example.com] created validation record" {
		t.Errorf("Expected info output == '[example.com] created validation record', got: %s", mockOutput.InfoMsgs[0])
	}
}

func TestCertificateValidateOperationFindCertificateError(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockRoute53Client := route53client.NewMockClient(mockCtrl)
	mockACMClient := acmclient.NewMockClient(mockCtrl)
	mockOutput := &mock.Output{}

	mockACMClient.EXPECT().ListCertificates().Return(acm.Certificates{}, errors.New("boom"))

	certificateValidateOperation{
		certificateOperation: certificateOperation{
			acm: mockACMClient,
		},
		domainName: "example.com",
		output:     mockOutput,
		route53:    mockRoute53Client,
	}.execute()

	if len(mockOutput.FatalMsgs) == 0 {
		t.Error("Expected fatal output, got none")
	} else if mockOutput.FatalMsgs[0].Msg != "Could not validate certificate" {
		t.Errorf("Expected fatal output == 'Could not validate certificate', got: %+v", mockOutput.FatalMsgs[0])
	}
}

func TestCertificateValidateOperationListHostedZonesError(t *testing.T) {
	certificates := acm.Certificates{
		acm.Certificate{
			ARN:        "arn:aws:acm:us-east-1:123456789012:certificate/12345678-1234-1234-1234-123456789012",
			DomainName: "example.com",
			Status:     "PENDING_VALIDATION",
		},
	}

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockRoute53Client := route53client.NewMockClient(mockCtrl)
	mockACMClient := acmclient.NewMockClient(mockCtrl)
	mockOutput := &mock.Output{}

	mockACMClient.EXPECT().ListCertificates().Return(certificates, nil)
	mockACMClient.EXPECT().InflateCertificate(certificates[0]).Return(certificates[0], nil)
	mockRoute53Client.EXPECT().ListHostedZones().Return(route53.HostedZones{}, errors.New("boom"))

	certificateValidateOperation{
		certificateOperation: certificateOperation{
			acm: mockACMClient,
		},
		domainName: "example.com",
		output:     mockOutput,
		route53:    mockRoute53Client,
	}.execute()

	if len(mockOutput.FatalMsgs) == 0 {
		t.Error("Expected fatal output, got none")
	} else if mockOutput.FatalMsgs[0].Msg != "Could not validate certificate" {
		t.Errorf("Expected fatal output == 'Could not validate certificate', got: %+v", mockOutput.FatalMsgs[0])
	}
}

func TestCertificateValidateOperationInvalidState(t *testing.T) {
	certificates := acm.Certificates{
		acm.Certificate{
			ARN:        "arn:aws:acm:us-east-1:123456789012:certificate/12345678-1234-1234-1234-123456789012",
			DomainName: "example.com",
		},
	}
	certificate := acm.Certificate{
		ARN:         "arn:aws:acm:us-east-1:123456789012:certificate/12345678-1234-1234-1234-123456789012",
		DomainName:  "example.com",
		Status:      "FAILED",
		Type:        "AMAZON_ISSUED",
		Validations: []acm.CertificateValidation{},
	}

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockRoute53Client := route53client.NewMockClient(mockCtrl)
	mockACMClient := acmclient.NewMockClient(mockCtrl)
	mockOutput := &mock.Output{}

	mockACMClient.EXPECT().ListCertificates().Return(certificates, nil)
	mockACMClient.EXPECT().InflateCertificate(certificates[0]).Return(certificate, nil)
	certificateValidateOperation{
		certificateOperation: certificateOperation{
			acm: mockACMClient,
		},
		domainName: "example.com",
		output:     mockOutput,
		route53:    mockRoute53Client,
	}.execute()

	if len(mockOutput.FatalMsgs) == 0 {
		t.Error("Expected fatal output, got none")
		t.FailNow()
	}

	if mockOutput.FatalMsgs[0].Msg != "Could not validate certificate" {
		t.Errorf("Expected fatal output == 'Could not validate certificate', got: %s", mockOutput.FatalMsgs[0])
	}

	if mockOutput.FatalMsgs[0].Errors[0].Error() != "certificate example.com is in state failed" {
		t.Errorf("Expected error == 'certificate example.com is in state failed', got: %s", mockOutput.FatalMsgs[0].Errors[0])
	}
}

func TestCertificateValidateOperationZoneNotFound(t *testing.T) {
	certificates := acm.Certificates{
		acm.Certificate{
			ARN:        "arn:aws:acm:us-east-1:123456789012:certificate/12345678-1234-1234-1234-123456789012",
			DomainName: "example.com",
		},
	}
	certificate := acm.Certificate{
		ARN:        "arn:aws:acm:us-east-1:123456789012:certificate/12345678-1234-1234-1234-123456789012",
		DomainName: "example.com",
		Status:     "PENDING_VALIDATION",
		Type:       "AMAZON_ISSUED",
		Validations: []acm.CertificateValidation{
			acm.CertificateValidation{
				Status:     "PENDING_VALIDATION",
				DomainName: "example.com",
			},
		},
	}

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockRoute53Client := route53client.NewMockClient(mockCtrl)
	mockACMClient := acmclient.NewMockClient(mockCtrl)
	mockOutput := &mock.Output{}

	mockACMClient.EXPECT().ListCertificates().Return(certificates, nil)
	mockACMClient.EXPECT().InflateCertificate(certificates[0]).Return(certificate, nil)
	mockRoute53Client.EXPECT().ListHostedZones().Return(route53.HostedZones{}, nil)

	certificateValidateOperation{
		certificateOperation: certificateOperation{
			acm: mockACMClient,
		},
		domainName: "example.com",
		output:     mockOutput,
		route53:    mockRoute53Client,
	}.execute()

	if len(mockOutput.WarnMsgs) == 0 {
		t.Error("Expected warn output, got none")
	} else if mockOutput.WarnMsgs[0] != "[example.com] could not find zone in Amazon Route 53" {
		t.Errorf("Expected warn output == '[example.com] could not find zone in Amaozn Route 53', got: %s", mockOutput.WarnMsgs[0])
	}
}

func TestCertificateValidateOperationValidationSuccess(t *testing.T) {
	certificates := acm.Certificates{
		acm.Certificate{
			ARN:        "arn:aws:acm:us-east-1:123456789012:certificate/12345678-1234-1234-1234-123456789012",
			DomainName: "example.com",
		},
	}
	certificate := acm.Certificate{
		ARN:        "arn:aws:acm:us-east-1:123456789012:certificate/12345678-1234-1234-1234-123456789012",
		DomainName: "example.com",
		Status:     "PENDING_VALIDATION",
		Type:       "AMAZON_ISSUED",
		Validations: []acm.CertificateValidation{
			acm.CertificateValidation{
				Status:     "SUCCESS",
				DomainName: "example.com",
			},
		},
	}

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockRoute53Client := route53client.NewMockClient(mockCtrl)
	mockACMClient := acmclient.NewMockClient(mockCtrl)
	mockOutput := &mock.Output{}

	mockACMClient.EXPECT().ListCertificates().Return(certificates, nil)
	mockACMClient.EXPECT().InflateCertificate(certificates[0]).Return(certificate, nil)
	mockRoute53Client.EXPECT().ListHostedZones().Return(route53.HostedZones{}, nil)

	certificateValidateOperation{
		certificateOperation: certificateOperation{
			acm: mockACMClient,
		},
		domainName: "example.com",
		output:     mockOutput,
		route53:    mockRoute53Client,
	}.execute()

	if len(mockOutput.InfoMsgs) == 0 {
		t.Error("Expected info output, got none")
	} else if mockOutput.InfoMsgs[0] != "[example.com] already validated" {
		t.Errorf("Expected info output == '[example.com] already validated', got: %s", mockOutput.InfoMsgs[0])
	}
}

func TestCertificateValidateOperationValidationFailed(t *testing.T) {
	certificates := acm.Certificates{
		acm.Certificate{
			ARN:        "arn:aws:acm:us-east-1:123456789012:certificate/12345678-1234-1234-1234-123456789012",
			DomainName: "example.com",
		},
	}
	certificate := acm.Certificate{
		ARN:        "arn:aws:acm:us-east-1:123456789012:certificate/12345678-1234-1234-1234-123456789012",
		DomainName: "example.com",
		Status:     "PENDING_VALIDATION",
		Type:       "AMAZON_ISSUED",
		Validations: []acm.CertificateValidation{
			acm.CertificateValidation{
				Status:     "FAILED",
				DomainName: "example.com",
			},
		},
	}

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockRoute53Client := route53client.NewMockClient(mockCtrl)
	mockACMClient := acmclient.NewMockClient(mockCtrl)
	mockOutput := &mock.Output{}

	mockACMClient.EXPECT().ListCertificates().Return(certificates, nil)
	mockACMClient.EXPECT().InflateCertificate(certificates[0]).Return(certificate, nil)
	mockRoute53Client.EXPECT().ListHostedZones().Return(route53.HostedZones{}, nil)

	certificateValidateOperation{
		certificateOperation: certificateOperation{
			acm: mockACMClient,
		},
		domainName: "example.com",
		output:     mockOutput,
		route53:    mockRoute53Client,
	}.execute()

	if len(mockOutput.FatalMsgs) == 0 {
		t.Error("Expected fatal output, got none")
	} else if mockOutput.FatalMsgs[0].Msg != "[example.com] failed validation" {
		t.Errorf("Expected fatal output == '[example.com] failed validation', got: %s", mockOutput.FatalMsgs[0])
	}
}

func TestCertificateValidateOperationValidationUnknown(t *testing.T) {
	certificates := acm.Certificates{
		acm.Certificate{
			ARN:        "arn:aws:acm:us-east-1:123456789012:certificate/12345678-1234-1234-1234-123456789012",
			DomainName: "example.com",
		},
	}
	certificate := acm.Certificate{
		ARN:        "arn:aws:acm:us-east-1:123456789012:certificate/12345678-1234-1234-1234-123456789012",
		DomainName: "example.com",
		Status:     "PENDING_VALIDATION",
		Type:       "AMAZON_ISSUED",
		Validations: []acm.CertificateValidation{
			acm.CertificateValidation{
				Status:     "SOME_UNKNOWN_STATUS",
				DomainName: "example.com",
			},
		},
	}

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockRoute53Client := route53client.NewMockClient(mockCtrl)
	mockACMClient := acmclient.NewMockClient(mockCtrl)
	mockOutput := &mock.Output{}

	mockACMClient.EXPECT().ListCertificates().Return(certificates, nil)
	mockACMClient.EXPECT().InflateCertificate(certificates[0]).Return(certificate, nil)
	mockRoute53Client.EXPECT().ListHostedZones().Return(route53.HostedZones{}, nil)

	certificateValidateOperation{
		certificateOperation: certificateOperation{
			acm: mockACMClient,
		},
		domainName: "example.com",
		output:     mockOutput,
		route53:    mockRoute53Client,
	}.execute()

	if len(mockOutput.WarnMsgs) == 0 {
		t.Error("Expected warn output, got none")
	} else if mockOutput.WarnMsgs[0] != "[example.com] unexpected status: some unknown status" {
		t.Errorf("Expected warn output == '[example.com] unexpected status: some unknown status', got: %s", mockOutput.WarnMsgs[0])
	}
}

func TestCertificateValidateOperationRecordSetError(t *testing.T) {
	resourceRecordType := "CNAME"
	resourceRecordName := "_beeed67ae3f2d83f6cd3e19a8064947b.staging.example.com"
	resourceRecordValue := "_6ddc33cd42c3fe3d5eca4cb075013a0a.acm-validations.aws."

	hostedZones := route53.HostedZones{
		route53.HostedZone{
			Name: "example.com.",
			ID:   "Z2FDTNDATAQYW2",
		},
	}
	certificates := acm.Certificates{
		acm.Certificate{
			ARN:        "arn:aws:acm:us-east-1:123456789012:certificate/12345678-1234-1234-1234-123456789012",
			DomainName: "example.com",
		},
	}
	certificate := acm.Certificate{
		ARN:        "arn:aws:acm:us-east-1:123456789012:certificate/12345678-1234-1234-1234-123456789012",
		DomainName: "example.com",
		Status:     "PENDING_VALIDATION",
		Type:       "AMAZON_ISSUED",
		Validations: []acm.CertificateValidation{
			acm.CertificateValidation{
				Status:     "PENDING_VALIDATION",
				DomainName: "example.com",
				ResourceRecord: acm.CertificateResourceRecord{
					Name:  resourceRecordName,
					Type:  resourceRecordType,
					Value: resourceRecordValue,
				},
			},
		},
	}

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockRoute53Client := route53client.NewMockClient(mockCtrl)
	mockACMClient := acmclient.NewMockClient(mockCtrl)
	mockOutput := &mock.Output{}

	createResourceRecordInput := route53.CreateResourceRecordInput{
		HostedZoneID: hostedZones[0].ID,
		RecordType:   resourceRecordType,
		Name:         resourceRecordName,
		Value:        resourceRecordValue,
	}

	mockACMClient.EXPECT().ListCertificates().Return(certificates, nil)
	mockACMClient.EXPECT().InflateCertificate(certificates[0]).Return(certificate, nil)
	mockRoute53Client.EXPECT().ListHostedZones().Return(hostedZones, nil)
	mockRoute53Client.EXPECT().CreateResourceRecord(createResourceRecordInput).Return("", errors.New("boom"))

	certificateValidateOperation{
		certificateOperation: certificateOperation{
			acm: mockACMClient,
		},
		domainName: "example.com",
		output:     mockOutput,
		route53:    mockRoute53Client,
	}.execute()

	if len(mockOutput.FatalMsgs) == 0 {
		t.Error("Expected fatal output, got none")
	} else if mockOutput.FatalMsgs[0].Msg != "Could not validate certificate" {
		t.Errorf("Expected fatal output == 'Could not validate certificate', got: %+v", mockOutput.FatalMsgs[0])
	}
}
