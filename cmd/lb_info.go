package cmd

import (
	"fmt"
	"os"
	"sort"
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
	Use:   "info <load-balancer-name>",
	Short: "Inspect load balancer",
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
	ecs := ECS.New(sess, clusterName)
	loadBalancer := elbv2.DescribeLoadBalancer(operation.LoadBalancerName)
	services := ecs.ListServices()

	console.KeyValue("Load Balancer Name", "%s\n", loadBalancer.Name)
	console.KeyValue("Status", "%s\n", util.Humanize(loadBalancer.State))
	console.KeyValue("Type", "%s\n", util.Humanize(loadBalancer.Type))
	console.KeyValue("DNS Name", "%s\n", loadBalancer.DNSName)
	console.KeyValue("Subnets", "%s\n", strings.Join(loadBalancer.SubnetIds, ", "))
	console.KeyValue("Security Groups", "%s\n", strings.Join(loadBalancer.SecurityGroupIds, ", "))
	console.KeyValue("Ports", "\n")

	for _, listener := range elbv2.GetListeners(loadBalancer.Arn) {
		var ruleCount int

		console.KeyValue("  "+listener.String(), "\n")

		if len(listener.CertificateArns) > 0 {
			certificateDomains := acm.ListCertificateDomainNames(listener.CertificateArns)
			console.KeyValue("    Certificates", "%s\n", strings.Join(certificateDomains, ", "))
		}

		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 4, 2, 2, ' ', 0)

		console.KeyValue("    Rules", "\n")

		rules := elbv2.DescribeRules(listener.Arn)

		sort.Slice(rules, func(i, j int) bool { return rules[i].Priority > rules[j].Priority })

		for _, rule := range rules {
			serviceName := fmt.Sprintf("Unknown (%s)", rule.TargetGroupArn)

			if strings.Contains(rule.TargetGroupArn, fmt.Sprintf("/%s-default/", loadBalancer.Name)) {
				continue
			}

			for _, service := range services {
				if service.TargetGroupArn == rule.TargetGroupArn {
					serviceName = service.Name
				}
			}

			fmt.Fprintf(w, "     %d\t%s\t%s\n", rule.Priority, rule.String(), serviceName)

			ruleCount++
		}

		if ruleCount == 0 {
			fmt.Println("      None")
		}

		w.Flush()
	}

}
