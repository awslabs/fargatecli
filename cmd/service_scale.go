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
	Ecs          ECS.ECS
}

func (o *ScaleServiceOperation) SetScale(scaleExpression string) {
	validScale := regexp.MustCompile(validScalePattern)

	if !validScale.MatchString(scaleExpression) {
		console.ErrorExit(fmt.Errorf("Invalid scale expression %s", scaleExpression), "Invalid command line argument")
	}

	if scaleExpression[0] == '+' || scaleExpression[0] == '-' {
		if s, err := strconv.ParseInt(scaleExpression[1:len(scaleExpression)], 10, 64); err == nil {
			currentDesiredCount := o.Ecs.GetDesiredCount(o.ServiceName)
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
	Use:   "scale <service name> <scalei expression>",
	Short: "Changes the number of tasks running for the service",
	Long: `Changes the number of tasks running for the service

For scale expression, specify either an absolute number e.g. 5 or a delta such
as +2 or -4.  The desired count will be changed to reflect the new scale of the
service. The new desired count must be above zero.`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		operation := &ScaleServiceOperation{
			ServiceName: args[0],
			Ecs:         ECS.New(sess),
		}

		operation.SetScale(args[1])

		scaleService(operation)
	},
}

func init() {
	serviceCmd.AddCommand(serviceScaleCmd)
}

func scaleService(operation *ScaleServiceOperation) {
	console.Info("Scaling service %s to %d", operation.ServiceName, operation.DesiredCount)
	operation.Ecs.SetDesiredCount(operation.ServiceName, operation.DesiredCount)
}
