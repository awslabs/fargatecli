package cmd

import (
	"github.com/jpignata/fargate/ec2"
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
