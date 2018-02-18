package cmd

import (
	"github.com/jpignata/fargate/acm"
	"github.com/spf13/cobra"
)

type certificateDestroyOperation struct {
	certificateOperation
	domainName string
	output     Output
}

func (o certificateDestroyOperation) execute() {
	certificate, err := o.findCertificate(o.domainName)

	if err != nil {
		switch err {
		case errCertificateNotFound:
			o.output.Fatal(err, "Could not find certificate for %s", o.domainName)
			return
		case errCertificateTooManyFound:
			o.output.Fatal(err, "Multiple certificates found for %s, for safety please destroy the one you intend via the AWS CLI", o.domainName)
			return
		default:
			o.output.Fatal(err, "Could not destroy certificate")
			return
		}
	}

	o.output.Debug("Deleting certificate [API=acm Action=DeleteCertificate ARN=%s]", certificate.ARN)
	if err := o.acm.DeleteCertificate(certificate.ARN); err != nil {
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
			certificateOperation: certificateOperation{acm: acm.New(sess), output: output},
			domainName:           args[0],
			output:               output,
		}.execute()
	},
}

func init() {
	certificateCmd.AddCommand(certificateDestroyCmd)
}
