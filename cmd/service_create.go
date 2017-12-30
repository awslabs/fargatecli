package cmd

import (
	"fmt"
	"regexp"
	"strings"

	CWL "github.com/jpignata/fargate/cloudwatchlogs"
	"github.com/jpignata/fargate/console"
	"github.com/jpignata/fargate/docker"
	EC2 "github.com/jpignata/fargate/ec2"
	ECR "github.com/jpignata/fargate/ecr"
	ECS "github.com/jpignata/fargate/ecs"
	ELBV2 "github.com/jpignata/fargate/elbv2"
	"github.com/jpignata/fargate/git"
	IAM "github.com/jpignata/fargate/iam"
	"github.com/spf13/cobra"
)

const validRuleTypesPattern = "(?i)^host|path$"

type ServiceCreateOperation struct {
	ServiceName      string
	Cpu              string
	Image            string
	Memory           string
	Port             Port
	LoadBalancerArn  string
	LoadBalancerName string
	Rules            []ELBV2.Rule
	Elbv2            ELBV2.ELBV2
	EnvVars          []ECS.EnvVar
	Num              int64
}

func (o *ServiceCreateOperation) SetPort(inputPort string) {
	var msgs []string

	port := inflatePort(inputPort)
	validProtocols := regexp.MustCompile(validProtocolsPattern)

	if !validProtocols.MatchString(port.Protocol) {
		msgs = append(msgs, fmt.Sprintf("Invalid protocol %s [specify TCP, HTTP, or HTTPS]", port.Protocol))
	}

	if port.Port < 1 || port.Port > 65535 {
		msgs = append(msgs, fmt.Sprintf("Invalid port %d [specify within 1 - 65535]", port.Port))
	}

	if len(msgs) > 0 {
		console.ErrorExit(fmt.Errorf(strings.Join(msgs, ", ")), "Invalid command line flags")
	}

	o.Port = port
}

func (o *ServiceCreateOperation) Validate() {
	err := validateCpuAndMemory(o.Cpu, o.Memory)

	if err != nil {
		console.ErrorExit(err, "Invalid settings: %s CPU units / %s MiB", o.Cpu, o.Memory)
	}

	if o.Num < 1 {
		console.ErrorExit(err, "Invalid number of tasks to keep running: %d, num must be > 1", o.Num)
	}
}

func (o *ServiceCreateOperation) SetLoadBalancer(lb string) {
	loadBalancer := o.Elbv2.DescribeLoadBalancer(lb)

	if loadBalancer.Type == "network" {
		if o.Port.Protocol != "TCP" {
			console.ErrorExit(fmt.Errorf("network load balancer %s only supports TCP", lb), "Invalid load balancer and protocol")
		}
	}

	if loadBalancer.Type == "application" {
		if !(o.Port.Protocol == "HTTP" || o.Port.Protocol == "HTTPS") {
			console.ErrorExit(fmt.Errorf("application load balancer %s only supports HTTP or HTTPS", lb), "Invalid load balancer and protocol")
		}
	}

	o.LoadBalancerName = lb
	o.LoadBalancerArn = loadBalancer.Arn
}

func (o *ServiceCreateOperation) SetRules(inputRules []string) {
	var rules []ELBV2.Rule
	var msgs []string

	validRuleTypes := regexp.MustCompile(validRuleTypesPattern)

	if len(inputRules) > 0 && o.LoadBalancerArn == "" {
		msgs = append(msgs, "lb must be configured if rules are specified")
	}

	for _, inputRule := range inputRules {
		splitInputRule := strings.SplitN(inputRule, "=", 2)

		if len(splitInputRule) != 2 {
			msgs = append(msgs, "rules must be in the form of type=value")
		}

		if !validRuleTypes.MatchString(splitInputRule[0]) {
			msgs = append(msgs, fmt.Sprintf("Invalid rule type %s [must be path or host]", splitInputRule[0]))
		}

		rules = append(rules,
			ELBV2.Rule{
				Type:  strings.ToUpper(splitInputRule[0]),
				Value: splitInputRule[1],
			},
		)
	}

	if len(msgs) > 0 {
		console.ErrorExit(fmt.Errorf(strings.Join(msgs, ", ")), "Invalid rule")
	}

	o.Rules = rules
}

