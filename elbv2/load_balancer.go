package elbv2

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	awselbv2 "github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/jpignata/fargate/console"
)

// LoadBalancer represents an Elastic Load Balancing (v2) load balancer.
type LoadBalancer struct {
	ARN              string
	DNSName          string
	HostedZoneID     string
	Listeners        Listeners
	Name             string
	SecurityGroupIDs []string
	Status           string
	SubnetIDs        []string
	Type             string
	VPCID            string
}

// LoadBalancers is a collection of Elastic Load Balancing (v2) load balancers.
type LoadBalancers []LoadBalancer

// CreateLoadBalancerParameters are the parameters required to create a new load balancer.
type CreateLoadBalancerParameters struct {
	Name             string
	SecurityGroupIDs []string
	SubnetIDs        []string
	Type             string
}

// CreateLoadBalancer creates a new load balancer. It returns the ARN of the load balancer if it is successfully
// created.
func (elbv2 SDKClient) CreateLoadBalancer(p CreateLoadBalancerParameters) (string, error) {
	sdki := &awselbv2.CreateLoadBalancerInput{
		Name:    aws.String(p.Name),
		Subnets: aws.StringSlice(p.SubnetIDs),
		Type:    aws.String(p.Type),
	}

	if p.Type == awselbv2.LoadBalancerTypeEnumApplication {
		sdki.SetSecurityGroups(aws.StringSlice(p.SecurityGroupIDs))
	}

	resp, err := elbv2.client.CreateLoadBalancer(sdki)

	if err != nil {
		return "", err
	}

	return aws.StringValue(resp.LoadBalancers[0].LoadBalancerArn), nil
}

// DescribeLoadBalancers returns all load balancers.
func (elbv2 SDKClient) DescribeLoadBalancers() (LoadBalancers, error) {
	return elbv2.describeLoadBalancers(&awselbv2.DescribeLoadBalancersInput{})
}

// DescribeLoadBalancersByName returns load balancers that match the given load balancer names.
func (elbv2 SDKClient) DescribeLoadBalancersByName(lbNames []string) (LoadBalancers, error) {
	return elbv2.describeLoadBalancers(
		&awselbv2.DescribeLoadBalancersInput{Names: aws.StringSlice(lbNames)},
	)
}

// DescribeLoadBalancersByARN returns load balancers that match the given load balancer ARNs.
func (elbv2 SDKClient) DescribeLoadBalancersByARN(lbARNs []string) (LoadBalancers, error) {
	return elbv2.describeLoadBalancers(
		&awselbv2.DescribeLoadBalancersInput{LoadBalancerArns: aws.StringSlice(lbARNs)},
	)
}

func (elbv2 SDKClient) DescribeLoadBalancer(lbName string) LoadBalancer {
	loadBalancers, _ := elbv2.DescribeLoadBalancersByName([]string{lbName})

	if len(loadBalancers) == 0 {
		console.ErrorExit(fmt.Errorf("%s not found", lbName), "Could not find ELB load balancer")
	}

	return loadBalancers[0]
}

func (elbv2 SDKClient) DescribeLoadBalancerByARN(lbARN string) LoadBalancer {
	loadBalancers, _ := elbv2.DescribeLoadBalancersByARN([]string{lbARN})

	if len(loadBalancers) == 0 {
		console.ErrorExit(fmt.Errorf("%s not found", lbARN), "Could not find ELB load balancer")
	}

	return loadBalancers[0]
}

func (elbv2 SDKClient) DeleteLoadBalancer(lbName string) {
	loadBalancer := elbv2.DescribeLoadBalancer(lbName)
	_, err := elbv2.client.DeleteLoadBalancer(
		&awselbv2.DeleteLoadBalancerInput{
			LoadBalancerArn: aws.String(loadBalancer.ARN),
		},
	)

	if err != nil {
		console.ErrorExit(err, "Could not destroy ELB load balancer")
	}
}

func (elbv2 SDKClient) describeLoadBalancers(i *awselbv2.DescribeLoadBalancersInput) (LoadBalancers, error) {
	var loadBalancers []LoadBalancer

	handler := func(resp *awselbv2.DescribeLoadBalancersOutput, lastPage bool) bool {
		for _, loadBalancer := range resp.LoadBalancers {
			var subnetIDs []string

			for _, availabilityZone := range loadBalancer.AvailabilityZones {
				subnetIDs = append(subnetIDs, aws.StringValue(availabilityZone.SubnetId))
			}

			loadBalancers = append(loadBalancers,
				LoadBalancer{
					ARN:              aws.StringValue(loadBalancer.LoadBalancerArn),
					DNSName:          aws.StringValue(loadBalancer.DNSName),
					HostedZoneID:     aws.StringValue(loadBalancer.CanonicalHostedZoneId),
					VPCID:            aws.StringValue(loadBalancer.VpcId),
					Name:             aws.StringValue(loadBalancer.LoadBalancerName),
					SecurityGroupIDs: aws.StringValueSlice(loadBalancer.SecurityGroups),
					Status:           aws.StringValue(loadBalancer.State.Code),
					SubnetIDs:        subnetIDs,
					Type:             aws.StringValue(loadBalancer.Type),
				},
			)
		}

		return true
	}

	err := elbv2.client.DescribeLoadBalancersPages(i, handler)

	return loadBalancers, err
}
