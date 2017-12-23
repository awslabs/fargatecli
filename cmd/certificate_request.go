package cmd

import (
	ACM "github.com/jpignata/fargate/acm"
	"github.com/jpignata/fargate/console"
	"github.com/spf13/cobra"
)

var (
	domainName string
	aliases    []string
)

var certificateCreateCmd = &cobra.Command{
	Use:   "request <domain name>",
	Short: "Request a new SSL certificate",
	Long: `Request a new SSL certificate

Creates an SSL certificate within AWS Certificate Manager. After the request
is made, domain ownership must be validated via DNS. The convenience command
_fargate certificate validate_ can be used to create these records if your
domain is hosted within Amazon Route53 and within the same AWS account.`,
	Args: cobra.ExactArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		validateDomainName(args[0])
	},
	Run: func(cmd *cobra.Command, args []string) {
		createCertificate(args[0])
	},
}

func init() {
	certificateCreateCmd.Flags().StringSliceVarP(&aliases, "alias", "a", []string{}, "Additional FQDNs to be includes in the Subject Alternative Name extension of the SSL certificate")

	certificateCmd.AddCommand(certificateCreateCmd)
}

func createCertificate(domainName string) {
	console.Info("Requesting certificate [%s]", domainName)

	acm := ACM.New(sess)
	acm.RequestCertificate(domainName, aliases)

	console.Info("[%s] You must validate ownership of the domain name for the certificate to be issued", domainName)
	console.Info("[%s] If your domain is hosted in Amazon Route53, you can do this by running: `fargate certificate validate %s`", domainName, domainName)
	console.Info("[%s] Otherwise you must manually create the DNS records, see: `fargate certificate info %s`", domainName, domainName)
}
