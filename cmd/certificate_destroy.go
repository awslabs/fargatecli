package cmd

import (
	ACM "github.com/jpignata/fargate/acm"
	"github.com/jpignata/fargate/console"
	"github.com/spf13/cobra"
)

var certificateDestroyCmd = &cobra.Command{
	Use:   "destroy <domain name>",
	Short: "Deletes an SSL certificate",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		destroyCertificate(args[0])
	},
}

func init() {
	certificateCmd.AddCommand(certificateDestroyCmd)
}

func destroyCertificate(domainName string) {
	console.Info("[%s] Destroying certificate", domainName)

	acm := ACM.New(sess)
	acm.DeleteCertificate(domainName)
}
