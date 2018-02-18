package sdk

import (
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/aws/aws-sdk-go/service/elbv2/elbv2iface"
)

type MockDescribeLoadBalancersClient struct {
	elbv2iface.ELBV2API
	Resp  *elbv2.DescribeLoadBalancersOutput
	Error error
}

func (m MockDescribeLoadBalancersClient) DescribeLoadBalancersPages(in *elbv2.DescribeLoadBalancersInput, fn func(*elbv2.DescribeLoadBalancersOutput, bool) bool) error {
	if m.Error != nil {
		return m.Error
	}

	fn(m.Resp, true)

	return nil
}
