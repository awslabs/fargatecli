package acm

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	awsacm "github.com/aws/aws-sdk-go/service/acm"
	"github.com/golang/mock/gomock"
	"github.com/jpignata/fargate/acm/mock/sdk"
)

func TestRequestCertificate(t *testing.T) {
	certificateArn := "arn:aws:acm:us-east-1:123456789012:certificate/12345678-1234-1234-1234-123456789012"
	domainName := "*.example.com"
	aliases := []string{"example-other.com"}
	validationMethod := "DNS"

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockACMAPI := sdk.NewMockACMAPI(mockCtrl)
	acm := SDKClient{client: mockACMAPI}

	i := &awsacm.RequestCertificateInput{
		DomainName:              aws.String(domainName),
		ValidationMethod:        aws.String(validationMethod),
		SubjectAlternativeNames: aws.StringSlice(aliases),
	}
	o := &awsacm.RequestCertificateOutput{
		CertificateArn: aws.String(certificateArn),
	}

	mockACMAPI.EXPECT().RequestCertificate(i).Return(o, nil)

	arn, err := acm.RequestCertificate(domainName, aliases)

	if err != nil {
		t.Error("Error; %+v", err)
	}

	if arn != certificateArn {
		t.Error("Invalid certificate ARN; want: %s, got: %s", certificateArn, arn)
	}
}

func TestRequestCertificateError(t *testing.T) {
	var aliases []string

	domainName := "*.example.com"
	validationMethod := "DNS"

	// Simulating a certificate request with more than 10 domains
	for i := 0; i < 10; i++ {
		aliases = append(aliases, fmt.Sprintf("example-%i.com", i))
	}

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockACMAPI := sdk.NewMockACMAPI(mockCtrl)
	acm := SDKClient{client: mockACMAPI}

	i := &awsacm.RequestCertificateInput{
		DomainName:              aws.String(domainName),
		ValidationMethod:        aws.String(validationMethod),
		SubjectAlternativeNames: aws.StringSlice(aliases),
	}
	o := &awsacm.RequestCertificateOutput{}

	mockACMAPI.EXPECT().RequestCertificate(i).Return(o, errors.New("Certificate has too many domains."))

	arn, err := acm.RequestCertificate(domainName, aliases)

	if err == nil {
		t.Errorf("No error; want: %+v", err)
	}

	if arn != "" {
		t.Errorf("Invalid certificate ARN; want: empty, got: %s", arn)
	}
}

func TestListCertificates(t *testing.T) {
	resp := &awsacm.ListCertificatesOutput{
		CertificateSummaryList: []*awsacm.CertificateSummary{
			&awsacm.CertificateSummary{
				DomainName:     aws.String("www.example.com"),
				CertificateArn: aws.String("arn:aws:acm:us-east-1:123456789012:certificate/12345678-1234-1234-1234-123456789012"),
			},
		},
	}

	mockClient := sdk.MockCertificateListPagesClient{Resp: resp}
	acm := SDKClient{client: mockClient}
	certificates, err := acm.ListCertificates()

	if err != nil {
		t.Errorf("Expected no error, got %+v", err)
	}

	if len(certificates) != 1 {
		t.Errorf("Expected 1 certificate, got %d", len(certificates))
	}

	if certificates[0].DomainName != "www.example.com" {
		t.Errorf("Expected certificate domain to be www.example.com, got %s", certificates[0].DomainName)
	}
}

func TestListCertificatesError(t *testing.T) {
	mockClient := sdk.MockCertificateListPagesClient{
		Resp:  &awsacm.ListCertificatesOutput{},
		Error: errors.New(":-("),
	}
	acm := SDKClient{client: mockClient}
	certificates, err := acm.ListCertificates()

	if err == nil {
		t.Errorf("Expected error, got nil")
	}

	if len(certificates) > 0 {
		t.Errorf("Expected no certificates, got %d", len(certificates))
	}
}

func TestDeleteCertificate(t *testing.T) {
	certificateArn := "arn:aws:acm:us-east-1:123456789012:certificate/12345678-1234-1234-1234-123456789012"

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockACMAPI := sdk.NewMockACMAPI(mockCtrl)

	acm := SDKClient{client: mockACMAPI}
	i := &awsacm.DeleteCertificateInput{CertificateArn: aws.String(certificateArn)}
	o := &awsacm.DeleteCertificateOutput{}

	mockACMAPI.EXPECT().DeleteCertificate(i).Return(o, nil)

	err := acm.DeleteCertificate(certificateArn)

	if err != nil {
		t.Error("Error; %+v", err)
	}
}

