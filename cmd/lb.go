package cmd

import (
	"errors"

	"github.com/jpignata/fargate/elbv2"
	"github.com/spf13/cobra"
)

const defaultTargetGroupFormat = "%s-default"

type lbOperation struct {
	elbv2 elbv2.Client
}

func (o lbOperation) findLB(lbName string, output Output) (elbv2.LoadBalancer, error) {
	output.Debug("Finding load balancer[API=elb2 Action=DescribeLoadBalancers]")
	loadBalancers, err := o.elbv2.DescribeLoadBalancersByName([]string{lbName})

	if err != nil {
		return elbv2.LoadBalancer{}, err
	}

	switch {
	case len(loadBalancers) == 0:
		return elbv2.LoadBalancer{}, errLBNotFound
	case len(loadBalancers) > 1:
		return elbv2.LoadBalancer{}, errLBTooManyFound
	}

	return loadBalancers[0], nil
}

var (
	errLBNotFound     = errors.New("load balancer not found")
	errLBTooManyFound = errors.New("too many load balancers found")

	lbCmd = &cobra.Command{
		Use:   "lb",
		Short: "Manage load balancers",
		Long: `Manage load balancers

Load balancers distribute incoming traffic between the tasks within a service
for HTTP/HTTPS and TCP applications. HTTP/HTTPS load balancers can route to
multiple services based upon rules you specify when you create a new service.`,
	}
)

func init() {
	rootCmd.AddCommand(lbCmd)
}
