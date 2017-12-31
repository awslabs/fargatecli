package cmd

import (
	"fmt"
	"regexp"
	"strings"

	ACM "github.com/jpignata/fargate/acm"
	"github.com/jpignata/fargate/console"
	EC2 "github.com/jpignata/fargate/ec2"
	ELBV2 "github.com/jpignata/fargate/elbv2"
	"github.com/jpignata/fargate/util"
	"github.com/spf13/cobra"
)

const (
	minPort int64 = 1
	maxPort int64 = 65535
)

type LbCreateOperation struct {
	LoadBalancerName string
	CertificateArns  []string
	Ports            []Port
	Type             string
	SecurityGroupIds []string
	SubnetIds        []string
}

func (o *LbCreateOperation) SetCertificateArns(certificateDomainNames []string) {
	var certificateArns []string

	acm := ACM.New(sess)

	for _, certificateDomainName := range certificateDomainNames {
		certificate := acm.DescribeCertificate(certificateDomainName)

		if certificate.IsIssued() {
			certificateArns = append(certificateArns, certificate.Arn)
		} else {
			err := fmt.Errorf("Certificate is in state %s", util.Humanize(certificate.Status))
			console.ErrorExit(err, "Couldn't use certificate %s", certificateDomainName)
		}
	}

	o.CertificateArns = certificateArns
}

func (o *LbCreateOperation) SetPorts(inputPorts []string) {
	var msgs []string
	var protocols []string

	ports := inflatePorts(inputPorts)
	validProtocols := regexp.MustCompile(validProtocolsPattern)

	if len(inputPorts) == 0 {
		msgs = append(msgs, fmt.Sprintf("at least one --port must be specified"))
	}

	for _, port := range ports {
		if !validProtocols.MatchString(port.Protocol) {
			msgs = append(msgs, fmt.Sprintf("Invalid protocol %s [specify TCP, HTTP, or HTTPS]", port.Protocol))
		}

		if port.Port < minPort || port.Port > maxPort {
			msgs = append(msgs, fmt.Sprintf("Invalid port %d [specify within 1 - 65535]", port.Port))
		}

		if port.Protocol == protocolTcp {
			for _, protocol := range protocols {
				if protocol == protocolHttp || protocol == protocolHttps {
					msgs = append(msgs, "load balancers do not support comingled groups of TCP and HTTP/HTTPS ports")
				}
			}
		}

		if port.Protocol == protocolHttp || port.Protocol == protocolHttps {
			for _, protocol := range protocols {
				if protocol == protocolTcp {
					msgs = append(msgs, "load balancers do not support comingled groups of TCP and HTTP/HTTPS ports")
				}
			}
		}

		protocols = append(protocols, port.Protocol)
	}

	if len(msgs) > 0 {
		console.ErrorExit(fmt.Errorf(strings.Join(msgs, ", ")), "Invalid command line flags")
	}

	o.Ports = ports
}

func (o *LbCreateOperation) SetTypeFromPorts() {
	if o.Ports[0].Protocol == protocolHttp || o.Ports[0].Protocol == protocolHttps {
		o.Type = typeApplication
	} else if o.Ports[0].Protocol == protocolTcp {
		o.Type = typeNetwork
	} else {
		console.ErrorExit(fmt.Errorf("Could not infer type; check port settings"), "Invalid command line flags")
	}
}

func (o *LbCreateOperation) SetSecurityGroupIds(securityGroupIds []string) {
	if o.Type != typeApplication {
		console.IssueExit("Security groups can only be specified for HTTP/HTTPS load balancers")
	}

	o.SecurityGroupIds = securityGroupIds
}

func (o *LbCreateOperation) SetSubnetIds(subnetIds []string) {
	if o.Type == typeApplication && len(subnetIds) < 2 {
		console.IssueExit("HTTP/HTTPS load balancers require two subnet IDs from unique availability zones")
	}

	o.SubnetIds = subnetIds
}

var (
	flagLbCreateCertificates     []string
	flagLbCreatePorts            []string
	flagLbCreateSecurityGroupIds []string
	flagLbCreateSubnetIds        []string
)

