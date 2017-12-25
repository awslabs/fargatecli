package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/jpignata/fargate/console"
	ECS "github.com/jpignata/fargate/ecs"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
)

const (
	clusterName   = "fargate"
	defaultRegion = "us-east-1"
	version       = "0.0.1"
)

var validRegions = []string{"us-east-1"}

var (
	region     string
	verbose    bool
	sess       *session.Session
	envVars    []ECS.EnvVar
	envVarsRaw []string
	noColor    bool
)

type Port struct {
	Port     int64
	Protocol string
}

var rootCmd = &cobra.Command{
	Use: "fargate",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		envAwsDefaultRegion := os.Getenv("AWS_DEFAULT_REGION")
		envAwsRegion := os.Getenv("AWS_REGION")

		if region == "" {
			if envAwsDefaultRegion != "" {
				region = envAwsDefaultRegion
			} else if envAwsRegion != "" {
				region = envAwsRegion
			} else {
				region = defaultRegion
			}
		}

		for _, validRegion := range validRegions {
			if region == validRegion {
				break
			}

			console.IssueExit("Invalid region: %s [valid regions: %s]", region, strings.Join(validRegions, ", "))
		}

		sess = session.Must(
			session.NewSession(
				&aws.Config{
					Region: aws.String(region),
				},
			),
		)

		ecs := ECS.New(sess)
		ecs.CreateCluster()

		if verbose {
			verbose = true
			console.Verbose = true
		}

		if noColor || !terminal.IsTerminal(int(os.Stdout.Fd())) {
			console.Color = false
		}
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	rootCmd.PersistentFlags().StringVar(&region, "region", "", "AWS Region (defaults to us-east-1)")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Suppress colors in output")
}

func Execute() {
	rootCmd.Version = version
	rootCmd.Execute()
}

func inflatePort(src string) (port Port) {
	ports := inflatePorts([]string{src})
	return ports[0]
}

func inflatePorts(src []string) (ports []Port) {
	for _, portRaw := range src {
		if portRaw == "80" {
			ports = append(ports,
				Port{
					Port:     80,
					Protocol: "HTTP",
				},
			)
		} else if portRaw == "443" {
			ports = append(ports,
				Port{
					Port:     443,
					Protocol: "HTTPS",
				},
			)
		} else if strings.Index(portRaw, ":") > 1 {
			portRawContents := strings.Split(portRaw, ":")
			protocol := strings.ToUpper(portRawContents[0])
			port, err := strconv.ParseInt(portRawContents[1], 10, 64)

			if err != nil {
				console.ErrorExit(err, "Invalid command line flags")
			}

			ports = append(ports,
				Port{
					Port:     port,
					Protocol: protocol,
				},
			)
		} else {
			port, err := strconv.ParseInt(portRaw, 10, 64)

			if err != nil {
				console.ErrorExit(err, "Invalid command line flags")
			}

			ports = append(ports,
				Port{
					Port:     port,
					Protocol: "TCP",
				},
			)
		}
	}

	return
}

func extractEnvVars() {
	if len(envVarsRaw) == 0 {
		return
	}

	for _, envVar := range envVarsRaw {
		splitEnvVar := strings.Split(envVar, "=")

		if len(splitEnvVar) != 2 {
			console.ErrorExit(fmt.Errorf("%s must be in the form of KEY=value", envVar), "Invalid environment variable")
		}

		envVar := ECS.EnvVar{
			Key:   strings.ToUpper(splitEnvVar[0]),
			Value: splitEnvVar[1],
		}

		envVars = append(envVars, envVar)
	}
}