func (o *ServiceCreateOperation) SetEnvVars(inputEnvVars []string) {
	o.EnvVars = extractEnvVars(inputEnvVars)
}

var (
	flagServiceCreateCpu     string
	flagServiceCreateEnvVars []string
	flagServiceCreateImage   string
	flagServiceCreateLb      string
	flagServiceCreateMemory  string
	flagServiceCreatePort    string
	flagServiceCreateRules   []string
	flagServiceCreateNum     int64
)

var serviceCreateCmd = &cobra.Command{
	Use:   "create <service-name>",
	Short: "Create a new service",
	Long: `Create a new service

CPU and memory settings can be optionally specified as CPU units and mebibytes
respectively using the --cpu and --memory flags. Every 1024 CPU units is
equivilent to a single vCPU. AWS Fargate only supports certain combinations of
CPU and memory configurations:

| CPU (CPU Units) | Memory (MiB)                          |
| --------------- | ------------------------------------- |
| 256             | 512, 1024, or 2048                    |
| 512             | 1024 through 4096 in 1GiB increments  |
| 1024            | 2048 through 8192 in 1GiB increments  |
| 2048            | 4096 through 16384 in 1GiB increments |
| 4096            | 8192 through 30720 in 1GiB increments |

If not specified, fargate will launch minimally sized tasks at 0.25 vCPU (256
CPU units) and 0.5GB (512 MiB) of memory.

The Docker container image to use in the service can be optionally specified
via the --image flag. If not specified, fargate will build a new Docker
container image from the current working directory and push it to Amazon ECR in
a repository named for the task group. If the current working directory is a
git repository, the container image will be tagged with the short ref of the
HEAD commit. If not, a timestamp in the format of YYYYMMDDHHMMSS will be used.

Services can optionally be configured to use a load balancer. To put a load
balancer in front a service, pass the --lb flag with the name of a load
balancer. If you specify a load balancer, you must also specify a port via the
--port flag to which the load balancer should forward requests. Optionally,
Application Load Balancers can be configured to route HTTP/HTTPS traffic to the
service based upon a rule. Rules are configured by passing one or more rules by
specifying the --rule flag along with a rule expression. Rule expressions are
in the format of TYPE=VALUE. Type can either be PATH or HOST. PATH matches the
PATH of the request and HOST matches the requested hostname in the HTTP
request. Both PATH and HOST types can include up to three wildcard characters:
* to match multiple characters and ? to match a single character.

Environment variables can be specified via the --env flag. Specify --env with a
key=value parameter multiple times to add multiple variables.

Specify the desired count of tasks the service should maintain by passing the
--num flag with a number. If you omit this flag, fargate will configure a
service with a desired number of tasks of 1.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		operation := &ServiceCreateOperation{
			Cpu:         flagServiceCreateCpu,
			Elbv2:       ELBV2.New(sess),
			Image:       flagServiceCreateImage,
			Memory:      flagServiceCreateMemory,
			Num:         flagServiceCreateNum,
			ServiceName: args[0],
		}

		if flagServiceCreatePort != "" {
			operation.SetPort(flagServiceCreatePort)
		}

		if flagServiceCreateLb != "" {
			operation.SetLoadBalancer(flagServiceCreateLb)
		}

		if len(flagServiceCreateRules) > 0 {
			operation.SetRules(flagServiceCreateRules)
		}

		if len(flagServiceCreateEnvVars) > 0 {
			operation.SetEnvVars(flagServiceCreateEnvVars)
		}

		operation.Validate()
		createService(operation)
	},
}

func init() {
	serviceCreateCmd.Flags().StringVarP(&flagServiceCreateCpu, "cpu", "c", "256", "Amount of cpu units to allocate for each task")
	serviceCreateCmd.Flags().StringVarP(&flagServiceCreateMemory, "memory", "m", "512", "Amount of MiB to allocate for each task")
	serviceCreateCmd.Flags().StringSliceVarP(&flagServiceCreateEnvVars, "env", "e", []string{}, "Environment variables to set [e.g. KEY=value] (can be specified multiple times)")
	serviceCreateCmd.Flags().StringVarP(&flagServiceCreatePort, "port", "p", "", "Port to listen on [e.g., 80, 443, http:8080, https:8443, tcp:1935]")
	serviceCreateCmd.Flags().StringVarP(&flagServiceCreateImage, "image", "i", "", "Docker image to run in the service; if omitted Fargate will build an image from the Dockerfile in the current directory")
	serviceCreateCmd.Flags().StringVarP(&flagServiceCreateLb, "lb", "l", "", "Name of a load balancer to use")
	serviceCreateCmd.Flags().StringSliceVarP(&flagServiceCreateRules, "rule", "r", []string{}, "Routing rule for the load balancer [e.g. host=api.example.com, path=/api/*]; if omitted service will be the default route")
	serviceCreateCmd.Flags().Int64VarP(&flagServiceCreateNum, "num", "n", 1, "Number of tasks instances to keep running")

	serviceCmd.AddCommand(serviceCreateCmd)
}

func createService(operation *ServiceCreateOperation) {
	cwl := CWL.New(sess)
	ec2 := EC2.New(sess)
	ecr := ECR.New(sess)
	ecs := ECS.New(sess, clusterName)
	iam := IAM.New(sess)

	var (
		targetGroupArn string
		repositoryUri  string
	)

	if ecr.IsRepositoryCreated(operation.ServiceName) {
		repositoryUri = ecr.GetRepositoryUri(operation.ServiceName)
	} else {
		repositoryUri = ecr.CreateRepository(operation.ServiceName)
	}

	repository := docker.Repository{Uri: repositoryUri}
	subnetIds := ec2.GetDefaultVpcSubnetIds()
	securityGroupId := ec2.GetDefaultSecurityGroupId()
	ecsTaskExecutionRoleArn := iam.CreateEcsTaskExecutionRole()
	logGroupName := cwl.CreateLogGroup(serviceLogGroupFormat, operation.ServiceName)

	if operation.Image == "" {
		var tag string

		username, password := ecr.GetUsernameAndPassword()

		if git.IsCwdGitRepo() {
			tag = git.GetShortSha()
		} else {
			tag = docker.GenerateTag()
		}

		repository.Login(username, password)
		repository.Build(tag)
		repository.Push(tag)

		operation.Image = repository.UriFor(tag)
	}

	if operation.LoadBalancerArn != "" {
		vpcId := ec2.GetDefaultVpcId()
		targetGroupArn = operation.Elbv2.CreateTargetGroup(
			&ELBV2.CreateTargetGroupInput{
				Name:     operation.LoadBalancerName + "-" + operation.ServiceName,
				Port:     operation.Port.Port,
				Protocol: operation.Port.Protocol,
				VpcId:    vpcId,
			},
		)

		if len(operation.Rules) > 0 {
			for _, rule := range operation.Rules {
				operation.Elbv2.AddRule(operation.LoadBalancerArn, targetGroupArn, rule)
			}
		} else {
			operation.Elbv2.ModifyLoadBalancerDefaultAction(operation.LoadBalancerArn, targetGroupArn)
		}
	}

	taskDefinitionArn := ecs.CreateTaskDefinition(
		&ECS.CreateTaskDefinitionInput{
			Cpu:              operation.Cpu,
			EnvVars:          operation.EnvVars,
			ExecutionRoleArn: ecsTaskExecutionRoleArn,
			Image:            operation.Image,
			Memory:           operation.Memory,
			Name:             operation.ServiceName,
			Port:             operation.Port.Port,
			LogGroupName:     logGroupName,
			LogRegion:        region,
			Type:             "service",
		},
	)

	ecs.CreateService(
		&ECS.CreateServiceInput{
			Cluster:           clusterName,
			Name:              operation.ServiceName,
			Port:              operation.Port.Port,
			SubnetIds:         subnetIds,
			TargetGroupArn:    targetGroupArn,
			TaskDefinitionArn: taskDefinitionArn,
			DesiredCount:      operation.Num,
			SecurityGroupId:   securityGroupId,
		},
	)

	console.Info("Created service %s", operation.ServiceName)
}
