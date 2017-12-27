package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	ACM "github.com/jpignata/fargate/acm"
	"github.com/jpignata/fargate/console"
	ECS "github.com/jpignata/fargate/ecs"
	ELBV2 "github.com/jpignata/fargate/elbv2"
	"github.com/jpignata/fargate/util"
	"github.com/spf13/cobra"
)

type LbInfoOperation struct {
	LoadBalancerName string
}

var lbInfoCmd = &cobra.Command{
	Use:   "info <load balancer name>",
	Short: "Display information about a load balancer",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		operation := &LbInfoOperation{
			LoadBalancerName: args[0],
		}

		getLoadBalancerInfo(operation)
	},
}

func init() {
	lbCmd.AddCommand(lbInfoCmd)
}

func getLoadBalancerInfo(operation *LbInfoOperation) {
	elbv2 := ELBV2.New(sess)
	acm := ACM.New(sess)
	ecs := ECS.New(sess)
	loadBalancer := elbv2.DescribeLoadBalancer(operation.LoadBalancerName)
	services := ecs.ListServices()

	console.KeyValue("Load Balancer Name", "%s\n", loadBalancer.Name)
	console.KeyValue("Status", "%s\n", util.Humanize(loadBalancer.State))
	console.KeyValue("Type", "%s\n", util.Humanize(loadBalancer.Type))
	console.KeyValue("DNS Name", "%s\n", loadBalancer.DNSName)
	console.KeyValue("Listeners", "\n")

	for _, listener := range elbv2.GetListeners(loadBalancer.Arn) {
		var ruleCount int

		console.KeyValue("  "+listener.String(), "\n")

		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 4, 2, 2, ' ', 0)

		console.KeyValue("    Rules", "\n")

		for _, rule := range elbv2.DescribeRules(listener.Arn) {
			var priority string

			serviceName := fmt.Sprintf("Unknown (%s)", rule.TargetGroupArn)

			if strings.Contains(rule.TargetGroupArn, fmt.Sprintf("/%s-default/", loadBalancer.Name)) {
				continue
			}

			for _, service := range services {
				if service.TargetGroupArn == rule.TargetGroupArn {
					serviceName = service.Name
				}
			}

			if rule.Priority != "" {
				priority = rule.Priority
			} else {
				priority = "0"
			}

			fmt.Fprintf(w, "     %s\t%s\t%s\n", priority, rule.String(), serviceName)

			ruleCount++
		}

		if ruleCount == 0 {
			fmt.Println("      None")
		}

		w.Flush()

		if len(listener.CertificateArns) > 0 {
			certificateDomains := acm.ListCertificateDomainNames(listener.CertificateArns)
			console.KeyValue("    Certificates", "%s\n", strings.Join(certificateDomains, ", "))
		}
	}

}
