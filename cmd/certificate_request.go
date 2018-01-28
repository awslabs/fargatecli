package cmd

import (
	ACM "github.com/jpignata/fargate/acm"
	"github.com/jpignata/fargate/console"
	"github.com/spf13/cobra"
)

type CertificateRequestOperation struct {
	ACM        ACM.ACMClient
	Aliases    []string
	DomainName string
}

func (o CertificateRequestOperation) Validate() {
	validateDomainName(o.DomainName)

	for _, alias := range o.Aliases {
		err := ACM.ValidateAlias(alias)

		if err != nil {
			console.ErrorExit(err, "Invalid domain name")
		}
	}
}

func (o CertificateRequestOperation) Execute() {
	_, err := o.ACM.RequestCertificate(o.DomainName, o.Aliases)

	if err != nil {
		console.ErrorExit(err, "Could not request certificate")
	}

	console.Info("Requested certificate for %s", o.DomainName)
	console.Info("  You must validate ownership of the domain name for the certificate to be issued")
	console.Info("  If your domain is hosted in Amazon Route53, you can do this by running: `fargate certificate validate %s`", o.DomainName)
	console.Info("  Otherwise you must manually create the DNS record, see: `fargate certificate info %s`", o.DomainName)
}

var flagCertificateRequestAliases []string

var certificateRequestCmd = &cobra.Command{
	Use:   "request <domain-name>",
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
			ACM:        ACM.New(sess),
		}

		operation.Validate()
		operation.Execute()
	},
}

func init() {
	certificateRequestCmd.Flags().StringSliceVarP(&flagCertificateRequestAliases, "alias", "a", []string{},
		`Additional domain names to be included in the certificate (can be specified multiple times)`)

	certificateCmd.AddCommand(certificateRequestCmd)
}
