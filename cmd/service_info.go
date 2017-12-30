package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	ACM "github.com/jpignata/fargate/acm"
	"github.com/jpignata/fargate/console"
	EC2 "github.com/jpignata/fargate/ec2"
	ECS "github.com/jpignata/fargate/ecs"
	ELBV2 "github.com/jpignata/fargate/elbv2"
	"github.com/jpignata/fargate/util"
	"github.com/spf13/cobra"
)

type ServiceInfoOperation struct {
	ServiceName string
}

var serviceInfoCmd = &cobra.Command{
	Use:   "info <service-name>",
	Short: "Inspect service",
	Long: `Inspect service

Show extended information for a service including load balancer configuration,
active deployments, and environment variables.

Deployments show active versions of your service that are running. Multiple
deployments are shown if a service is transitioning due to a deployment or
update to configuration such a CPU, memory, or environment variables.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		operation := &ServiceInfoOperation{
			ServiceName: args[0],
		}

		getServiceInfo(operation)
	},
}

func init() {
	serviceCmd.AddCommand(serviceInfoCmd)
}

func getServiceInfo(operation *ServiceInfoOperation) {
	var eniIds []string

	acm := ACM.New(sess)
	ecs := ECS.New(sess, clusterName)
	ec2 := EC2.New(sess)
	elbv2 := ELBV2.New(sess)
	service := ecs.DescribeService(operation.ServiceName)
	tasks := ecs.DescribeTasksForService(operation.ServiceName)

	console.KeyValue("Service Name", "%s\n", operation.ServiceName)
	console.KeyValue("Status", "\n")
	console.KeyValue("  Desired", "%d\n", service.DesiredCount)
	console.KeyValue("  Running", "%d\n", service.RunningCount)
	console.KeyValue("  Pending", "%d\n", service.PendingCount)
	console.KeyValue("Image", "%s\n", service.Image)
	console.KeyValue("Cpu", "%s\n", service.Cpu)
	console.KeyValue("Memory", "%s\n", service.Memory)

	if service.TargetGroupArn != "" {
		loadBalancerArn := elbv2.GetTargetGroupLoadBalancerArn(service.TargetGroupArn)
		loadBalancer := elbv2.DescribeLoadBalancerByArn(loadBalancerArn)
		listeners := elbv2.GetListeners(loadBalancerArn)

		console.KeyValue("Load Balancer", "\n")
		console.KeyValue("  Name", "%s\n", loadBalancer.Name)
		console.KeyValue("  DNS Name", "%s\n", loadBalancer.DNSName)

		if len(listeners) > 0 {
			console.KeyValue("  Listeners", "\n")
		}

		for _, listener := range listeners {
			var ruleOutput []string

			for _, rule := range elbv2.DescribeRules(listener.Arn) {
				if rule.TargetGroupArn == service.TargetGroupArn {
					ruleOutput = append(ruleOutput, rule.String())
				}
			}

			console.KeyValue("    "+listener.String(), "\n")
			console.KeyValue("      Rules", "%s\n", strings.Join(ruleOutput, ", "))

			if len(listener.CertificateArns) > 0 {
				certificateDomains := acm.ListCertificateDomainNames(listener.CertificateArns)
				console.KeyValue("      Certificates", "%s\n", strings.Join(certificateDomains, ", "))
			}
		}

		if len(service.EnvVars) > 0 {
			console.KeyValue("Environment Variables", "\n")

			for _, envVar := range service.EnvVars {
				fmt.Printf("   %s=%s\n", envVar.Key, envVar.Value)
			}
		}
	}

	if len(tasks) > 0 {
		console.Header("Tasks")

		for _, task := range tasks {
			if task.EniId != "" {
				eniIds = append(eniIds, task.EniId)
			}
		}

		enis := ec2.DescribeNetworkInterfaces(eniIds)
		w := new(tabwriter.Writer)

		w.Init(os.Stdout, 0, 8, 1, '\t', 0)
		fmt.Fprintln(w, "ID\tIMAGE\tSTATUS\tRUNNING\tIP\tCPU\tMEMORY\tDEPLOYMENT\t")

		for _, t := range tasks {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
				t.TaskId,
				t.Image,
				util.Humanize(t.LastStatus),
				t.RunningFor(),
				enis[t.EniId].PublicIpAddress,
				t.Cpu,
				t.Memory,
				t.DeploymentId,
			)
		}

		w.Flush()
	}

	if len(service.Deployments) > 0 {
		console.Header("Deployments")

		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 0, 8, 1, '\t', 0)
		fmt.Fprintln(w, "ID\tIMAGE\tSTATUS\tCREATED\tDESIRED\tRUNNING\tPENDING")

		for _, d := range service.Deployments {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\t%d\t%d\n",
				d.Id,
				d.Image,
				util.Humanize(d.Status),
				d.CreatedAt,
				d.DesiredCount,
				d.RunningCount,
				d.PendingCount,
			)
		}

		w.Flush()
	}

	if len(service.Events) > 0 {
		console.Header("Events")

		for i, event := range service.Events {
			fmt.Printf("[%s] %s\n", event.CreatedAt, event.Message)

			if i == 10 && !verbose {
				break
			}
		}
	}
}
