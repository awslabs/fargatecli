package cmd

import (
	"fmt"
	"os"
	"runtime"
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
	version = "0.2.3"

	defaultClusterName = "fargate"
	defaultRegion      = "us-east-1"

	mebibytesInGibibyte   = 1024
	protocolHttp          = "HTTP"
	protocolHttps         = "HTTPS"
	protocolTcp           = "TCP"
	runtimeMacOS          = "darwin"
	typeApplication       = "application"
	typeNetwork           = "network"
	validRuleTypesPattern = "(?i)^host|path$"

	describeRequestLimitRate = 10
)

var InvalidCpuAndMemoryCombination = fmt.Errorf(`Invalid CPU and Memory settings

CPU (CPU Units)    Memory (MiB)
---------------    ------------
256                512, 1024, or 2048
512                1024 through 4096 in 1GiB increments
1024               2048 through 8192 in 1GiB increments
2048               4096 through 16384 in 1GiB increments
4096               8192 through 30720 in 1GiB increments
`)

var validRegions = []string{"us-east-1"}

var (
	clusterName string
	noColor     bool
	noEmoji     bool
	output      ConsoleOutput
	region      string
	sess        *session.Session
	verbose     bool
)

var rootCmd = &cobra.Command{
	Use:   "fargate",
	Short: "Deploy serverless containers onto the cloud from your command line",
	Long: `Deploy serverless containers onto the cloud from your command line

fargate is a command-line interface to deploy containers to AWS Fargate that
makes it easy to run containers in AWS as one-off tasks or managed, highly
available services secured by free TLS certificates. It bundles the power of AWS
including Amazon Elastic Container Service (ECS), Amazon Elastic Container
Registry (ECR), Elastic Load Balancing, AWS Certificate Manager, Amazon
CloudWatch Logs, and Amazon Route 53 into an easy-to-use CLI.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		output = ConsoleOutput{}

		if cmd.Parent().Name() == "fargate" {
			return
		}

		if verbose {
			verbose = true
			console.Verbose = true
			output.Verbose = true
		}

		if terminal.IsTerminal(int(os.Stdout.Fd())) {
			if !noColor {
				console.Color = true
				output.Color = true
			}

			if runtime.GOOS == runtimeMacOS && !noEmoji {
				output.Emoji = true
			}
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

		config := &aws.Config{
			Region: aws.String(region),
		}

		if verbose {
			config.LogLevel = aws.LogLevel(aws.LogDebugWithHTTPBody)
		}

		sess = session.Must(
			session.NewSession(config),
		)

		_, err := sess.Config.Credentials.Get()

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
				console.ErrorExit(err, "Could not create create AWS session")
			}
		}

		if clusterName == "" {
			clusterName = defaultClusterName
			ecs := ECS.New(sess, clusterName)

			output.Debug("Creating default cluster [API=ecs Action=CreateCluster]")

			arn, err := ecs.CreateCluster()

			if err == nil {
				output.Debug("Created default cluster [ARN=%s]", arn)
			} else {
				output.Fatal(err, "Could not create default cluster")
			}
		}
	},
}

func Execute() {
	rootCmd.Version = version
	rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	rootCmd.PersistentFlags().StringVar(&region, "region", "", `AWS region (default "us-east-1")`)
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable color output")
	rootCmd.PersistentFlags().StringVar(&clusterName, "cluster", "", `ECS cluster name (default "fargate")`)

	if runtime.GOOS == runtimeMacOS {
		rootCmd.PersistentFlags().BoolVar(&noEmoji, "no-emoji", false, "Disable emoji output")
	}
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
