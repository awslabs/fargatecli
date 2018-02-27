package acm

import (
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	awsacm "github.com/aws/aws-sdk-go/service/acm"
	"github.com/golang/mock/gomock"
	"github.com/jpignata/fargate/acm/mock/sdk"
)

func TestValidateAlias(t *testing.T) {
	var tests = []struct {
		in  string
		out error
	}{
		{"valid.example.com", nil},
		{"invalid", errors.New("An alias requires at least 2 octets")},
		{strings.Repeat(".", 253), errors.New("An alias cannot exceed 253 octets")},
		{strings.Repeat("a", 255), errors.New("An alias must be between 1 and 253 characters in length")},
		{"", errors.New("An alias must be between 1 and 253 characters in length")},
	}

	for _, test := range tests {
		err := ValidateAlias(test.in)

		switch {
		case test.out == nil:
			if err != nil {
				t.Errorf("Expected nil, got %s", err)
			}
		default:
			if !strings.Contains(err.Error(), test.out.Error()) {
				t.Errorf("Expected contains %s, got %s", test.out.Error(), err.Error())
			}
		}
	}
}

func TestValidateDomainName(t *testing.T) {
	var tests = []struct {
		in  string
		out error
	}{
		{"valid.example.com", nil},
		{"invalid", errors.New("The domain name requires at least 2 octets")},
		{strings.Repeat(".", 63), errors.New("The domain name cannot exceed 63 octets")},
		{strings.Repeat("a", 255), errors.New("The domain name must be between 1 and 253 characters in length")},
		{"", errors.New("The domain name must be between 1 and 253 characters in length")},
	}

	for _, test := range tests {
		err := ValidateDomainName(test.in)

		switch {
		case test.out == nil:
			if err != nil {
				t.Errorf("Expected nil, got %s", err)
			}
		default:
			if !strings.Contains(err.Error(), test.out.Error()) {
				t.Errorf("Expected contains %s, got %s", test.out.Error(), err.Error())
			}
		}
	}
}

func TestCertificateIsPendingValidation(t *testing.T) {
	var tests = []struct {
		in  Certificate
		out bool
	}{
		{Certificate{Status: "PENDING_VALIDATION"}, true},
		{Certificate{Status: "ISSUED"}, false},
		{Certificate{Status: ""}, false},
	}

	for _, test := range tests {
		if test.in.IsPendingValidation() != test.out {
			t.Errorf("Expected %s to be IsPendingValidation", test.in.Status)
		}
	}
}

func TestCertificateIsIssued(t *testing.T) {
	var tests = []struct {
		in  Certificate
		out bool
	}{
		{Certificate{Status: "ISSUED"}, true},
		{Certificate{Status: "PENDING_VALIDATION"}, false},
		{Certificate{Status: ""}, false},
	}

	for _, test := range tests {
		if test.in.IsIssued() != test.out {
			t.Errorf("Expected %s to be IsIssued", test.in.Status)
		}
	}
}

func TestCertificateValidationIsPendingValidation(t *testing.T) {
	var tests = []struct {
		in  CertificateValidation
		out bool
	}{
		{CertificateValidation{Status: "PENDING_VALIDATION"}, true},
		{CertificateValidation{Status: "SUCCESS"}, false},
		{CertificateValidation{Status: ""}, false},
	}

	for _, test := range tests {
		if test.in.IsPendingValidation() != test.out {
			t.Errorf("Expected %s to be IsPendingValidation", test.in.Status)
		}
	}
}

func TestCertificateValidationIsSuccess(t *testing.T) {
	var tests = []struct {
		in  CertificateValidation
		out bool
	}{
		{CertificateValidation{Status: "SUCCESS"}, true},
		{CertificateValidation{Status: "PENDING_VALIDATION"}, false},
		{CertificateValidation{Status: ""}, false},
	}

	for _, test := range tests {
		if test.in.IsSuccess() != test.out {
			t.Errorf("Expected %s to be IsSuccess", test.in.Status)
		}
	}
}

func TestCertificateValidationIsFailed(t *testing.T) {
	var tests = []struct {
		in  CertificateValidation
		out bool
	}{
		{CertificateValidation{Status: "FAILED"}, true},
		{CertificateValidation{Status: "PENDING_VALIDATION"}, false},
		{CertificateValidation{Status: ""}, false},
	}

	for _, test := range tests {
		if test.in.IsFailed() != test.out {
			t.Errorf("Expected %s to be IsFailed", test.in.Status)
		}
	}
}

