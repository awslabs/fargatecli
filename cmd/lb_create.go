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

type LbCreateOperation struct {
	LoadBalancerName string
	CertificateArns  []string
	Ports            []Port
	Type             string
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

		if port.Port < 1 || port.Port > 65535 {
			msgs = append(msgs, fmt.Sprintf("Invalid port %d [specify within 1 - 65535]", port.Port))
		}

		if port.Protocol == "TCP" {
			for _, protocol := range protocols {
				if protocol == "HTTP" || protocol == "HTTPS" {
					msgs = append(msgs, "load balancers do not support comingled groups of TCP and HTTP/HTTPS ports")
				}
			}
		}

		if port.Protocol == "HTTP" || port.Protocol == "HTTPS" {
			for _, protocol := range protocols {
				if protocol == "TCP" {
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
	if o.Ports[0].Protocol == "HTTP" || o.Ports[0].Protocol == "HTTPS" {
		o.Type = "application"
	} else if o.Ports[0].Protocol == "TCP" {
		o.Type = "network"
	} else {
		console.ErrorExit(fmt.Errorf("Could not infer type; check port settings"), "Invalid command line flags")
	}
}

var (
	flagLbCreateCertificates []string
	flagLbCreatePorts        []string
)

var lbCreateCmd = &cobra.Command{
	Use:  "create <load balancer name>",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		operation := &LbCreateOperation{
			LoadBalancerName: args[0],
		}

		operation.SetCertificateArns(flagLbCreateCertificates)
		operation.SetPorts(flagLbCreatePorts)
		operation.SetTypeFromPorts()

		createLoadBalancer(operation)
	},
}

func init() {
	lbCreateCmd.Flags().StringSliceVarP(&flagLbCreateCertificates, "certificate", "c", []string{}, "Name of certificate to add (can be specified multiple times)")
	lbCreateCmd.Flags().StringSliceVarP(&flagLbCreatePorts, "port", "p", []string{}, "Port to listen on [e.g., 80, 443, http:8080, https:8443, tcp:1935] (can be specified multiple times)")

	lbCmd.AddCommand(lbCreateCmd)
}

func createLoadBalancer(operation *LbCreateOperation) {
	elbv2 := ELBV2.New(sess)
	ec2 := EC2.New(sess)

	subnetIds := ec2.GetDefaultVpcSubnetIds()
	vpcId := ec2.GetDefaultVpcId()

	loadBalancerArn := elbv2.CreateLoadBalancer(
		&ELBV2.CreateLoadBalancerInput{
			Name:      operation.LoadBalancerName,
			SubnetIds: subnetIds,
			Type:      operation.Type,
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

		if port.Protocol == "HTTPS" {
			input.SetCertificateArns(operation.CertificateArns)
		}

		elbv2.CreateListener(input)
	}

	console.Info("Created load balancer %s", operation.LoadBalancerName)

}
