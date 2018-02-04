package route53

import (
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	awsroute53 "github.com/aws/aws-sdk-go/service/route53"
)

const defaultTTL = 86400

// HostedZone is a zone hosted in Amazon Route 53.
type HostedZone struct {
	Name string
	ID   string
}

func (h HostedZone) isSuperDomainOf(fqdn string) bool {
	return strings.HasSuffix(fqdn+".", h.Name)
}

// HostedZones is a collection of HostedZones.
type HostedZones []HostedZone

// FindSuperDomainOf searches a HostedZones collection for the zone that is the superdomain of the
// given fully qualified domain name. Returns a HostedZone and a boolean indicating whether a
// match was found.
func (h HostedZones) FindSuperDomainOf(fqdn string) (HostedZone, bool) {
	sort.Slice(h, func(i, j int) bool {
		return len(h[i].Name) > len(h[j].Name)
	})

	for _, zone := range h {
		if zone.isSuperDomainOf(fqdn) {
			return zone, true
		}
	}

	return HostedZone{}, false
}

// CreateResourceRecord creates a DNS record in an Amazon Route 53 hosted zone.
func (route53 SDKClient) CreateResourceRecord(h HostedZone, recordType, name, value string) (string, error) {
	change := &awsroute53.Change{
		Action: aws.String(awsroute53.ChangeActionUpsert),
		ResourceRecordSet: &awsroute53.ResourceRecordSet{
			Name: aws.String(name),
			Type: aws.String(recordType),
			TTL:  aws.Int64(defaultTTL),
			ResourceRecords: []*awsroute53.ResourceRecord{
				&awsroute53.ResourceRecord{
					Value: aws.String(value),
				},
			},
		},
	}

	resp, err := route53.client.ChangeResourceRecordSets(
		&awsroute53.ChangeResourceRecordSetsInput{
			HostedZoneId: aws.String(h.ID),
			ChangeBatch: &awsroute53.ChangeBatch{
				Changes: []*awsroute53.Change{change},
			},
		},
	)

	return aws.StringValue(resp.ChangeInfo.Id), err
}

// CreateAlias creates an alias record in an Amazon Route 53 hosted zone.
func (route53 SDKClient) CreateAlias(h HostedZone, recordType, name, target, targetHostedZoneID string) (string, error) {
	change := &awsroute53.Change{
		Action: aws.String(awsroute53.ChangeActionUpsert),
		ResourceRecordSet: &awsroute53.ResourceRecordSet{
			Name: aws.String(name),
			Type: aws.String(recordType),
			AliasTarget: &awsroute53.AliasTarget{
				DNSName:              aws.String(target),
				EvaluateTargetHealth: aws.Bool(false),
				HostedZoneId:         aws.String(targetHostedZoneID),
			},
		},
	}

	resp, err := route53.client.ChangeResourceRecordSets(
		&awsroute53.ChangeResourceRecordSetsInput{
			HostedZoneId: aws.String(h.ID),
			ChangeBatch: &awsroute53.ChangeBatch{
				Changes: []*awsroute53.Change{change},
			},
		},
	)

	return aws.StringValue(resp.ChangeInfo.Id), err
}

// ListHostedZones returns all Amazon Route 53 zones in the caller's account.
func (route53 SDKClient) ListHostedZones() (HostedZones, error) {
	var hostedZones HostedZones

	input := &awsroute53.ListHostedZonesInput{}
	handler := func(resp *awsroute53.ListHostedZonesOutput, lastPage bool) bool {
		for _, hostedZone := range resp.HostedZones {
			hostedZones = append(
				hostedZones,
				HostedZone{
					Name: aws.StringValue(hostedZone.Name),
					ID:   aws.StringValue(hostedZone.Id),
				},
			)
		}

		return true
	}

	err := route53.client.ListHostedZonesPages(input, handler)

	return hostedZones, err
}
