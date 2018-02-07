package cmd

import (
	"github.com/jpignata/fargate/acm"
	"github.com/spf13/cobra"
)

type certificateRequestOperation struct {
	acm        acm.Client
	aliases    []string
	output     Output
	domainName string
}

func (o certificateRequestOperation) execute() {
	if errs := o.validate(); len(errs) > 0 {
		o.output.Fatals(errs, "Invalid certificate request parameters")

		return
	}

	o.output.Debug("Requesting certificate [API=acm Action=RequestCertificate]")

	if arn, err := o.acm.RequestCertificate(o.domainName, o.aliases); err == nil {
		o.output.Debug("Requested certificate [ARN=%s]", arn)
	} else {
		o.output.Fatal(err, "Could not request certificate")

		return
	}

	o.output.Info("Requested certificate for %s", o.domainName)
	o.output.LineBreak()
	o.output.Say("You must validate ownership of the domain name for the certificate to be issued.", 0)
	o.output.LineBreak()
	o.output.Say("If your domain is hosted using Amazon Route 53, this can be done automatically by running:", 0)
	o.output.Say("fargate certificate validate %s", 1, o.domainName)
	o.output.LineBreak()
	o.output.Say("If not, you must manually create the DNS records returned by running:", 0)
	o.output.Say("fargate certificate info %s", 1, o.domainName)
}

func (o certificateRequestOperation) validate() []error {
	var errors []error

	if err := acm.ValidateDomainName(o.domainName); err != nil {
		errors = append(errors, err)
	}

	for _, alias := range o.aliases {
		if err := acm.ValidateAlias(alias); err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}

var (
	flagCertificateRequestAliases []string

	certificateRequestCmd = &cobra.Command{
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
			certificateRequestOperation{
				acm:        acm.New(sess),
				aliases:    flagCertificateRequestAliases,
				output:     output,
				domainName: args[0],
			}.execute()
		},
	}
)

func init() {
	certificateRequestCmd.Flags().StringSliceVarP(&flagCertificateRequestAliases, "alias", "a", []string{},
		`Additional domain names to be included in the certificate (can be specified multiple times)`)

	certificateCmd.AddCommand(certificateRequestCmd)
}
