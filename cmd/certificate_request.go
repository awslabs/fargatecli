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
	Short: "Request a new SSL certificate",
	Long: `Request a new SSL certificate

Creates an SSL certificate within AWS Certificate Manager. After the request
is made, domain ownership must be validated via DNS. The convenience command
_fargate certificate validate_ can be used to create these records if your
domain is hosted within Amazon Route53 and within the same AWS account.`,
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
