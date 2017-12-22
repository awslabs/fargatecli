package cmd

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/jpignata/fargate/console"
	ECS "github.com/jpignata/fargate/ecs"
	"github.com/spf13/cobra"
)

var desiredCount int64

var validScale = regexp.MustCompile("[-\\+]?[0-9]+")

var serviceScaleCmd = &cobra.Command{
	Use:   "scale <service name> <scale>",
	Short: "Changes the number of containers running for the service",
	Long: `Changes the number of containers running for the service

For scale, specify either an absolute number e.g. 5 or a delta such as +2 or -4.
The desired count will be changed to reflect the new scale of the service. If
your new scale would bring your desired count under 0, and error will be
returned.`,
	Args: cobra.ExactArgs(2),
	PreRun: func(cmd *cobra.Command, args []string) {
		validateScale(args[1])
		setDesiredCount(args[0], args[1])
	},
	Run: func(cmd *cobra.Command, args []string) {
		scaleService(args[0])
	},
}

func init() {
	serviceCmd.AddCommand(serviceScaleCmd)
}

func validateScale(scale string) {
	if !validScale.MatchString(scale) {
		console.ErrorExit(fmt.Errorf("Invalid scale expression %s", scale), "Invalid command line argument")
	}
}

func setDesiredCount(serviceName, scale string) {
	if scale[0] == '+' || scale[0] == '-' {
		if s, err := strconv.ParseInt(scale[1:len(scale)], 10, 64); err == nil {
			ecs := ECS.New()
			desiredCount = ecs.GetDesiredCount(serviceName)

			if scale[0] == '+' {
				desiredCount = desiredCount + s
			} else if scale[0] == '-' {
				desiredCount = desiredCount - s

				if desiredCount < 0 {
					console.ErrorExit(fmt.Errorf("requested scale %d < 0", desiredCount), "Invalid command line argument")
				}
			}

			return
		}
	}

	if s, err := strconv.ParseInt(scale, 10, 64); err == nil {
		desiredCount = s
		return
	}

	console.ErrorExit(fmt.Errorf("Invalid scale expression %s", scale), "Invalid command line argument")
}

func scaleService(serviceName string) {
	console.Info("Scaling service %s to %d", serviceName, desiredCount)

	ecs := ECS.New()
	ecs.SetDesiredCount(serviceName, desiredCount)
}
