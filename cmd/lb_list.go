package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/jpignata/fargate/console"
	ELBV2 "github.com/jpignata/fargate/elbv2"
	"github.com/spf13/cobra"
)

var lbListCmd = &cobra.Command{
	Use:   "list",
	Short: "List load balancers",
	Run: func(cmd *cobra.Command, args []string) {
		listLoadBalancers()
	},
}

func init() {
	lbCmd.AddCommand(lbListCmd)
}

func listLoadBalancers() {
	elbv2 := ELBV2.New(sess)
	loadBalancers := elbv2.DescribeLoadBalancers(ELBV2.DescribeLoadBalancersInput{})

	if len(loadBalancers) == 0 {
		console.InfoExit("No load balancers found")
	}

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 1, '\t', 0)
	fmt.Fprintln(w, "NAME\tTYPE\tSTATUS\tDNS NAME\tPORTS")

	for _, loadBalancer := range loadBalancers {
		var listeners []string

		for _, listener := range elbv2.GetListeners(loadBalancer.Arn) {
			listeners = append(listeners, listener.String())
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			loadBalancer.Name,
			Humanize(loadBalancer.Type),
			Humanize(loadBalancer.State),
			loadBalancer.DNSName,
			strings.Join(listeners, ", "),
		)
	}

	w.Flush()
}
