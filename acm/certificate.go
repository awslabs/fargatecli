package acm

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	awsacm "github.com/aws/aws-sdk-go/service/acm"
	"github.com/jpignata/fargate/console"
	"golang.org/x/time/rate"
)

type Certificate struct {
	Arn                     string
	Status                  string
	SubjectAlternativeNames []string
	DomainName              string
	Validations             []CertificateValidation
	Type                    string
}

func (c *Certificate) AddValidation(v CertificateValidation) {
	c.Validations = append(c.Validations, v)
}

func (c *Certificate) Inflate(d *awsacm.CertificateDetail) *Certificate {
	c.Status = aws.StringValue(d.Status)
	c.SubjectAlternativeNames = aws.StringValueSlice(d.SubjectAlternativeNames)
	c.Type = aws.StringValue(d.Type)

	for _, domainValidation := range d.DomainValidationOptions {
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

	return c
}

func (c *Certificate) IsIssued() bool {
	return c.Status == awsacm.CertificateStatusIssued
}

type CertificateValidation struct {
	Status         string
	DomainName     string
	ResourceRecord CertificateResourceRecord
}

func (v *CertificateValidation) IsPendingValidation() bool {
	return v.Status == awsacm.DomainStatusPendingValidation
}

func (v *CertificateValidation) IsSuccess() bool {
	return v.Status == awsacm.DomainStatusSuccess
}

func (v *CertificateValidation) IsFailed() bool {
	return v.Status == awsacm.DomainStatusFailed
}

func (v *CertificateValidation) ResourceRecordString() string {
	if v.ResourceRecord.Type == "" {
		return ""
	}

	return fmt.Sprintf("%s %s -> %s",
		v.ResourceRecord.Type,
		v.ResourceRecord.Name,
		v.ResourceRecord.Value,
	)
}

type CertificateResourceRecord struct {
	Type  string
	Name  string
	Value string
}

func (c *Certificate) IsPendingValidation() bool {
	return c.Status == awsacm.CertificateStatusPendingValidation
}

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

type Certificates []Certificate

func (cs Certificates) GetCertificateArns(domainName string) []string {
	var certificateArns []string

	for _, c := range cs {
		if c.DomainName == domainName {
			certificateArns = append(certificateArns, c.Arn)
		}
	}

	return certificateArns
}

func ValidateAlias(alias string) error {
	if len(alias) < 1 || len(alias) > 253 {
		return fmt.Errorf("%s: An alias must be between 1 and 253 characters in length", alias)
	}

	if strings.Count(alias, ".") > 252 {
		return fmt.Errorf("%s: An alias domain name cannot exceed 253 octets", alias)
	}

	if strings.Count(alias, ".") == 0 {
		return fmt.Errorf("%s: An alias requires at least 2 octets", alias)
	}

	return nil
}

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

func (acm *SDKClient) ListCertificates() []Certificate {
	var wg sync.WaitGroup

	ctx := context.Background()
	ch := make(chan Certificate)
	certificates, _ := acm.ListCertificates2()
	limiter := rate.NewLimiter(10, 1)

	for i := 0; i < 4; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for c := range ch {
				if err := limiter.Wait(ctx); err == nil {
					certificateDetail := acm.describeCertificate(c.Arn)
					c.Inflate(certificateDetail)
				}
			}
		}()
	}

	for _, c := range certificates {
		ch <- c
	}

	close(ch)

	wg.Wait()

	return certificates
}

func (acm *SDKClient) DescribeCertificate(domainName string) *Certificate {
	var certificate *Certificate

	certificates, _ := acm.ListCertificates2()

	for _, c := range certificates {
		if c.DomainName == domainName {
			certificateDetail := acm.describeCertificate(c.Arn)
			certificate = c.Inflate(certificateDetail)

			break
		}
	}

	if certificate == nil {
		err := fmt.Errorf("Could not find ACM certificate for %s", domainName)
		console.ErrorExit(err, "Couldn't describe ACM certificate")
	}

	return certificate
}

func (acm *SDKClient) ListCertificateDomainNames(certificateArns []string) []string {
	var domainNames []string

	certificates, _ := acm.ListCertificates2()

	for _, certificate := range certificates {
		for _, certificateArn := range certificateArns {
			if certificate.Arn == certificateArn {
				domainNames = append(domainNames, certificate.DomainName)
			}
		}
	}

	return domainNames
}

func (acm *SDKClient) ImportCertificate(certificate, privateKey, certificateChain []byte) {
	console.Debug("Importing ACM certificate")

	input := &awsacm.ImportCertificateInput{
		Certificate: certificate,
		PrivateKey:  privateKey,
	}

	if len(certificateChain) != 0 {
		input.SetCertificateChain(certificateChain)
	}

	_, err := acm.client.ImportCertificate(input)

	if err != nil {
		console.ErrorExit(err, "Couldn't import certificate")
	}
}

func (acm SDKClient) DeleteCertificate(arn string) error {
	input := &awsacm.DeleteCertificateInput{
		CertificateArn: aws.String(arn),
	}

	if _, err := acm.client.DeleteCertificate(input); err != nil {
		return err
	}

	return nil
}

func (acm *SDKClient) describeCertificate(arn string) *awsacm.CertificateDetail {
	resp, err := acm.client.DescribeCertificate(
		&awsacm.DescribeCertificateInput{
			CertificateArn: aws.String(arn),
		},
	)

	if err != nil {
		console.ErrorExit(err, "Couldn't describe ACM certificate")
	}

	return resp.Certificate
}

func (acm SDKClient) ListCertificates2() (Certificates, error) {
	var certificates Certificates

	input := &awsacm.ListCertificatesInput{}
	handler := func(resp *awsacm.ListCertificatesOutput, lastPage bool) bool {
		for _, cs := range resp.CertificateSummaryList {
			c := Certificate{
				Arn:        aws.StringValue(cs.CertificateArn),
				DomainName: aws.StringValue(cs.DomainName),
			}

			certificates = append(certificates, c)
		}

		return true
	}

	if err := acm.client.ListCertificatesPages(input, handler); err != nil {
		return certificates, err
	}

	return certificates, nil
}
