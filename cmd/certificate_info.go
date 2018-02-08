package cmd

import (
	"strings"

	"github.com/jpignata/fargate/acm"
	"github.com/spf13/cobra"
)

type certificateInfoOperation struct {
	certificateOperation
	domainName string
	output     Output
}

func (o certificateInfoOperation) execute() {
	certificate, err := o.findCertificate(o.domainName, o.output)

	if err != nil {
		switch err {
		case errCertificateNotFound:
			o.output.Info("No certificate found for %s", o.domainName)
		case errCertificateTooManyFound:
			o.output.Fatal(nil, "Multiple certificates found for %s", o.domainName)
		default:
			o.output.Fatal(err, "Could not find certificate for %s", o.domainName)
		}

		return
	}

	o.display(certificate)
}

func (o certificateInfoOperation) display(certificate acm.Certificate) {
	o.output.KeyValue("Domain Name", certificate.DomainName, 0)
	o.output.KeyValue("Status", Humanize(certificate.Status), 0)
	o.output.KeyValue("Type", Humanize(certificate.Type), 0)
	o.output.KeyValue("Subject Alternative Names", strings.Join(certificate.SubjectAlternativeNames, ", "), 0)

	if len(certificate.Validations) > 0 {
		rows := [][]string{
			[]string{"DOMAIN NAME", "STATUS", "RECORD"},
		}

		for _, v := range certificate.Validations {
			rows = append(rows, []string{v.DomainName, Humanize(v.Status), v.ResourceRecordString()})
		}

		o.output.LineBreak()
		o.output.Table("Validations", rows)
	}
}

var certificateInfoCmd = &cobra.Command{
	Use:   "info <domain-name>",
	Short: "Inspect certificate",
	Long: `Inspect certificate

Show extended information for a certificate. Includes each validation for the
certificate which shows DNS records which must be created to validate domain
ownership.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		certificateInfoOperation{
			certificateOperation: certificateOperation{
				acm: acm.New(sess),
			},
			domainName: args[0],
			output:     output,
		}.execute()
	},
}

func init() {
	certificateCmd.AddCommand(certificateInfoCmd)
}
