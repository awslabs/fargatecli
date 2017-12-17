package cmd

import (
	ACM "github.com/jpignata/fargate/acm"
	"github.com/jpignata/fargate/console"
	"github.com/spf13/cobra"
)

var certificateCmd = &cobra.Command{
	Use:   "certificate",
	Short: "Manage SSL certificates",
	Long: `Manages SSL certificate for use in load balancers

Creates, imports, and validates SSL certificates in AWS Certificate Mananger
for use within Elastic Load Balancing load balancers.
`,
}

func init() {
	rootCmd.AddCommand(certificateCmd)
}

func validateDomainName(domainName string) {
	err := ACM.ValidateDomainName(domainName)

	if err != nil {
		console.ErrorExit(err, "Invalid domain name")
	}
}
