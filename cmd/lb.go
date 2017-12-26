package cmd

import (
	"github.com/spf13/cobra"
)

const defaultTargetGroupFormat = "%s-default"

var lbCmd = &cobra.Command{
	Use:   "lb",
	Short: "Manage load balancers",
}

func init() {
	rootCmd.AddCommand(lbCmd)
}
