package cmd

import (
	"fmt"

	"github.com/jpignata/fargate/console"
	ECS "github.com/jpignata/fargate/ecs"
	ELBV2 "github.com/jpignata/fargate/elbv2"
	"github.com/spf13/cobra"
)

type ServiceDestroyOperation struct {
	ServiceName string
}

var serviceDestroyCmd = &cobra.Command{
	Use:   "destroy <service-name>",
	Short: "Destroy a service",
	Long: `Destroy service

In order to destroy a service, it must first be scaled to 0 running tasks.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		operation := &ServiceDestroyOperation{
			ServiceName: args[0],
		}

		destroyService(operation)
	},
}

func init() {
	serviceCmd.AddCommand(serviceDestroyCmd)
}

func destroyService(operation *ServiceDestroyOperation) {
	elbv2 := ELBV2.New(sess)
	ecs := ECS.New(sess, clusterName)
	service := ecs.DescribeService(operation.ServiceName)

	if service.TargetGroupArn != "" {
		loadBalancerArn := elbv2.GetTargetGroupLoadBalancerArn(service.TargetGroupArn)
		loadBalancer := elbv2.DescribeLoadBalancerByArn(loadBalancerArn)
		listeners := elbv2.GetListeners(loadBalancerArn)

		for _, listener := range listeners {
			for _, rule := range elbv2.DescribeRules(listener.Arn) {
				if rule.TargetGroupArn == service.TargetGroupArn {
					if rule.IsDefault {
						defaultTargetGroupName := fmt.Sprintf(defaultTargetGroupFormat, loadBalancer.Name)
						defaultTargetGroupArn := elbv2.GetTargetGroupArn(defaultTargetGroupName)

						if defaultTargetGroupArn == "" {
							defaultTargetGroupArn = elbv2.CreateTargetGroup(
								&ELBV2.CreateTargetGroupInput{
									Name:     defaultTargetGroupName,
									Port:     listeners[0].Port,
									Protocol: listeners[0].Protocol,
									VpcId:    loadBalancer.VpcId,
								},
							)
						}

						elbv2.ModifyLoadBalancerDefaultAction(loadBalancer.Arn, defaultTargetGroupArn)
					} else {
						elbv2.DeleteRule(rule.Arn)
					}
				}
			}
		}

		elbv2.DeleteTargetGroupByArn(service.TargetGroupArn)
	}

	ecs.DestroyService(operation.ServiceName)
	console.Info("Destroyed service %s", operation.ServiceName)
}
