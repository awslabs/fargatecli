package cmd

import (
	"fmt"

	"github.com/jpignata/fargate/acm"
	"github.com/jpignata/fargate/route53"
	"github.com/spf13/cobra"
)

type certificateValidateOperation struct {
	certificateOperation
	domainName string
	output     Output
	route53    route53.Client
}

func (o certificateValidateOperation) execute() {
	certificate, err := o.findCertificate(o.domainName)

	if err != nil {
		o.output.Fatal(err, "Could not validate certificate")
		return
	}

	if !certificate.IsPendingValidation() {
		o.output.Fatal(fmt.Errorf("certificate %s is in state %s", o.domainName, Humanize(certificate.Status)), "Could not validate certificate")
		return
	}

	o.output.Debug("Listing hosted zones [API=route53 Action=ListHostedZones]")
	hostedZones, err := o.route53.ListHostedZones()

	if err != nil {
		o.output.Fatal(err, "Could not validate certificate")
		return
	}

	for _, v := range certificate.Validations {
		switch {
		case v.IsPendingValidation():
			if zone, ok := hostedZones.FindSuperDomainOf(v.DomainName); ok {
				o.output.Debug("Creating resource record [API=route53 Action=ChangeResourceRecordSets HostedZone=%s]", zone.ID)
				id, err := o.route53.CreateResourceRecord(
					route53.CreateResourceRecordInput{
						HostedZoneID: zone.ID,
						RecordType:   v.ResourceRecord.Type,
						Name:         v.ResourceRecord.Name,
						Value:        v.ResourceRecord.Value,
					},
				)

				if err != nil {
					o.output.Fatal(err, "Could not validate certificate")
					return
				}

				o.output.Debug("Created resource record [ChangeID=%s]", id)
				o.output.Info("[%s] created validation record", v.DomainName)
			} else {
				o.output.Warn("[%s] could not find zone in Amazon Route 53", v.DomainName)
			}
		case v.IsSuccess():
			o.output.Info("[%s] already validated", v.DomainName)
		case v.IsFailed():
			o.output.Fatal(nil, "[%s] failed validation", v.DomainName)
			return
		default:
			o.output.Warn("[%s] unexpected status: %s", v.DomainName, Humanize(v.Status))
		}
	}
}

var certificateValidateCmd = &cobra.Command{
	Use:   "validate <domain-name>",
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
		certificateValidateOperation{
			certificateOperation: certificateOperation{acm: acm.New(sess), output: output},
			domainName:           args[0],
			output:               output,
			route53:              route53.New(sess),
		}.execute()
	},
}

func init() {
	certificateCmd.AddCommand(certificateValidateCmd)
}
