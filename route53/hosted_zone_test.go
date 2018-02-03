package route53

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	awsroute53 "github.com/aws/aws-sdk-go/service/route53"
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
