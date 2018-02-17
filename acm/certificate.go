package acm

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	awsacm "github.com/aws/aws-sdk-go/service/acm"
)

// Certificate is a certificate hosted in AWS Certificate Manager.
type Certificate struct {
	ARN                     string
	Status                  string
	SubjectAlternativeNames []string
	DomainName              string
	Validations             []CertificateValidation
	Type                    string
}

// AddValidation adds a certificate validation to a certificate.
func (c *Certificate) AddValidation(v CertificateValidation) {
	c.Validations = append(c.Validations, v)
}

// IsIssued returns true if the certificate's status is ISSUED.
func (c Certificate) IsIssued() bool {
	return c.Status == awsacm.CertificateStatusIssued
}

// IsPendingValidation returns true if the certificate's status is PENDING_VALIDATION.
func (c Certificate) IsPendingValidation() bool {
	return c.Status == awsacm.CertificateStatusPendingValidation
}

// CertificateValidation holds details about how to validate a certificate.
type CertificateValidation struct {
	Status         string
	DomainName     string
	ResourceRecord CertificateResourceRecord
}

// IsFailed returns true if a certificate validation's status is FAILED.
func (v CertificateValidation) IsFailed() bool {
	return v.Status == awsacm.DomainStatusFailed
}

// IsPendingValidation returns true if a certificate validation's status is PENDING_VALIDATION.
func (v CertificateValidation) IsPendingValidation() bool {
	return v.Status == awsacm.DomainStatusPendingValidation
}

// IsSuccess returns true if a certificate validation's status is SUCCESS
func (v CertificateValidation) IsSuccess() bool {
	return v.Status == awsacm.DomainStatusSuccess
}

// ResourceRecordString returns the resource record for display.
func (v CertificateValidation) ResourceRecordString() string {
	if v.ResourceRecord.Type == "" {
		return ""
	}

	return fmt.Sprintf("%s %s -> %s",
		v.ResourceRecord.Type,
		v.ResourceRecord.Name,
		v.ResourceRecord.Value,
	)
}

// CertificateResourceRecord contains the DNS record used to validate a certificate.
type CertificateResourceRecord struct {
	Type  string
	Name  string
	Value string
}

// Certificates is a collection of certificates.
type Certificates []Certificate

// GetCertificates returns certificates for the given domain name.
func (cs Certificates) GetCertificates(domainName string) Certificates {
	var certificates Certificates

	for _, c := range cs {
		if c.DomainName == domainName {
			certificates = append(certificates, c)
		}
	}

	return certificates
}

// ValidateAlias checks an alias' length and number of octets for validity.
func ValidateAlias(alias string) error {
	if len(alias) < 1 || len(alias) > 253 {
		return fmt.Errorf("%s: An alias must be between 1 and 253 characters in length", alias)
	}

	if strings.Count(alias, ".") > 252 {
		return fmt.Errorf("%s: An alias cannot exceed 253 octets", alias)
	}

	if strings.Count(alias, ".") == 0 {
		return fmt.Errorf("%s: An alias requires at least 2 octets", alias)
	}

	return nil
}

// ValidateDomainName checks a domain names length and number of octets for validity.
func ValidateDomainName(domainName string) error {
	if len(domainName) < 1 || len(domainName) > 253 {
		return fmt.Errorf("%s: The domain name must be between 1 and 253 characters in length", domainName)
	}

	if strings.Count(domainName, ".") > 62 {
		return fmt.Errorf("%s: The domain name cannot exceed 63 octets", domainName)
	}

	if strings.Count(domainName, ".") == 0 {
		return fmt.Errorf("%s: The domain name requires at least 2 octets", domainName)
	}

	return nil
}

// DeleteCertificate deletes the certificate identified by the given ARN.
func (acm SDKClient) DeleteCertificate(arn string) error {
	input := &awsacm.DeleteCertificateInput{
		CertificateArn: aws.String(arn),
	}

	if _, err := acm.client.DeleteCertificate(input); err != nil {
		return err
	}

	return nil
}

// ImportCertificate creates a new certificate from the provided certificate, private key, and
// optional certificate chain.
func (acm SDKClient) ImportCertificate(certificate, privateKey, certificateChain []byte) (string, error) {
	input := &awsacm.ImportCertificateInput{
		Certificate: certificate,
		PrivateKey:  privateKey,
	}

	if len(certificateChain) != 0 {
		input.SetCertificateChain(certificateChain)
	}

	resp, err := acm.client.ImportCertificate(input)

	return aws.StringValue(resp.CertificateArn), err
}

// InflateCertificate uses a partially hydrated certificate to fetch the rest of its details and
// return the full certificate.
func (acm SDKClient) InflateCertificate(c Certificate) (Certificate, error) {
	resp, err := acm.client.DescribeCertificate(
		&awsacm.DescribeCertificateInput{
			CertificateArn: aws.String(c.ARN),
		},
	)

	if err != nil {
		return c, err
	}

	c.Status = aws.StringValue(resp.Certificate.Status)
	c.SubjectAlternativeNames = aws.StringValueSlice(resp.Certificate.SubjectAlternativeNames)
	c.Type = aws.StringValue(resp.Certificate.Type)

	for _, domainValidation := range resp.Certificate.DomainValidationOptions {
		validation := CertificateValidation{
			Status:     aws.StringValue(domainValidation.ValidationStatus),
			DomainName: aws.StringValue(domainValidation.DomainName),
		}

		if domainValidation.ResourceRecord != nil {
			validation.ResourceRecord = CertificateResourceRecord{
				Type:  aws.StringValue(domainValidation.ResourceRecord.Type),
				Name:  aws.StringValue(domainValidation.ResourceRecord.Name),
				Value: aws.StringValue(domainValidation.ResourceRecord.Value),
			}
		}

		c.AddValidation(validation)
	}

	return c, nil
}

// ListCertificates returns all certificates associated with the caller's account.
func (acm SDKClient) ListCertificates() (Certificates, error) {
	var certificates Certificates

	input := &awsacm.ListCertificatesInput{}
	handler := func(resp *awsacm.ListCertificatesOutput, lastPage bool) bool {
		for _, cs := range resp.CertificateSummaryList {
			c := Certificate{
				ARN:        aws.StringValue(cs.CertificateArn),
				DomainName: aws.StringValue(cs.DomainName),
			}

			certificates = append(certificates, c)
		}

		return true
	}

	err := acm.client.ListCertificatesPages(input, handler)

	return certificates, err
}

// RequestCertificate creates a new certificate.
func (acm SDKClient) RequestCertificate(domainName string, aliases []string) (string, error) {
	requestCertificateInput := &awsacm.RequestCertificateInput{
		DomainName:       aws.String(domainName),
		ValidationMethod: aws.String(awsacm.ValidationMethodDns),
	}

	if len(aliases) > 0 {
		requestCertificateInput.SetSubjectAlternativeNames(aws.StringSlice(aliases))
	}

	resp, err := acm.client.RequestCertificate(requestCertificateInput)

	if err != nil {
		return "", err
	}

	return aws.StringValue(resp.CertificateArn), nil
}

// ListCertificateDomainNames is bunk and will be refactored out of existence soon.
func (acm *SDKClient) ListCertificateDomainNames(certificateARNs []string) []string {
	var domainNames []string

	certificates, _ := acm.ListCertificates()

	for _, certificate := range certificates {
		for _, certificateARN := range certificateARNs {
			if certificate.ARN == certificateARN {
				domainNames = append(domainNames, certificate.DomainName)
			}
		}
	}

	return domainNames
}
