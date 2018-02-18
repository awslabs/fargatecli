package sdk

import (
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/aws/aws-sdk-go/service/route53/route53iface"
)

type MockListHostedZonesPagesClient struct {
	route53iface.Route53API
	Resp  *route53.ListHostedZonesOutput
	Error error
}

func (m MockListHostedZonesPagesClient) ListHostedZonesPages(in *route53.ListHostedZonesInput, fn func(*route53.ListHostedZonesOutput, bool) bool) error {
	if m.Error != nil {
		return m.Error
	}

	fn(m.Resp, true)

	return nil
}
