package cmd

/*
import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/jpignata/fargate/console"
	ECS "github.com/jpignata/fargate/ecs"
	"github.com/jpignata/fargate/util"
	"github.com/spf13/cobra"
)

var lbListCmd = &cobra.Command{
	Use:   "list",
	Short: "List services",
	Run: func(cmd *cobra.Command, args []string) {
		listServices()
	},
}

func init() {
	serviceCmd.AddCommand(serviceListCmd)
}

func listServices() {
	ecs = ECS.New()

	services := ecs.DescribeServices()

	if len(services) > 0 {
		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 0, 8, 1, '\t', 0)
		fmt.Fprintln(w, "Name\ttStatus\tDNS Name\tListeners")

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
*/
