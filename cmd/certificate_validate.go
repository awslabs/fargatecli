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
	Short: "Validate certificate ownership",
	Long: `Validate certificate ownership

fargate will automatically create DNS validation record to verify ownership for
any domain names that are hosted within Amazon Route 53. If your certificate
has aliases, a validation record will be attempted per alias. Any records whose
domains are hosted in other DNS hosting providers or in other DNS accounts
and cannot be automatically validated will have the necessary records output.
These records are also available in fargate certificate info \<domain-name>.

AWS Certificate Manager may take up to several hours after the DNS records are
created to complete validation and issue the certificate.`,
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
	var successfulValidations int

	route53 := Route53.New(sess)
	acm := ACM.New(sess)

	hostedZones := route53.ListHostedZones()
	certificate := acm.DescribeCertificate(operation.DomainName)

	if !certificate.IsPendingValidation() {
		console.ErrorExit(fmt.Errorf("Certificate status is %s", util.Humanize(certificate.Status)), "Could not validate certificate")
	}

	for _, certificateValidation := range certificate.Validations {
		if certificateValidation.IsPendingValidation() {
			if createResourceRecord(certificateValidation, route53, hostedZones) {
				successfulValidations++
			}
		} else if certificateValidation.IsSuccess() {
			console.Info("[%s] Domain has been validated", certificateValidation.DomainName)
		} else if certificateValidation.IsFailed() {
			console.Info("[%s] Domain has failed validation; please delete the certificate and re-request", certificateValidation.DomainName)
		}
	}

	if successfulValidations > 0 {
		console.Info("[%s] Record validation could take up to several hours to complete", operation.DomainName)
		console.Info("[%s] To view the status of pending validations, run: `fargate certificate info %s`", operation.DomainName, operation.DomainName)
	}
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
	console.Info("[%s]   If you're hosting this domain elsewhere or in another AWS account, please manually create the validation record:", v.DomainName)
	console.Info("[%s]   %s %s -> %s", v.DomainName, v.ResourceRecord.Type, v.ResourceRecord.Name, v.ResourceRecord.Value)

	return false
}
