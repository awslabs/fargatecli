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
		validateType()
	},
	Run: func(cmd *cobra.Command, args []string) {
		createLb(args[0])
	},
}

func init() {
	lbCreateCmd.Flags().StringSliceVarP(&certificateDomainNames, "certificate", "c", []string{}, "Name of certificate to add (can be specified multiple times)")
	lbCreateCmd.Flags().StringVarP(&lbType, "type", "t", "application", "Type of load balancer [application, network]")
	lbCreateCmd.Flags().StringSliceVarP(&portsRaw, "port", "p", []string{}, "Port to listen on [e.g., 80, 443, http:8080, https:8443, tcp:1935] (can be specified multiple times)")

	lbCmd.AddCommand(lbCreateCmd)
}

func normalizeFields() {
	lbType = strings.ToLower(lbType)
}

func getCertificateArns() {
	acm := ACM.New()

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

func validateType() {
	if !validTypes.MatchString(lbType) {
		console.ErrorExit(fmt.Errorf("Invalid type %s [specify application or network]", lbType), "Invalid command line flags")
	}
}

func createLb(lbName string) {
	console.Info("Creating load balancer [%s]", lbName)

	elbv2 := ELBV2.New()
	ec2 := EC2.New()

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
