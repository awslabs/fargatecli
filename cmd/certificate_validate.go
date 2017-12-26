package cmd

import (
	"fmt"

	ACM "github.com/jpignata/fargate/acm"
	"github.com/jpignata/fargate/console"
	Route53 "github.com/jpignata/fargate/route53"
	"github.com/jpignata/fargate/util"
	"github.com/spf13/cobra"
)

type CertificateValidateOperation struct {
	DomainName string
}

func (o *CertificateValidateOperation) Validate() {
	validateDomainName(o.DomainName)
}

var certificateValidateCmd = &cobra.Command{
	Use:   "validate <domain name>",
	Args:  cobra.ExactArgs(1),
	Short: "Validate an SSL certificate request",
	Long: `Validate an SSL certificate request

Creates a validation record for domains pending validation that are hosted
within Amazon Route53 and within the same AWS account. If you're using another
DNS provider, this command will return the DNS records you must add to your
domain in order to validate domain ownership to complete the SSL certificate
request.`,
	Run: func(cmd *cobra.Command, args []string) {
		operation := &CertificateValidateOperation{
			DomainName: args[0],
		}

		validateCertificate(operation)
	},
}

func init() {
	certificateCmd.AddCommand(certificateValidateCmd)
}

func validateCertificate(operation *CertificateValidateOperation) {
	console.Info("Validating certificate [%s]", operation.DomainName)

	route53 := Route53.New(sess)
	acm := ACM.New(sess)

	hostedZones := route53.ListHostedZones()
	certificate := acm.DescribeCertificate(operation.DomainName)

	if !certificate.IsPendingValidation() {
		console.ErrorExit(fmt.Errorf("Certificate status is %s", util.Humanize(certificate.Status)), "Could not validate certificate")
	}

	for _, certificateValidation := range certificate.Validations {
		if certificateValidation.IsPendingValidation() {
			createResourceRecord(certificateValidation, route53, hostedZones)
		} else if certificateValidation.IsSuccess() {
			console.Info("[%s] Domain has been validated", certificateValidation.DomainName)
		} else if certificateValidation.IsFailed() {
			console.Info("[%s] Domain has failed validation; please delete the certificate and re-request", certificateValidation.DomainName)
		}
	}

	console.Info("[%s] Record validation could take up to several hours to complete", operation.DomainName)
	console.Info("[%s] To view the status of pending validations, run: `fargate certificate info %s`", operation.DomainName, operation.DomainName)
}

func createResourceRecord(v ACM.CertificateValidation, route53 Route53.Route53, hostedZones []Route53.HostedZone) bool {
	for _, hostedZone := range hostedZones {
		if hostedZone.IsSuperDomainOf(v.DomainName) {
			console.Debug("[%s] Found Route53 hosted zone", v.DomainName)
			console.Debug("[%s] Creating %s %s -> %s",
				v.DomainName,
				v.ResourceRecord.Type,
				v.ResourceRecord.Name,
				v.ResourceRecord.Value,
			)

			route53.CreateResourceRecord(
				hostedZone,
				v.ResourceRecord.Type,
				v.ResourceRecord.Name,
				v.ResourceRecord.Value,
			)

			console.Info("[%s] Created validation record", v.DomainName)

			return true
		}
	}

	console.Issue("[%s] Could not find Route53 hosted zone", v.DomainName)
	console.Info("[%s] If you're hosting this domain elsewhere or in another AWS account, please manually create the validation record:", v.DomainName)
	console.Info("[%s] Name: %s  Type: %s  Value: %s", v.DomainName, v.ResourceRecord.Name, v.ResourceRecord.Type, v.ResourceRecord.Value)

	return false
}
