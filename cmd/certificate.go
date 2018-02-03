package cmd

import (
	"github.com/spf13/cobra"
)

var certificateCmd = &cobra.Command{
	Use:   "certificate",
	Short: "Manage certificates",
	Long: `Manages certificate

Certificates are TLS certificates issued by or imported into AWS Certificate
Manager for use in securing traffic between load balancers and end users. ACM
provides TLS certificates free of charge for use within AWS resources.`,
}

func init() {
	rootCmd.AddCommand(certificateCmd)
}
