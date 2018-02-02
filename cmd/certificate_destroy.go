package cmd

import (
	"github.com/jpignata/fargate/acm"
	"github.com/spf13/cobra"
)

type certificateDestroyOperation struct {
	acm        acm.Client
	domainName string
	output     Output
}

func (o certificateDestroyOperation) execute() {
	o.output.Debug("Listing certificates [API=acm Action=ListCertificate]")

	certificates, err := o.acm.ListCertificates()

	if err != nil {
		o.output.Fatal(err, "Could not destroy certificate")
		return
	}

	certificateArns := certificates.GetCertificateArns(o.domainName)

	switch {
	case len(certificateArns) == 0:
		o.output.Fatal(nil, "Could not find certificate for %s", o.domainName)
		return
	case len(certificateArns) > 1:
		o.output.Fatal(
			nil,
			"Found %d certificates for %s, for safety please destroy the one you intend via the AWS CLI or AWS Management Console",
			len(certificateArns),
			o.domainName,
		)
		return
	}

	o.output.Debug("Deleting certificate [API=acm Action=DeleteCertificate ARN=%s]", certificateArns[0])

	if err := o.acm.DeleteCertificate(certificateArns[0]); err != nil {
		o.output.Fatal(err, "Could not destroy certificate")
		return
	}

	o.output.Info("Destroyed certificate %s", o.domainName)
}

var certificateDestroyCmd = &cobra.Command{
	Use:   "destroy <domain-name>",
	Short: "Destroy certificate",
	Long: `Destroy certificate

In order to destroy a certificate, it must not be in use by any load balancers or
any other AWS resources.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		certificateDestroyOperation{
			acm:        acm.New(sess),
			domainName: args[0],
			output:     output,
		}.execute()
	},
}

func init() {
	certificateCmd.AddCommand(certificateDestroyCmd)
}
