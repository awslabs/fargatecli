package cmd

import (
	"github.com/jpignata/fargate/console"
	ELBV2 "github.com/jpignata/fargate/elbv2"
	"github.com/spf13/cobra"
)

var loadBalancerDestroyCmd = &cobra.Command{
	Use:   "destroy <load balancer name>",
	Short: "Deletes an SSL load balancer",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		destroyLoadBalancer(args[0])
	},
}

func init() {
	lbCmd.AddCommand(loadBalancerDestroyCmd)
}

func destroyLoadBalancer(domainName string) {
	console.Info("[%s] Destroying load balancer", domainName)

	elbv2 := ELBV2.New(sess)
	elbv2.DeleteLoadBalancer(domainName)
	elbv2.DeleteTargetGroup(domainName + "-" + "default")
}
