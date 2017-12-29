package cmd

import (
	ACM "github.com/jpignata/fargate/acm"
	"github.com/jpignata/fargate/console"
	"github.com/spf13/cobra"
)

type CertificateRequestOperation struct {
	DomainName string
	Aliases    []string
}

func (o *CertificateRequestOperation) Validate() {
	validateDomainName(o.DomainName)

	for _, alias := range o.Aliases {
		err := ACM.ValidateAlias(alias)

		if err != nil {
			console.ErrorExit(err, "Invalid domain name")
		}
	}
}

var flagCertificateRequestAliases []string

var certificateRequestCmd = &cobra.Command{
	Use:   "request <domain name>",
	Short: "Request a certificate",
	Long: `Request a certificate

Certificates can be for a fully qualified domain name (e.g. www.example.com) or
a wildcard domain name (e.g. *.example.com). You can add aliases to a
certificate by specifying additional domain names via the --alias flag. To add
multiple aliases, pass --alias multiple times. By default, AWS Certificate
Manager has a limit of 10 domain names per certificate, but this limit can be
raised by AWS support.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		operation := &CertificateRequestOperation{
			DomainName: args[0],
			Aliases:    flagCertificateRequestAliases,
		}

		operation.Validate()

		requestCertificate(operation)
	},
}

func init() {
	certificateRequestCmd.Flags().StringSliceVarP(&flagCertificateRequestAliases, "alias", "a", []string{}, "Additional FQDNs to be includes in the Subject Alternative Name extension of the SSL certificate")

	certificateCmd.AddCommand(certificateRequestCmd)
}

func requestCertificate(operation *CertificateRequestOperation) {
	acm := ACM.New(sess)
	acm.RequestCertificate(operation.DomainName, operation.Aliases)

	console.Info("Requested certificate for %s", operation.DomainName)
	console.Info("  You must validate ownership of the domain name for the certificate to be issued")
	console.Info("  If your domain is hosted in Amazon Route53, you can do this by running: `fargate certificate validate %s`", operation.DomainName)
	console.Info("  Otherwise you must manually create the DNS record, see: `fargate certificate info %s`", operation.DomainName)
}
