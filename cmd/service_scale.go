package cmd

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/jpignata/fargate/console"
	ECS "github.com/jpignata/fargate/ecs"
	"github.com/spf13/cobra"
)

const validScalePattern = "[-\\+]?[0-9]+"

type ScaleServiceOperation struct {
	ServiceName  string
	DesiredCount int64
}

func (o *ScaleServiceOperation) SetScale(scaleExpression string) {
	ecs := ECS.New(sess, clusterName)
	validScale := regexp.MustCompile(validScalePattern)

	if !validScale.MatchString(scaleExpression) {
		console.ErrorExit(fmt.Errorf("Invalid scale expression %s", scaleExpression), "Invalid command line argument")
	}

	if scaleExpression[0] == '+' || scaleExpression[0] == '-' {
		if s, err := strconv.ParseInt(scaleExpression[1:len(scaleExpression)], 10, 64); err == nil {
			currentDesiredCount := ecs.GetDesiredCount(o.ServiceName)
			if scaleExpression[0] == '+' {
				o.DesiredCount = currentDesiredCount + s
			} else if scaleExpression[0] == '-' {
				o.DesiredCount = currentDesiredCount - s
			}
		}
	} else if s, err := strconv.ParseInt(scaleExpression, 10, 64); err == nil {
		o.DesiredCount = s
	} else {
		console.ErrorExit(fmt.Errorf("Invalid scale expression %s", scaleExpression), "Invalid command line argument")
	}

	if o.DesiredCount < 0 {
		console.ErrorExit(fmt.Errorf("requested scale %d < 0", o.DesiredCount), "Invalid command line argument")
	}
}

var serviceScaleCmd = &cobra.Command{
	Use:   "scale <service-name> <scale-expression>",
	Short: "Changes the number of tasks running for the service",
	Long: `Scale number of tasks in a service

Changes the number of desired tasks to be run in a service by the given scale
expression. A scale expression can either be an absolute number or a delta
specified with a sign such as +5 or -2.`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		operation := &ScaleServiceOperation{
			ServiceName: args[0],
		}

		operation.SetScale(args[1])

		scaleService(operation)
	},
}

func init() {
	serviceCmd.AddCommand(serviceScaleCmd)
}

func scaleService(operation *ScaleServiceOperation) {
	ecs := ECS.New(sess, clusterName)

	ecs.SetDesiredCount(operation.ServiceName, operation.DesiredCount)
	console.Info("Scaled service %s to %d", operation.ServiceName, operation.DesiredCount)
}
