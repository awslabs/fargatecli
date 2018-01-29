package cmd

import (
	ACM "github.com/jpignata/fargate/acm"
	"github.com/jpignata/fargate/console"
	"github.com/spf13/cobra"
)

type CertificateDestroyOperation struct {
	DomainName string
}

var certificateDestroyCmd = &cobra.Command{
	Use:   "destroy <domain-name>",
	Short: "Destroy certificate",
	Long: `Destroy certificate

In order to destroy a certificate, it must not be in use by any load balancers or
any other AWS resources.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		operation := &CertificateDestroyOperation{
			DomainName: args[0],
		}

		destroyCertificate(operation)
	},
}

func init() {
	certificateCmd.AddCommand(certificateDestroyCmd)
}

func destroyCertificate(operation *CertificateDestroyOperation) {
	acm := ACM.New(sess)

	acm.DeleteCertificate(operation.DomainName)
	console.Info("Destroyed certificate %s", operation.DomainName)
}
