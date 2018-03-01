package cmd

import (
	"fmt"

	"github.com/jpignata/fargate/elbv2"
	"github.com/spf13/cobra"
)

type lbDestroyOperation struct {
	lbOperation
	output Output
	elbv2  elbv2.Client
	lbName string
}

func (o lbDestroyOperation) execute() {
	loadBalancer, err := o.findLB(o.lbName)

	if err != nil {
		o.output.Fatal(err, "Could not destroy load balancer")
		return
	}

	o.output.Debug("Describing Listeners [API=elbv2 Action=DescribeListeners ARN=%s]", loadBalancer.ARN)
	if err := o.elbv2.InflateListeners(&loadBalancer); err != nil {
		o.output.Fatal(err, "Could not destroy load balancer")
		return
	}

	for _, listener := range loadBalancer.Listeners {
		o.output.Debug("Deleting Listener [API=elbv2 Action=DeleteListener ARN=%s]", listener.ARN)
		o.elbv2.DeleteListener(listener.ARN)
	}

	if err := o.elbv2.DeleteLoadBalancer(o.lbName); err != nil {
		o.output.Fatal(err, "Could not destroy load balancer")
		return
	}

	o.elbv2.DeleteTargetGroup(fmt.Sprintf(defaultTargetGroupFormat, o.lbName))

	o.output.Info("Destroyed load balancer %s", o.lbName)
}

var lbDestroyCmd = &cobra.Command{
	Use:   "destroy <load-balancer-name>",
	Short: "Destroy load balancer",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		lbDestroyOperation{
			elbv2:       elbv2.New(sess),
			lbName:      args[0],
			lbOperation: lbOperation{elbv2: elbv2.New(sess), output: output},
			output:      output,
		}.execute()
	},
}

func init() {
	lbCmd.AddCommand(lbDestroyCmd)
}
