package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

const (
	serviceLogGroupFormat      = "/fargate/service/%s"
	serviceEnvironmentVariable = "FARGATE_SERVICE"
)

var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "Manage services",
	Long: `Manage services

Services manage long-lived instances of your containers that are run on AWS
Fargate. If your container exits for any reason, the service scheduler will
restart your containers and ensure your service has the desired number of
tasks running. Services can be used in concert with a load balancer to
distribute traffic amongst the tasks in your service.`,
}

func init() {
	rootCmd.AddCommand(serviceCmd)
}

var serviceName string

//look for service first from cli args then envvars
func setServiceName(cmd *cobra.Command, args []string) {
	result := ""
	if len(args) > 0 {
		result = args[0]
	} else {
		result = os.Getenv(serviceEnvironmentVariable)
	}
	if result == "" {
		cmd.Usage()
		os.Exit(-1)
	}
	serviceName = result
}
