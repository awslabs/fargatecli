package cmd

import (
	"fmt"
	"regexp"
	"strconv"
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
	"github.com/jpignata/fargate/service"
	"github.com/spf13/cobra"
)

const logGroupFormat = "/fargate/service/%s"

var validRuleTypes = regexp.MustCompile("(?i)^host|path$")

var cpu int16
var envVars map[string]string
var envVarsRaw []string
var image string
var memory int16
var portRaw string
var port Port
var lbName string
var lbArn string
var rules []ELBV2.Rule
var rulesRaw []string

var serviceCreateCmd = &cobra.Command{
	Use:  "create <service name>",
	Args: cobra.ExactArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		if portRaw != "" {
			port = inflatePort(portRaw)
			validatePort()
		}

		if lbName != "" {
			validateLb()
		}

		validateCpuAndMemory()
		extractEnvVars()
		validateAndExtractRules()
	},
	Run: func(cmd *cobra.Command, args []string) {
		createService(args[0])
	},
}

func init() {
	serviceCreateCmd.Flags().Int16VarP(&cpu, "cpu", "c", 256, "Amount of cpu units to allocate for each task")
	serviceCreateCmd.Flags().Int16VarP(&memory, "memory", "m", 512, "Amount of MiB to allocate for each task")
	serviceCreateCmd.Flags().StringSliceVarP(&envVarsRaw, "env", "e", []string{}, "Environment variables [e.g. KEY=value]")
	serviceCreateCmd.Flags().StringVarP(&portRaw, "port", "p", "", "Port to listen on [e.g., 80, 443, http:8080, https:8443, tcp:1935]")
	serviceCreateCmd.Flags().StringVarP(&image, "image", "i", "", "Docker image to run in the service; if omitted Fargate will build an image from the Dockerfile in the current directory")
	serviceCreateCmd.Flags().StringVarP(&lbName, "lb", "l", "", "Name of a load balancer to use")
	serviceCreateCmd.Flags().StringSliceVarP(&rulesRaw, "rule", "r", []string{}, "Routing rule for the load balancer [e.g. host=api.example.com, path=/api/*]; if omitted service will be the default route")

	serviceCmd.AddCommand(serviceCreateCmd)
}

func validateLb() {
	elbv2 := ELBV2.New(sess)
	loadBalancer := elbv2.DescribeLoadBalancer(lbName)

	if loadBalancer.Type == "network" {
		if port.Protocol != "TCP" {
			console.ErrorExit(fmt.Errorf("network load balancer only supports TCP"), "Invalid load balancer and protocol")
		}
	}

	if loadBalancer.Type == "application" {
		if !(port.Protocol == "HTTP" || port.Protocol == "HTTPS") {
			console.ErrorExit(fmt.Errorf("application load balancer only supports HTTP or HTTPS"), "Invalid load balancer and protocol")
		}
	}

	lbArn = loadBalancer.Arn
}

func validatePort() {
	var msgs []string

	for _, port := range ports {
		if !validProtocols.MatchString(port.Protocol) {
			msgs = append(msgs, fmt.Sprintf("Invalid protocol %s [specify TCP, HTTP, or HTTPS]", port.Protocol))
		}

		if port.Port < 1 || port.Port > 65535 {
			msgs = append(msgs, fmt.Sprintf("Invalid port %d [specify within 1 - 65535]", port.Port))
		}

		if len(msgs) > 0 {
			console.ErrorExit(fmt.Errorf(strings.Join(msgs, ", ")), "Invalid command line flags")
		}
	}
}

func validateAndExtractRules() {
	var msgs []string

	if len(rulesRaw) > 0 && lbName == "" {
		msgs = append(msgs, "lb must be configured if rules are specified")
	}

	for _, ruleRaw := range rulesRaw {
		splitRuleRaw := strings.Split(ruleRaw, "=")

		if len(splitRuleRaw) != 2 {
			msgs = append(msgs, "rules must be in the form of type=value")
		}

		if !validRuleTypes.MatchString(splitRuleRaw[0]) {
			msgs = append(msgs, fmt.Sprintf("Invalid rule type %s [must be path or host]", splitRuleRaw[0]))
		}

		rules = append(rules,
			ELBV2.Rule{
				Type:  splitRuleRaw[0],
				Value: splitRuleRaw[1],
			},
		)
	}

	if len(msgs) > 0 {
		console.ErrorExit(fmt.Errorf(strings.Join(msgs, ", ")), "Invalid rule")
	}
}

func validateCpuAndMemory() {
	err := service.ValidateCpuAndMemory(cpu, memory)

	if err != nil {
		console.ErrorExit(err, "Invalid command line flags")
	}
}

func extractEnvVars() {
	if len(envVarsRaw) == 0 {
		return
	}

	envVars = make(map[string]string)

	for _, envVar := range envVarsRaw {
		splitEnvVar := strings.Split(envVar, "=")

		if len(splitEnvVar) != 2 {
			console.ErrorExit(fmt.Errorf("%s must be in the form of KEY=value", envVar), "Invalid environment variable")
		}

		envVars[splitEnvVar[0]] = splitEnvVar[1]
	}
}

func createService(serviceName string) {
	console.Info("Creating %s", serviceName)

	cwl := CWL.New(sess)
	ec2 := EC2.New(sess)
	ecr := ECR.New(sess)
	ecs := ECS.New(sess)
	elbv2 := ELBV2.New(sess)
	iam := IAM.New(sess)

	var (
		targetGroupArn string
		repositoryUri  string
	)

	if ecr.IsRepositoryCreated(serviceName) {
		repositoryUri = ecr.GetRepositoryUri(serviceName)
	} else {
		repositoryUri = ecr.CreateRepository(serviceName)
	}

	clusterName := ecs.CreateCluster()
	repository := docker.Repository{
		Uri: repositoryUri,
	}
	subnetIds := ec2.GetDefaultVpcSubnetIds()
	ecsTaskExecutionRoleArn := iam.CreateEcsTaskExecutionRole()
	logGroupName := cwl.CreateLogGroup(logGroupFormat, serviceName)

	if image == "" {
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

		image = repository.UriFor(tag)
	}

	if lbArn != "" {
		vpcId := ec2.GetDefaultVpcId()
		targetGroupArn = elbv2.CreateTargetGroup(
			&ELBV2.CreateTargetGroupInput{
				Name:     lbName + "-" + serviceName,
				Port:     port.Port,
				Protocol: port.Protocol,
				VpcId:    vpcId,
			},
		)

		if len(rules) > 0 {
			for _, rule := range rules {
				elbv2.AddRule(lbArn, targetGroupArn, rule)
			}
		} else {
			elbv2.ModifyLoadBalancerDefaultAction(lbArn, targetGroupArn)
		}
	}

	taskDefinitionArn := ecs.CreateTaskDefinition(
		&ECS.CreateTaskDefinitionInput{
			Cpu:              strconv.FormatInt(int64(cpu), 10),
			EnvVars:          envVars,
			ExecutionRoleArn: ecsTaskExecutionRoleArn,
			Image:            image,
			Memory:           strconv.FormatInt(int64(memory), 10),
			Name:             serviceName,
			Port:             port.Port,
			LogGroupName:     logGroupName,
			LogRegion:        region,
		},
	)
	ecs.CreateService(
		&ECS.CreateServiceInput{
			Cluster:           clusterName,
			Name:              serviceName,
			Port:              port.Port,
			SubnetIds:         subnetIds,
			TargetGroupArn:    targetGroupArn,
			TaskDefinitionArn: taskDefinitionArn,
		},
	)
}
