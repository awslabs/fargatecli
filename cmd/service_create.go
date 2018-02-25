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

const typeService = "service"

type ServiceCreateOperation struct {
	Cpu              string
	EnvVars          []ECS.EnvVar
	Image            string
	LoadBalancerArn  string
	LoadBalancerName string
	Memory           string
	Num              int64
	Port             Port
	Rules            []ELBV2.Rule
	SecurityGroupIds []string
	ServiceName      string
	SubnetIds        []string
	TaskRole         string
}

func (o *ServiceCreateOperation) SetPort(inputPort string) {
	var msgs []string

	port, _ := inflatePort(inputPort)

	if !validProtocol.MatchString(port.Protocol) {
		msgs = append(msgs, fmt.Sprintf("Invalid protocol %s [specify TCP, HTTP, or HTTPS]", port.Protocol))
	}

	if port.Number < 1 || port.Number > 65535 {
		msgs = append(msgs, fmt.Sprintf("Invalid port %d [specify within 1 - 65535]", port.Number))
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
	if o.Port.Empty() {
		console.IssueExit("Setting a load balancer requires a port")
	}

	elbv2 := ELBV2.New(sess)
	loadBalancer := elbv2.DescribeLoadBalancer(lb)

	if loadBalancer.Type == typeNetwork {
		if o.Port.Protocol != protocolTcp {
			console.ErrorExit(fmt.Errorf("network load balancer %s only supports TCP", lb), "Invalid load balancer and protocol")
		}
	}

	if loadBalancer.Type == typeApplication {
		if !(o.Port.Protocol == protocolHttp || o.Port.Protocol == protocolHttps) {
			console.ErrorExit(fmt.Errorf("application load balancer %s only supports HTTP or HTTPS", lb), "Invalid load balancer and protocol")
		}
	}

	o.LoadBalancerName = lb
	o.LoadBalancerArn = loadBalancer.ARN
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

func (o *ServiceCreateOperation) SetSecurityGroupIds(securityGroupIds []string) {
	o.SecurityGroupIds = securityGroupIds
}

var (
	flagServiceCreateCpu              string
	flagServiceCreateEnvVars          []string
	flagServiceCreateImage            string
	flagServiceCreateLb               string
	flagServiceCreateMemory           string
	flagServiceCreateNum              int64
	flagServiceCreatePort             string
	flagServiceCreateRules            []string
	flagServiceCreateSecurityGroupIds []string
	flagServiceCreateSubnetIds        []string
	flagServiceCreateTaskRole         string
)

var serviceCreateCmd = &cobra.Command{
	Use:   "create <service-name>",
	Short: "Create a service",
	Long: `Create a service

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

To use the service with a load balancer, a port must be specified when the
service is created. Specify a port by passing the --port flag and a port
expression of protocol:port-number. For example, if the service listens on port
80 and uses HTTP, specify HTTP:80.  Valid protocols are HTTP, HTTPS, and TCP.
You can only specify a single port.

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
* to match multiple characters and ? to match a single character. If rules are
omitted, the service will be the load balancer's default action.

Environment variables can be specified via the --env flag. Specify --env with a
key=value parameter multiple times to add multiple variables.

Specify the desired count of tasks the service should maintain by passing the
--num flag with a number. If you omit this flag, fargate will configure a
service with a desired number of tasks of 1.

Security groups can optionally be specified for the service by passing the
--security-group-id flag with a security group ID. To add multiple security
groups, pass --security-group-id with a security group ID multiple times. If
--security-group-id is omitted, a permissive security group will be applied to
the service.

By default, the service will be created in the default VPC and attached
to the default VPC subnets for each availability zone. You can override this by
specifying explicit subnets by passing the --subnet-id flag with a subnet ID.

A task role can be optionally specified via the --task-role flag by providing
eith a full IAM role ARN or the name of an IAM role. The tasks run by the
service will be able to assume this role.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		operation := &ServiceCreateOperation{
			Cpu:              flagServiceCreateCpu,
			Image:            flagServiceCreateImage,
			Memory:           flagServiceCreateMemory,
			Num:              flagServiceCreateNum,
			SecurityGroupIds: flagServiceCreateSecurityGroupIds,
			ServiceName:      args[0],
			SubnetIds:        flagServiceCreateSubnetIds,
			TaskRole:         flagServiceCreateTaskRole,
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
	serviceCreateCmd.Flags().StringSliceVarP(&flagServiceCreateRules, "rule", "r", []string{}, "Routing rule for the load balancer [e.g. host=api.example.com, path=/api/*]; if omitted service will be the default route (can be specified multiple times)")
	serviceCreateCmd.Flags().Int64VarP(&flagServiceCreateNum, "num", "n", 1, "Number of tasks instances to keep running")
	serviceCreateCmd.Flags().StringSliceVar(&flagServiceCreateSecurityGroupIds, "security-group-id", []string{}, "ID of a security group to apply to the service (can be specified multiple times)")
	serviceCreateCmd.Flags().StringSliceVar(&flagServiceCreateSubnetIds, "subnet-id", []string{}, "ID of a subnet in which to place the service (can be specified multiple times)")
	serviceCreateCmd.Flags().StringVarP(&flagServiceCreateTaskRole, "task-role", "", "", "Name or ARN of an IAM role that the service's tasks can assume")

	serviceCmd.AddCommand(serviceCreateCmd)
}

func createService(operation *ServiceCreateOperation) {
	var targetGroupArn string

	cwl := CWL.New(sess)
	ec2 := EC2.New(sess)
	ecr := ECR.New(sess)
	elbv2 := ELBV2.New(sess)
	ecs := ECS.New(sess, clusterName)
	iam := IAM.New(sess)
	ecsTaskExecutionRoleArn := iam.CreateEcsTaskExecutionRole()
	logGroupName := cwl.CreateLogGroup(serviceLogGroupFormat, operation.ServiceName)

	if len(operation.SecurityGroupIds) == 0 {
		defaultSecurityGroupID, _ := ec2.GetDefaultSecurityGroupID()
		operation.SecurityGroupIds = []string{defaultSecurityGroupID}
	}

	if len(operation.SubnetIds) == 0 {
		operation.SubnetIds, _ = ec2.GetDefaultSubnetIDs()
	}

	if operation.Image == "" {
		var tag, repositoryUri string

		if ecr.IsRepositoryCreated(operation.ServiceName) {
			repositoryUri = ecr.GetRepositoryUri(operation.ServiceName)
		} else {
			repositoryUri = ecr.CreateRepository(operation.ServiceName)
		}

		if git.IsCwdGitRepo() {
			tag = git.GetShortSha()
		} else {
			tag = docker.GenerateTag()
		}

		repository := docker.NewRepository(repositoryUri)
		username, password := ecr.GetUsernameAndPassword()

		repository.Login(username, password)
		repository.Build(tag)
		repository.Push(tag)

		operation.Image = repository.UriFor(tag)
	}

	if operation.LoadBalancerArn != "" {
		vpcId, _ := ec2.GetSubnetVPCID(operation.SubnetIds[0])
		targetGroupArn, _ = elbv2.CreateTargetGroup(
			ELBV2.CreateTargetGroupParameters{
				Name:     fmt.Sprintf("%s-%s", clusterName, operation.ServiceName),
				Port:     operation.Port.Number,
				Protocol: operation.Port.Protocol,
				VPCID:    vpcId,
			},
		)

		if len(operation.Rules) > 0 {
			for _, rule := range operation.Rules {
				elbv2.AddRule(operation.LoadBalancerArn, targetGroupArn, rule)
			}
		} else {
			elbv2.ModifyLoadBalancerDefaultAction(operation.LoadBalancerArn, targetGroupArn)
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
			Port:             operation.Port.Number,
			LogGroupName:     logGroupName,
			LogRegion:        region,
			TaskRole:         operation.TaskRole,
			Type:             typeService,
		},
	)

	ecs.CreateService(
		&ECS.CreateServiceInput{
			Cluster:           clusterName,
			DesiredCount:      operation.Num,
			Name:              operation.ServiceName,
			Port:              operation.Port.Number,
			SecurityGroupIds:  operation.SecurityGroupIds,
			SubnetIds:         operation.SubnetIds,
			TargetGroupArn:    targetGroupArn,
			TaskDefinitionArn: taskDefinitionArn,
		},
	)

	console.Info("Created service %s", operation.ServiceName)
}
