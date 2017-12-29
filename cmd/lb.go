package cmd

import (
	"github.com/spf13/cobra"
)

const defaultTargetGroupFormat = "%s-default"

var lbCmd = &cobra.Command{
	Use:   "lb",
	Short: "Manage load balancers",
	Long: `Manage load balancers

Load balancers distribute incoming traffic between the tasks within a service
for HTTP/HTTPS and TCP applications. HTTP/HTTPS load balancers can route to
multiple services based upon rules you specify when you create a new service.`,
}

func init() {
	rootCmd.AddCommand(lbCmd)
}
