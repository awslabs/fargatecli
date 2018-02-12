package cmd

import (
	"fmt"

	"github.com/jpignata/fargate/acm"
	"github.com/jpignata/fargate/ec2"
	"github.com/jpignata/fargate/elbv2"
	"github.com/spf13/cobra"
)

type lbCreateOperation struct {
	certificateOperation
	certificateARNs  []string
	ec2              ec2.Client
	elbv2            elbv2.Client
	lbType           string
	loadBalancerName string
	output           Output
	ports            []Port
	securityGroupIDs []string
	subnetIDs        []string
	vpcID            string
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

func (o *lbCreateOperation) inferType() error {
	if len(o.ports) > 0 {
		switch o.ports[0].Protocol {
		case "HTTP", "HTTPS":
			o.lbType = "application"
		case "TCP":
			o.lbType = "network"
		default:
			return fmt.Errorf("could not infer type from port settings")
		}
	}

	return nil
}

func (o *lbCreateOperation) setCertificateArns(certificateDomainNames []string) []error {
	var (
		certificateARNs []string
		errs            []error
	)

	for _, certificateDomainName := range certificateDomainNames {
		if certificate, err := o.findCertificate(certificateDomainName, output); err == nil {
			if certificate.IsIssued() {
				certificateARNs = append(certificateARNs, certificate.Arn)
			} else {
				errs = append(errs, fmt.Errorf("certificate %s is in state %s", certificateDomainName, Humanize(certificate.Status)))
			}
		} else {
			switch err {
			case errCertificateNotFound:
				errs = append(errs, fmt.Errorf("no certificate found for %s", certificateDomainName))
			case errCertificateTooManyFound:
				errs = append(errs, fmt.Errorf("multiple certificates found for %s", certificateDomainName))
			default:
				errs = append(errs, err)
			}
		}
	}

	if len(errs) == 0 {
		o.certificateARNs = certificateARNs
	}

	return errs
}

func (o *lbCreateOperation) setSubnetIDs(subnetIDs []string) error {
	if o.lbType == "application" && len(subnetIDs) < 2 {
		return fmt.Errorf("HTTP/HTTPS load balancers require two subnet IDs from unique availability zones")
	}

	o.output.Debug("Finding VPC ID [API=ec2 Action=DescribeSubnets]")
	vpcID, err := o.ec2.GetSubnetVPCID(subnetIDs[0])

	if err != nil {
		return err
	}

	o.subnetIDs = subnetIDs
	o.vpcID = vpcID

	return nil
}

func (o *lbCreateOperation) setSecurityGroupIDs(securityGroupIDs []string) error {
	if o.lbType != "application" {
		return fmt.Errorf("security groups can only be specified for HTTP/HTTPS load balancers")
	}

	o.securityGroupIDs = securityGroupIDs

	return nil
}

func (o *lbCreateOperation) setDefaultSecurityGroupID() error {
	o.output.Debug("Finding default security group [API=ec2 Action=DescribeSecurityGroups]")
	defaultSecurityGroupID, err := o.ec2.GetDefaultSecurityGroupID()

	if err != nil {
		return err
	}

	if defaultSecurityGroupID == "" {
		o.output.Debug("Creating default security group [API=ec2 Action=CreateSecurityGroup]")
		defaultSecurityGroupID, err = o.ec2.CreateDefaultSecurityGroup()

		if err != nil {
			return err
		}

		o.output.Debug("Created default security group [ID=%s]", defaultSecurityGroupID)

		o.output.Debug("Configuring default security group [API=ec2 Action=AuthorizeSecurityGroupIngress]")
		if err := o.ec2.AuthorizeAllSecurityGroupIngress(defaultSecurityGroupID); err != nil {
			return err
		}
	}

	o.securityGroupIDs = []string{defaultSecurityGroupID}

	return nil
}

func (o *lbCreateOperation) setDefaultSubnetIDs() error {
	o.output.Debug("Finding default subnets [API=ec2 Action=DescribeSubnets]")
	subnetIDs, err := o.ec2.GetDefaultSubnetIDs()

	if err != nil {
		return err
	}

	o.output.Debug("Finding VPC ID [API=ec2 Action=DescribeSubnets]")
	vpcID, err := o.ec2.GetSubnetVPCID(subnetIDs[0])

	if err != nil {
		return err
	}

	o.subnetIDs = subnetIDs
	o.vpcID = vpcID

	return nil
}

func (o *lbCreateOperation) execute() {
	defaultTargetGroupName := fmt.Sprintf(defaultTargetGroupFormat, o.loadBalancerName)

	loadBalancerArn, err := o.elbv2.CreateLoadBalancer(
		elbv2.CreateLoadBalancerInput{
			Name:             o.loadBalancerName,
			SecurityGroupIDs: o.securityGroupIDs,
			SubnetIDs:        o.subnetIDs,
			Type:             o.lbType,
		},
	)

	if err != nil {
		o.output.Fatal(err, "Could not create load balancer")
		return
	}

	o.output.Debug("Creating target group [Name=%s]", defaultTargetGroupName)
	defaultTargetGroupARN, err := o.elbv2.CreateTargetGroup(
		elbv2.CreateTargetGroupInput{
			Name:     defaultTargetGroupName,
			Port:     o.ports[0].Number,
			Protocol: o.ports[0].Protocol,
			VPCID:    o.vpcID,
		},
	)

	if err != nil {
		o.output.Fatal(err, "Could not create default target group")
		return
	}

	o.output.Debug("Created target group [ARN=%s]", defaultTargetGroupARN)

	for _, port := range o.ports {
		o.output.Debug("Creating listener [Port=%s Protocol=%s]", port.Number, port.Protocol)
		listenerARN, err := o.elbv2.CreateListener(
			elbv2.CreateListenerInput{
				CertificateARNs:       o.certificateARNs,
				DefaultTargetGroupARN: defaultTargetGroupARN,
				LoadBalancerARN:       loadBalancerArn,
				Port:                  port.Number,
				Protocol:              port.Protocol,
			},
		)

		if err != nil {
			o.output.Fatal(err, "Could not create listener")
			return
		}

		o.output.Debug("Created listener [ARN=%s]", listenerARN)
	}

	o.output.Info("Created load balancer %s", o.loadBalancerName)
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
			var errors []error

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
				errors = append(errors, errs...)
			}

			if err := operation.inferType(); err != nil {
				errors = append(errors, err)
			}

			if len(flagLbCreateCertificates) > 0 {
				if errs := operation.setCertificateArns(flagLbCreateCertificates); len(errs) > 0 {
					errors = append(errors, errs...)
				}
			}

			if len(flagLbCreateSecurityGroupIds) > 0 {
				if err := operation.setSecurityGroupIDs(flagLbCreateSecurityGroupIds); err != nil {
					errors = append(errors, err)
				}
			} else {
				if err := operation.setDefaultSecurityGroupID(); err != nil {
					errors = append(errors, err)
				}
			}

			if len(flagLbCreateSubnetIds) > 0 {
				if err := operation.setSubnetIDs(flagLbCreateSubnetIds); err != nil {
					errors = append(errors, err)
				}
			}

			if len(errors) == 0 {
				operation.execute()
			} else {
				output.Fatals(errors, "Invalid command line flags")
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
