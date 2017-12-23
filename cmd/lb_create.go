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

var validProtocols = regexp.MustCompile("^TCP|HTTP(S)?")
var validTypes = regexp.MustCompile("(?i)^application|network")

var (
	certificateDomainNames []string
	certificateArns        []string
	lbType                 string
	ports                  []Port
	portsRaw               []string
)

var lbCreateCmd = &cobra.Command{
	Use:  "create <load balancer name>",
	Args: cobra.ExactArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		ports = inflatePorts(portsRaw)
		normalizeFields()
		getCertificateArns()
		validatePorts()
		inferType()
	},
	Run: func(cmd *cobra.Command, args []string) {
		createLb(args[0])
	},
}

func init() {
	lbCreateCmd.Flags().StringSliceVarP(&certificateDomainNames, "certificate", "c", []string{}, "Name of certificate to add (can be specified multiple times)")
	lbCreateCmd.Flags().StringSliceVarP(&portsRaw, "port", "p", []string{}, "Port to listen on [e.g., 80, 443, http:8080, https:8443, tcp:1935] (can be specified multiple times)")

	lbCmd.AddCommand(lbCreateCmd)
}

func normalizeFields() {
	lbType = strings.ToLower(lbType)
}

func getCertificateArns() {
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
}

func validatePorts() {
	var msgs []string
	var protocols []string

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

		if len(msgs) > 0 {
			console.ErrorExit(fmt.Errorf(strings.Join(msgs, ", ")), "Invalid command line flags")
		}

		protocols = append(protocols, port.Protocol)
	}
}

func inferType() {
	if ports[0].Protocol == "HTTP" || ports[0].Protocol == "HTTPS" {
		lbType = "application"
	} else if ports[0].Protocol == "TCP" {
		lbType = "network"
	} else {
		console.ErrorExit(fmt.Errorf("Could not infer type; check port settings"), "Invalid command line flags")
	}
}

func createLb(lbName string) {
	console.Info("Creating load balancer [%s]", lbName)

	elbv2 := ELBV2.New(sess)
	ec2 := EC2.New(sess)

	subnetIds := ec2.GetDefaultVpcSubnetIds()
	vpcId := ec2.GetDefaultVpcId()

	lbArn := elbv2.CreateLoadBalancer(
		&ELBV2.CreateLoadBalancerInput{
			Name:      lbName,
			SubnetIds: subnetIds,
			Type:      lbType,
		},
	)

	defaultTargetGroupArn := elbv2.CreateTargetGroup(
		&ELBV2.CreateTargetGroupInput{
			Name:     lbName + "-default",
			Port:     ports[0].Port,
			Protocol: ports[0].Protocol,
			VpcId:    vpcId,
		},
	)

	for _, port := range ports {
		input := &ELBV2.CreateListenerInput{
			Protocol:              port.Protocol,
			Port:                  port.Port,
			LoadBalancerArn:       lbArn,
			DefaultTargetGroupArn: defaultTargetGroupArn,
		}

		if port.Protocol == "HTTPS" {
			input.SetCertificateArns(certificateArns)
		}

		elbv2.CreateListener(input)
	}
}
