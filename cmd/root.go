package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/jpignata/fargate/console"
	ECS "github.com/jpignata/fargate/ecs"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
)

const (
	clusterName   = "fargate"
	defaultRegion = "us-east-1"
	version       = "0.1.0"

	mebibytesInGibibyte = 1024
	cpuUnitsInVCpu      = 1024

	validProtocolsPattern = "(?i)\\ATCP|HTTP(S)?\\z"
)

var InvalidCpuAndMemoryCombination = fmt.Errorf(`Invalid CPU and Memory settings

CPU (CPU Units)    Memory (MiB)
---------------    ------------
256                512, 1024, or 2048
512                1024 through 4096 in 1GB increments
1024               2048 through 8192 in 1GB increments
2048               4096 through 16384 in 1GB increments
4096               8192 through 30720 in 1GB increments
`)

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
		if cmd.Parent().Name() == "fargate" {
			return
		}

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
		err := ecs.CreateCluster()

		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case "NoCredentialProviders":
				console.Issue("Could not find your AWS credentials")
				console.Info("Your AWS credentials could not be found. Please configure your environment with your access key")
				console.Info("   ID and secret access key using either the shared configuration file or environment variables.")
				console.Info("   See http://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html#specifying-credentials")
				console.Info("   for more details.")
				console.Exit(1)
			default:
				console.ErrorExit(err, "Could not create ECS cluster")
			}
		}

		if verbose {
			verbose = true
			console.Verbose = true
		}

		if noColor || !terminal.IsTerminal(int(os.Stdout.Fd())) {
			console.Color = false
		}
	},
}

func Execute() {
	rootCmd.Version = version
	rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	rootCmd.PersistentFlags().StringVar(&region, "region", "", "AWS Region (defaults to us-east-1)")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Suppress colors in output")
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

func extractEnvVars(inputEnvVars []string) []ECS.EnvVar {
	var envVars []ECS.EnvVar

	if len(inputEnvVars) == 0 {
		return envVars
	}

	for _, inputEnvVar := range inputEnvVars {
		splitInputEnvVar := strings.SplitN(inputEnvVar, "=", 2)

		if len(splitInputEnvVar) != 2 {
			console.ErrorExit(fmt.Errorf("%s must be in the form of KEY=value", inputEnvVar), "Invalid environment variable")
		}

		envVar := ECS.EnvVar{
			Key:   strings.ToUpper(splitInputEnvVar[0]),
			Value: splitInputEnvVar[1],
		}

		envVars = append(envVars, envVar)
	}

	return envVars
}

func validateCpuAndMemory(inputCpuUnits, inputMebibytes string) error {
	cpuUnits, err := strconv.ParseInt(inputCpuUnits, 10, 16)

	if err != nil {
		return err
	}

	mebibytes, err := strconv.ParseInt(inputMebibytes, 10, 16)

	if err != nil {
		return err
	}

	switch cpuUnits {
	case 256:
		if mebibytes == 512 || validateMebibytes(mebibytes, 1024, 2048) {
			return nil
		}
	case 512:
		if validateMebibytes(mebibytes, 1024, 4096) {
			return nil
		}
	case 1024:
		if validateMebibytes(mebibytes, 2048, 8192) {
			return nil
		}
	case 2048:
		if validateMebibytes(mebibytes, 4096, 16384) {
			return nil
		}
	case 4096:
		if validateMebibytes(mebibytes, 8192, 30720) {
			return nil
		}
	}

	return InvalidCpuAndMemoryCombination
}

func validateMebibytes(mebibytes, min, max int64) bool {
	return mebibytes >= min && mebibytes <= max && mebibytes%mebibytesInGibibyte == 0
}
