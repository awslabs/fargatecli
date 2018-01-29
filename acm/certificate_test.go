package acm

import (
	"errors"
	"fmt"
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

	acm := SDKClient{
		client: mockACMAPI,
	}

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

func TestRequestWithLimitError(t *testing.T) {
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

	acm := SDKClient{
		client: mockACMAPI,
	}

	i := &awsacm.RequestCertificateInput{
		DomainName:              aws.String(domainName),
		ValidationMethod:        aws.String(validationMethod),
		SubjectAlternativeNames: aws.StringSlice(aliases),
	}
	o := &awsacm.RequestCertificateOutput{}

	mockACMAPI.EXPECT().RequestCertificate(i).Return(o, errors.New("Certificate has too many domains."))

	arn, err := acm.RequestCertificate(domainName, aliases)

	if err == nil {
		t.Error("No error; want: %+v", err)
	}

	if arn != "" {
		t.Error("Invalid certificate ARN; want: empty, got: %s", arn)
	}
}