func TestCertificateValidationResourceRecordString(t *testing.T) {
	var tests = []struct {
		in  CertificateValidation
		out string
	}{
		{CertificateValidation{}, ""},
		{CertificateValidation{
			ResourceRecord: CertificateResourceRecord{
				Type:  "CNAME",
				Name:  "name",
				Value: "value",
			},
		}, "CNAME name -> value"},
	}

	for _, test := range tests {
		if test.in.ResourceRecordString() != test.out {
			t.Errorf("Expected ResourceRecordString() == CNAME name -> value, got: %s", test.in.ResourceRecordString())
		}
	}
}

func TestCertificatesGetCertificates(t *testing.T) {
	certificates := Certificates{
		Certificate{DomainName: "staging.example.com", ARN: "staging.example.com-1"},
		Certificate{DomainName: "www.example.com", ARN: "www.example.com-1"},
		Certificate{DomainName: "www.example.com", ARN: "www.example.com-2"},
	}

	var empty Certificates
	var tests = []struct {
		in  string
		out Certificates
	}{
		{"staging.example.com", Certificates{certificates[0]}},
		{"www.example.com", Certificates{certificates[1], certificates[2]}},
		{"www.amazon.com", empty},
	}

	for _, test := range tests {
		if !reflect.DeepEqual(certificates.GetCertificates(test.in), test.out) {
			t.Errorf("Expected %+v, got: %+v", test.out, certificates.GetCertificates(test.in))
		}
	}
}

func TestRequestCertificate(t *testing.T) {
	certificateARN := "arn:aws:acm:us-east-1:123456789012:certificate/12345678-1234-1234-1234-123456789012"
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
		CertificateArn: aws.String(certificateARN),
	}

	mockACMAPI.EXPECT().RequestCertificate(i).Return(o, nil)

	arn, err := acm.RequestCertificate(domainName, aliases)

	if err != nil {
		t.Errorf("Error; %+v", err)
	}

	if arn != certificateARN {
		t.Errorf("Invalid certificate ARN; want: %s, got: %s", certificateARN, arn)
	}
}

