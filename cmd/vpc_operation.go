package cmd

import (
	"github.com/awslabs/fargatecli/ec2"
)

type vpcOperation struct {
	ec2              ec2.Client
	output           Output
	securityGroupIDs []string
	subnetIDs        []string
	vpcID            string
}

func (o *vpcOperation) setSubnetIDs(subnetIDs []string) error {
	o.output.Debug("Finding VPC ID [API=ec2 Action=DescribeSubnets]")
	vpcID, err := o.ec2.GetSubnetVPCID(subnetIDs[0])

	if err != nil {
		return err
	}

	o.subnetIDs = subnetIDs
	o.vpcID = vpcID

	return nil
}

func (o *vpcOperation) setSecurityGroupIDs(securityGroupIDs []string) {
	o.securityGroupIDs = securityGroupIDs
}

func (o *vpcOperation) setDefaultSecurityGroupID() error {

	// setting of default security group id is delegated to the ec2 module
	defaultSecurityGroupID, err := o.ec2.SetDefaultSecurityGroupID()
	if err != nil {
		return err
	}

	o.securityGroupIDs = []string{defaultSecurityGroupID}

	return nil
}

func (o *vpcOperation) setDefaultSubnetIDs() error {
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