var lbCreateCmd = &cobra.Command{
	Use:   "create <load-balancer-name> --port <port-expression>",
	Args:  cobra.ExactArgs(1),
	Short: "Create a load balancer",
	Long: `Create a load balancer

At least one port must be specified for the load balancer listener via the
--port flag and a port expression of protocol:port-number. For example, if you
wanted an HTTP load balancer to listen on port 80, you would specify HTTP:80.
Valid protocols are HTTP, HTTPS, and TCP. You can specify multiple listeners by
passing the --port flag with a port expression multiple times. You cannot mix
TCP ports with HTTP/HTTPS ports on a single load balancer.

You can optionally include certificates to secure HTTPS ports by passed the
--certificate flag along with a certificate name. This option can be specified
multiple times to add additional certificates to a single load balancer which
uses Service Name Identification (SNI) to select the appropriate certificate
for the request.

By default, the load balancer will be created in the default VPC and attached
to the default VPC subnets for each availability zone. You can override this by
specifying explicit subnets by passing the --subnet-id flag with a subnet ID.
HTTP/HTTPS load balancers require at least two subnets attached while a TCP
load balancer requires only one. You may only specify a single subnet from each
availability zone.

Security groups can optionally be specified for HTTP/HTTPS load balancers by
passing the --security-group-id flag with a security group ID. To add multiple
security groups, pass --security-group-id with a security group ID multiple
times. If --security-group-id is omitted, a permissive security group will be
applied to the load balancer.`,
	Run: func(cmd *cobra.Command, args []string) {
		operation := &LbCreateOperation{
			LoadBalancerName: args[0],
		}

		if len(flagLbCreateCertificates) > 0 {
			operation.SetCertificateArns(flagLbCreateCertificates)
		}

		operation.SetPorts(flagLbCreatePorts)
		operation.SetTypeFromPorts()

		if len(flagLbCreateSecurityGroupIds) > 0 {
			operation.SetSecurityGroupIds(flagLbCreateSecurityGroupIds)
		}

		if len(flagLbCreateSubnetIds) > 0 {
			operation.SetSubnetIds(flagLbCreateSubnetIds)
		}

		createLoadBalancer(operation)
	},
}

func init() {
	lbCreateCmd.Flags().StringSliceVarP(&flagLbCreateCertificates, "certificate", "c", []string{}, "Name of certificate to add (can be specified multiple times)")
	lbCreateCmd.Flags().StringSliceVarP(&flagLbCreatePorts, "port", "p", []string{}, "Port to listen on [e.g., 80, 443, http:8080, https:8443, tcp:1935] (can be specified multiple times)")
	lbCreateCmd.Flags().StringSliceVar(&flagLbCreateSecurityGroupIds, "security-group-id", []string{}, "ID of a security group to apply to the load balancer (can be specified multiple times)")
	lbCreateCmd.Flags().StringSliceVar(&flagLbCreateSubnetIds, "subnet-id", []string{}, "ID of a subnet to attach to the load balancer (can be specified multiple times)")

	lbCmd.AddCommand(lbCreateCmd)
}

func createLoadBalancer(operation *LbCreateOperation) {
	elbv2 := ELBV2.New(sess)
	ec2 := EC2.New(sess)

	if len(operation.SecurityGroupIds) == 0 {
		operation.SecurityGroupIds = []string{ec2.GetDefaultSecurityGroupId()}
	}

	if len(operation.SubnetIds) == 0 {
		operation.SubnetIds = ec2.GetDefaultVpcSubnetIds()
	}

	vpcId := ec2.GetSubnetVpcId(operation.SubnetIds[0])
	loadBalancerArn := elbv2.CreateLoadBalancer(
		&ELBV2.CreateLoadBalancerInput{
			Name:             operation.LoadBalancerName,
			SecurityGroupIds: operation.SecurityGroupIds,
			SubnetIds:        operation.SubnetIds,
			Type:             operation.Type,
		},
	)
	defaultTargetGroupArn := elbv2.CreateTargetGroup(
		&ELBV2.CreateTargetGroupInput{
			Name:     fmt.Sprintf(defaultTargetGroupFormat, operation.LoadBalancerName),
			Port:     operation.Ports[0].Port,
			Protocol: operation.Ports[0].Protocol,
			VpcId:    vpcId,
		},
	)

	for _, port := range operation.Ports {
		input := &ELBV2.CreateListenerInput{
			Protocol:              port.Protocol,
			Port:                  port.Port,
			LoadBalancerArn:       loadBalancerArn,
			DefaultTargetGroupArn: defaultTargetGroupArn,
		}

		if port.Protocol == protocolHttps {
			input.SetCertificateArns(operation.CertificateArns)
		}

		elbv2.CreateListener(input)
	}

	console.Info("Created load balancer %s", operation.LoadBalancerName)
}
