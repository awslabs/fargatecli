package cmd

import (
	"errors"
	"strings"

	"github.com/jpignata/fargate/acm"
	"github.com/spf13/cobra"
)

var errCertificateNotFound = errors.New("Certificate not found")

type certificateInfoOperation struct {
	acm        acm.Client
	domainName string
	output     Output
}

func (o certificateInfoOperation) execute() {
	certificate, err := o.find()

	if err != nil {
		switch err {
		case errCertificateNotFound:
			o.output.Info("No certificate found for %s", o.domainName)
		default:
			o.output.Fatal(err, "Could not find certificate for %s", o.domainName)
		}

		return
	}

	o.display(certificate)
}

func (o certificateInfoOperation) find() (acm.Certificate, error) {
	o.output.Debug("Listing certificates [API=acm Action=ListCertificate]")
	certificates, err := o.acm.ListCertificates()

	if err != nil {
		return acm.Certificate{}, err
	}

	for _, certificate := range certificates {
		if certificate.DomainName == o.domainName {
			o.output.Debug("Describing certificate [API=acm Action=DescribeCertificate ARN=%s]", certificate.Arn)
			certificate, err := o.acm.InflateCertificate(certificate)

			if err != nil {
				return acm.Certificate{}, err
			}

			return certificate, nil
		}
	}

	return acm.Certificate{}, errCertificateNotFound
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
			acm:        acm.New(sess),
			domainName: args[0],
			output:     output,
		}.execute()
	},
}

func init() {
	certificateCmd.AddCommand(certificateInfoCmd)
}
