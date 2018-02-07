package cmd

import (
	"fmt"

	"github.com/jpignata/fargate/acm"
	"github.com/jpignata/fargate/console"
	"github.com/jpignata/fargate/ec2"
	"github.com/jpignata/fargate/elbv2"
	"github.com/spf13/cobra"
)

type lbCreateOperation struct {
	certificateOperation
	elbv2            elbv2.SDKClient
	ec2              ec2.EC2
	loadBalancerName string
	certificateArns  []string
	ports            []Port
	lbType           string
	securityGroupIds []string
	subnetIds        []string
	output           Output
}

func (o *lbCreateOperation) setPorts(inputPorts []string) []error {
	var (
		errs      []error
		protocols []string
	)

	if len(inputPorts) == 0 {
		return append(errs, fmt.Errorf("at least one --port must be specified"))
	}

	ports, errs := inflatePorts(inputPorts)

	if len(errs) > 0 {
		return errs
	}

	for _, port := range ports {
		errs = append(errs, validatePort(port)...)
		protocols = append(protocols, port.Protocol)
	}

	for _, protocol := range protocols {
		if protocol == "TCP" {
			for _, protocol := range protocols {
				if protocol == "HTTP" || protocol == "HTTPS" {
					return append(errs, fmt.Errorf("load balancers do not support commingled TCP and HTTP/HTTPS ports"))
				}
			}
		}
	}

	if len(errs) == 0 {
		o.ports = ports
	}

	return errs
}

func (o *lbCreateOperation) InferType() error {
	var err error

	switch o.ports[0].Protocol {
	case "HTTP", "HTTPS":
		o.lbType = "application"
	case "TCP":
		o.lbType = "network"
	default:
		err = fmt.Errorf("Could not infer type; check port settings")
	}

	return err
}

func (o *lbCreateOperation) setCertificateArns(certificateDomainNames []string) []error {
	var (
		certificateArns []string
		errs            []error
	)

	for _, certificateDomainName := range certificateDomainNames {
		if certificate, err := o.findCertificate(certificateDomainName, output); err == nil {
			if certificate.IsIssued() {
				certificateArns = append(certificateArns, certificate.Arn)
			} else {
				errs = append(errs, fmt.Errorf("Certificate %s is in state %s", certificateDomainName, Humanize(certificate.Status)))
			}
		} else {
			errs = append(errs, err)
		}
	}

	if len(errs) == 0 {
		o.certificateArns = certificateArns
	}

	return errs
}

func (o *lbCreateOperation) setSubnetIDs(subnetIds []string) error {
	if o.lbType == "application" && len(subnetIds) < 2 {
		return fmt.Errorf("HTTP/HTTPS load balancers require two subnet IDs from unique availability zones")
	}

	o.subnetIds = subnetIds
	return nil
}

func (o *lbCreateOperation) setSecurityGroupIDs(securityGroupIds []string) error {
	if o.lbType != "application" {
		return fmt.Errorf("Security groups can only be specified for HTTP/HTTPS load balancers")
	}

	o.securityGroupIds = securityGroupIds
	return nil
}

func (o *lbCreateOperation) execute() {
	if len(o.securityGroupIds) == 0 {
		o.securityGroupIds = []string{o.ec2.GetDefaultSecurityGroupId()}
	}

	if len(o.subnetIds) == 0 {
		o.subnetIds = o.ec2.GetDefaultVpcSubnetIds()
	}

	vpcID := o.ec2.GetSubnetVpcId(o.subnetIds[0])
	loadBalancerArn := o.elbv2.CreateLoadBalancer(
		&elbv2.CreateLoadBalancerInput{
			Name:             o.loadBalancerName,
			SecurityGroupIds: o.securityGroupIds,
			SubnetIds:        o.subnetIds,
			Type:             o.lbType,
		},
	)
	defaultTargetGroupArn := o.elbv2.CreateTargetGroup(
		&elbv2.CreateTargetGroupInput{
			Name:     fmt.Sprintf(defaultTargetGroupFormat, o.loadBalancerName),
			Port:     o.ports[0].Port,
			Protocol: o.ports[0].Protocol,
			VpcId:    vpcID,
		},
	)

	for _, port := range o.ports {
		input := &elbv2.CreateListenerInput{
			Protocol:              port.Protocol,
			Port:                  port.Port,
			LoadBalancerArn:       loadBalancerArn,
			DefaultTargetGroupArn: defaultTargetGroupArn,
		}

		if port.Protocol == "HTTPS" {
			input.SetCertificateArns(o.certificateArns)
		}

		o.elbv2.CreateListener(input)
	}

	console.Info("Created load balancer %s", o.loadBalancerName)
}

var (
	flagLbCreateCertificates     []string
	flagLbCreatePorts            []string
	flagLbCreateSecurityGroupIds []string
	flagLbCreateSubnetIds        []string

	lbCreateCmd = &cobra.Command{
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
			var errs []error

			operation := &lbCreateOperation{
				certificateOperation: certificateOperation{
					acm: acm.New(sess),
				},
				ec2:              ec2.New(sess),
				elbv2:            elbv2.New(sess),
				loadBalancerName: args[0],
				output:           output,
			}

			if errs := operation.setPorts(flagLbCreatePorts); len(errs) > 0 {
				errs = append(errs, errs...)
			}

			if err := operation.InferType(); err != nil {
				errs = append(errs, err)
			}

			if len(flagLbCreateCertificates) > 0 {
				if errs := operation.setCertificateArns(flagLbCreateCertificates); len(errs) > 0 {
					errs = append(errs, errs...)
				}
			}

			if len(flagLbCreateSecurityGroupIds) > 0 {
				if err := operation.setSecurityGroupIDs(flagLbCreateSecurityGroupIds); err != nil {
					errs = append(errs, err)
				}
			}

			if len(flagLbCreateSubnetIds) > 0 {
				if err := operation.setSubnetIDs(flagLbCreateSubnetIds); err != nil {
					errs = append(errs, err)
				}
			}

			if len(errs) == 0 {
				operation.execute()
			} else {
				output.Fatals(errs, "Invalid command line flags")
			}
		},
	}
)

func init() {
	lbCreateCmd.Flags().StringSliceVarP(&flagLbCreateCertificates, "certificate", "c", []string{},
		"Name of certificate to add (can be specified multiple times)")
	lbCreateCmd.Flags().StringSliceVarP(&flagLbCreatePorts, "port", "p", []string{},
		"Port to listen on [e.g., 80, 443, http:8080, https:8443, tcp:1935] (can be specified multiple times)")
	lbCreateCmd.Flags().StringSliceVar(&flagLbCreateSecurityGroupIds, "security-group-id", []string{},
		"ID of a security group to apply to the load balancer (can be specified multiple times)")
	lbCreateCmd.Flags().StringSliceVar(&flagLbCreateSubnetIds, "subnet-id", []string{},
		"ID of a subnet to place the load balancer (can be specified multiple times)")

	lbCmd.AddCommand(lbCreateCmd)
}
