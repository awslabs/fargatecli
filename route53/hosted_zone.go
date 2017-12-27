package route53

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	awsroute53 "github.com/aws/aws-sdk-go/service/route53"
	"github.com/jpignata/fargate/console"
)

const (
	maxItems   = "100"
	defaultTtl = 86400
)

type HostedZone struct {
	Name string
	Id   string
}

func (h HostedZone) IsSuperDomainOf(fqdn string) bool {
	return strings.HasSuffix(fqdn+".", h.Name)
}

func (route53 *Route53) CreateResourceRecord(h HostedZone, Type, Name, Value string) {
	change := &awsroute53.Change{
		Action: aws.String(awsroute53.ChangeActionUpsert),
		ResourceRecordSet: &awsroute53.ResourceRecordSet{
			Name: aws.String(Name),
			Type: aws.String(Type),
			TTL:  aws.Int64(defaultTtl),
			ResourceRecords: []*awsroute53.ResourceRecord{
				&awsroute53.ResourceRecord{
					Value: aws.String(Value),
				},
			},
		},
	}

	_, err := route53.svc.ChangeResourceRecordSets(
		&awsroute53.ChangeResourceRecordSetsInput{
			HostedZoneId: aws.String(h.Id),
			ChangeBatch: &awsroute53.ChangeBatch{
				Changes: []*awsroute53.Change{change},
			},
		},
	)

	if err != nil {
		console.ErrorExit(err, "Could not create Route53 resource record")
	}
}

func (route53 *Route53) CreateAlias(h HostedZone, recordType, name, target, targetHostedZone string) {
	change := &awsroute53.Change{
		Action: aws.String(awsroute53.ChangeActionUpsert),
		ResourceRecordSet: &awsroute53.ResourceRecordSet{
			Name: aws.String(name),
			Type: aws.String(recordType),
			AliasTarget: &awsroute53.AliasTarget{
				DNSName:              aws.String(target),
				EvaluateTargetHealth: aws.Bool(false),
				HostedZoneId:         aws.String(targetHostedZone),
			},
		},
	}

	_, err := route53.svc.ChangeResourceRecordSets(
		&awsroute53.ChangeResourceRecordSetsInput{
			HostedZoneId: aws.String(h.Id),
			ChangeBatch: &awsroute53.ChangeBatch{
				Changes: []*awsroute53.Change{change},
			},
		},
	)

	if err != nil {
		console.ErrorExit(err, "Could not create Route53 resource record")
	}
}

func (route53 *Route53) ListHostedZones() []HostedZone {
	console.Debug("Listing Route53 hosted zones")

	hostedZones := []HostedZone{}

	err := route53.svc.ListHostedZonesPagesWithContext(
		context.Background(),
		&awsroute53.ListHostedZonesInput{
			MaxItems: aws.String(maxItems),
		},
		func(resp *awsroute53.ListHostedZonesOutput, lastPage bool) bool {
			for _, hostedZone := range resp.HostedZones {
				hostedZones = append(
					hostedZones,
					HostedZone{
						Name: aws.StringValue(hostedZone.Name),
						Id:   aws.StringValue(hostedZone.Id),
					},
				)
			}

			return true
		},
	)

	if err != nil {
		console.ErrorExit(err, "")
	}

	return hostedZones
}