func TestDeleteCertificateError(t *testing.T) {
	certificateArn := "arn:aws:acm:us-east-1:123456789012:certificate/12345678-1234-1234-1234-123456789012"

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockACMAPI := sdk.NewMockACMAPI(mockCtrl)

	acm := SDKClient{client: mockACMAPI}
	i := &awsacm.DeleteCertificateInput{CertificateArn: aws.String(certificateArn)}
	o := &awsacm.DeleteCertificateOutput{}

	mockACMAPI.EXPECT().DeleteCertificate(i).Return(o, errors.New(":-("))

	err := acm.DeleteCertificate(certificateArn)

	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestInflateCertificate(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	var tests = []struct {
		DescribeCertificateOutput *awsacm.DescribeCertificateOutput
		InCertificate             Certificate
		OutCertificate            Certificate
	}{
		{
			&awsacm.DescribeCertificateOutput{
				Certificate: &awsacm.CertificateDetail{
					Type:                    aws.String("AMAZON_ISSUED"),
					Status:                  aws.String("PENDING_VALIDATION"),
					SubjectAlternativeNames: aws.StringSlice([]string{"staging.example.com"}),
					DomainValidationOptions: []*awsacm.DomainValidation{
						&awsacm.DomainValidation{
							ValidationStatus: aws.String("SUCCESS"),
							DomainName:       aws.String("staging.example.com"),
							ResourceRecord: &awsacm.ResourceRecord{
								Name:  aws.String("_beeed67ae3f2d83f6cd3e19a8064947b.staging.example.com"),
								Type:  aws.String("CNAME"),
								Value: aws.String("_6ddc33cd42c3fe3d5eca4cb075013a0a.acm-validations.aws."),
							},
						},
						&awsacm.DomainValidation{
							ValidationStatus: aws.String("PENDING_VALIDATION"),
							DomainName:       aws.String("www.example.com"),
							ResourceRecord: &awsacm.ResourceRecord{
								Name:  aws.String("_beeed67ae3f2d83f6cd3e19a8064947b.www.example.com"),
								Type:  aws.String("CNAME"),
								Value: aws.String("_6ddc33cd42c3fe3d5eca4cb075013a0a.acm-validations.aws."),
							},
						},
					},
				},
			},
			Certificate{
				DomainName: "www.example.com",
				Arn:        "arn:aws:acm:us-east-1:123456789012:certificate/12345678-1234-1234-1234-123456789012",
			},
			Certificate{
				DomainName:              "www.example.com",
				Arn:                     "arn:aws:acm:us-east-1:123456789012:certificate/12345678-1234-1234-1234-123456789012",
				Type:                    "AMAZON_ISSUED",
				Status:                  "PENDING_VALIDATION",
				SubjectAlternativeNames: []string{"staging.example.com"},
				Validations: []CertificateValidation{
					CertificateValidation{
						Status:     "SUCCESS",
						DomainName: "staging.example.com",
						ResourceRecord: CertificateResourceRecord{
							Name:  "_beeed67ae3f2d83f6cd3e19a8064947b.staging.example.com",
							Type:  "CNAME",
							Value: "_6ddc33cd42c3fe3d5eca4cb075013a0a.acm-validations.aws.",
						},
					},
					CertificateValidation{
						Status:     "PENDING_VALIDATION",
						DomainName: "www.example.com",
						ResourceRecord: CertificateResourceRecord{
							Name:  "_beeed67ae3f2d83f6cd3e19a8064947b.www.example.com",
							Type:  "CNAME",
							Value: "_6ddc33cd42c3fe3d5eca4cb075013a0a.acm-validations.aws.",
						},
					},
				},
			},
		},
		{
			&awsacm.DescribeCertificateOutput{
				Certificate: &awsacm.CertificateDetail{
					Type:   aws.String("AMAZON_ISSUED"),
					Status: aws.String("FAILED"),
					DomainValidationOptions: []*awsacm.DomainValidation{
						&awsacm.DomainValidation{
							ValidationStatus: aws.String("FAILED"),
							DomainName:       aws.String("staging.example.com"),
						},
					},
				},
			},
			Certificate{
				DomainName: "staging.example.com",
				Arn:        "arn:aws:acm:us-east-1:123456789012:certificate/12345678-1234-1234-1234-123456789012",
			},
			Certificate{
				DomainName:              "staging.example.com",
				Arn:                     "arn:aws:acm:us-east-1:123456789012:certificate/12345678-1234-1234-1234-123456789012",
				Type:                    "AMAZON_ISSUED",
				Status:                  "FAILED",
				SubjectAlternativeNames: []string{},
				Validations: []CertificateValidation{
					CertificateValidation{
						Status:     "FAILED",
						DomainName: "staging.example.com",
					},
				},
			},
		},
	}

	for _, test := range tests {
		mockACMAPI := sdk.NewMockACMAPI(mockCtrl)

		acm := SDKClient{client: mockACMAPI}
		i := &awsacm.DescribeCertificateInput{
			CertificateArn: aws.String(test.InCertificate.Arn),
		}

		mockACMAPI.EXPECT().DescribeCertificate(i).Return(test.DescribeCertificateOutput, nil)

		certificate, err := acm.InflateCertificate(test.InCertificate)

		if err != nil {
			t.Errorf("Expected no error, got %+v", err)
		}

		if certificate.DomainName != test.OutCertificate.DomainName {
			t.Errorf("Expected DomainName %s, got %s", test.OutCertificate.DomainName, certificate.DomainName)
		}

		if certificate.Arn != test.OutCertificate.Arn {
			t.Errorf("Expected Arn %s, got %s", test.OutCertificate.Arn, certificate.Arn)
		}

		if certificate.Type != test.OutCertificate.Type {
			t.Errorf("Expected Type %s, got %s", test.OutCertificate.Type, certificate.Type)
		}

		if certificate.Status != test.OutCertificate.Status {
			t.Errorf("Expected Status %s, got %s", test.OutCertificate.Status, certificate.Status)
		}

		if !reflect.DeepEqual(certificate.SubjectAlternativeNames, test.OutCertificate.SubjectAlternativeNames) {
			t.Errorf("Expected SubjectAlternativeNames %+v, got %+v", test.OutCertificate.SubjectAlternativeNames, certificate.SubjectAlternativeNames)
		}

		if len(certificate.Validations) != len(test.OutCertificate.Validations) {
			t.Errorf("Expected %d Validations, got %d", len(test.OutCertificate.Validations), len(certificate.Validations))
		}

		for i, v := range certificate.Validations {
			if v.Status != test.OutCertificate.Validations[i].Status {
				t.Errorf("Expected Validation Type %s, got %s", test.OutCertificate.Validations[i].Status, v.Status)
			}

			if v.DomainName != test.OutCertificate.Validations[i].DomainName {
				t.Errorf("Expected Validation DomainName %s, got %s", test.OutCertificate.Validations[i].DomainName, v.DomainName)
			}

			if v.ResourceRecordString() != test.OutCertificate.Validations[i].ResourceRecordString() {
				t.Errorf("Expected Validation ResourceRecord %s, got %s", test.OutCertificate.Validations[i].ResourceRecordString(), v.ResourceRecordString())
			}
		}
	}
}

func TestInflateCertificateError(t *testing.T) {
	inCertificate := Certificate{
		Arn:        "arn:aws:acm:us-east-1:123456789012:certificate/12345678-1234-1234-1234-123456789012",
		DomainName: "www.example.com",
	}

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockACMAPI := sdk.NewMockACMAPI(mockCtrl)

	acm := SDKClient{client: mockACMAPI}
	i := &awsacm.DescribeCertificateInput{CertificateArn: aws.String(inCertificate.Arn)}
	o := &awsacm.DescribeCertificateOutput{}

	mockACMAPI.EXPECT().DescribeCertificate(i).Return(o, errors.New(":-("))

	outCertificate, err := acm.InflateCertificate(inCertificate)

	if err == nil {
		t.Error("Expected error, got nil")
	}

	if !reflect.DeepEqual(outCertificate, inCertificate) {
		t.Errorf("Certificate modified, expected %+v, got %+v", inCertificate, outCertificate)
	}
}
