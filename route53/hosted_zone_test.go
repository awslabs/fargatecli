package route53

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	awsroute53 "github.com/aws/aws-sdk-go/service/route53"
	"github.com/golang/mock/gomock"
	"github.com/jpignata/fargate/route53/mock/sdk"
)

func TestHostedZonesFindSuperDomainOf(t *testing.T) {
	examplecom := HostedZone{Name: "example.com."}
	amazoncom := HostedZone{Name: "amazon.com."}
	intexamplecom := HostedZone{Name: "int.example.com."}

	var tests = []struct {
		fqdn  string
		zones HostedZones
		zone  HostedZone
	}{
		{"staging.example.com", HostedZones{examplecom, amazoncom}, examplecom},
		{"mail.int.example.com", HostedZones{examplecom, intexamplecom, amazoncom}, intexamplecom},
		{"www.amazon.com", HostedZones{examplecom, intexamplecom, amazoncom}, amazoncom},
	}

	for _, test := range tests {
		zone, ok := test.zones.FindSuperDomainOf(test.fqdn)

		if !ok {
			t.Errorf("No match found for %s, expected %s", test.fqdn, test.zone.Name)
		} else if zone != test.zone {
			t.Errorf("Expected %s to be superdomain of %s, got: %s", test.zone.Name, test.fqdn, zone.Name)
		}
	}
}

func TestHostedZonesFindSuperDomainOfNotFound(t *testing.T) {
	zones := HostedZones{
		HostedZone{Name: "zombo.com."},
	}

	zone, ok := zones.FindSuperDomainOf("www.example.com")

	if ok {
		t.Errorf("%s matched, expected none", zone.Name)
	}
}

func TestListHostedZones(t *testing.T) {
	resp := &awsroute53.ListHostedZonesOutput{
		HostedZones: []*awsroute53.HostedZone{
			&awsroute53.HostedZone{Id: aws.String("1"), Name: aws.String("example.com.")},
			&awsroute53.HostedZone{Id: aws.String("2"), Name: aws.String("amazon.com.")},
		},
	}

	mockClient := sdk.MockListHostedZonesPagesClient{Resp: resp}
	route53 := SDKClient{client: mockClient}
	hostedZones, err := route53.ListHostedZones()

	if err != nil {
		t.Errorf("Expected no error, got %+v", err)
	}

	if len(hostedZones) != 2 {
		t.Errorf("Expected 2 hosted zones, got %d", len(hostedZones))
	}

	if hostedZones[0].Name != "example.com." {
		t.Errorf("Expected hosted zone name to be example.com., got %s", hostedZones[0].Name)
	}
}

func TestListHostedZonesError(t *testing.T) {
	mockClient := sdk.MockListHostedZonesPagesClient{
		Error: errors.New("boom"),
		Resp:  &awsroute53.ListHostedZonesOutput{},
	}
	route53 := SDKClient{client: mockClient}
	hostedZones, err := route53.ListHostedZones()

	if err == nil {
		t.Error("Expected error, got none")
	}

	if len(hostedZones) > 0 {
		t.Errorf("Expected no hosted zones, got %d", len(hostedZones))
	}
}

func TestCreateResourceRecord(t *testing.T) {
	hostedZoneID := "zone1"
	hostedZone := HostedZone{Name: "example.com", ID: hostedZoneID}
	recordType := "CNAME"
	name := "www.example.com"
	value := "example.hosted-websites.com"

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockRoute53API := sdk.NewMockRoute53API(mockCtrl)
	route53 := SDKClient{client: mockRoute53API}

	i := &awsroute53.ChangeResourceRecordSetsInput{
		HostedZoneId: aws.String(hostedZoneID),
		ChangeBatch: &awsroute53.ChangeBatch{
			Changes: []*awsroute53.Change{
				&awsroute53.Change{
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
				},
			},
		},
	}
	o := &awsroute53.ChangeResourceRecordSetsOutput{
		ChangeInfo: &awsroute53.ChangeInfo{
			Id: aws.String("1"),
		},
	}

	mockRoute53API.EXPECT().ChangeResourceRecordSets(i).Return(o, nil)

	id, err := route53.CreateResourceRecord(hostedZone, recordType, name, value)

	if id != "1" {
		t.Errorf("Expected id == 1, got %s", id)
	}

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestCreateAliasRecord(t *testing.T) {
	hostedZoneID := "zone1"
	targetHostedZoneID := "zone2"
	hostedZone := HostedZone{Name: "example.com", ID: hostedZoneID}
	recordType := "A"
	name := "www.example.com"
	target := "example.load-balancers.com"

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockRoute53API := sdk.NewMockRoute53API(mockCtrl)
	route53 := SDKClient{client: mockRoute53API}

	i := &awsroute53.ChangeResourceRecordSetsInput{
		HostedZoneId: aws.String(hostedZoneID),
		ChangeBatch: &awsroute53.ChangeBatch{
			Changes: []*awsroute53.Change{
				&awsroute53.Change{
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
				},
			},
		},
	}
	o := &awsroute53.ChangeResourceRecordSetsOutput{
		ChangeInfo: &awsroute53.ChangeInfo{
			Id: aws.String("2"),
		},
	}

	mockRoute53API.EXPECT().ChangeResourceRecordSets(i).Return(o, nil)

	id, err := route53.CreateAlias(hostedZone, recordType, name, target, targetHostedZoneID)

	if id != "2" {
		t.Errorf("Expected id == 2, got %s", id)
	}

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}
