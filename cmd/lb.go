package cmd

import (
	"github.com/spf13/cobra"
)

var lbCmd = &cobra.Command{
	Use:   "lb",
	Short: "Manage load balancers",
}

func init() {
	rootCmd.AddCommand(lbCmd)
}
