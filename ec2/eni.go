package ec2

import (
	"github.com/aws/aws-sdk-go/aws"
	awsec2 "github.com/aws/aws-sdk-go/service/ec2"
	"github.com/jpignata/fargate/console"
)

type Eni struct {
	PublicIpAddress  string
	EniId            string
	SecurityGroupIds []string
}

func (ec2 *EC2) DescribeNetworkInterfaces(eniIds []string) map[string]Eni {
	enis := make(map[string]Eni)

	resp, err := ec2.svc.DescribeNetworkInterfaces(
		&awsec2.DescribeNetworkInterfacesInput{
			NetworkInterfaceIds: aws.StringSlice(eniIds),
		},
	)

	if err != nil {
		console.ErrorExit(err, "Could not describe network interfaces")
	}

	for _, e := range resp.NetworkInterfaces {
		var securityGroupIds []*string

		for _, group := range e.Groups {
			securityGroupIds = append(securityGroupIds, group.GroupId)
		}

		if e.Association != nil {
			eni := Eni{
				EniId:            aws.StringValue(e.NetworkInterfaceId),
				PublicIpAddress:  aws.StringValue(e.Association.PublicIp),
				SecurityGroupIds: aws.StringValueSlice(securityGroupIds),
			}

			enis[eni.EniId] = eni
		}
	}

	return enis
}
