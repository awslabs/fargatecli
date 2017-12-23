package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/jpignata/fargate/console"
	ELBV2 "github.com/jpignata/fargate/elbv2"
	"github.com/jpignata/fargate/util"
	"github.com/spf13/cobra"
)

var lbListCmd = &cobra.Command{
	Use:   "list",
	Short: "List requested and imported lbs",
	Long: `List requested and imported lbs

Shows all load balancers within Elastic Load Balancing and their attributes
such as type, status, and DNS name.

e.g.:

  $ fargate lb list
  NAME   TYPE          STATUS  DNS NAME                                    LISTENERS
  web    Application   Active  web-1186393493.us-east-1.elb.amazonaws.com  HTTP:80
`,
	Run: func(cmd *cobra.Command, args []string) {
		listLoadBalancers()
	},
}

func init() {
	lbCmd.AddCommand(lbListCmd)
}

func listLoadBalancers() {
	elbv2 := ELBV2.New()
	loadBalancers := elbv2.DescribeLoadBalancers([]string{})

	if len(loadBalancers) > 0 {
		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 0, 8, 1, '\t', 0)
		fmt.Fprintln(w, "NAME\tTYPE\tSTATUS\tDNS NAME\tLISTENERS")

		for _, loadBalancer := range loadBalancers {
			var listeners []string

			for _, listener := range elbv2.GetListeners(loadBalancer.Arn) {
				listeners = append(listeners, fmt.Sprintf("%s:%d", *listener.Protocol, *listener.Port))
			}

			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				loadBalancer.Name,
				util.Humanize(loadBalancer.Type),
				util.Humanize(loadBalancer.State),
				loadBalancer.DNSName,
				strings.Join(listeners, ", "),
			)
		}

		w.Flush()
	} else {
		console.Info("No load balancers found")
	}
}
