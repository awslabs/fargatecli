package cmd

import (
	"fmt"

	"github.com/jpignata/fargate/console"
	ELBV2 "github.com/jpignata/fargate/elbv2"
	"github.com/spf13/cobra"
)

type LoadBalancerDestroyOperation struct {
	LoadBalancerName string
}

var loadBalancerDestroyCmd = &cobra.Command{
	Use:   "destroy <load balancer name>",
	Short: "Deletes an SSL load balancer",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		operation := &LoadBalancerDestroyOperation{
			LoadBalancerName: args[0],
		}

		destroyLoadBalancer(operation)
	},
}

func init() {
	lbCmd.AddCommand(loadBalancerDestroyCmd)
}

func destroyLoadBalancer(operation *LoadBalancerDestroyOperation) {
	console.Info("[%s] Destroying load balancer", operation.LoadBalancerName)

	elbv2 := ELBV2.New(sess)
	elbv2.DeleteLoadBalancer(operation.LoadBalancerName)
	elbv2.DeleteTargetGroup(fmt.Sprintf(defaultTargetGroupFormat, operation.LoadBalancerName))
}