func TestRequestCertificateError(t *testing.T) {
	var aliases []string

	domainName := "*.example.com"
	validationMethod := "DNS"

	// Simulating a certificate request with more than 10 domains
	for i := 0; i < 10; i++ {
		aliases = append(aliases, fmt.Sprintf("example-%d.com", i))
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

	mockACMAPI.EXPECT().RequestCertificate(i).Return(o, errors.New("certificate has too many domains"))

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

	mockClient := sdk.MockListCertificatesPagesClient{Resp: resp}
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
	mockClient := sdk.MockListCertificatesPagesClient{
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
	certificateARN := "arn:aws:acm:us-east-1:123456789012:certificate/12345678-1234-1234-1234-123456789012"

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockACMAPI := sdk.NewMockACMAPI(mockCtrl)

	acm := SDKClient{client: mockACMAPI}
	i := &awsacm.DeleteCertificateInput{CertificateArn: aws.String(certificateARN)}
	o := &awsacm.DeleteCertificateOutput{}

	mockACMAPI.EXPECT().DeleteCertificate(i).Return(o, nil)

	err := acm.DeleteCertificate(certificateARN)

	if err != nil {
		t.Errorf("Error; %+v", err)
	}
}

func TestDeleteCertificateError(t *testing.T) {
	certificateARN := "arn:aws:acm:us-east-1:123456789012:certificate/12345678-1234-1234-1234-123456789012"

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockACMAPI := sdk.NewMockACMAPI(mockCtrl)

	acm := SDKClient{client: mockACMAPI}
	i := &awsacm.DeleteCertificateInput{CertificateArn: aws.String(certificateARN)}
	o := &awsacm.DeleteCertificateOutput{}

	mockACMAPI.EXPECT().DeleteCertificate(i).Return(o, errors.New(":-("))

	err := acm.DeleteCertificate(certificateARN)

	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestInflateCertificate(t *testing.T) {
	domainName := "www.example.com"
	certificateARN := "arn:aws:acm:us-east-1:123456789012:certificate/12345678-1234-1234-1234-123456789012"
	certificateType := "AMAZON_ISSUED"
	certificateStatus := "PENDING_VALIDATION"
	certificateSANs := []string{"staging.example.com"}

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockACMAPI := sdk.NewMockACMAPI(mockCtrl)

	acm := SDKClient{client: mockACMAPI}
	i := &awsacm.DescribeCertificateInput{
		CertificateArn: aws.String(certificateARN),
	}
	o := &awsacm.DescribeCertificateOutput{
		Certificate: &awsacm.CertificateDetail{
			Type:                    aws.String(certificateType),
			Status:                  aws.String(certificateStatus),
			SubjectAlternativeNames: aws.StringSlice(certificateSANs),
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
					DomainName:       aws.String(domainName),
					ResourceRecord: &awsacm.ResourceRecord{
						Name:  aws.String("_beeed67ae3f2d83f6cd3e19a8064947b.www.example.com"),
						Type:  aws.String("CNAME"),
						Value: aws.String("_6ddc33cd42c3fe3d5eca4cb075013a0a.acm-validations.aws."),
					},
				},
			},
		},
	}
	certificate := Certificate{
		DomainName: domainName,
		ARN:        certificateARN,
	}

	mockACMAPI.EXPECT().DescribeCertificate(i).Return(o, nil)

	if err := acm.InflateCertificate(&certificate); err != nil {
		t.Errorf("Expected no error, got %+v", err)
	}

	if certificate.DomainName != domainName {
		t.Errorf("Expected DomainName %s, got %s", domainName, certificate.DomainName)
	}

	if certificate.ARN != certificateARN {
		t.Errorf("Expected ARN %s, got %s", certificateARN, certificate.ARN)
	}

	if certificate.Type != certificateType {
		t.Errorf("Expected type %s, got %s", certificateType, certificate.Type)
	}

	if certificate.Status != certificateStatus {
		t.Errorf("Expected status %s, got %s", certificateStatus, certificate.Status)
	}

	if !reflect.DeepEqual(certificateSANs, certificate.SubjectAlternativeNames) {
		t.Errorf("Expected subject alternative names %v, got %v", certificateSANs, certificate.SubjectAlternativeNames)
	}

	if len(certificate.Validations) != 2 {
		t.Errorf("Expected 2 validations, got %d", len(certificate.Validations))
	}
}

func TestInflateCertificateError(t *testing.T) {
	certificate := Certificate{
		ARN:        "arn:aws:acm:us-east-1:123456789012:certificate/12345678-1234-1234-1234-123456789012",
		DomainName: "www.example.com",
	}

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockACMAPI := sdk.NewMockACMAPI(mockCtrl)

	acm := SDKClient{client: mockACMAPI}
	i := &awsacm.DescribeCertificateInput{CertificateArn: aws.String(certificate.ARN)}
	o := &awsacm.DescribeCertificateOutput{}

	mockACMAPI.EXPECT().DescribeCertificate(i).Return(o, errors.New(":-("))

	err := acm.InflateCertificate(&certificate)

	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestImportCertificate(t *testing.T) {
	dummy := make([]byte, 10)
	rand.Read(dummy)

	certificateARN := "arn:aws:acm:us-east-1:123456789012:certificate/12345678-1234-1234-1234-123456789012"

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockACMAPI := sdk.NewMockACMAPI(mockCtrl)

	acm := SDKClient{client: mockACMAPI}
	i := &awsacm.ImportCertificateInput{
		Certificate:      dummy,
		CertificateChain: dummy,
		PrivateKey:       dummy,
	}
	o := &awsacm.ImportCertificateOutput{
		CertificateArn: aws.String(certificateARN),
	}

	mockACMAPI.EXPECT().ImportCertificate(i).Return(o, nil)

	arn, err := acm.ImportCertificate(dummy, dummy, dummy)

	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}

	if arn != certificateARN {
		t.Errorf("Expected ARN == %s, got: %s", certificateARN, arn)
	}
}

func TestImportCertificateError(t *testing.T) {
	var empty []byte

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockACMAPI := sdk.NewMockACMAPI(mockCtrl)

	acm := SDKClient{client: mockACMAPI}
	i := &awsacm.ImportCertificateInput{
		Certificate:      empty,
		CertificateChain: empty,
		PrivateKey:       empty,
	}
	o := &awsacm.ImportCertificateOutput{}

	mockACMAPI.EXPECT().ImportCertificate(i).Return(o, errors.New(":-("))

	_, err := acm.ImportCertificate(empty, empty, empty)

	if err == nil {
		t.Error("Expected error, got nil")
	}
}
